-- Миграция 005: refresh-токены (идемпотентно дублирует схему из 002,
-- чтобы её можно было применить и к БД, где 002 не выполнялась)

CREATE TABLE IF NOT EXISTS refresh_tokens (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    token VARCHAR(255) NOT NULL UNIQUE,
    user_uuid UUID NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked BOOLEAN DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_uuid);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires ON refresh_tokens(expires_at) WHERE revoked = false;

-- Чистим просроченные и отозванные токены
DELETE FROM refresh_tokens WHERE expires_at < now() OR revoked = true;

COMMENT ON TABLE refresh_tokens IS 'Хранит SHA-256 хэши refresh токенов для ротации и отзыва';
