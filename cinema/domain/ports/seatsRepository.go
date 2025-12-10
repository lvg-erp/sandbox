package ports

import "cinema/domain/entities"

type SeatsRepository interface {
	CreateSeat(hallID uint, row, number int) error
	GetSeats(sessionID int) ([]entities.Seat, error)
	BookSeat(userID, seatID int) error
}
