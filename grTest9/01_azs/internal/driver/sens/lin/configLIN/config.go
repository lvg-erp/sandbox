package configLIN

import (
	"encoding/json"
	"fmt"
	"fuelazs/internal/driver/sens"
	"os"
)

type SensLinRsUsbLan struct {
	Address            string                       `json:"Address"`
	SettingsParameters map[string]SettingsParameter `json:"SettingsParameters"`
	DriverName         string                       `json:"DriverName"`
	DevicePNumber      string                       `json:"DevicePNumber"`
}

type SettingsParameter struct {
	Comment string  `json:"Comment"`
	Value   float32 `json:"Value"`
}

// Вспомогательная функция для чтения и парсинга конфига
func ReadConfig(filePath string) (*SensLinRsUsbLan, error) {
	if !sens.FileExists(filePath) {
		return nil, fmt.Errorf("file %s not found", filePath)
	}
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file %s error: %w", filePath, err)
	}

	var config SensLinRsUsbLan
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, fmt.Errorf("parse json file %s error: %w", filePath, err)
	}
	return &config, nil
}
