package handler

import (
	"HoBot_Backend/internal/service/voting"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type VotingHandler struct {
	votingService *voting.VotingService
}

func NewVotingHandler(votingService *voting.VotingService) *VotingHandler {
	return &VotingHandler{votingService: votingService}
}

func (s *VotingHandler) StartVoting(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)

	var votingRequest voting.VotingRequest
	err := c.BodyParser(&votingRequest)
	if err != nil {
		log.Error("Voting: Error while parsing body:", err)
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	s.votingService.StartVoting(userId, votingRequest)

	return nil
}

func (s *VotingHandler) GetVotingState(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)
	result := s.votingService.GetVotingStatus(userId)
	return c.JSON(result)
}

func (s *VotingHandler) StopVoting(c *fiber.Ctx) error {
	userId := parseUserIdFromRequest(c)
	s.votingService.StopVoting(userId)
	return nil
}
