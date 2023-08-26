package handler

import (
	"HoBot_Backend/pkg/model"
	usetService "HoBot_Backend/pkg/service"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
)

func Register(c *fiber.Ctx) error {
	user := new(model.User)

	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	fmt.Printf("%+v\n", user)
	s, _ := json.MarshalIndent(user, "", "\t")
	fmt.Print(string(s))

	res, err := usetService.Registration(*user)
	if err != nil {
		return err
	}

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
