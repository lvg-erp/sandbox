package sens_PMP_118_Modbus

import (
	"fmt"
	"fuelazs/internal/driver/sens"
	"fuelazs/internal/driver/sens/lc/configLC"
	"path/filepath"
)

// sendLinCommand Вспомогательная функция для отправки команды
func (lc *LCDriver) sendLinCommand(devicePNumber string, address byte, commandCode byte, payload []byte, deviceType string) ([]byte, error) {
	if !(*lc.Adapter)["1"].IsOpen() {
		return nil, fmt.Errorf("LC port is not open for command 0x%X", commandCode)
	}

	cmd := []byte{SyncByte, address, byte(len(payload)), commandCode}
	cmd = append(cmd, payload...)
	crc := sens.SENSCalculateCRC(cmd)
	cmd = append(cmd, crc)

	res, err := (*lc.Adapter)["1"].SendCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("SendCommand error for command 0x%X: %w", commandCode, err)
	}

	return res, nil
}

// sendLinCommand Вспомогательная функция для отправки команды
func (lc *LCDriver) sendLCCommandWithoutRead(devicePNumber string, address byte, commandCode byte, payload []byte, deviceType string) error {
	if !(*lc.Adapter)["1"].IsOpen() {
		return fmt.Errorf("LC port is not open for command 0x%X", commandCode)
	}

	cmd := []byte{SyncByte, address, byte(len(payload)), commandCode}
	cmd = append(cmd, payload...)
	crc := sens.SENSCalculateCRC(cmd)
	cmd = append(cmd, crc)

	err := (*lc.Adapter)["1"].SendCommandWithoutRead(cmd)
	if err != nil {
		return fmt.Errorf("SendCommand error for command 0x%X: %w", commandCode, err)
	}

	return nil
}

// isAddressUnique Вспомогательная функция для проверки уникальности адреса
func (lc *LCDriver) isAddressUnique(newAddress string, currentDevicePNumber string) error {
	files, err := filepath.Glob(DriverName + "_*.json") // Найти все файлы драйвера
	if err != nil {
		return fmt.Errorf("error searching for driver files: %w", err)
	}

	for _, filePath := range files {
		baseName := filepath.Base(filePath)
		if baseName == "lc/"+DriverName+"_"+currentDevicePNumber+".json" || baseName == "lc"+DriverName+"_default.json" {
			continue
		}

		existingConfig, err := configLC.ReadConfig(filePath)
		if err != nil {
			fmt.Printf("Warning: could not read/parse config %s: %v\n", filePath, err)
			continue
		}

		if existingConfig.Address == newAddress {
			return fmt.Errorf("address %s is already used by device in file %s", newAddress, baseName)
		}
	}
	return nil
}

// Валидация структуры и типов dataJSON (пример, нужно адаптировать)
func validateSettingsStructure(template *configLC.SensPMP118Modbus, data *configLC.SensPMP118Modbus) error {
	if template == nil || data == nil {
		return fmt.Errorf("cannot validate nil config")
	}
	for key := range template.SettingsParameters {
		if _, exists := data.SettingsParameters[key]; !exists {
			return fmt.Errorf("missing setting parameter key: %s", key)
		}
	}
	// Проверка отсутствия лишних ключей в данных
	for key := range data.SettingsParameters {
		if _, exists := template.SettingsParameters[key]; !exists {
			return fmt.Errorf("extra setting parameter key found: %s", key)
		}
	}
	// Проверка наличия всех ключей в таблицах
	for key := range template.Tables {
		if _, exists := data.Tables[key]; !exists {
			return fmt.Errorf("missing tables parameter key: %s", key)
		}
	}
	// Проверка отсутствия лишних ключей в таблицах
	for key := range data.Tables {
		if _, exists := template.Tables[key]; !exists {
			return fmt.Errorf("extra tables parameter key found: %s", key)
		}
	}
	return nil
}
