package common

import (
	"HoBot_Backend/internal/model"
	DB "HoBot_Backend/internal/mongo"
	"context"
	"time"

	"github.com/gofiber/fiber/v2/log"
)

func AddFeedback(ctx context.Context, userId, feedbackText string) error {
	feedback := model.Feedback{
		UserId:  userId,
		Text:    feedbackText,
		AddedAt: time.Now().Format(time.DateTime),
	}

	_, err := DB.GetCollection(DB.Feedback).InsertOne(ctx, feedback)
	if err != nil {
		log.Error("Error while inserting feedback:", err)
		return err
	}
	return nil
}
