package controller

import "github.com/gofiber/fiber/v2"

func Register(c *fiber.Ctx) error {

	return nil
}

func Login(c *fiber.Ctx) error {
	return c.SendString("123")
}

func Logout(c *fiber.Ctx) error {

	return nil
}

func Refresh(c *fiber.Ctx) error {

	return nil
}

func Users(c *fiber.Ctx) error {

	return nil
}
