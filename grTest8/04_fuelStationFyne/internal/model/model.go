package model

import "database/sql"

type FuelOperation struct {
	ColumnID       string        `json:"column_id"`
	FuelType       string        `json:"fuel_type"`
	Liters         float64       `json:"liters"`
	Action         string        `json:"action"`
	FillTimestamp  sql.NullInt64 `json:"fill_timestamp"`
	DrainTimestamp sql.NullInt64 `json:"drain_timestamp"`
	Timestamp      int64         `json:"timestamp"`
}
