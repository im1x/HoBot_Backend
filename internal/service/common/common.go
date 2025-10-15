package common

import (
	"HoBot_Backend/internal/model"
	repoFeedback "HoBot_Backend/internal/repository/feedback"
	"context"
	"time"

	"github.com/gofiber/fiber/v2/log"
)

type commonService struct {
	ctxApp       context.Context
	feedbackRepo repoFeedback.Repository
}

func NewCommonService(ctx context.Context, feedbackRepo repoFeedback.Repository) CommonService {
	return &commonService{
		ctxApp:       ctx,
		feedbackRepo: feedbackRepo,
	}
}

func (s *commonService) AddFeedback(ctx context.Context, userId, feedbackText string) error {
	feedback := model.Feedback{
		UserId:  userId,
		Text:    feedbackText,
		AddedAt: time.Now().Format(time.DateTime),
	}

	err := s.feedbackRepo.AddFeedback(ctx, feedback)
	if err != nil {
		log.Error("Error while inserting feedback:", err)
		return err
	}
	return nil
}
