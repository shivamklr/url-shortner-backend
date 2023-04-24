package main

import (
	"context"
	"log"
	"os"
	"strconv"
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

type ShortenUrlModel struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OriginalUrl string             `json:"original_url" bson:"original_url,omitempty"`
	ExpireIn    time.Duration      `json:"expire_in" bson:"expire,omitempty"`
	Shorten     string             `json:"shorten" bson:"shorten,omitempty"`
	ShortenUrl  string             `json:"shorten_url" bson:"omitempty"`
}

type ShortenUrlCreateModel struct {
	OriginalUrl string        `json:"original_url" validate:"required"`
	ExpireIn    time.Duration `json:"expire_in" validate:"required"`
}

type ShortenResponse struct {
	Status  int        `json:"status"`
	Message string     `json:"message"`
	Data    *fiber.Map `json:"data"`
}

func createShortenURL(ctx *fiber.Ctx) error {

	request := new(ShortenUrlCreateModel)

	if err := ctx.BodyParser(request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse JSON",
		})
	}

	// Create a struct instance
	urlRecord := &ShortenUrlModel{
		OriginalUrl: request.OriginalUrl,
		ExpireIn:    request.ExpireIn,
		Shorten:     uuid.New().String()[:8],
	}

	result, err := urlCollection.InsertOne(context.Background(), urlRecord)
	if err != nil {
		log.Fatal(err)

	}
	recordId, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		log.Fatal("Type Assertion Failed")
	}
	urlRecord.Id = recordId

	redisClient := connectToRedis(0)
	defer redisClient.Close()

	redisTtl, _ := strconv.Atoi(os.Getenv("REDIS_TTL"))

	if err := redisClient.Set(context.Background(), urlRecord.Shorten, urlRecord.OriginalUrl, time.Duration(redisTtl)*60*time.Second).Err(); err != nil {
		log.Fatal(err)
	}
	//  add ttl of 2 minutes.
	urlRecord.ShortenUrl = os.Getenv("DOMAIN") + "/" + urlRecord.Shorten
	return ctx.Status(fiber.StatusCreated).JSON(ShortenResponse{Status: fiber.StatusCreated, Message: "success", Data: &fiber.Map{"data": urlRecord}})
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
		var urlFound ShortenUrlModel
		filter := ShortenUrlModel{
			Shorten: url,
		}
		if err := urlCollection.FindOne(context.Background(), filter).Decode(&urlFound); err != nil {
			if err == mongo.ErrNoDocuments {
				return ctx.SendStatus(fiber.StatusNotFound)
			}
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}
		redisTtl, _ := strconv.Atoi(os.Getenv("REDIS_TTL"))
		if err := redisClient.Set(context.Background(), urlFound.Shorten, urlFound.OriginalUrl, time.Duration(redisTtl)*60*time.Second).Err(); err != nil {
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
		Keys:    bson.D{{Key: "shorten", Value: -1}},
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
