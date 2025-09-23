package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type TRKConfig struct {
	Address           string                       `json:"Address"`
	GeneralParameters map[string]SettingsParameter `json:"GeneralParameters"`
	DriverName        string                       `json:"DriverName"`
	DevicePNumber     string                       `json:"DevicePNumber"`
	AdminPass         string                       `json:"AdminPass"`
}

type SettingsParameter struct {
	Comment string  `json:"Comment"`
	Value   float32 `json:"Value"`
}

// Вспомогательная функция для чтения и парсинга конфига
func ReadConfig(filePath string) (*TRKConfig, error) {
	if !FileExists(filePath) {
		return nil, fmt.Errorf("file %s not found", filePath)
	}
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file %s error: %w", filePath, err)
	}

	var config TRKConfig
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, fmt.Errorf("parse json file %s error: %w", filePath, err)
	}
	return &config, nil
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true // файл существует
	}
	if os.IsNotExist(err) {
		return false // файл не существует
	}
	return false // произошла другая ошибка (например, нет прав доступа)
}
