package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"fuelazs/internal/usecase/models"
	"strings"
)

type (
	FuelGive struct {
		conn *sql.DB
	}
)

func NewFuelGive(conn *sql.DB) *FuelGive {
	return &FuelGive{
		conn: conn,
	}
}

func (f *FuelGive) InsertFuelGiveTransaction(params models.InsertFuelGiveTransactionDep) error {
	if params.FuelGiveID == "" ||
		params.JarNumber == "" ||
		params.FuelType == "" ||
		params.DocNumber == "" ||
		params.KazsNumber == "" {
		return fmt.Errorf("required fields (FuelGiveID, JarNumber, FuelType, DocNumber) cannot be empty for insert")
	}

	columns := []string{
		"fuel_give_id",
		"kazs_number",
		"jar_number",
		"start_time",
		"fuel_type",
		"doc_number",
		"send_status",
		"fuel_liters_plan",
	}
	args := []interface{}{
		params.FuelGiveID,
		params.KazsNumber,
		params.JarNumber,
		params.StartTime,
		params.FuelType,
		params.DocNumber,
		params.SendStatus,
		params.FuelLitersPlan,
	}

	if params.EndTime != nil {
		columns = append(columns, "end_time")
		args = append(args, *params.EndTime)
	}

	if params.SensorBeforeGive != nil {
		columns = append(columns, "sensor_before_give")
		sensorBeforeGive, err := json.Marshal(params.SensorBeforeGive)
		if err != nil {
			sensorBeforeGive = []byte{}
		}
		args = append(args, sensorBeforeGive)
	}

	if params.SensorAfterGive != nil {
		columns = append(columns, "sensor_after_give")
		sensorAfterGive, err := json.Marshal(params.SensorAfterGive)
		if err != nil {
			sensorAfterGive = []byte{}
		}
		args = append(args, sensorAfterGive)
	}

	if params.FuelLiters != nil {
		columns = append(columns, "fuel_liters")
		args = append(args, *params.FuelLiters)
	}

	if params.Errors != nil {
		columns = append(columns, "errors")
		args = append(args, *params.Errors)
	}

	// Формируем плейсхолдеры $1, $2, ... для PostgreSQL
	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(
		"INSERT INTO fuel_give_transactions (%s) VALUES (%s)",
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := f.conn.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to insert new transaction with FuelGiveID %s: %w", params.FuelGiveID, err)
	}
	return nil
}

func (f *FuelGive) UpdateFuelGiveTransaction(params models.UpdateFuelGiveTransactionDep) error {
	if params.FuelGiveID == "" {
		return fmt.Errorf("fuelGiveID cannot be empty")
	}

	updates := []string{}
	args := []interface{}{}

	// Счётчик для индексов плейсхолдеров
	// индекс плейсхолдера будет len(args)+1, потому что args всё время растёт

	if params.KazsNumber != nil {
		updates = append(updates, fmt.Sprintf("kazs_number = $%d", len(args)+1))
		args = append(args, *params.KazsNumber)
	}
	if params.JarNumber != nil {
		updates = append(updates, fmt.Sprintf("jar_number = $%d", len(args)+1))
		args = append(args, *params.JarNumber)
	}
	if params.StartTime != nil {
		updates = append(updates, fmt.Sprintf("start_time = $%d", len(args)+1))
		args = append(args, *params.StartTime)
	}
	if params.EndTime != nil {
		updates = append(updates, fmt.Sprintf("end_time = $%d", len(args)+1))
		args = append(args, *params.EndTime)
	}
	if params.FuelType != nil {
		updates = append(updates, fmt.Sprintf("fuel_type = $%d", len(args)+1))
		args = append(args, *params.FuelType)
	}
	if params.DocNumber != nil {
		updates = append(updates, fmt.Sprintf("doc_number = $%d", len(args)+1))
		args = append(args, *params.DocNumber)
	}
	if params.SensorBeforeGive != nil {
		updates = append(updates, fmt.Sprintf("sensor_before_give = $%d", len(args)+1))
		sensorBeforeGive, err := json.Marshal(params.SensorBeforeGive)
		if err != nil {
			sensorBeforeGive = []byte{}
		}
		args = append(args, sensorBeforeGive)
	}
	if params.SensorAfterGive != nil {
		updates = append(updates, fmt.Sprintf("sensor_after_give = $%d", len(args)+1))
		sensorAfterGive, err := json.Marshal(params.SensorAfterGive)
		if err != nil {
			sensorAfterGive = []byte{}
		}
		args = append(args, sensorAfterGive)
	}
	if params.FuelLiters != nil {
		updates = append(updates, fmt.Sprintf("fuel_liters = $%d", len(args)+1))
		args = append(args, *params.FuelLiters)
	}
	if params.FuelLitersPlan != nil {
		updates = append(updates, fmt.Sprintf("fuel_liters_plan = $%d", len(args)+1))
		args = append(args, *params.FuelLitersPlan)
	}
	if params.SendStatus != nil {
		updates = append(updates, fmt.Sprintf("send_status = $%d", len(args)+1))
		args = append(args, *params.SendStatus)
	}
	if params.Errors != nil {
		updates = append(updates, fmt.Sprintf("errors = $%d", len(args)+1))
		args = append(args, *params.Errors)
	}

	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Ключ в WHERE ставим тоже с плейсхолдером
	query := fmt.Sprintf("UPDATE fuel_give_transactions SET %s WHERE fuel_give_id = $%d",
		strings.Join(updates, ", "),
		len(args)+1,
	)

	args = append(args, params.FuelGiveID)

	_, err := f.conn.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update transaction for FuelGiveID %s: %w", params.FuelGiveID, err)
	}
	return nil
}

// getTransactionByFuelGiveID получает транзакцию по fuelGiveID
func (f *FuelGive) GetFuelGiveTransactionByID(fuelGiveID string) (models.FuelGiveReceipt, error) {
	query := `
	SELECT
		fuel_give_id,
		kazs_number,
		jar_number,
		start_time,
		end_time,
		fuel_type,
		doc_number,
		sensor_before_give,
		sensor_after_give,
		fuel_liters_plan,
		fuel_liters,
		send_status,
        errors
	FROM fuel_give_transactions
	WHERE fuel_give_id = $1
	`
	row := f.conn.QueryRow(query, fuelGiveID)

	var (
		fuelGiveIDStr, kazsNumber, jarNumber, fuelType, docNumber, errorsStr sql.NullString
		startTime, endTime                                                   sql.NullInt64
		sensorBeforeGiveStr, sensorAfterGiveStr                              sql.NullString
		fuelLitersPlan                                                       sql.NullFloat64
		fuelLiters                                                           sql.NullFloat64
		sendStatus                                                           bool
	)

	err := row.Scan(
		&fuelGiveIDStr,
		&kazsNumber,
		&jarNumber,
		&startTime,
		&endTime,
		&fuelType,
		&docNumber,
		&sensorBeforeGiveStr,
		&sensorAfterGiveStr,
		&fuelLitersPlan,
		&fuelLiters,
		&sendStatus,
		&errorsStr,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.FuelGiveReceipt{}, fmt.Errorf("no transaction found for fuelGiveID: %s", fuelGiveID)
		}
		return models.FuelGiveReceipt{}, fmt.Errorf("failed to get transaction for fuelGiveID %s: %w", fuelGiveID, err)
	}

	receipt := models.FuelGiveReceipt{
		KazsNumber:       kazsNumber.String,
		JarId:            jarNumber.String,
		FuelType:         fuelType.String,
		StartTime:        startTime.Int64,
		DocNumber:        docNumber.String,
		FuelLitersPlan:   fuelLitersPlan.Float64,
		SendStatus:       sendStatus,
		Errors:           "",
		FuelLiter:        0,
		EndTime:          0,
		SensorBeforeGive: models.SensorInfoGive{},
		SensorAfterGive:  models.SensorInfoGive{},
	}

	if endTime.Valid {
		receipt.EndTime = endTime.Int64
	} else {
		receipt.EndTime = 0
	}

	if fuelLiters.Valid {
		receipt.FuelLiter = fuelLiters.Float64
	} else {
		receipt.FuelLiter = 0
	}

	if errorsStr.Valid {
		receipt.Errors = errorsStr.String
	}

	// Обрабатываем sensorBeforeGive
	if sensorBeforeGiveStr.Valid && sensorBeforeGiveStr.String != "" {
		if err := json.Unmarshal([]byte(sensorBeforeGiveStr.String), &receipt.SensorBeforeGive); err != nil {
			// можно логировать ошибку
			receipt.SensorBeforeGive = models.SensorInfoGive{}
		}
	}

	// Обрабатываем sensorAfterGive
	if sensorAfterGiveStr.Valid && sensorAfterGiveStr.String != "" {
		if err := json.Unmarshal([]byte(sensorAfterGiveStr.String), &receipt.SensorAfterGive); err != nil {
			// логирование
			receipt.SensorAfterGive = models.SensorInfoGive{}
		}
	}

	return receipt, nil
}

// getLastFuelGiveTransaction по jarNumber
func (f *FuelGive) GetLastFuelGiveTransaction(jarNumber string) (models.LastFuelGiveReceipt, error) {
	if jarNumber == "" {
		return models.LastFuelGiveReceipt{}, fmt.Errorf("jarNumber cannot be empty")
	}

	query := `
	SELECT
		fuel_give_id,
		kazs_number,
		jar_number,
		start_time,
		end_time,
		fuel_type,
		doc_number,
		sensor_before_give,
		sensor_after_give,
		fuel_liters_plan,
		fuel_liters,
		send_status,
        errors
	FROM fuel_give_transactions
	WHERE jar_number = $1 AND end_time = 0
	ORDER BY start_time DESC
	LIMIT 1
	`
	var (
		fuelGiveID, kazsNumber, jarId, fuelType, docNumber, errorsStr sql.NullString
		startTime, endTime                                            sql.NullInt64
		sensorBeforeGiveStr, sensorAfterGiveStr                       sql.NullString
		fuelLitersPlan                                                sql.NullFloat64
		fuelLiters                                                    sql.NullFloat64
		sendStatus                                                    bool
	)

	row := f.conn.QueryRow(query, jarNumber)
	err := row.Scan(
		&fuelGiveID,
		&kazsNumber,
		&jarId,
		&startTime,
		&endTime,
		&fuelType,
		&docNumber,
		&sensorBeforeGiveStr,
		&sensorAfterGiveStr,
		&fuelLitersPlan,
		&fuelLiters,
		&sendStatus,
		&errorsStr,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.LastFuelGiveReceipt{}, sql.ErrNoRows
		}
		return models.LastFuelGiveReceipt{}, fmt.Errorf("failed to get transaction for jar number %s: %w", jarNumber, err)
	}

	receipt := models.LastFuelGiveReceipt{
		FuelGiveID:     fuelGiveID.String,
		KazsNumber:     kazsNumber.String,
		JarId:          jarId.String,
		FuelType:       fuelType.String,
		StartTime:      startTime.Int64,
		FuelLitersPlan: fuelLitersPlan.Float64,
	}

	if endTime.Valid {
		receipt.EndTime = endTime.Int64
	}

	if fuelLiters.Valid {
		receipt.FuelLiter = fuelLiters.Float64
	}

	if errorsStr.Valid {
		receipt.Errors = errorsStr.String
	}

	if sensorBeforeGiveStr.Valid && sensorBeforeGiveStr.String != "" {
		if err := json.Unmarshal([]byte(sensorBeforeGiveStr.String), &receipt.SensorBeforeGive); err != nil {
			// можно логировать
		}
	}

	if sensorAfterGiveStr.Valid && sensorAfterGiveStr.String != "" {
		if err := json.Unmarshal([]byte(sensorAfterGiveStr.String), &receipt.SensorAfterGive); err != nil {
			// можно логировать
		}
	}

	return receipt, nil
}

// getUnsentFuelGiveTransactions - получает все транзакции, которых еще не отправляли (send_status = false)
func (f *FuelGive) GetUnsentFuelGiveTransactions() ([]models.LastFuelGiveReceipt, error) {
	query := `
	SELECT
		fuel_give_id,
		kazs_number,
		jar_number,
		start_time,
		end_time,
		fuel_type,
		doc_number,
		sensor_before_give,
		sensor_after_give,
		fuel_liters_plan,
		fuel_liters,
		send_status,
        errors
	FROM fuel_give_transactions
	WHERE send_status = false AND end_time <> 0
	ORDER BY start_time DESC
	`

	rows, err := f.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query for unsent transactions: %w", err)
	}
	defer rows.Close()

	var receipts []models.LastFuelGiveReceipt

	for rows.Next() {
		var (
			fuelGiveID, kazsNumber, jarId, fuelType, docNumber, errorsStr sql.NullString
			startTime, endTime                                            sql.NullInt64
			sensorBeforeGiveStr, sensorAfterGiveStr                       sql.NullString
			fuelLitersPlan                                                sql.NullFloat64
			fuelLiters                                                    sql.NullFloat64
			sendStatus                                                    bool
		)

		if err := rows.Scan(
			&fuelGiveID,
			&kazsNumber,
			&jarId,
			&startTime,
			&endTime,
			&fuelType,
			&docNumber,
			&sensorBeforeGiveStr,
			&sensorAfterGiveStr,
			&fuelLitersPlan,
			&fuelLiters,
			&sendStatus,
			&errorsStr,
		); err != nil {
			return receipts, fmt.Errorf("failed to scan transaction row: %w", err)
		}

		receipt := models.LastFuelGiveReceipt{
			FuelGiveID:     fuelGiveID.String,
			KazsNumber:     kazsNumber.String,
			JarId:          jarId.String,
			FuelType:       fuelType.String,
			StartTime:      startTime.Int64,
			FuelLitersPlan: fuelLitersPlan.Float64,
		}

		if endTime.Valid {
			receipt.EndTime = endTime.Int64
		}
		if fuelLiters.Valid {
			receipt.FuelLiter = fuelLiters.Float64
		}
		if errorsStr.Valid {
			receipt.Errors = errorsStr.String
		}

		// Обрабатываем sensorBeforeGive
		if sensorBeforeGiveStr.Valid && sensorBeforeGiveStr.String != "" {
			if err := json.Unmarshal([]byte(sensorBeforeGiveStr.String), &receipt.SensorBeforeGive); err != nil {
				// логировать ошибку
			}
		}

		// Обрабатываем sensorAfterGive
		if sensorAfterGiveStr.Valid && sensorAfterGiveStr.String != "" {
			if err := json.Unmarshal([]byte(sensorAfterGiveStr.String), &receipt.SensorAfterGive); err != nil {
				// логировать ошибку
			}
		}

		receipts = append(receipts, receipt)
	}

	if err = rows.Err(); err != nil {
		return receipts, fmt.Errorf("error during rows iteration: %w", err)
	}

	return receipts, nil
}
