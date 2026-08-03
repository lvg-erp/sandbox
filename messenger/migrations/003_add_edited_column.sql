-- Миграция 003: Добавление колонки edited в таблицу messages
ALTER TABLE messages ADD COLUMN IF NOT EXISTS edited BOOLEAN DEFAULT FALSE;

-- Добавляем комментарий для документации
COMMENT ON COLUMN messages.edited IS 'Флаг, указывающий, было ли сообщение отредактировано';