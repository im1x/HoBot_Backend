package feedback

import (
	"HoBot_Backend/internal/model"
	"HoBot_Backend/internal/mongodb"
	"context"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type feedbackRepository struct {
	col *mongo.Collection
}

func NewFeedbackRepository(client *mongodb.Client) Repository {
	return &feedbackRepository{
		col: client.GetCollection(mongodb.Feedback),
	}
}

func (r *feedbackRepository) AddFeedback(ctx context.Context, feedback model.Feedback) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, err := r.col.InsertOne(ctx, feedback)
	if err != nil {
		log.Error("Error while inserting feedback:", err)
		return err
	}
	return nil
}
