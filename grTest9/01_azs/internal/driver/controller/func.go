package controller

import (
	"encoding/json"
	"fmt"
	"go.bug.st/serial"
	"os"
	"path/filepath"
	"runtime"
)

func getFilePath() string {
	// Предполагаем, что корень проекта — две директории выше от sensAdapter
	wd, _ := os.Getwd() // Текущая рабочая директория
	return filepath.Join(wd, "internal", "driver", "controller", "config", "ControllerMaping.json")
}

func getControllerConnectionInfo() (*ControllerMaping, error) {
	//TODO
	//filePath := "ControllerMaping.json"
	filePath := getFilePath()
	if !fileExists(filePath) {
		return nil, fmt.Errorf("file " + filePath + " does not exist")
	}

	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("file reading error " + filePath)
	}

	var connectionInfo *ControllerMaping
	err = json.Unmarshal(file, &connectionInfo)
	if err != nil {
		return nil, fmt.Errorf("json unmarshalling error " + filePath)
	}

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
	return false
}

// getOSPortName Преобразуем номер порта в имя, зависящее от ОС
func getOSPortName(comNumber int, isUsb bool) string {
	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("COM%d", comNumber)
	case "linux":
		if isUsb {
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
