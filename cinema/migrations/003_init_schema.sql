-- +goose Up

CREATE TABLE IF NOT EXISTS cinemas (
           id SERIAL PRIMARY KEY,
           name TEXT NOT NULL,
           address TEXT,
           city TEXT,
           phone TEXT,
           total_seats INT,
           poster TEXT,
           created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS halls (
         id SERIAL PRIMARY KEY,
         cinema_id INT NOT NULL REFERENCES cinemas(id) ON DELETE CASCADE,
         name VARCHAR(50) NOT NULL,
         rows INT NOT NULL,
         seats_per_row INT NOT NULL,
         created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS films (
         id SERIAL PRIMARY KEY,
         title TEXT NOT NULL,
         poster TEXT,
         description TEXT,
         duration INT,
         added_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS film_sessions (
         id SERIAL PRIMARY KEY,
         film_id INT REFERENCES films(id),
         cinema_id INT REFERENCES cinemas(id),
         hall_id INT REFERENCES halls(id),
         start_time TIMESTAMP NOT NULL,
         end_date DATE NOT NULL,
         created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS seats (
         id SERIAL PRIMARY KEY,
         hall_id INT NOT NULL REFERENCES halls(id) ON DELETE CASCADE,
         row INT NOT NULL,
         number INT NOT NULL,
         reserved BOOLEAN DEFAULT FALSE,
         UNIQUE(hall_id, row, number)
);

-- +goose Down
DROP TABLE IF EXISTS seats CASCADE;
DROP TABLE IF EXISTS film_sessions CASCADE;
DROP TABLE IF EXISTS films CASCADE;
DROP TABLE IF EXISTS halls CASCADE;
DROP TABLE IF EXISTS cinemas CASCADE;
