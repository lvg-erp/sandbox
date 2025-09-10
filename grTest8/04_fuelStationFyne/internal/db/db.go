package db

import (
	"database/sql"
	"fuelstation/internal/model"
)

func ConnectDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func InsertOperation(db *sql.DB, op model.FuelOperation) error {
	query := `
		INSERT INTO fuel_operations (column_id, fuel_type, liters, action, fill_timestamp, drain_timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := db.Exec(query, op.ColumnID, op.FuelType, op.Liters, op.Action, op.FillTimestamp, op.DrainTimestamp)
	return err
}

func GetFuelStats(db *sql.DB) (float64, float64, float64, float64, error) {
	var petrolFill, dieselFill, petrolDrain, dieselDrain float64

	// Заправка (fill)
	queryFill := `
		SELECT fuel_type, SUM(liters)
		FROM fuel_operations
		WHERE action = 'fill' AND fill_timestamp IS NOT NULL
		GROUP BY fuel_type
	`
	rows, err := db.Query(queryFill)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var fuelType string
		var liters float64
		if err := rows.Scan(&fuelType, &liters); err != nil {
			return 0, 0, 0, 0, err
		}
		if fuelType == "petrol" {
			petrolFill = liters
		} else if fuelType == "diesel" {
			dieselFill = liters
		}
	}

	// Слив (drain)
	queryDrain := `
		SELECT fuel_type, SUM(liters)
		FROM fuel_operations
		WHERE action = 'drain' AND drain_timestamp IS NOT NULL
		GROUP BY fuel_type
	`
	rows, err = db.Query(queryDrain)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var fuelType string
		var liters float64
		if err := rows.Scan(&fuelType, &liters); err != nil {
			return 0, 0, 0, 0, err
		}
		if fuelType == "petrol" {
			petrolDrain = liters
		} else if fuelType == "diesel" {
			dieselDrain = liters
		}
	}

	return petrolFill, dieselFill, petrolDrain, dieselDrain, nil
}
