-- +goose Up
-- DROP TABLE IF EXISTS bookings, seats, film_sessions, films, cinemas, sessions, users, halls CASCADE;

CREATE TABLE IF NOT EXISTS users (
                                     id SERIAL PRIMARY KEY,
                                     email TEXT NOT NULL UNIQUE,
                                     pass TEXT NOT NULL,
                                     role TEXT NOT NULL DEFAULT 'user'
);


CREATE TABLE IF NOT EXISTS sessions (
                                        id SERIAL PRIMARY KEY,
                                        user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                        token TEXT NOT NULL UNIQUE,
                                        expires_at TIMESTAMP NOT NULL,
                                        created_at TIMESTAMP DEFAULT NOW()
);

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
                                             start_time TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS halls (
                                     id SERIAL PRIMARY KEY,
                                     cinema_id INTEGER NOT NULL REFERENCES cinemas(id) ON DELETE CASCADE,
                                     name VARCHAR(50) NOT NULL,
                                     rows INTEGER NOT NULL,
                                     seats_per_row INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS seats (
                                     id SERIAL PRIMARY KEY,
                                     hall_id INTEGER NOT NULL REFERENCES halls(id) ON DELETE CASCADE,
                                     row INTEGER NOT NULL,
                                     number INTEGER NOT NULL,
                                     reserved BOOLEAN DEFAULT false,
                                     UNIQUE(hall_id, row, number)
);

CREATE TABLE IF NOT EXISTS bookings (
                                        id SERIAL PRIMARY KEY,
                                        user_id INT REFERENCES users(id),
                                        seat_id INT REFERENCES seats(id),
                                        UNIQUE(seat_id)
);

-- +goose Down
-- DROP TABLE IF EXISTS bookings, seats, film_sessions, films, cinemas, sessions, users, halls CASCADE;