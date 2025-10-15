package privilegedlasqakp

import (
	"HoBot_Backend/internal/model"
	"HoBot_Backend/internal/mongodb"
	"context"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type lasqaKpRepository struct {
	col *mongo.Collection
}

func NewLasqaKpRepository(client *mongodb.Client) Repository {
	return &lasqaKpRepository{
		col: client.GetCollection(mongodb.PrivilegedLasqaKp),
	}
}

func (r *lasqaKpRepository) GetMovies() ([]model.MovieKp, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		log.Error("Error while finding movies:", err)
		return nil, err
	}

	var movies []model.MovieKp
	if err = cursor.All(ctx, &movies); err != nil {
		log.Error("Error while decoding movies:", err)
		return nil, err
	}
	return movies, nil
}
