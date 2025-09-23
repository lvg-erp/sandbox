package configBK

import (
	"encoding/json"
	"fmt"
	"fuelazs/internal/driver/sens"
	"os"
)

type SensBk2P struct {
	Address            string                       `json:"Address"`
	SettingsParameters map[string]SettingsParameter `json:"SettingsParameters"`
	Tables             map[string]SettingsTables    `json:"Tables"`
	DriverName         string                       `json:"DriverName"`
	DevicePNumber      string                       `json:"DevicePNumber"`
}

type SettingsTables struct {
	Comment string    `json:"Comment"`
	Value   []float32 `json:"Value"`
}

type SettingsParameter struct {
	Comment string  `json:"Comment"`
	Value   float32 `json:"Value"`
}

// ReadConfig Вспомогательная функция для чтения и парсинга конфига
func ReadConfig(filePath string) (*SensBk2P, error) {
	if !sens.FileExists(filePath) {
		return nil, fmt.Errorf("file %s not found", filePath)
	}
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file %s error: %w", filePath, err)
	}

	var config SensBk2P
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, fmt.Errorf("parse json file %s error: %w", filePath, err)
	}
	return &config, nil
}
