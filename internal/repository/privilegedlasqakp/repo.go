package privilegedlasqakp

import "HoBot_Backend/internal/model"

type Repository interface {
	GetMovies() ([]model.MovieKp, error)
}
