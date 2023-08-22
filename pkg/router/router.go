package router

import (
	"HoBot_Backend/pkg/controller"
	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/register", controller.Register)
	api.Post("/login", controller.Login)
	api.Post("/logout", controller.Logout)
	api.Get("/refresh", controller.Refresh)
	api.Get("/users", controller.Users)
}
