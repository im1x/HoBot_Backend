package feedback

import (
	"HoBot_Backend/internal/model"
	"context"
)

type Repository interface {
	AddFeedback(ctx context.Context, feedback model.Feedback) error
}
