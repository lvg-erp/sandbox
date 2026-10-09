package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"messanger/internal/domain/entity"
	"messanger/internal/domain/repository"
)

type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) repository.TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	query := `
        INSERT INTO refresh_tokens (uuid, token, user_uuid, created_at, expires_at)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := r.db.ExecContext(ctx, query,
		token.UUID,
		token.Token,
		token.UserUUID,
		token.CreatedAt,
		token.ExpiresAt,
	)
	return err
}

func (r *TokenRepository) GetByToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	query := `
        SELECT uuid, token, user_uuid, created_at, expires_at, revoked
        FROM refresh_tokens
        WHERE token = $1
    `
	var rt entity.RefreshToken
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&rt.UUID, &rt.Token, &rt.UserUUID, &rt.CreatedAt, &rt.ExpiresAt, &rt.Revoked,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *TokenRepository) Revoke(ctx context.Context, token string) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE token = $1 AND revoked = false`
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

func (r *TokenRepository) RevokeAllForUser(ctx context.Context, userUUID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE user_uuid = $1 AND revoked = false`
	_, err := r.db.ExecContext(ctx, query, userUUID)
	return err
}
