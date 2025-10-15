package handler

import (
	"HoBot_Backend/internal/service/common"
	"HoBot_Backend/internal/telegram"
	"os"

	"github.com/gofiber/fiber/v2"
)

type CommonHandler struct {
	commonService common.CommonService
}

func NewCommonHandler(commonService common.CommonService) *CommonHandler {
	return &CommonHandler{commonService: commonService}
}

func (s *CommonHandler) Feedback(c *fiber.Ctx) error {
	requestBody := c.Body()
	feedbackText := string(requestBody)

	if feedbackText == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Feedback text is empty"})
	}

	userId := parseUserIdFromRequest(c)

	telegram.SendMessage(userId + ": " + feedbackText)

	err := s.commonService.AddFeedback(c.Context(), userId, feedbackText)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(err.Error())
	}

	return nil
}

func (s *CommonHandler) TerminateApp(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if id == os.Getenv("TERMINATE_CODE") {
		os.Exit(0)
	}

	return c.SendStatus(fiber.StatusOK)
}
