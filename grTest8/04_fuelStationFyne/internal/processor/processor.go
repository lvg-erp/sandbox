package processor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"fuelstation/internal/gui"
	"fuelstation/internal/model"
	"log"
	"os"
	"time"
)

// Маппинг UUID на номера колонок
var columnIDToJarNumber = map[string]string{
	"123e4567-e89b-12d3-a456-426614174000": "1",
	"987fcdeb-51a2-437b-9f3d-8a7b3c2d1e45": "2",
}

// jarNumberToColumnID для обратного маппинга
var jarNumberToColumnID = map[string]string{
	"1": "123e4567-e89b-12d3-a456-426614174000",
	"2": "987fcdeb-51a2-437b-9f3d-8a7b3c2d1e45",
}

// insertOperation вставляет операцию в базу данных
func insertOperation(db *sql.DB, op model.FuelOperation) error {
	query := `
		INSERT INTO fuel_operations (column_id, fuel_type, liters, action, fill_timestamp, drain_timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := db.Exec(query,
		op.ColumnID,
		op.FuelType,
		op.Liters,
		op.Action,
		op.FillTimestamp.Int64,
		op.DrainTimestamp.Int64,
	)
	if err != nil {
		log.Printf("insertOperation: Ошибка SQL: %v, операция: %+v", err, op)
		return fmt.Errorf("ошибка вставки операции в базу данных: %w", err)
	}
	log.Printf("insertOperation: Операция успешно вставлена: %+v", op)
	return nil
}

func ProcessJSONFile(ctx context.Context, g *gui.Gui, db *sql.DB, filePath string, action string, jarNumber string) error {
	log.Printf("ProcessJSONFile: Начало обработки JSON файла для action=%s, jarNumber=%s", action, jarNumber)
	operations, err := readJSONFile(filePath)
	if err != nil {
		return fmt.Errorf("ошибка чтения JSON файла: %w", err)
	}

	// Ищем операцию с нужным column_id и action
	columnID := jarNumberToColumnID[jarNumber]
	if columnID == "" {
		return fmt.Errorf("неизвестный jarNumber %s", jarNumber)
	}

	var selectedOp *model.FuelOperation
	for _, op := range operations {
		if op.ColumnID == columnID && op.Action == action {
			selectedOp = &op
			break
		}
	}
	if selectedOp == nil {
		return fmt.Errorf("операция с column_id=%s и action=%s не найдена в JSON", columnID, action)
	}

	// Устанавливаем текущее время Unix
	now := time.Now().UnixMilli()
	if action == "fill" {
		selectedOp.FillTimestamp.Int64 = now
		selectedOp.FillTimestamp.Valid = true
		selectedOp.DrainTimestamp.Int64 = 0
		selectedOp.DrainTimestamp.Valid = false
	} else if action == "drain" {
		selectedOp.DrainTimestamp.Int64 = now
		selectedOp.DrainTimestamp.Valid = true
		selectedOp.FillTimestamp.Int64 = 0
		selectedOp.FillTimestamp.Valid = false
	}

	// Записываем операцию в базу данных
	if err := insertOperation(db, *selectedOp); err != nil {
		return fmt.Errorf("ошибка записи операции в базу данных: %w", err)
	}

	// Обновляем GUI
	if action == "fill" {
		g.CreateFuelGiveStartScreen(jarNumber, float32(selectedOp.Liters), selectedOp.FuelType, 30)
		time.Sleep(time.Second)
		g.CreateFuelGiveInProgressScreen(jarNumber, selectedOp.FuelType, float32(selectedOp.Liters), float32(selectedOp.Liters))
		time.Sleep(time.Second)
		g.CreateFuelGiveCompleteScreen(jarNumber, selectedOp.FuelType, "DOC123", float32(selectedOp.Liters), float32(selectedOp.Liters), selectedOp.FillTimestamp.Int64, now)
	} else if action == "drain" {
		g.CreateFuelGetStartScreen(jarNumber, selectedOp.FuelType, 100, 200, float32(selectedOp.Liters), 30)
		time.Sleep(time.Second)
		g.CreateFuelGetInProgressScreen(jarNumber, selectedOp.FuelType, float32(selectedOp.Liters), float32(selectedOp.Liters), 100, 300, 5)
		time.Sleep(time.Second)
		g.CreateFuelGetCompleteScreen(jarNumber, selectedOp.FuelType, "DOC124", 100, 200, float32(selectedOp.Liters), float32(selectedOp.Liters), selectedOp.DrainTimestamp.Int64, now, 10)
	}

	log.Printf("ProcessJSONFile: Завершение обработки для action=%s, jarNumber=%s", action, jarNumber)
	return nil
}

func readJSONFile(filePath string) ([]model.FuelOperation, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл %s: %w", filePath, err)
	}
	var operations []model.FuelOperation
	if err := json.Unmarshal(data, &operations); err != nil {
		return nil, err
	}
	return operations, nil
}
