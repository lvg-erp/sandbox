package topazAdapter

import (
	"encoding/json"
	"fmt"
	"go.bug.st/serial"
	"os"
	"runtime"
)

// TOPAZCalculateChecksum Метод вычисления контрольной суммы
func TOPAZCalculateChecksum(data []byte) byte {
	if len(data) < 4 {
		return 0 // Ошибка: слишком короткий пакет
	}

	var checksum byte = 0

	// XOR только для байтов с нечётными индексами (после STX)
	for i := 2; i < len(data)-1; i += 2 {
		checksum ^= data[i]
	}

	// Добавляем 0x40
	checksum |= 0x40
	return checksum
}

// getTRKConnectionInfo Получаем информацию о подключение устройств ТРК
func getTRKConnectionInfo() (*TRKMaping, error) {
	filePath := "TopazMaping.json"

	if !fileExists(filePath) {
		return nil, fmt.Errorf("file " + filePath + " does not exist")
	}

	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("file reading error " + filePath)
	}

	var connectionInfo *TRKMaping
	err = json.Unmarshal(file, &connectionInfo)
	if err != nil {
		return nil, fmt.Errorf("json unmarshalling error " + filePath)
	}

	return connectionInfo, nil
}

// fileExists Вспомогательный метод для проверки наличия файла
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true // файл существует
	}
	if os.IsNotExist(err) {
		return false // файл не существует
	}
	return false // произошла другая ошибка (например, нет прав доступа)
}

// getOSPortName Преобразуем номер порта в имя, зависящее от ОС
func getOSPortName(comNumber int, isUSB bool) string {
	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("COM%d", comNumber)
	case "linux":
		if isUSB {
			return fmt.Sprintf("/dev/ttyUSB%d", comNumber-1)
		} else {
			return fmt.Sprintf("/dev/ttyS%d", comNumber-1)
		}
	default:
		return fmt.Sprintf("COM%d", comNumber)
	}
}

// parity Метод для конвертации четности
func parity(parity string) serial.Parity {
	switch parity {
	case "Even":
		return serial.EvenParity
	case "Odd":
		return serial.OddParity
	case "Mark":
		return serial.MarkParity
	case "Space":
		return serial.SpaceParity
	default:
		return serial.NoParity
	}
}

// stopBits Метод для конвертации стоп битов
func stopBits(stopBits float32) serial.StopBits {
	switch stopBits {
	case 2:
		return serial.TwoStopBits
	case 1.5:
		return serial.OnePointFiveStopBits
	default:
		return serial.OneStopBit
	}
}
