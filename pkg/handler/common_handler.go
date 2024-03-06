package handler

import (
	commonService "HoBot_Backend/pkg/service/common"
	"github.com/gofiber/fiber/v2"
)

func Feedback(c *fiber.Ctx) error {
	requestBody := c.Body()
	feedbackText := string(requestBody)

	if feedbackText == "" {
		return c.Status(fiber.StatusBadRequest).JSON("Feedback text is empty")
	}

	userId := parseUserIdFromRequest(c)

	err := commonService.AddFeedback(c.Context(), userId, feedbackText)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	return nil
}
