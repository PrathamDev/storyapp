package main

import (
	"os"

	"github.com/PrathamDev/storyapp/database"
	"github.com/PrathamDev/storyapp/router"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New(fiber.Config{BodyLimit: 20 * 1024 * 1024})

	app.Use(cors.New())

	database.ConnectDB()

	router.SetupRouter(app)

	app.Listen(":" + os.Getenv("PORT"))

}
