package application

import (
	"fmt"
	"fuelazs/internal/logger"
	"fuelazs/internal/repository/postgres"
)

func NewRegistry(dbPath string, logger *logger.Logger) (*postgres.Registry, error) {
	postgresConn, err := postgres.OpenPostgresDB(logger, dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	err = postgres.InitializeDatabaseTables(postgresConn, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database tables: %w", err)
	}

	fuelGive := postgres.NewFuelGive(postgresConn)

	fuelGet := postgres.NewFuelGet(postgresConn)

	errorLogs := postgres.NewErrors(postgresConn)

	telemetry := postgres.NewTelemetry(postgresConn)

	activation := postgres.NewActivation(postgresConn)

	return &postgres.Registry{
		FuelGive:   fuelGive,
		FuelGet:    fuelGet,
		ErrorLogs:  errorLogs,
		Telemetry:  telemetry,
		Activation: activation,
	}, nil
}
