-- Создание расширения для UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
                                     uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                     username VARCHAR(100) NOT NULL UNIQUE,
                                     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                     updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                     last_seen TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Таблица чатов
CREATE TABLE IF NOT EXISTS chats (
                                     uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                     name VARCHAR(255),
                                     type VARCHAR(20) NOT NULL, -- 'personal', 'group'
                                     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                     updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Таблица участников чатов
CREATE TABLE IF NOT EXISTS chat_participants (
                                                 chat_uuid UUID REFERENCES chats(uuid) ON DELETE CASCADE,
                                                 user_uuid UUID REFERENCES users(uuid) ON DELETE CASCADE,
                                                 joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                                 PRIMARY KEY (chat_uuid, user_uuid)
);

-- Таблица сообщений
CREATE TABLE IF NOT EXISTS messages (
                                        uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                        chat_uuid UUID REFERENCES chats(uuid) ON DELETE CASCADE,
                                        sender_uuid UUID REFERENCES users(uuid) ON DELETE CASCADE,
                                        body TEXT NOT NULL,
                                        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                        deleted BOOLEAN DEFAULT FALSE
);

-- Индексы для производительности
CREATE INDEX idx_messages_chat_created ON messages(chat_uuid, created_at);
CREATE INDEX idx_chat_participants_user ON chat_participants(user_uuid);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_last_seen ON users(last_seen);

-- Триггер для обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_chats_updated_at BEFORE UPDATE ON chats
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_messages_updated_at BEFORE UPDATE ON messages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();