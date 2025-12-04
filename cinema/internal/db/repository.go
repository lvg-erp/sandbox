// internal/db/repository.go
package db

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"golang.org/x/exp/rand"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// Repository — интерфейс для всех операций с БД
type Repository interface {
	// Пользователи
	CreateUser(User) error
	GetUser(email string) (User, error)
	GetUserByID(id int) (User, error)

	// Фильмы
	CreateFilm(Film) error
	ListFilms() ([]Film, error)
	GetFilm(id int) (Film, error)

	// Кинотеатры
	CreateCinema(Cinema) error
	ListCinemas() ([]Cinema, error)
	GetCinema(id int) (Cinema, error)

	// Сеансы
	GetAllSessions() ([]SessionInfo, error)
	CreateSession(userID int, expires time.Time) (string, error)
	GetSession(token string) (Session, error)
	DeleteSession(token string) error
	//CreateFilmSession(filmID, cinemaID int, startTime time.Time, price float64) (int, error)
	CreateFilmSession(filmID, cinemaID int, start time.Time) (int, error)
	CreateBooking(Booking) error
	DeleteUserSessions(userID int) error

	GetSeats(sessionID int) ([]Seat, error)
	BookSeat(userID, seatID int) error

	ListSessionsForCinema(cinemaID int) ([]map[string]interface{}, error)
}

// repo — реализация
type repo struct {
	db *sql.DB
}

// NewRepo — конструктор
func NewRepo(db *sql.DB) Repository {
	return &repo{db: db}
}

// === ПОЛЬЗОВАТЕЛИ ===

// CreateUser — регистрация
func (r *repo) CreateUser(u User) error {
	_, err := r.db.Exec(`
		INSERT INTO users (email, pass, role) VALUES ($1, $2, $3)`,
		u.Email, u.Pass, "user",
	)
	return err
}

// GetUser — по email
func (r *repo) GetUser(email string) (User, error) {
	var u User
	err := r.db.QueryRow(`
		SELECT id, email, pass, role FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.Pass, &u.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return u, errors.New("user not found")
		}
		return u, err
	}
	return u, nil
}

// GetUserByID — по ID
func (r *repo) GetUserByID(id int) (User, error) {
	var u User
	err := r.db.QueryRow(`
		SELECT id, email, pass, role FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.Pass, &u.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return u, errors.New("user not found")
		}
		return u, err
	}
	return u, nil
}

// === ФИЛЬМЫ ===

// CreateFilm — добавить фильм
func (r *repo) CreateFilm(f Film) error {
	_, err := r.db.Exec(`
		INSERT INTO films (title, poster, description, duration)
		VALUES ($1, $2, $3, $4)`,
		f.Title, f.Poster, f.Description, f.Duration,
	)
	return err
}

// ListFilms — все фильмы
func (r *repo) ListFilms() ([]Film, error) {
	rows, err := r.db.Query(`
		SELECT id, title, poster, description, duration
		FROM films ORDER BY added_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var films []Film
	for rows.Next() {
		var f Film
		if err := rows.Scan(&f.ID, &f.Title, &f.Poster, &f.Description, &f.Duration); err != nil {
			return nil, err
		}
		films = append(films, f)
	}
	return films, rows.Err()
}

// GetFilm — по ID
func (r *repo) GetFilm(id int) (Film, error) {
	var f Film
	err := r.db.QueryRow(`
		SELECT id, title, poster, description, duration
		FROM films WHERE id = $1`, id,
	).Scan(&f.ID, &f.Title, &f.Poster, &f.Description, &f.Duration)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return f, errors.New("film not found")
		}
		return f, err
	}
	return f, nil
}

// === КИНОТЕАТРЫ ===

// CreateCinema — добавить кинотеатр
func (r *repo) CreateCinema(cinema Cinema) error {
	_, err := r.db.Exec(`
		INSERT INTO cinemas (name, address, city, phone, total_seats, poster)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		cinema.Name, cinema.Address, cinema.City, cinema.Phone, cinema.TotalSeats, cinema.Poster,
	)
	return err
}

// ListCinemas — все кинотеатры
func (r *repo) ListCinemas() ([]Cinema, error) {
	rows, err := r.db.Query(`
		SELECT id, name, address, city, phone, total_seats, poster, created_at
		FROM cinemas ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cinemas []Cinema
	for rows.Next() {
		var c Cinema
		if err := rows.Scan(&c.ID, &c.Name, &c.Address, &c.City, &c.Phone, &c.TotalSeats, &c.Poster, &c.CreatedAt); err != nil {
			return nil, err
		}
		cinemas = append(cinemas, c)
	}
	return cinemas, rows.Err()
}

// GetCinema — по ID
func (r *repo) GetCinema(id int) (Cinema, error) {
	var c Cinema
	err := r.db.QueryRow(`
		SELECT id, name, address, city, phone, total_seats, poster, created_at
		FROM cinemas WHERE id = $1`, id,
	).Scan(&c.ID, &c.Name, &c.Address, &c.City, &c.Phone, &c.TotalSeats, &c.Poster, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c, errors.New("cinema not found")
		}
		return c, err
	}
	return c, nil
}

// === СЕАНСЫ ===

// GenerateSessionToken — случайный токен 32 байта
func GenerateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CreateSession — создаёт сессию в БД

func (r *repo) CreateSession(userID int, expires time.Time) (string, error) {
	token, err := GenerateSessionToken()
	if err != nil {
		return "", err
	}
	_, err = r.db.Exec(`
        INSERT INTO sessions (user_id, token, expires_at)
        VALUES ($1, $2, $3)`,
		userID, token, expires,
	)
	return token, err
}

func (r *repo) GetSession(token string) (Session, error) {
	var s Session
	err := r.db.QueryRow(`
        SELECT id, user_id, token, expires_at FROM sessions WHERE token = $1`, token,
	).Scan(&s.ID, &s.UserID, &s.Token, &s.ExpiresAt)
	if err != nil {
		return s, err
	}
	if time.Now().After(s.ExpiresAt) {
		r.DeleteSession(token)
		return s, errors.New("session expired")
	}
	return s, nil
}

func (r *repo) DeleteSession(token string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE token = $1`, token)
	return err
}

// CreateBooking — бронирование
func (r *repo) CreateBooking(b Booking) error {
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

func (r *repo) DeleteUserSessions(userID int) error {
	log.Printf("DeleteUserSessions: deleting for user_id=%d", userID)
	result, err := r.db.Exec(`DELETE FROM sessions WHERE user_id = $1`, userID)
	if err != nil {
		log.Printf("DeleteUserSessions: ERROR: %v", err)
		return err
	}
	rows, _ := result.RowsAffected()
	log.Printf("DeleteUserSessions: deleted %d rows", rows)
	return nil
}

// Создать сеанс без мест
func (r *repo) CreateFilmSession(filmID, cinemaID int, start time.Time) (int, error) {
	var sessionID int
	err := r.db.QueryRow(`
        INSERT INTO film_sessions (film_id, cinema_id, start_time)
        VALUES ($1, $2, $3) RETURNING id`, filmID, cinemaID, start).Scan(&sessionID)
	return sessionID, err
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

func (r *repo) GetSeats(sessionID int) ([]Seat, error) {
	rows, err := r.db.Query(`
        SELECT s.id, s.row, s.col, b.id IS NOT NULL AS booked
        FROM seats s
        LEFT JOIN bookings b ON s.id = b.seat_id
        WHERE s.session_id = $1
        ORDER BY s.row, s.col`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seats []Seat
	for rows.Next() {
		var s Seat
		if err := rows.Scan(&s.ID, &s.HallID, &s.Row, &s.Number, &s.Reserved); err != nil {
			return nil, err
		}
		seats = append(seats, s)
	}
	return seats, nil
}

// BookSeat — бронирует место, если оно свободно
func (r *repo) BookSeat(userID, seatID int) error {
	tx, err := r.db.Begin() // ← r.db
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM bookings WHERE seat_id = $1)`, seatID).Scan(&exists) // ← r.db
	if err != nil {
		return err
	}
	if exists {
		return errors.New("seat already booked")
	}

	_, err = tx.Exec(`INSERT INTO bookings (user_id, seat_id) VALUES ($1, $2)`, userID, seatID) // ← r.db
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *repo) GetAllSessions() ([]SessionInfo, error) {
	rows, err := r.db.Query(`
        SELECT fs.id, f.title, fs.start_time
        FROM film_sessions fs
        JOIN films f ON fs.film_id = f.id
        ORDER BY fs.start_time`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []SessionInfo
	for rows.Next() {
		var s SessionInfo
		if err := rows.Scan(&s.ID, &s.FilmTitle, &s.StartTime); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *repo) CreateHall(cinemaID uint, name string, rows, seatsPerRow int) (uint, error) {
	var hallID uint
	err := r.db.QueryRow(`
		INSERT INTO halls (cinema_id, name, rows, seats_per_row) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id`,
		cinemaID, name, rows, seatsPerRow,
	).Scan(&hallID)
	return hallID, err
}

func (r *repo) CreateSeat(hallID uint, row, number int) error {
	_, err := r.db.Exec(`
		INSERT INTO seats (hall_id, row, number, reserved) 
		VALUES ($1, $2, $3, false)`,
		hallID, row, number,
	)
	return err
}

func (r *repo) ListSessionsForCinema(cinemaID int) ([]map[string]interface{}, error) {
	rows, err := r.db.Query(`
        SELECT fs.id, f.title as film_title, f.duration, fs.start_time, c.name as cinema_name 
        FROM film_sessions fs 
        JOIN films f ON fs.film_id = f.id 
        JOIN cinemas c ON fs.cinema_id = c.id 
        WHERE fs.cinema_id = $1`, cinemaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []map[string]interface{}
	for rows.Next() {
		var s struct {
			ID         int
			FilmTitle  string
			Duration   int
			StartTime  string
			CinemaName string
		}
		err := rows.Scan(&s.ID, &s.FilmTitle, &s.Duration, &s.StartTime, &s.CinemaName)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, map[string]interface{}{
			"id":          s.ID,
			"film_title":  s.FilmTitle,
			"duration":    s.Duration,
			"start_time":  s.StartTime,
			"cinema_name": s.CinemaName,
		})
	}
	return sessions, nil
}

func (r *repo) ListCinemasWithSessions() ([]map[string]interface{}, error) {
	rows, err := r.db.Query(`
        SELECT c.id, c.name, COUNT(fs.id) as session_count
        FROM cinemas c LEFT JOIN film_sessions fs ON c.id = fs.cinema_id
        GROUP BY c.id, c.name
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cinemas []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		var count int
		err := rows.Scan(&id, &name, &count)
		if err != nil {
			return nil, err
		}
		cinemas = append(cinemas, map[string]interface{}{
			"id":            id,
			"name":          name,
			"session_count": float64(count), // для JSON float64
		})
	}
	return cinemas, nil
}
