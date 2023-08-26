package handler

import (
	"HoBot_Backend/pkg/model"
	usetService "HoBot_Backend/pkg/service"
	"github.com/gofiber/fiber/v2"
	"time"
)

func Register(c *fiber.Ctx) error {
	user := new(model.User)

	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	res, err := usetService.Registration(*user)
	if err != nil {
		return err
	}

	cookie := new(fiber.Cookie)
	cookie.Name = "refreshToken"
	cookie.Value = res.RefreshToken
	cookie.Expires = time.Now().Add(24 * time.Hour)
	c.Cookie(cookie)

	return c.JSON(res)
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
