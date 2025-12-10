package repository

import (
	"cinema/domain/entities"
	"cinema/domain/ports"
	"database/sql"
	"errors"
	_ "github.com/lib/pq"
)

// repo — реализация
type bookingRepo struct {
	db *sql.DB
}

// NewRepo — конструктор
func NewBookingRepo(db *sql.DB) ports.BookingRepository {
	return &bookingRepo{db: db}
}

// CreateBooking — бронирование
func (r *bookingRepo) CreateBooking(b entities.Booking) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Проверяем и бронируем каждое место
	for _, seatID := range b.SeatIDs {
		// Проверяем, что место свободно и принадлежит сеансу
		var exists bool
		err := tx.QueryRow(`
            SELECT EXISTS(
                SELECT 1 FROM seats s
                LEFT JOIN bookings b ON s.id = b.seat_id
                WHERE s.id = $1 AND s.session_id = $2 AND b.id IS NULL
            )`, seatID, b.SessionID,
		).Scan(&exists)
		if err != nil {
			return err
		}
		if !exists {
			return errors.New("seat already taken or invalid")
		}

		// 2. Создаём бронь
		_, err = tx.Exec(`
            INSERT INTO bookings (user_id, seat_id) VALUES ($1, $2)`,
			b.UserID, seatID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Создать сеанс и места
//func (r *repo) CreateFilmSession(filmID, cinemaID int, start time.Time, rows, cols int) (int, error) {
//	var sessionID int
//	err := r.db.QueryRow(`
//        INSERT INTO film_sessions (film_id, cinema_id, start_time)
//        VALUES ($1, $2, $3) RETURNING id`, filmID, cinemaID, start,
//	).Scan(&sessionID)
//	if err != nil {
//		return 0, err
//	}
//
//	stmt, err := r.db.Prepare(`INSERT INTO seats (session_id, row, col) VALUES ($1, $2, $3)`)
//	if err != nil {
//		return 0, err
//	}
//	defer stmt.Close()
//
//	for row := 1; row <= rows; row++ {
//		for col := 1; col <= cols; col++ {
//			if _, err := stmt.Exec(sessionID, row, col); err != nil {
//				return 0, err
//			}
//		}
//	}
//
//	return sessionID, nil
//}

// Получить места для сеанса
