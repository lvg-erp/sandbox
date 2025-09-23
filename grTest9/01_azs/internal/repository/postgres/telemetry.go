package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"fuelazs/internal/integration"
	"sync"
)

type (
	Telemetry struct {
		conn *sql.DB
		mu   sync.Mutex
	}
)

func NewTelemetry(conn *sql.DB) *Telemetry {
	return &Telemetry{
		conn: conn,
		mu:   sync.Mutex{},
	}
}

func (f *Telemetry) InsertTelemetry(t *integration.TelemetryRequest) error {
	query := `
	INSERT INTO telemetry (
		status_time, 
		json
	) VALUES  (?, ?);`

	req, err := json.Marshal(t)
	if err != nil {
		return err
	}

	_, err = f.conn.Exec(query, t.StatusTime, req)

	if err != nil {
		return fmt.Errorf("insert telemetry failed: %v", err)
	}

	return nil
}

func (f *Telemetry) GetAllTelemetry(limit int) ([]integration.TelemetryRequest, error) {
	query := fmt.Sprintf("SELECT json FROM telemetry ORDER BY status_time DESC LIMIT %d;", limit)

	rows, err := f.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("get telemetry failed: %v", err)
	}
	defer rows.Close()

	var reports []integration.TelemetryRequest
	var jsonString string

	for rows.Next() {
		err := rows.Scan(&jsonString)
		if err != nil {
			return nil, fmt.Errorf("get telemetry failed: %v", err)
		}

		var t integration.TelemetryRequest
		err = json.Unmarshal([]byte(jsonString), &t)
		if err != nil {
			return nil, fmt.Errorf("get telemetry failed: %v", err)
		}

		reports = append(reports, t)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("get telemetry failed: %v", err)
	}

	return reports, nil
}

func (f *Telemetry) DeleteOneTelemetry(statusTime int64) error {
	query := `DELETE FROM telemetry WHERE status_time = ?;`

	_, err := f.conn.Exec(query, statusTime)
	if err != nil {
		return fmt.Errorf("delete telemetry failed: %v", err)
	}

	return nil
}

func (f *Telemetry) GetTelemetryRows() (int64, error) {
	query := `SELECT COUNT(*) FROM telemetry;`

	var count int64
	err := f.conn.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get telemetry rows failed: %v", err)
	}

	return count, nil
}
