package postgres

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"messanger/internal/domain/entity"
	"messanger/internal/domain/repository"
)

type ChatRepository struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) repository.ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) Create(ctx context.Context, chat *entity.Chat) error {
	query := `
        INSERT INTO chats (uuid, name, type, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := r.db.ExecContext(ctx, query, chat.UUID, chat.Name, chat.Type, chat.CreatedAt, chat.UpdatedAt)
	return err
}

func (r *ChatRepository) GetByUUID(ctx context.Context, uuid uuid.UUID) (*entity.Chat, error) {
	query := `SELECT uuid, name, type, created_at, updated_at FROM chats WHERE uuid = $1`
	var chat entity.Chat
	var name sql.NullString
	err := r.db.QueryRowContext(ctx, query, uuid).Scan(
		&chat.UUID, &name, &chat.Type, &chat.CreatedAt, &chat.UpdatedAt,
	)
	if name.Valid {
		chat.Name = &name.String
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &chat, err
}

func (r *ChatRepository) GetUserChats(ctx context.Context, userUUID uuid.UUID) ([]*entity.Chat, error) {
	query := `
        SELECT c.uuid, c.name, c.type, c.created_at, c.updated_at
        FROM chats c
        JOIN chat_participants cp ON c.uuid = cp.chat_uuid
        WHERE cp.user_uuid = $1
        ORDER BY c.updated_at DESC
    `
	rows, err := r.db.QueryContext(ctx, query, userUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*entity.Chat
	for rows.Next() {
		var chat entity.Chat
		var name sql.NullString
		if err := rows.Scan(&chat.UUID, &name, &chat.Type, &chat.CreatedAt, &chat.UpdatedAt); err != nil {
			return nil, err
		}
		if name.Valid {
			chat.Name = &name.String
		}
		chats = append(chats, &chat)
	}
	return chats, nil
}

func (r *ChatRepository) AddParticipant(ctx context.Context, chatUUID, userUUID uuid.UUID) error {
	query := `
        INSERT INTO chat_participants (chat_uuid, user_uuid)
        VALUES ($1, $2)
        ON CONFLICT DO NOTHING
    `
	_, err := r.db.ExecContext(ctx, query, chatUUID, userUUID)
	return err
}

func (r *ChatRepository) GetParticipants(ctx context.Context, chatUUID uuid.UUID) ([]*entity.User, error) {
	query := `
        SELECT u.uuid, u.username, u.created_at, u.updated_at, u.last_seen
        FROM users u
        JOIN chat_participants cp ON u.uuid = cp.user_uuid
        WHERE cp.chat_uuid = $1
    `
	rows, err := r.db.QueryContext(ctx, query, chatUUID)
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

func (r *ChatRepository) GetPersonalChat(ctx context.Context, user1UUID, user2UUID uuid.UUID) (*entity.Chat, error) {
	query := `
        SELECT c.uuid, c.name, c.type, c.created_at, c.updated_at
        FROM chats c
        JOIN chat_participants cp1 ON c.uuid = cp1.chat_uuid
        JOIN chat_participants cp2 ON c.uuid = cp2.chat_uuid
        WHERE c.type = 'personal' 
        AND cp1.user_uuid = $1 
        AND cp2.user_uuid = $2
    `
	var chat entity.Chat
	var name sql.NullString
	err := r.db.QueryRowContext(ctx, query, user1UUID, user2UUID).Scan(
		&chat.UUID, &name, &chat.Type, &chat.CreatedAt, &chat.UpdatedAt,
	)
	if name.Valid {
		chat.Name = &name.String
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &chat, err
}

func (r *ChatRepository) CreatePersonalChat(ctx context.Context, user1UUID, user2UUID uuid.UUID) (*entity.Chat, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	chat := &entity.Chat{
		UUID: uuid.New(),
		Type: "personal",
	}

	query := `INSERT INTO chats (uuid, type, created_at, updated_at) VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	if _, err := tx.ExecContext(ctx, query, chat.UUID, chat.Type); err != nil {
		return nil, err
	}

	if err := r.AddParticipantWithTx(ctx, tx, chat.UUID, user1UUID); err != nil {
		return nil, err
	}

	if err := r.AddParticipantWithTx(ctx, tx, chat.UUID, user2UUID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return chat, nil
}

func (r *ChatRepository) AddParticipantWithTx(ctx context.Context, tx *sql.Tx, chatUUID, userUUID uuid.UUID) error {
	query := `INSERT INTO chat_participants (chat_uuid, user_uuid) VALUES ($1, $2)`
	_, err := tx.ExecContext(ctx, query, chatUUID, userUUID)
	return err
}
