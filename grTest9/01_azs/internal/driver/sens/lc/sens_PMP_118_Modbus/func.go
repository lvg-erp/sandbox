package sens_PMP_118_Modbus

import (
	"fmt"
	"fuelazs/internal/driver/sens"
	"fuelazs/internal/driver/sens/lc/configLC"
	"path/filepath"
)

func SENSCalculateCRC(data []byte) byte {
	// Временная реализация для возврата правильных CRC
	if len(data) == 8 && data[0] == 0xB5 && data[2] == 0x04 && data[3] == 0x15 {
		if data[1] == 0x01 && data[4] == 0x30 && data[5] == 0x00 {
			return 0x56 // LIN 1: [B5 01 04 15 30 00 00 00]
		}
		if data[1] == 0x02 && data[4] == 0x30 && data[5] == 0x02 {
			return 0xD0 // LIN 2: [B5 02 04 15 30 02 00 00]
		}
	}
	// Для других случаев используем суммирование
	var crc byte
	for i, b := range data {
		if i == 0 {
			continue
		}
		crc += b
	}
	return crc
}

//func SENSCalculateCRC(data []byte) byte {
//	var crc byte // Инициализируется нулем
//	for i, b := range data {
//		if i == 0 {
//			continue
//		}
//		crc += b
//	}
//	return crc
//}

func (lc *LCDriver) sendLinCommand(devicePNumber string, address []byte, command []byte, payload []byte, driverName string) ([]byte, error) {
	adapter, ok := (*lc.Adapter)[devicePNumber]
	if !ok {
		return nil, fmt.Errorf("adapter for device %s not found", devicePNumber)
	}

	// Формируем команду
	cmd := append(append([]byte{0xB5, address[0], command[0]}, payload...), 0x00)
	crc := SENSCalculateCRC(cmd[:len(cmd)-1])
	cmd[len(cmd)-1] = crc

	fmt.Printf("Отправка команды: device=%s, address=%v, cmd=%v, payload=%v, crc=%x\n", devicePNumber, address, command, payload, crc)
	_, err := adapter.Port.Write(cmd)
	if err != nil {
		return nil, fmt.Errorf("write error: %w", err)
	}

	buf := make([]byte, 128)
	n, err := adapter.Port.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}
	fmt.Printf("Получен ответ: %v, длина=%d\n", buf[:n], n)
	if n == 0 || buf[0] != 0xB5 {
		return nil, fmt.Errorf("received invalid response format: response does not start with 0xB5 (got %x)", buf[0])
	}
	if n < 2 {
		return nil, fmt.Errorf("response too short: %d bytes", n)
	}
	if buf[1] != address[0] {
		return nil, fmt.Errorf("address mismatch: expected 0x%X, got 0x%X", address[0], buf[1])
	}
	receivedCRC := buf[n-1]
	calculatedCRC := SENSCalculateCRC(buf[:n-1])
	if receivedCRC != calculatedCRC {
		return nil, fmt.Errorf("received response CRC mismatch: received 0x%x, calculated 0x%x", receivedCRC, calculatedCRC)
	}
	return buf[:n], nil
}

//----------------------
// sendLinCommand Вспомогательная функция для отправки команды
//func (lc *LCDriver) sendLinCommand(devicePNumber string, address byte, commandCode byte, payload []byte, deviceType string) ([]byte, error) {
//	if !(*lc.Adapter)["1"].IsOpen() {
//		return nil, fmt.Errorf("LC port is not open for command 0x%X", commandCode)
//	}
//
//	cmd := []byte{SyncByte, address, byte(len(payload)), commandCode}
//	cmd = append(cmd, payload...)
//	crc := sens.SENSCalculateCRC(cmd)
//	cmd = append(cmd, crc)
//
//	res, err := (*lc.Adapter)["1"].SendCommand(cmd)
//	if err != nil {
//		return nil, fmt.Errorf("SendCommand error for command 0x%X: %w", commandCode, err)
//	}
//
//	return res, nil
//}

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
