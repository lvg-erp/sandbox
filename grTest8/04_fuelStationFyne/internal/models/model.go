package models

import "database/sql"

//type ScannerResponse struct {
//	TID string
//}

type TRKResponse struct {
	ValueStr   string
	ValueFloat float32
	Err        error
}

type CaptureErrorDep struct {
	TID        string
	Operation  string
	KazsNumber string
	JarNumber  string
}

type ScannerResponse struct {
	TID      string  `json:"tid"`
	Action   string  `json:"action"`
	FuelType string  `json:"fuel_type"` // Добавлено поле
	Liters   float64 `json:"liters"`    // Добавлено поле
}

type FuelOperation struct {
	ColumnID       string        `json:"column_id"`
	FuelType       string        `json:"fuel_type"`
	Liters         float64       `json:"liters"`
	Action         string        `json:"action"`
	FillTimestamp  sql.NullInt64 `json:"fill_timestamp"`
	DrainTimestamp sql.NullInt64 `json:"drain_timestamp"`
	JarNumber      string        `json:"jar_number"` // Добавлено для соответствия GUI
}

//type FuelOperation struct {
//	TID            string
//	JarNumber      string
//	FuelType       string
//	Liters         float32
//	Status         string
//	ColumnID       string
//	Action         string        // Добавлено поле Action
//	FillTimestamp  sql.NullInt64 // Изменено на sql.NullInt64
//	DrainTimestamp sql.NullInt64 // Изменено на sql.NullInt64
//}
