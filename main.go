package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
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

//	type CreateShortenUrlResponse struct {
//		// create fields to be a part of response instance
//	}
var urlCollection *mongo.Collection = GetCollection(DB, "urls")

func createShortenURL(ctx *fiber.Ctx) error {

	body := new(ShortenUrl)

	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse JSON",
		})
	}
	// insert data to mongodb database,
	res, err := urlCollection.InsertOne(context.Background(), body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Inserted document with _id: %v\n", res.InsertedID)
	// add ttl of 2 hours
	// if data is inserted add cache data to redis add ttl of 1 minute.
	return ctx.SendStatus(fiber.StatusCreated)
}

// func resolveURL(ctx *fiber.Ctx) error {
// 	// parse body from context and check for errors

// 	// check whether the shorten URL key exist in redis cache

// 	// if it does, return that as response

// 	// if it does not add it in the cache and return it

// }
var DB *mongo.Client = connectToMongodb()

func main() {

	app := fiber.New()
	app.Use(logger.New())

	defer DB.Disconnect(context.Background())

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello World")
	})
	app.Post("/api/v1", createShortenURL)

	app.Listen(":3000")
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
