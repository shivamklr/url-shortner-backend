package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type CreateShortenUrlRequest struct {
	originalUrl string `json:"original_url"`
	expireIn    string `json:"expire_in"`
}

type CreateShortenUrlResponse struct {
	// create fields to be a part of response instance
}

func createShortenURL(ctx *fiber.Ctx) error {

	body := new(CreateShortenUrlRequest)

	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse JSON",
		})
	}
	// insert data to mongodb database, add ttl of 2 hours
	// if data is inserted add cache data to redis add ttl of 1 minute.

}

func resolveURL(ctx *fiber.Ctx) error {
	// parse body from context and check for errors

	// check whether the shorten URL key exist in redis cache

	// if it does, return that as response

	// if it does not add it in the cache and return it

}

func main() {

	app := fiber.New()
	app.Use(logger.New())

	// connect to mongodb instance
	// connect to redis instance

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello World")
	})

	app.Post("/api/v1", createShortenURL)

	app.Listen(":3000")
}
