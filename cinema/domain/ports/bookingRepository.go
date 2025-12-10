package ports

import "cinema/domain/entities"

type BookingRepository interface {
	CreateBooking(entities.Booking) error
}
