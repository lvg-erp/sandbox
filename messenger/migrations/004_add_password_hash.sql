-- Миграция 004: пароли пользователей (bcrypt-хэши)

ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255) NOT NULL DEFAULT '';

-- Легаси-пользователям без пароля назначаем дев-пароль "changeme123".
-- Только для локальной разработки: для production удалите этот UPDATE
-- и заводите пользователей через /api/auth/register.
UPDATE users SET password_hash = '$2a$10$sn7ig.kCSu2WSdcqJqWime9RA5d9RIwP4hhUTM783ON0MAwuByYmu'
WHERE password_hash = '';

COMMENT ON COLUMN users.password_hash IS 'bcrypt-хэш пароля; пустая строка — пароль не задан, вход запрещён';
