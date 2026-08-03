package postgres

import (
	"context"
	"database/sql"
	"messanger/internal/domain/entity"
	"messanger/internal/domain/repository"

	"github.com/google/uuid"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) repository.MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(ctx context.Context, message *entity.Message) error {
	query := `
        INSERT INTO messages (uuid, chat_uuid, sender_uuid, body, created_at, updated_at, deleted, edited)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
	_, err := r.db.ExecContext(ctx, query,
		message.UUID,
		message.ChatUUID,
		message.SenderUUID,
		message.Body,
		message.CreatedAt, // Убедитесь, что это не nil
		message.UpdatedAt,
		message.Deleted,
		message.Edited,
	)
	return err
}

func (r *MessageRepository) GetByUUID(ctx context.Context, uuid uuid.UUID) (*entity.Message, error) {
	query := `SELECT uuid, chat_uuid, sender_uuid, body, created_at, updated_at, deleted FROM messages WHERE uuid = $1`
	var message entity.Message
	err := r.db.QueryRowContext(ctx, query, uuid).Scan(
		&message.UUID, &message.ChatUUID, &message.SenderUUID, &message.Body, &message.CreatedAt, &message.UpdatedAt, &message.Deleted,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &message, err
}

func (r *MessageRepository) GetChatMessages(ctx context.Context, chatUUID uuid.UUID, limit, offset int) ([]*entity.Message, error) {
	query := `
        SELECT uuid, chat_uuid, sender_uuid, body, created_at, updated_at, deleted
        FROM messages
        WHERE chat_uuid = $1 AND deleted = false
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `
	rows, err := r.db.QueryContext(ctx, query, chatUUID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*entity.Message
	for rows.Next() {
		var message entity.Message
		if err := rows.Scan(&message.UUID, &message.ChatUUID, &message.SenderUUID, &message.Body, &message.CreatedAt, &message.UpdatedAt, &message.Deleted); err != nil {
			return nil, err
		}
		messages = append(messages, &message)
	}
	return messages, nil
}

func (r *MessageRepository) Delete(ctx context.Context, uuid uuid.UUID) error {
	query := `UPDATE messages SET deleted = true, updated_at = CURRENT_TIMESTAMP WHERE uuid = $1`
	_, err := r.db.ExecContext(ctx, query, uuid)
	return err
}

func (r *MessageRepository) GetLastMessages(ctx context.Context, chatUUID uuid.UUID, limit int) ([]*entity.Message, error) {
	query := `
        SELECT uuid, chat_uuid, sender_uuid, body, created_at, updated_at, deleted
        FROM messages
        WHERE chat_uuid = $1 AND deleted = false
        ORDER BY created_at DESC
        LIMIT $2
    `
	rows, err := r.db.QueryContext(ctx, query, chatUUID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*entity.Message
	for rows.Next() {
		var message entity.Message
		if err := rows.Scan(&message.UUID, &message.ChatUUID, &message.SenderUUID, &message.Body, &message.CreatedAt, &message.UpdatedAt, &message.Deleted); err != nil {
			return nil, err
		}
		messages = append(messages, &message)
	}
	return messages, nil
}
