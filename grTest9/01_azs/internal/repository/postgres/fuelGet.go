package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"fuelazs/internal/usecase/models"
	"strconv"
	"strings"
	"time"
)

type (
	FuelGet struct {
		conn *sql.DB
	}
)

func NewFuelGet(conn *sql.DB) *FuelGet {
	return &FuelGet{
		conn: conn,
	}
}

func (f *FuelGet) InsertFuelGetTransaction(params models.InsertFuelGetTransactionDep) error {
	// Проверка обязательных полей
	if params.FuelGetID == "" ||
		params.JarNumber == "" ||
		params.FuelType == "" ||
		params.DocNumber == "" ||
		params.KazsNumber == "" {
		return fmt.Errorf("required fields (FuelGetID, JarNumber, FuelType, DocNumber) cannot be empty for insert")
	}

	columns := []string{
		"fuel_get_id",
		"kazs_number",
		"jar_number",
		"start_time",
		"fuel_type",
		"doc_number",
		"send_status",
		"fuel_liters_plan",
	}
	placeholders := []string{
		"$1", "$2", "$3", "$4", "$5", "$6", "$7", "$8",
	}
	args := []interface{}{
		params.FuelGetID,
		params.KazsNumber,
		params.JarNumber,
		params.StartTime,
		params.FuelType,
		params.DocNumber,
		params.SendStatus,
		params.FuelLiterPlan,
	}

	argIndex := 9 // следующий индекс для параметров, добавляемых позднее

	if params.EndTime != nil {
		columns = append(columns, "end_time")
		placeholders = append(placeholders, "$"+strconv.Itoa(argIndex))
		args = append(args, *params.EndTime)
		argIndex++
	}
	if params.MonitoringFinishTime != nil {
		columns = append(columns, "monitoring_finish_time")
		placeholders = append(placeholders, "$"+strconv.Itoa(argIndex))
		args = append(args, *params.MonitoringFinishTime)
		argIndex++
	}
	if params.SensorBeforeGive != nil {
		columns = append(columns, "sensor_before_give")
		placeholders = append(placeholders, "$"+strconv.Itoa(argIndex))
		sensorBeforeGive, err := json.Marshal(params.SensorBeforeGive)
		if err != nil {
			sensorBeforeGive = []byte{}
		}
		args = append(args, sensorBeforeGive)
		argIndex++
	}
	if params.SensorAfterGive != nil {
		columns = append(columns, "sensor_after_give")
		placeholders = append(placeholders, "$"+strconv.Itoa(argIndex))
		sensorAfterGive, err := json.Marshal(params.SensorAfterGive)
		if err != nil {
			sensorAfterGive = []byte{}
		}
		args = append(args, sensorAfterGive)
		argIndex++
	}
	if params.FuelLiters != nil {
		columns = append(columns, "fuel_liters")
		placeholders = append(placeholders, "$"+strconv.Itoa(argIndex))
		args = append(args, *params.FuelLiters)
		argIndex++
	}
	if params.Errors != nil {
		columns = append(columns, "errors")
		placeholders = append(placeholders, "$"+strconv.Itoa(argIndex))
		args = append(args, *params.Errors)
		argIndex++
	}
	if params.Speed != nil {
		columns = append(columns, "speed")
		placeholders = append(placeholders, "$"+strconv.Itoa(argIndex))
		args = append(args, *params.Speed)
		argIndex++
	}

	query := fmt.Sprintf(
		"INSERT INTO fuel_get_transactions (%s) VALUES (%s)",
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := f.conn.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to insert new transaction with FuelGetID %s: %w", params.FuelGetID, err)
	}

	return nil
}

func (f *FuelGet) UpdateFuelGetTransaction(params models.UpdateFuelGetTransactionDep) error {
	if params.FuelGetID == "" {
		return fmt.Errorf("FuelGetID cannot be empty")
	}

	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if params.KazsNumber != nil {
		updates = append(updates, "kazs_number = $"+strconv.Itoa(argIndex))
		args = append(args, *params.KazsNumber)
		argIndex++
	}
	if params.JarNumber != nil {
		updates = append(updates, "jar_number = $"+strconv.Itoa(argIndex))
		args = append(args, *params.JarNumber)
		argIndex++
	}
	if params.StartTime != nil {
		updates = append(updates, "start_time = $"+strconv.Itoa(argIndex))
		args = append(args, *params.StartTime)
		argIndex++
	}
	if params.EndTime != nil {
		updates = append(updates, "end_time = $"+strconv.Itoa(argIndex))
		args = append(args, *params.EndTime)
		argIndex++
	}
	if params.MonitoringFinishTime != nil {
		updates = append(updates, "monitoring_finish_time = $"+strconv.Itoa(argIndex))
		args = append(args, *params.MonitoringFinishTime)
		argIndex++
	}
	if params.FuelType != nil {
		updates = append(updates, "fuel_type = $"+strconv.Itoa(argIndex))
		args = append(args, *params.FuelType)
		argIndex++
	}
	if params.DocNumber != nil {
		updates = append(updates, "doc_number = $"+strconv.Itoa(argIndex))
		args = append(args, *params.DocNumber)
		argIndex++
	}
	if params.SensorBeforeGive != nil {
		updates = append(updates, "sensor_before_give = $"+strconv.Itoa(argIndex))
		sensorBeforeGive, err := json.Marshal(params.SensorBeforeGive)
		if err != nil {
			sensorBeforeGive = []byte{}
		}
		args = append(args, sensorBeforeGive)
		argIndex++
	}
	if params.SensorAfterGive != nil {
		updates = append(updates, "sensor_after_give = $"+strconv.Itoa(argIndex))
		sensorAfterGive, err := json.Marshal(params.SensorAfterGive)
		if err != nil {
			sensorAfterGive = []byte{}
		}
		args = append(args, sensorAfterGive)
		argIndex++
	}
	if params.FuelLiters != nil {
		updates = append(updates, "fuel_liters = $"+strconv.Itoa(argIndex))
		args = append(args, *params.FuelLiters)
		argIndex++
	}
	if params.FuelLitersPlans != nil {
		updates = append(updates, "fuel_liters_plan = $"+strconv.Itoa(argIndex))
		args = append(args, *params.FuelLitersPlans)
		argIndex++
	}
	if params.Speed != nil {
		updates = append(updates, "speed = $"+strconv.Itoa(argIndex))
		args = append(args, *params.Speed)
		argIndex++
	}
	if params.SendStatus != nil {
		updates = append(updates, "send_status = $"+strconv.Itoa(argIndex))
		args = append(args, *params.SendStatus)
		argIndex++
	}
	if params.Errors != nil {
		updates = append(updates, "errors = $"+strconv.Itoa(argIndex))
		args = append(args, *params.Errors)
		argIndex++
	}

	if len(updates) == 0 {
		return fmt.Errorf("no fields to update for FuelGetID %s", params.FuelGetID)
	}

	query := fmt.Sprintf(
		"UPDATE fuel_get_transactions SET %s WHERE fuel_get_id = $%d",
		strings.Join(updates, ", "),
		argIndex,
	)

	args = append(args, params.FuelGetID)

	_, err := f.conn.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update transaction for FuelGetID %s: %w", params.FuelGetID, err)
	}

	return nil
}

func (f *FuelGet) GetFuelGetTransaction(fuelGetID string) (models.FuelGetReceipt, error) {
	var receipt models.FuelGetReceipt

	if fuelGetID == "" {
		return receipt, fmt.Errorf("fuelGetID cannot be empty")
	}

	var (
		kazsNumber           string
		jarNumber            string
		startTime            int64
		endTime              sql.NullInt64
		monitoringFinishTime sql.NullInt64
		fuelType             string
		docNumber            string
		sensorBeforeGive     sql.NullString
		sensorAfterGive      sql.NullString
		liters               sql.NullFloat64
		fuelLitersPlan       float64
		speed                float64
		errorsStr            sql.NullString
		sendStatus           bool
	)

	query := `
		SELECT
			jar_number,
			kazs_number,
			start_time,
			end_time,
			monitoring_finish_time,
			fuel_type,
			doc_number,
			sensor_before_give,
			sensor_after_give,
			fuel_liters,
			fuel_liters_plan,
			speed,
			errors,
			send_status
		FROM
			fuel_get_transactions
		WHERE
			fuel_get_id = $1
	`

	row := f.conn.QueryRow(query, fuelGetID)
	err := row.Scan(
		&jarNumber,
		&kazsNumber,
		&startTime,
		&endTime,
		&monitoringFinishTime,
		&fuelType,
		&docNumber,
		&sensorBeforeGive,
		&sensorAfterGive,
		&liters,
		&fuelLitersPlan,
		&speed,
		&errorsStr,
		&sendStatus,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return receipt, fmt.Errorf("no transaction found for fuelGetID: %s", fuelGetID)
		}
		return receipt, fmt.Errorf("failed to get transaction for fuelGetID %s: %w", fuelGetID, err)
	}

	receipt.KazsNumber = kazsNumber
	receipt.JarId = jarNumber
	receipt.FuelType = fuelType
	receipt.StartTime = startTime
	receipt.DocNumber = docNumber
	receipt.FuelLitersPlan = fuelLitersPlan
	receipt.Speed = speed

	if endTime.Valid {
		receipt.EndTime = endTime.Int64
	} else {
		receipt.EndTime = time.Now().Unix()
	}

	if monitoringFinishTime.Valid {
		receipt.MonitoringFinishTime = monitoringFinishTime.Int64
	} else {
		receipt.MonitoringFinishTime = 0
	}

	if liters.Valid {
		receipt.FuelLiter = liters.Float64
	} else {
		receipt.FuelLiter = 0.0
	}

	if errorsStr.Valid {
		receipt.Errors = errorsStr.String
	} else {
		receipt.Errors = ""
	}

	// Обработка sensorBeforeGive
	if sensorBeforeGive.Valid && sensorBeforeGive.String != "" {
		err = json.Unmarshal([]byte(sensorBeforeGive.String), &receipt.SensorBeforeGet)
		if err != nil {
			receipt.SensorBeforeGet = models.SensorInfo{}
		}
	} else {
		receipt.SensorBeforeGet = models.SensorInfo{}
	}

	// Обработка sensorAfterGive
	if sensorAfterGive.Valid && sensorAfterGive.String != "" {
		err = json.Unmarshal([]byte(sensorAfterGive.String), &receipt.SensorAfterGet)
		if err != nil {
			receipt.SensorAfterGet = models.SensorInfo{}
		}
	} else {
		receipt.SensorAfterGet = models.SensorInfo{}
	}

	return receipt, nil
}

func (f *FuelGet) GetLastFuelGetTransaction(jarNumber string) (models.LastFuelGetReceipt, error) {
	if jarNumber == "" {
		return models.LastFuelGetReceipt{}, fmt.Errorf("jarNumber cannot be empty")
	}

	query := `
       SELECT
           fuel_get_id,
           kazs_number,
           jar_number,
           start_time,
           end_time,
           monitoring_finish_time,
           fuel_type,
           doc_number,
           sensor_before_give,
           sensor_after_give,
           fuel_liters,
           fuel_liters_plan,
           speed,
           send_status,
           errors
       FROM
           fuel_get_transactions
       WHERE
           jar_number = $1 AND (end_time IS NULL OR monitoring_finish_time IS NULL)
       ORDER BY
           start_time DESC
       LIMIT 1;
    `

	var t struct {
		FuelGetID            string
		KazsNumber           string
		JarNumber            string
		StartTime            int64
		EndTime              sql.NullInt64
		MonitoringFinishTime sql.NullInt64
		FuelType             string
		DocNumber            string
		SensorBeforeGive     sql.NullString
		SensorAfterGive      sql.NullString
		FuelLiters           sql.NullFloat64
		FuelLitersPlan       float64
		Speed                float64
		SendStatus           bool
		Errors               sql.NullString
	}

	row := f.conn.QueryRow(query, jarNumber)
	err := row.Scan(
		&t.FuelGetID,
		&t.KazsNumber,
		&t.JarNumber,
		&t.StartTime,
		&t.EndTime,
		&t.MonitoringFinishTime,
		&t.FuelType,
		&t.DocNumber,
		&t.SensorBeforeGive,
		&t.SensorAfterGive,
		&t.FuelLiters,
		&t.FuelLitersPlan,
		&t.Speed,
		&t.SendStatus,
		&t.Errors,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.LastFuelGetReceipt{}, sql.ErrNoRows
		}
		return models.LastFuelGetReceipt{}, fmt.Errorf("failed to get transaction for jar number %s: %w", jarNumber, err)
	}

	receipt := models.LastFuelGetReceipt{
		FuelGetID:      t.FuelGetID,
		KazsNumber:     t.KazsNumber,
		JarId:          jarNumber,
		FuelType:       t.FuelType,
		StartTime:      t.StartTime,
		DocNumber:      t.DocNumber,
		FuelLitersPlan: t.FuelLitersPlan,
		Speed:          t.Speed,
	}

	if t.EndTime.Valid {
		receipt.EndTime = t.EndTime.Int64
	}
	if t.MonitoringFinishTime.Valid {
		receipt.MonitoringFinishTime = t.MonitoringFinishTime.Int64
	} else {
		receipt.MonitoringFinishTime = 0
	}
	if t.FuelLiters.Valid {
		receipt.FuelLiter = t.FuelLiters.Float64
	}
	if t.Errors.Valid {
		receipt.Errors = t.Errors.String
	}
	if t.SensorBeforeGive.Valid && t.SensorBeforeGive.String != "" {
		_ = json.Unmarshal([]byte(t.SensorBeforeGive.String), &receipt.SensorBeforeGive)
	}
	if t.SensorAfterGive.Valid && t.SensorAfterGive.String != "" {
		_ = json.Unmarshal([]byte(t.SensorAfterGive.String), &receipt.SensorAfterGive)
	}

	return receipt, nil
}

func (f *FuelGet) GetUnsentFuelGetTransactions() ([]models.LastFuelGetReceipt, error) {
	query := `
       SELECT
           fuel_get_id,
           kazs_number,
           jar_number,
           start_time,
           end_time,
           fuel_type,
           doc_number,
           sensor_before_give,
           sensor_after_give,
           fuel_liters,
           fuel_liters_plan,
           send_status,
           errors
       FROM
           fuel_get_transactions
       WHERE
           send_status = 0 AND end_time IS NOT NULL
       ORDER BY
           start_time DESC
    `

	rows, err := f.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query for unsent transactions: %w", err)
	}
	defer rows.Close()

	var receipts []models.LastFuelGetReceipt

	for rows.Next() {
		var t struct {
			FuelGetID        string
			KazsNumber       string
			JarNumber        string
			StartTime        int64
			EndTime          sql.NullInt64
			FuelType         string
			DocNumber        string
			SensorBeforeGive sql.NullString
			SensorAfterGive  sql.NullString
			FuelLiters       sql.NullFloat64
			FuelLitersPlan   float64
			SendStatus       bool
			Errors           sql.NullString
		}

		if err := rows.Scan(
			&t.FuelGetID, &t.KazsNumber, &t.JarNumber, &t.StartTime, &t.EndTime,
			&t.FuelType, &t.DocNumber, &t.SensorBeforeGive, &t.SensorAfterGive,
			&t.FuelLiters, &t.FuelLitersPlan, &t.SendStatus, &t.Errors,
		); err != nil {
			return receipts, fmt.Errorf("failed to scan transaction row: %w", err)
		}

		receipt := models.LastFuelGetReceipt{
			FuelGetID:      t.FuelGetID,
			KazsNumber:     t.KazsNumber,
			JarId:          t.JarNumber,
			FuelType:       t.FuelType,
			StartTime:      t.StartTime,
			DocNumber:      t.DocNumber,
			FuelLitersPlan: t.FuelLitersPlan,
		}
		if t.EndTime.Valid {
			receipt.EndTime = t.EndTime.Int64
		}
		if t.FuelLiters.Valid {
			receipt.FuelLiter = t.FuelLiters.Float64
		}
		if t.Errors.Valid {
			receipt.Errors = t.Errors.String
		}

		if t.SensorBeforeGive.Valid && t.SensorBeforeGive.String != "" {
			_ = json.Unmarshal([]byte(t.SensorBeforeGive.String), &receipt.SensorBeforeGive)
		}

		if t.SensorAfterGive.Valid && t.SensorAfterGive.String != "" {
			_ = json.Unmarshal([]byte(t.SensorAfterGive.String), &receipt.SensorAfterGive)
		}
		receipts = append(receipts, receipt)
	}

	if err = rows.Err(); err != nil {
		return receipts, fmt.Errorf("error during rows iteration: %w", err)
	}

	return receipts, nil
}
