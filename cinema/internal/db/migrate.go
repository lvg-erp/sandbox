package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func Migrate(dbURL string) error {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return err
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	err = goose.SetDialect("postgres")
	if err != nil {
		return err
	}
	return goose.Up(db, "migrations")
}

func Connect(dbURL string) *sql.DB {
	db, _ := sql.Open("postgres", dbURL)
	return db
}
