package sens_BK_2P

import (
	"fmt"
	"fuelazs/internal/driver/sens"
	"fuelazs/internal/driver/sens/bk/configBK"
	"path/filepath"
)

// Вспомогательная функция для отправки команды
func (bk *BK) sendLinCommand(devicePNumber string, address byte, commandCode byte, payload []byte, deviceType string) ([]byte, error) {
	if !(*bk.Adapter)["1"].IsOpen() {
		return nil, fmt.Errorf("BK port is not open for command 0x%X", commandCode)
	}

	cmd := []byte{SyncByte, address, byte(len(payload)), commandCode}
	cmd = append(cmd, payload...)
	crc := sens.SENSCalculateCRC(cmd)
	cmd = append(cmd, crc)

	res, err := (*bk.Adapter)["1"].SendCommand(cmd)
	if err != nil {
		return nil, fmt.Errorf("SendCommand error for command 0x%X: %w", commandCode, err)
	}

	return res, nil
}

// Вспомогательная функция для проверки уникальности адреса
func (bk *BK) isAddressUnique(newAddress string, currentDevicePNumber string) error {
	files, err := filepath.Glob(DriverName + "_*.json")
	if err != nil {
		return fmt.Errorf("error searching for driver files: %w", err)
	}

	for _, filePath := range files {
		baseName := filepath.Base(filePath)
		if baseName == DriverName+"_"+currentDevicePNumber+".json" || baseName == DriverName+"_default.json" {
			continue
		}

		existingConfig, err := configBK.ReadConfig(filePath)
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
func validateSettingsStructure(template *configBK.SensBk2P, data *configBK.SensBk2P) error {
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

// encodeTableData Метод для кодирования таблиц (float32 -> []byte)
func encodeTableData(values []float32) ([]byte, error) {
	buf := make([]byte, 0, len(values))
	for _, v := range values {
		buf = append(buf, byte(v))
	}
	if len(values) > 0 && len(buf) == 0 {
		return nil, fmt.Errorf("table data encoding resulted in empty buffer (likely incorrect encoding logic)")
	}
	return buf, nil
}

// decodeTableData Метод для декодирования таблиц ([]byte -> float32)
func decodeTableData(data []byte) ([]float32, error) {
	if len(data)%2 != 0 {
		if len(data) > 0 {
			data = data[:len(data)-1]
		}
	}
	result := make([]float32, 0, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		val1 := float32(data[i])
		val2 := float32(data[i+1])
		result = append(result, val1, val2)
	}
	return result, nil
}
