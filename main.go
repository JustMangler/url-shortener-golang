package main

import (
	"go-url-short/database"
	"go-url-short/handler"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	if err := database.Connect(); err != nil {
		log.Fatal(err)
	}

	app := fiber.New()

	app.Get("/api/:id", handler.GetShortURL)

	app.Post("/api/create", handler.CreateShortURL)

	log.Fatalln(app.Listen(":3001"))
}
