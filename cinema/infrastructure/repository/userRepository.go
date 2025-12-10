package repository

import (
	"cinema/domain/entities"
	"cinema/domain/ports"
	"database/sql"
	"errors"
	"log"
)

type userRepo struct {
	db *sql.DB
}

// NewRepo — конструктор
func NewUserRepo(db *sql.DB) ports.UserRepository {
	return &userRepo{db: db}
}

// CreateUser — регистрация
func (r *userRepo) CreateUser(u entities.User) error {
	_, err := r.db.Exec(`
		INSERT INTO users (email, pass, role) VALUES ($1, $2, $3)`,
		u.Email, u.Pass, "user",
	)
	return err
}

// GetUser — по email
func (r *userRepo) GetUser(email string) (entities.User, error) {
	var u entities.User
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
func (r *userRepo) GetUserByID(id int) (entities.User, error) {
	var u entities.User
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

func (r *userRepo) DeleteUserSessions(userID int) error {
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
