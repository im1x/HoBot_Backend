package common

import "context"

type CommonService interface {
	AddFeedback(ctx context.Context, userId, feedbackText string) error
}
