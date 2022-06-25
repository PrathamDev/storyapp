package router

import (
	"github.com/PrathamDev/storyapp/handler"
	"github.com/PrathamDev/storyapp/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func SetupRouter(app *fiber.App) {
	api := app.Group("/api", logger.New())

	//Auth
	auth := api.Group("/auth")
	auth.Post("/signup", handler.SignupWithEmailAndPassword)
	auth.Post("/login", handler.Login)
	auth.Post("/token", handler.Token)

	//User
	user := api.Group("/user")
	user.Post("/", handler.CreateUser)
	user.Get("/:id", middleware.CheckToken(), middleware.IsAdmin(), handler.GetUser)
	user.Post("/profile", middleware.CheckToken(), handler.UploadProfilePicture)
}
