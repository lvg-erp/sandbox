package ports

import "cinema/domain/entities"

type CinemaRepository interface {
	// Кинотеатры
	CreateCinema(entities.Cinema) error
	ListCinemas() ([]entities.Cinema, error)
	GetCinema(id int) (entities.Cinema, error)
}
