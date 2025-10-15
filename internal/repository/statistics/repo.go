package statistics

import "context"

type Repository interface {
	IncField(ctx context.Context, userId string, fieldName UpdateName)
}
