package handler

import (
	"HoBot_Backend/pkg/model"
	userService "HoBot_Backend/pkg/service"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"time"
)

var validate = validator.New()

func Register(c *fiber.Ctx) error {
	user := new(model.User)

	if err := c.BodyParser(&user); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if err := validate.Struct(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	res, err := userService.Registration(*user)
	if err != nil {
		log.Info(err)
		return err
	}

	cookie := new(fiber.Cookie)
	cookie.Name = "refreshToken"
	cookie.Value = res.RefreshToken
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.HTTPOnly = true
	c.Cookie(cookie)

	return c.JSON(res)
}

func Login(c *fiber.Ctx) error {
	user := new(model.User)

	if err := c.BodyParser(&user); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if err := validate.Struct(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}

	res, err := userService.Login(*user)
	if err != nil {
		return err
	}

	cookie := new(fiber.Cookie)
	cookie.Name = "refreshToken"
	cookie.Value = res.RefreshToken
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.HTTPOnly = true
	c.Cookie(cookie)

	return c.JSON(res)
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
