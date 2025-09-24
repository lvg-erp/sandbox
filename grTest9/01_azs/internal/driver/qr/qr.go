package qr

import (
	"bufio"
	"encoding/json"
	"fmt"
	"fuelazs/internal/usecase/models"
	"go.bug.st/serial"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// SerialPort интерфейс для работы с COM-портом
type SerialPort interface {
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	Close() error
	Drain() error
	ResetInputBuffer() error
	ResetOutputBuffer() error
	SetReadTimeout(time.Duration) error
}

// MockSerialPort заглушка для эмуляции COM-порта
type MockSerialPort struct {
	responseQueue [][]byte
	readIndex     int
	mutex         sync.Mutex
}

func NewMockSerialPort(responses [][]byte) *MockSerialPort {
	return &MockSerialPort{
		responseQueue: responses,
		readIndex:     0,
	}
}

func (m *MockSerialPort) Write(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("empty command")
	}
	return len(data), nil
}

func (m *MockSerialPort) Read(buf []byte) (int, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.readIndex >= len(m.responseQueue) {
		return 0, fmt.Errorf("no more responses")
	}

	response := m.responseQueue[m.readIndex]
	copy(buf, response)
	m.readIndex++
	return len(response), nil
}

func (m *MockSerialPort) Close() error {
	return nil
}

func (m *MockSerialPort) Drain() error {
	return nil
}

func (m *MockSerialPort) ResetInputBuffer() error {
	return nil
}

func (m *MockSerialPort) ResetOutputBuffer() error {
	return nil
}

func (m *MockSerialPort) SetReadTimeout(_ time.Duration) error {
	return nil
}

type QRAdapter struct {
	Port              SerialPort
	QRAdapterSettings QRAdapterSettings
	mutex             sync.Mutex
	Maping            *QRMaping
}

type QRAdapterSettings struct {
	PortName    string
	BaudRate    int
	DataBits    int
	Parity      serial.Parity
	StopBits    serial.StopBits
	ReadTimeout int
	Token       string
}

func NewQRAdapter(portFactory func(string, *serial.Mode) (SerialPort, error)) (*QRAdapter, error) {
	if portFactory == nil {
		portFactory = func(_ string, _ *serial.Mode) (SerialPort, error) {
			return NewMockSerialPort([][]byte{
				[]byte(`{"code":"test_qr_code"}`),
			}), nil
		}
	}

	QRInfo, err := getConnectionInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get QR connection info: %w", err)
	}

	// Используем QRInfo.QR как структуру Settings
	qrConfig := QRInfo.QR

	comName := getOSPortName(qrConfig.COMPort, qrConfig.IsUSB, qrConfig.IsACM)

	port, err := portFactory(comName, &serial.Mode{
		BaudRate: qrConfig.BaudRate,
		DataBits: qrConfig.DataBits,
		Parity:   parity(qrConfig.Parity),
		StopBits: stopBits(qrConfig.StopBits),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port: %w", err)
	}

	err = port.SetReadTimeout(time.Duration(qrConfig.ReadTimeout) * time.Second)
	if err != nil {
		port.Close()
		return nil, fmt.Errorf("failed to set read timeout: %w", err)
	}

	return &QRAdapter{
		Port: port,
		QRAdapterSettings: QRAdapterSettings{
			PortName:    comName,
			BaudRate:    qrConfig.BaudRate,
			DataBits:    qrConfig.DataBits,
			Parity:      parity(qrConfig.Parity),
			StopBits:    stopBits(qrConfig.StopBits),
			ReadTimeout: qrConfig.ReadTimeout,
			Token:       "", // Если токен нужен, добавьте его в QRMaping
		},
		Maping: QRInfo,
		mutex:  sync.Mutex{},
	}, nil
}

func (r *QRAdapter) Read(dataChan chan<- models.ScannerResponse, stopChan <-chan struct{}, activeChan <-chan bool) {
	if r.Port == nil {
		fmt.Println("QRReader.Read: Последовательный порт не инициализирован.")
		return
	}

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	scanner := bufio.NewScanner(r.Port)
	isPaused := false
	for {
		if isPaused {
			select {
			case <-stopChan:
				fmt.Println("QRReader.Read: Получен сигнал остановки во время паузы.")
				return
			case newActiveState := <-activeChan:
				if newActiveState {
					isPaused = false
					fmt.Println("QRReader.Read: Работа возобновлена.")
				}
				continue
			case <-ticker.C:
				_ = r.Reopen()
			}
		}

		select {
		case <-stopChan:
			fmt.Println("QRReader.Read: Получен сигнал остановки.")
			return
		case newActiveState := <-activeChan:
			if !newActiveState {
				isPaused = true
				fmt.Println("QRReader.Read: Работа приостановлена.")
			}
		case <-ticker.C:
			_ = r.Reopen()
			scanner = bufio.NewScanner(r.Port)
		default:
			scanSucces := scanner.Scan()

			if !scanSucces {
				err := scanner.Err()
				if strings.Contains(err.Error(), "multiple Read calls return no data or error") {
					continue
				}
				continue
			} else {
				line := scanner.Text()
				slog.Info("QR Read:", "qr_info", line)
				cleanedLine := strings.TrimSpace(line)
				if len(cleanedLine) == 0 {
					continue
				}

				var response models.ScannerResponse
				if err := json.Unmarshal([]byte(cleanedLine), &response); err != nil {
					fmt.Printf("QRReader.Read: Ошибка десериализации JSON '%s': %v\n", cleanedLine, err)
					continue
				}

				select {
				case dataChan <- response:
				case <-stopChan:
					fmt.Println("QRReader.Read: Получен сигнал остановки при отправке данных.")
					return
				}
			}
		}
	}
}

func (c *QRAdapter) Reopen() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.IsOpen() {
		err := c.Port.Close()
		if err != nil {
			return fmt.Errorf("close port error: %s", err)
		}
		c.Port = nil
	}

	newPort := NewMockSerialPort([][]byte{
		[]byte(`{"code":"test_qr_code"}`),
	})

	err := newPort.SetReadTimeout(time.Duration(c.QRAdapterSettings.ReadTimeout) * time.Second)
	if err != nil {
		newPort.Close()
		return fmt.Errorf("set read timeout error: %s", err)
	}

	c.Port = newPort
	return nil
}

func (c *QRAdapter) IsOpen() bool {
	return c.Port != nil
}
