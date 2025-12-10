package ports

type HallRepository interface {
	CreateHall(cinemaID uint, name string, rows, seatsPerRow int) (uint, error)
}
