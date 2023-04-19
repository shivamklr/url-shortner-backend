package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ShortenUrl struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OriginalUrl string             `json:"original_url" bson:"original_url,omitempty"`
	ExpireIn    time.Duration      `json:"expire_in" bson:"expire,omitempty"`
	ShortenUrl  string             `json:"shorten_url" bson:"shorten_url,omitempty"`
}

type ShortenResponse struct {
	Status  int        `json:"status"`
	Message string     `json:"message"`
	Data    *fiber.Map `json:"data"`
}

func createShortenURL(ctx *fiber.Ctx) error {

	body := new(ShortenUrl)

	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse JSON",
		})
	}
	// insert data to mongodb database,
	body.ShortenUrl = uuid.New().String()[:8]
	body.ExpireIn = 24
	result, err := urlCollection.InsertOne(context.Background(), body)
	if err != nil {
		log.Fatal(err)

	}
	bodyId, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		log.Fatal("Type Assertion Failed")
	}
	body.Id = bodyId
	// fmt.Printf("Inserted document with _id: %v\n", res.InsertedID)
	// add ttl of 2 hours
	// if data is inserted add cache data to redis
	redisClient := connectToRedis(0)
	defer redisClient.Close()
	if err := redisClient.Set(context.Background(), body.ShortenUrl, body.OriginalUrl, 1*60*time.Second).Err(); err != nil {
		log.Fatal(err)
	}
	//  add ttl of 1 minute.
	body.ShortenUrl = os.Getenv("DOMAIN") + "/" + body.ShortenUrl
	return ctx.Status(fiber.StatusCreated).JSON(ShortenResponse{Status: fiber.StatusCreated, Message: "success", Data: &fiber.Map{"data": body}})
}

func resolveURL(ctx *fiber.Ctx) error {
	// get the short from url params
	url := ctx.Params("url")

	// check whether the shorten URL key exist in redis cache
	redisClient := connectToRedis(0)
	defer redisClient.Close()
	val, err := redisClient.Get(context.Background(), url).Result()
	if err == redis.Nil {
		// if it does not, db lookup, add it in the cache and return it
		var urlFound ShortenUrl
		filter := ShortenUrl{
			ShortenUrl: url,
		}
		if err := urlCollection.FindOne(context.Background(), filter).Decode(&urlFound); err != nil {
			if err == mongo.ErrNoDocuments {
				return ctx.SendStatus(fiber.StatusNotFound)
			}
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}
		if err := redisClient.Set(context.Background(), urlFound.ShortenUrl, urlFound.OriginalUrl, 1*60*time.Second).Err(); err != nil {
			log.Fatal(err)
		}
		return ctx.Redirect(urlFound.OriginalUrl, 302)
	} else if err != nil {
		log.Fatal(err)
	}
	// if it does exist in cache, return that as response
	return ctx.Redirect(val, 302)
}

var DB *mongo.Client = connectToMongodb()
var urlCollection *mongo.Collection = GetCollection(DB, "urls")

func main() {

	app := fiber.New()
	app.Use(logger.New())

	defer DB.Disconnect(context.Background())
	createMongoDbIndex()
	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello World")
	})
	app.Post("/api/v1", createShortenURL)
	app.Get("/:url", resolveURL)
	log.Fatal(app.Listen(os.Getenv("APP_PORT")))
}

func connectToMongodb() *mongo.Client {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("You must set your 'MONGODB_URI' environmental variable.")
	}
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	return client
}

func createMongoDbIndex() {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "shorten_url", Value: -1}},
		Options: options.Index().SetUnique(true),
	}
	name, err := urlCollection.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		panic(err)
	}
	log.Printf("Created MongoDB Index %s", name)
}

func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("golangAPI").Collection(collectionName)
	return collection
}

func connectToRedis(dbNo int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URI"),
		Password: os.Getenv("REDIS_PASS"),
		DB:       dbNo,
	})
	return client
}
