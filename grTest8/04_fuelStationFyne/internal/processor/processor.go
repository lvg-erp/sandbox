package processor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"fuelstation/internal/gui"
	"fuelstation/internal/models"
	"fuelstation/internal/usecase"
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
func insertOperation(db *sql.DB, op models.FuelOperation) error {
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

type Processor struct {
	gui        gui.SectionInterface
	processing *usecase.Processing
}

func NewProcessor(gui gui.SectionInterface) *Processor {
	return &Processor{
		gui:        gui,
		processing: usecase.NewProcessing(gui),
	}
}

func (p *Processor) ProcessQRCode(qrInfo models.ScannerResponse) error {
	log.Printf("Processing QR code: TID=%s, FuelType=%s, Liters=%v", qrInfo.TID, qrInfo.FuelType, qrInfo.Liters)

	if qrInfo.Action == "FuelGive" {
		return p.processing.FuelGive(qrInfo)
	} else if qrInfo.Action == "FuelGet" {
		return p.processing.FuelGet(qrInfo)
	}

	log.Printf("Unknown action: %s", qrInfo.Action)
	return nil
}

func (p *Processor) ProcessJSONFile(ctx context.Context, db *sql.DB, filePath string, action string, jarNumber string) error {
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

	var selectedOp *models.FuelOperation
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
		p.gui.CreateFuelGiveStartScreen(jarNumber, float32(selectedOp.Liters), selectedOp.FuelType, 30)
		time.Sleep(time.Second)
		p.gui.CreateFuelGiveInProgressScreen(jarNumber, 0, float32(selectedOp.Liters))
		time.Sleep(time.Second)
		p.gui.CreateFuelGiveCompleteScreen(jarNumber, float32(selectedOp.Liters), selectedOp.FuelType)
	} else if action == "drain" {
		p.gui.CreateFuelGetStartScreen(jarNumber, float32(selectedOp.Liters), 0, 0, float32(selectedOp.Liters), 30)
		time.Sleep(time.Second)
		expectedAmount := float32(selectedOp.Liters)
		drainedAmount := float32(0) // Примерное значение, замените на реальное
		fuelVolume := float32(selectedOp.Liters)
		jarVolume := float32(100) // Примерное значение, замените на реальное
		getTimer := 5
		p.gui.CreateFuelGetInProgressScreen(jarNumber, expectedAmount, drainedAmount, fuelVolume, jarVolume, getTimer)
		time.Sleep(time.Second)
		p.gui.CreateFuelGiveCompleteScreen(jarNumber, float32(selectedOp.Liters), selectedOp.FuelType)
	}

	log.Printf("ProcessJSONFile: Завершение обработки для action=%s, jarNumber=%s", action, jarNumber)
	return nil
}

// EmulateQRScan эмулирует сканирование QR-кода
func (p *Processor) EmulateQRScan(action, fuelType string, liters float32, jarNumber string) error {
	qrInfo := models.ScannerResponse{
		TID:      "TID_" + time.Now().Format("20060102150405"),
		FuelType: fuelType,
		Liters:   float64(liters),
	}
	if action == "FuelGive" {
		p.processing.OneJarActive = jarNumber == "1"
		p.processing.TwoJarActive = jarNumber == "2"
		return p.processing.FuelGive(qrInfo)
	} else if action == "FuelGet" {
		p.processing.OneJarActive = jarNumber == "1"
		p.processing.TwoJarActive = jarNumber == "2"
		return p.processing.FuelGet(qrInfo)
	}
	return nil
}

func readJSONFile(filePath string) ([]models.FuelOperation, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл %s: %w", filePath, err)
	}
	var operations []models.FuelOperation
	if err := json.Unmarshal(data, &operations); err != nil {
		return nil, err
	}
	return operations, nil
}

func (p *Processor) IsJarActive(jarNumber string) bool {
	if jarNumber == "1" {
		return p.processing.OneJarActive
	} else if jarNumber == "2" {
		return p.processing.TwoJarActive
	}
	return false
}
