package ports

import "cinema/domain/entities"

type CinemaRepository interface {
	// Кинотеатры
	CreateCinema(cinema entities.Cinema) (int, error)
	ListCinemas() ([]entities.Cinema, error)
	GetCinema(id int) (entities.Cinema, error)
}
