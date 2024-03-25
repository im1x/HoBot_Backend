package handler

import (
	commonService "HoBot_Backend/pkg/service/common"
	"github.com/gofiber/fiber/v2"
	"os"
)

func Feedback(c *fiber.Ctx) error {
	requestBody := c.Body()
	feedbackText := string(requestBody)

	if feedbackText == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Feedback text is empty"})
	}

	userId := parseUserIdFromRequest(c)

	err := commonService.AddFeedback(c.Context(), userId, feedbackText)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	return nil
}

func TerminateApp(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if id == os.Getenv("TERMINATE_CODE") {
		os.Exit(0)
	}

	return c.SendStatus(fiber.StatusOK)
}
