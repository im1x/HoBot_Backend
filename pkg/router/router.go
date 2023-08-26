package router

import (
	userHandler "HoBot_Backend/pkg/handler"
	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/register", userHandler.Register)
	api.Post("/login", userHandler.Login)
	api.Post("/logout", userHandler.Logout)
	api.Get("/refresh", userHandler.Refresh)
	api.Get("/users", userHandler.Users)
}
