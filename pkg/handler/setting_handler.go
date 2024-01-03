package handler

import (
	"HoBot_Backend/pkg/model"
	"HoBot_Backend/pkg/service"
	"HoBot_Backend/pkg/service/vkplay"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"time"
)

func GetCommandsList(c *fiber.Ctx) error {
	// Record start time
	startTime := time.Now()

	commandsList, err := service.GetCommandsList()

	// Record end time
	endTime := time.Now()
	// Calculate and print the execution time in milliseconds
	executionTime := endTime.Sub(startTime).Milliseconds()
	fmt.Printf("Function execution time: %d ms\n", executionTime)

	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(commandsList.Commands)
}

func AddCommandAndAlias(c *fiber.Ctx) error {
	newCommand := new(model.NewCommand)

	if err := c.BodyParser(&newCommand); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if err := validate.Struct(newCommand); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}
	userId := parseUserIdFromRequest(c)

	fmt.Println("userId: ", userId)

	commandList, err := vkplay.AddCommandForUser(userId, newCommand)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	return c.JSON(commandList)
}
