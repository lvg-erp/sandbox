package qr

import (
	"encoding/json"
	_ "errors"
	"fmt"
	"go.bug.st/serial"
	"os"
	"path/filepath"
	"runtime"
)

//// QRMaping структура для десериализации JSON
//type QRMaping struct {
//	QR map[string]struct {
//		IsUSB       bool    `json:"IsUSB"`
//		IsACM       bool    `json:"IsACM"`
//		COMPort     int     `json:"COMPort"`
//		BaudRate    int     `json:"BaudRate"`
//		DataBits    int     `json:"DataBits"`
//		Parity      string  `json:"Parity"`
//		StopBits    float32 `json:"StopBits"`
//		ReadTimeout int     `json:"ReadTimeout"`
//	} `json:"QR"`
//}

func getFilePath() string {
	// Получаем путь к корню проекта
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	// Предполагаем, что QRMaping.json находится в корне проекта
	return filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(dir))), "QRMaping.json")
}

func getConnectionInfo() (*QRMaping, error) {
	filePath := getFilePath()
	fmt.Println(filePath)
	if !fileExists(filePath) {
		// Эмуляция данных, если файл не найден
		data := []byte(`{
			"QR": {
				"1": {
					"IsUSB": false,
					"IsACM": false,
					"COMPort": 4,
					"BaudRate": 9600,
					"DataBits": 8,
					"Parity": "None",
					"StopBits": 1,
					"ReadTimeout": 2
				}
			}
		}`)

		var connectionInfo QRMaping
		err := json.Unmarshal(data, &connectionInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal default QRMaping.json: %w", err)
		}
		return &connectionInfo, nil
	}

	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var connectionInfo QRMaping
	err = json.Unmarshal(file, &connectionInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON from %s: %w", filePath, err)
	}

	return &connectionInfo, nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
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

// stopBits Метод для конвертации стоп-битов
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
