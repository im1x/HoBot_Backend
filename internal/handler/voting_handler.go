package handler

import (
	"HoBot_Backend/internal/service/voting"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func StartVoting(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)

	var votingRequest voting.VotingRequest
	err := c.BodyParser(&votingRequest)
	if err != nil {
		log.Error("Voting: Error while parsing body:", err)
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	voting.StartVoting(userId, votingRequest)

	return nil
}

func GetVotingState(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)
	result := voting.GetVotingStatus(userId)
	return c.JSON(result)
}

func StopVoting(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)
	voting.StopVoting(userId)
	return nil
}
