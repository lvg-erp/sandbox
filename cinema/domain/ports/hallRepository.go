package ports

import "cinema/domain/entities"

type HallRepository interface {
	CreateHall(cinemaID uint, name string, rows, seatsPerRow int) (uint, error)
	ListHalls() ([]entities.Hall, error)
	GenerateSeats(hallID uint, rows, seatsPerRow int) error
}
