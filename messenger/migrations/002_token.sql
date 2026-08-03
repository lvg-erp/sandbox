-- Таблица API ключей
CREATE TABLE IF NOT EXISTS api_keys (
                                        uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                        key VARCHAR(255) NOT NULL UNIQUE,
                                        name VARCHAR(100) NOT NULL,
                                        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                        expires_at TIMESTAMP WITH TIME ZONE,
                                        is_active BOOLEAN DEFAULT TRUE
);

-- Таблица refresh токенов
CREATE TABLE IF NOT EXISTS refresh_tokens (
                                              uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                              token VARCHAR(255) NOT NULL UNIQUE,
                                              user_uuid UUID NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
                                              created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                              expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
                                              revoked BOOLEAN DEFAULT FALSE
);

-- Индексы для refresh_tokens
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_uuid);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires ON refresh_tokens(expires_at) WHERE revoked = false;

-- Индексы для api_keys
CREATE INDEX IF NOT EXISTS idx_api_keys_key ON api_keys(key);
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(is_active) WHERE is_active = true;

-- Создаем тестовый API ключ
INSERT INTO api_keys (uuid, key, name, is_active)
VALUES (uuid_generate_v4(), 'pfujkjdrf', 'test-key', true)
ON CONFLICT (key) DO NOTHING;

-- Создаем тестового пользователя если его нет
INSERT INTO users (username) VALUES ('admin')
ON CONFLICT (username) DO NOTHING;

-- Комментарий для документирования
COMMENT ON TABLE refresh_tokens IS 'Хранит refresh токены для JWT авторизации';
COMMENT ON TABLE api_keys IS 'Хранит API ключи для доступа к REST API';