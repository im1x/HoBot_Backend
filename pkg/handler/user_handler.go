package handler

import (
	"HoBot_Backend/pkg/model"
	"HoBot_Backend/pkg/service"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
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

	res, err := service.Registration(c.Context(), *user)
	if err != nil {
		log.Info(err)
		return err
	}

	cookie := new(fiber.Cookie)
	cookie.Name = "refreshToken"
	cookie.Value = res.RefreshToken
	cookie.Expires = time.Now().Add(1440 * time.Hour)
	cookie.HTTPOnly = true
	cookie.SameSite = "None"
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

	res, err := service.Login(c.Context(), *user)
	if err != nil {
		return err
	}

	cookie := new(fiber.Cookie)
	cookie.Name = "refreshToken"
	cookie.Value = res.RefreshToken
	cookie.Expires = time.Now().Add(1440 * time.Hour)
	cookie.HTTPOnly = true
	c.Cookie(cookie)

	return c.JSON(res)
}

func Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refreshToken")
	if refreshToken == "" {
		return fiber.NewError(fiber.StatusUnauthorized)
	}

	err := service.Logout(c.Context(), refreshToken)
	if err != nil {
		return err
	}
	c.ClearCookie()

	return nil
}

func Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refreshToken")
	if refreshToken == "" {
		return fiber.NewError(fiber.StatusUnauthorized)
	}
	res, err := service.RefreshToken(c.Context(), refreshToken)
	if err != nil {
		return err
	}

	cookie := new(fiber.Cookie)
	cookie.Name = "refreshToken"
	cookie.Value = res.RefreshToken
	cookie.Expires = time.Now().Add(1440 * time.Hour)
	cookie.HTTPOnly = true
	c.Cookie(cookie)

	return c.JSON(res)
}

func Users(c *fiber.Ctx) error {
	name := parseUserIdFromRequest(c)
	return c.SendString("Welcome " + name)
}

func parseUserIdFromRequest(c *fiber.Ctx) string {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	return claims["id"].(string)
}
