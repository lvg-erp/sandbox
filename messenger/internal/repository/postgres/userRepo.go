package postgres

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"messanger/internal/domain/entity"
	"messanger/internal/domain/repository"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
        INSERT INTO users (uuid, username, created_at, updated_at, last_seen)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := r.db.ExecContext(ctx, query, user.UUID, user.Username, user.CreatedAt, user.UpdatedAt, user.LastSeen)
	return err
}

func (r *UserRepository) GetByUUID(ctx context.Context, uuid uuid.UUID) (*entity.User, error) {
	query := `SELECT uuid, username, created_at, updated_at, last_seen FROM users WHERE uuid = $1`
	var user entity.User
	err := r.db.QueryRowContext(ctx, query, uuid).Scan(
		&user.UUID, &user.Username, &user.CreatedAt, &user.UpdatedAt, &user.LastSeen,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	query := `SELECT uuid, username, created_at, updated_at, last_seen FROM users WHERE username = $1`
	var user entity.User
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.UUID, &user.Username, &user.CreatedAt, &user.UpdatedAt, &user.LastSeen,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `UPDATE users SET username = $1, updated_at = $2 WHERE uuid = $3`
	_, err := r.db.ExecContext(ctx, query, user.Username, user.UpdatedAt, user.UUID)
	return err
}

func (r *UserRepository) UpdateLastSeen(ctx context.Context, uuid uuid.UUID) error {
	query := `UPDATE users SET last_seen = CURRENT_TIMESTAMP WHERE uuid = $1`
	_, err := r.db.ExecContext(ctx, query, uuid)
	return err
}

func (r *UserRepository) GetAll(ctx context.Context) ([]*entity.User, error) {
	query := `SELECT uuid, username, created_at, updated_at, last_seen FROM users ORDER BY username`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.UUID, &user.Username, &user.CreatedAt, &user.UpdatedAt, &user.LastSeen); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}
