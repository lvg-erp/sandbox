package sens_LIN_RS_USB_LAN

import (
	"fmt"
	"fuelazs/internal/driver/sens"
	"fuelazs/internal/driver/sens/lin/configLIN"
	"path/filepath"
)

// Вспомогательная функция для отправки команды
func (l *LinDriver) sendLinCommand(devicePNumber string, address byte, commandCode byte, payload []byte, deviceType string) ([]byte, error) {
	if !(*l.Adapter)["1"].IsOpen() {
		return nil, fmt.Errorf("LinDriver port is not open for command 0x%X", commandCode)
	}

	cmd := []byte{SyncByte, address, byte(len(payload)), commandCode}
	cmd = append(cmd, payload...)
	crc := sens.SENSCalculateCRC(cmd)
	cmd = append(cmd, crc)

	res, err := (*l.Adapter)["1"].SendCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("SendCommand error for command 0x%X: %w", commandCode, err)
	}

	return res, nil
}

// Вспомогательная функция для проверки уникальности адреса
func (l *LinDriver) isAddressUnique(newAddress string, currentDevicePNumber string) error {
	files, err := filepath.Glob(DriverName + "_*.json")
	if err != nil {
		return fmt.Errorf("error searching for driver files: %w", err)
	}

	for _, filePath := range files {
		baseName := filepath.Base(filePath)
		if baseName == DriverName+"_"+currentDevicePNumber+".json" || baseName == DriverName+"_default.json" {
			continue
		}

		existingConfig, err := configLIN.ReadConfig(filePath)
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

// Валидация структуры и типов dataJSON
func validateSettingsStructure(template *configLIN.SensLinRsUsbLan, data *configLIN.SensLinRsUsbLan) error {
	if template == nil || data == nil {
		return fmt.Errorf("cannot validate nil config")
	}
	// Проверка наличия всех ключей из шаблона в данных
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
	return nil
}
