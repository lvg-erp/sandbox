package configLC

import (
	"encoding/json"
	"fmt"
	"fuelazs/internal/driver/sens"
	"os"
)

type SensPMP118Modbus struct {
	Address            string                       `json:"Address"`
	MainRead           MainRead                     `json:"MainRead"`
	OtherRead          MainRead                     `json:"OtherRead"`
	SettingsParameters map[string]SettingsParameter `json:"SettingsParameters"`
	Tables             map[string]SettingsTables    `json:"Tables"`
	DriverName         string                       `json:"DriverName"`
	DevicePNumber      string                       `json:"DevicePNumber"`
}

type MainRead struct {
	Parameters map[string]SettingsReadParameter `json:"Parameters"`
	Tables     map[string]SettingsReadTables    `json:"Tables"`
}

type SettingsParameter struct {
	Comment string  `json:"Comment"`
	Value   float32 `json:"Value"`
}

type SettingsTables struct {
	Comment string    `json:"Comment"`
	Value   []float32 `json:"Value"`
}

type SettingsReadParameter struct {
	Comment string `json:"Comment"`
}

type SettingsReadTables struct {
	Comment string `json:"Comment"`
}

// ReadConfig Вспомогательная функция для чтения и парсинга конфига
func ReadConfig(filePath string) (*SensPMP118Modbus, error) {
	if !sens.FileExists(filePath) {
		return nil, fmt.Errorf("file %s not found", filePath)
	}
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file %s error: %w", filePath, err)
	}

	var config SensPMP118Modbus
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, fmt.Errorf("parse json file %s error: %w", filePath, err)
	}
	return &config, nil
}
