-- +goose Up

ALTER TABLE film_sessions
    RENAME COLUMN start_time TO start_time_full;

ALTER TABLE film_sessions
    ADD COLUMN date DATE NOT NULL DEFAULT CURRENT_DATE,
    ADD COLUMN start_time TIME NOT NULL DEFAULT '00:00:00',
    ADD COLUMN end_time TIME NOT NULL DEFAULT '00:00:00';

-- Миграция существующих данных (если нужно)
UPDATE film_sessions
SET date = start_time_full::DATE,
    start_time = start_time_full::TIME,
    end_time = (start_time_full + INTERVAL '2 hours')::TIME;  -- пример +2ч

-- +goose Down

ALTER TABLE film_sessions
    DROP COLUMN date,
    DROP COLUMN start_time,
    DROP COLUMN end_time;

ALTER TABLE film_sessions
    RENAME COLUMN start_time_full TO start_time;