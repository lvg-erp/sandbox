package postgres

import (
	"database/sql"
	"fuelazs/internal/usecase/models"
)

type (
	Errors struct {
		conn *sql.DB
	}
)

func NewErrors(conn *sql.DB) *Errors {
	return &Errors{
		conn: conn,
	}
}

func (errs *Errors) InsertError(msg models.Errors) error {
	query := `INSERT INTO errors (
		timestamp, 
		handler, 
		id, 
		error
 	) VALUES (?, ?, ?, ?);`

	_, err := errs.conn.Exec(query, msg.Time, msg.Handler, msg.Id, msg.Error)

	if err != nil {
		return err
	}

	return nil
}
