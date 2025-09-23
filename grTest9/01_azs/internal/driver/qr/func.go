package qr

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.bug.st/serial"
	"os"
	"runtime"
)

func getConnectionInfo() (*QRMaping, error) {
	filePath := "QRMaping.json"

	if !fileExists(filePath) {
		return nil, errors.New("file " + filePath + " does not exist")
	}

	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.New("file reading error " + filePath)
	}

	var connectionInfo *QRMaping
	err = json.Unmarshal(file, &connectionInfo)
	if err != nil {
		return nil, errors.New("json unmarshalling error " + filePath)
	}

	// Парсим номер виртуального уровнемера/номер ком порта
	return connectionInfo, nil
}

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
func getOSPortName(comNumber int, isUSB bool, isACM bool) string {
	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("COM%d", comNumber)
	case "linux":
		portIndex := comNumber - 1
		switch {
		case isUSB:
			return fmt.Sprintf("/dev/ttyUSB%d", portIndex)
		case isACM:
			return fmt.Sprintf("/dev/ttyACM%d", portIndex)
		default:
			return fmt.Sprintf("/dev/ttyS%d", portIndex)
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
