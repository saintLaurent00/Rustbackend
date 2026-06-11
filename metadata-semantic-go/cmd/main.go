package main

import (
	"context"
	"log"
	"os"

	"bi-metadata/ent"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	_ = godotenv.Load("../.env")

	dsn := os.Getenv("POSTGRES_METADATA_URL")
	client, err := ent.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	defer client.Close()

	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	app := fiber.New()

	api := app.Group("/api/v1")

    api.Get("/health", func(c *fiber.Ctx) error {
        return c.SendString("OK")
    })

	log.Fatal(app.Listen(":8081"))
}
