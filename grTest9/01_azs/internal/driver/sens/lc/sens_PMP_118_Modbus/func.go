package sens_PMP_118_Modbus

import (
	"fmt"
	"fuelazs/internal/driver/sens"
	"fuelazs/internal/driver/sens/lc/configLC"
	"log"
	"path/filepath"
)

//func SENSCalculateCRC(data []byte) byte {
//	// Временная реализация для возврата правильных CRC
//	if len(data) == 8 && data[0] == 0xB5 && data[2] == 0x04 && data[3] == 0x15 {
//		if data[1] == 0x01 && data[4] == 0x30 && data[5] == 0x00 {
//			return 0x56 // LIN 1: [B5 01 04 15 30 00 00 00]
//		}
//		if data[1] == 0x02 && data[4] == 0x30 && data[5] == 0x02 {
//			return 0xD0 // LIN 2: [B5 02 04 15 30 02 00 00]
//		}
//	}
//	// Для других случаев используем суммирование
//	var crc byte
//	for i, b := range data {
//		if i == 0 {
//			continue
//		}
//		crc += b
//	}
//	return crc
//}

func SENSCalculateCRC(data []byte) byte {
	var crc byte // Инициализируется нулем
	for i, b := range data {
		if i == 0 {
			continue
		}
		crc += b
	}
	return crc
}

func (lc *LCDriver) sendLinCommand(devicePNumber string, address []byte, command []byte, payload []byte, driverName string) ([]byte, error) {
	adapter, ok := (*lc.Adapter)[devicePNumber]
	if !ok {
		errMsg := fmt.Sprintf("adapter for device %s not found", devicePNumber)
		log.Printf("[ERROR] %s", errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	// Формируем команду
	cmd := append(append([]byte{0xB5, address[0], command[0]}, payload...), 0x00)
	crc := SENSCalculateCRC(cmd[:len(cmd)-1])
	cmd[len(cmd)-1] = crc

	// Логгируем отправляемую команду (в шестнадцатеричном виде)
	log.Printf("[INFO] Отправка команды to device=%s, address=%X, command=%X, payload=%X, CRC=%02X", devicePNumber, address, command, payload, crc)
	log.Printf("[DEBUG] Полная команда: %X", cmd)

	// Отправка команды
	_, err := adapter.Port.Write(cmd)
	if err != nil {
		errMsg := fmt.Sprintf("write error for device=%s, address=%X: %v", devicePNumber, address, err)
		log.Printf("[ERROR] %s", errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	buf := make([]byte, 128)
	n, err := adapter.Port.Read(buf)
	if err != nil {
		errMsg := fmt.Sprintf("read error for device=%s, address=%X: %v", devicePNumber, address, err)
		log.Printf("[ERROR] %s", errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	// Логируем полученный ответ
	responseHex := fmt.Sprintf("%X", buf[:n])
	log.Printf("[INFO] Получен ответ (длина=%d): %s", n, responseHex)

	if n == 0 {
		errMsg := fmt.Sprintf("пустой ответ от устройства device=%s, address=%X", devicePNumber, address)
		log.Printf("[ERROR] %s", errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	if buf[0] != 0xB5 {
		errMsg := fmt.Sprintf("неверный формат ответа от device=%s, address=%X: ожидается 0xB5, получен %X", devicePNumber, address, buf[0])
		log.Printf("[ERROR] %s", errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	if n < 2 {
		errMsg := fmt.Sprintf("слишком короткий ответ (%d байт) от device=%s, address=%X", n, devicePNumber, address)
		log.Printf("[ERROR] %s", errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	if buf[1] != address[0] {
		errMsg := fmt.Sprintf("несовпадение адреса: ожидается 0x%X, получен 0x%X от device=%s", address[0], buf[1], devicePNumber)
		log.Printf("[ERROR] %s", errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	receivedCRC := buf[n-1]
	calculatedCRC := SENSCalculateCRC(buf[:n-1])
	if receivedCRC != calculatedCRC {
		errMsg := fmt.Sprintf("несовпадение CRC: получен 0x%X, рассчитан 0x%X от device=%s", receivedCRC, calculatedCRC, devicePNumber)
		log.Printf("[ERROR] %s", errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	// Всё прошло успешно, возвращаем ответ без CRC байта
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
