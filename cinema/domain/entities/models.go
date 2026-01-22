package entities

import "time"

// User — модель пользователя
type User struct {
	ID    int    `json:"id,omitempty"`
	Email string `json:"email"`
	Pass  string `json:"password,omitempty"` // не отдаём в JSON
	Role  string `json:"role,omitempty"`     // "user" или "admin"
}

// Film — модель фильма
type Film struct {
	ID          int    `json:"id,omitempty"`
	Title       string `json:"title"`
	Poster      string `json:"poster,omitempty"`
	Description string `json:"description,omitempty"`
	Duration    int    `json:"duration,omitempty"` // в минутах
	AddedAt     string `json:"added_at,omitempty"` // ISO string, если нужно
}

// Session — модель сеанса
type Session struct {
	ID        int       `json:"id,omitempty"`
	UserID    int       `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// кинотеатры
type Cinema struct {
	ID         int    `json:"id,omitempty"`
	Name       string `json:"name"`
	Address    string `json:"address,omitempty"`
	City       string `json:"city,omitempty"`
	Phone      string `json:"phone,omitempty"`
	TotalSeats int    `json:"total_seats,omitempty"` // общее количество мест
	Poster     string `json:"poster,omitempty"`      // логотип/фото кинотеатра
	CreatedAt  string `json:"created_at,omitempty"`  // ISO время
}

type SessionFilm struct {
	ID        int    `json:"id"`
	FilmTitle string `json:"film_title"`
	Duration  int    `json:"duration,omitempty"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type Booking struct {
	ID        int
	UserID    int
	SessionID int
	SeatIDs   []int `json:"seat_ids"` // ← ОБЯЗАТЕЛЬНО!
}

type Hall struct {
	ID          uint   `json:"id"`
	CinemaID    uint   `json:"cinema_id"`
	Name        string `json:"name"`
	Rows        int    `json:"rows"`
	SeatsPerRow int    `json:"seats_per_row"`
}

type Seat struct {
	ID       uint `json:"id"`
	HallID   uint `json:"hall_id"`
	Row      int  `json:"row"`
	Number   int  `json:"number"`
	Reserved bool `json:"reserved"`
}

//type Booking struct {
//	ID        int    `json:"id" db:"id"`
//	UserID    int    `json:"user_id" db:"user_id"`
//	SessionID int    `json:"session_id" db:"session_id"`
//	SeatID    int    `json:"seat_id" db:"seat_id"`
//	FilmTitle string `json:"film_title" db:"film_title"`
//	HallID    int    `json:"hall_id" db:"hall_id"`
//	Row       int    `json:"row" db:"row"`
//	Seat      int    `json:"seat" db:"seat"`
//	Date      string `json:"date" db:"date"` // "2025-11-15"
//	Time      string `json:"time" db:"time"` // "18:30"
//	BookedAt  string `json:"booked_at" db:"booked_at"`
//}

// Booking — модель бронирования (если будет)
//type Booking struct {
//	UserID    int
//	SessionID int
//	SeatIDs   []int // ← ID мест из seats.id
//}
