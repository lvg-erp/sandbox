-- +goose Up
ALTER TABLE film_sessions ADD COLUMN hall_id INT REFERENCES halls(id);
ALTER TABLE film_sessions ADD COLUMN end_date DATE;

-- +goose Down
ALTER TABLE film_sessions DROP COLUMN hall_id;
ALTER TABLE film_sessions DROP COLUMN end_date;