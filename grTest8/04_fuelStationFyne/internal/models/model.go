package models

import "database/sql"

type ScannerResponse struct {
	TID string
}

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

type FuelOperation struct {
	TID            string
	JarNumber      string
	FuelType       string
	Liters         float32
	Status         string
	ColumnID       string
	Action         string        // Добавлено поле Action
	FillTimestamp  sql.NullInt64 // Изменено на sql.NullInt64
	DrainTimestamp sql.NullInt64 // Изменено на sql.NullInt64
}
