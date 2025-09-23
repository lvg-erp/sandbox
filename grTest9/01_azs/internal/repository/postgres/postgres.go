package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"fuelazs/internal/logger"
)

type Registry struct {
	FuelGive   *FuelGive
	FuelGet    *FuelGet
	ErrorLogs  *Errors
	Telemetry  *Telemetry
	Activation *Activation
}

// OpenPostgresDB открывает подключение к PostgreSQL
func OpenPostgresDB(logger *logger.Logger, connStr string) (*sql.DB, error) {
	db, err := sql.Open("pgx", connStr) // или "postgres" если используете driver postgres
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("db.Ping: %w", err)
	}

	logger.Info("Соединение с базой данных PostgreSQL успешно установлено.", "connStr", connStr)

	return db, nil
}

// InitializeDatabaseTables создает таблицы, если их нет, и добавляет отсутствующие столбцы
func InitializeDatabaseTables(db *sql.DB, logger *logger.Logger) error {
	createTablesSQL := `
	CREATE TABLE IF NOT EXISTS Activation(
	    last_modification TIMESTAMP,
	    kazs_api_key TEXT NOT NULL,
	    reset_password TEXT NOT NULL,
	    kazs_id TEXT NOT NULL,
	    url TEXT NOT NULL,
	    config_hash TEXT NOT NULL,
	    kazs_number TEXT NOT NULL,
	    kazs_timezone TEXT NOT NULL,
	    ntp_server TEXT NOT NULL,
	    support_number TEXT NOT NULL,
	    logo TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS fuel_give_transactions (
	    fuel_give_id TEXT PRIMARY KEY,
	    kazs_number TEXT NOT NULL,
	    jar_number TEXT NOT NULL,
	    start_time BIGINT NOT NULL,
	    end_time BIGINT,
	    fuel_type TEXT NOT NULL,
	    doc_number TEXT NOT NULL,
	    sensor_before_give TEXT,
	    sensor_after_give TEXT,
	    fuel_liters_plan REAL NOT NULL,
	    fuel_liters REAL,
		send_status BOOLEAN NOT NULL,
		errors TEXT
	);

	CREATE TABLE IF NOT EXISTS fuel_get_transactions (
	    fuel_get_id TEXT PRIMARY KEY,
	    kazs_number TEXT NOT NULL,
	    jar_number TEXT NOT NULL,
	    start_time BIGINT NOT NULL,
	    end_time BIGINT,
	    monitoring_finish_time BIGINT,
	    fuel_type TEXT NOT NULL,
	    doc_number TEXT NOT NULL,
	    sensor_before_give TEXT,
	    sensor_after_give TEXT,
	    fuel_liters_plan REAL NOT NULL,
	    fuel_liters REAL,
	    speed REAL,
	    send_status BOOLEAN NOT NULL,
	    errors TEXT
	);

	CREATE TABLE IF NOT EXISTS errors (
	    timestamp BIGINT NOT NULL,
	    handler TEXT NOT NULL,
	    id TEXT NOT NULL PRIMARY KEY,
	    error TEXT NOT NULL
  );

	CREATE TABLE IF NOT EXISTS telemetry (
	    status_time BIGINT NOT NULL,
	    json TEXT NOT NULL
	);
`

	_, err := db.Exec(createTablesSQL)
	if err != nil {
		return fmt.Errorf("db.Exec create tables: %w", err)
	}

	// === Добавляем отсутствующие столбцы, если их нет, для fuel_get_transactions ===
	err = addColumnIfNotExists(db, "fuel_get_transactions", "monitoring_finish_time", "BIGINT")
	if err != nil {
		return fmt.Errorf("addColumnIfNotExists monitoring_finish_time: %w", err)
	}

	err = addColumnIfNotExists(db, "fuel_get_transactions", "speed", "REAL")
	if err != nil {
		return fmt.Errorf("addColumnIfNotExists speed: %w", err)
	}

	logger.Info("Проверка и инициализация таблиц PostgreSQL завершена успешно.")
	return nil
}

// addColumnIfNotExists проверяет, есть ли столбец в таблице, и добавляет его, если отсутствует
func addColumnIfNotExists(db *sql.DB, tableName, columnName, columnType string) error {
	ctx := context.Background()
	var exists bool

	query := `
	SELECT EXISTS (
	    SELECT 1
	    FROM information_schema.columns
	    WHERE table_name=$1 AND column_name=$2);`

	err := db.QueryRowContext(ctx, query, tableName, columnName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check column existence error: %w", err)
	}

	if !exists {
		alterQuery := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnType)
		_, err := db.ExecContext(ctx, alterQuery)
		if err != nil {
			return fmt.Errorf("add column error: %w", err)
		}
	}

	return nil
}
