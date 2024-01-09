package handler

import (
	"HoBot_Backend/pkg/model"
	"HoBot_Backend/pkg/service"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"time"
)

func GetCommands(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)
	commands, err := service.GetCommands(c.Context(), userId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(commands)
}
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
	newCommand := new(model.CommonCommand)

	if err := c.BodyParser(&newCommand); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if err := validate.Struct(newCommand); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err.Error())
	}
	userId := parseUserIdFromRequest(c)

	fmt.Println("userId: ", userId)

	commandList, err := service.AddCommandForUser(c.Context(), userId, newCommand)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}
	return c.JSON(commandList)
}
