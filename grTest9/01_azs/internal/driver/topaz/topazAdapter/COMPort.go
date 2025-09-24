package topazAdapter

import (
	"bytes"
	"errors"
	"fmt"
	"go.bug.st/serial"
	"sync"
	"time"
)

// Константы
const (
	StartByte = 0x7F
	ErrByte   = 0x15
	ErrByte2  = 0x18
	STX       = 0x02
	ETX       = 0x03
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
	if len(data) == 0 || data[0] != StartByte {
		return 0, ErrInvalidSendCommand
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

var (
	openedPorts     = make(map[string]SerialPort)
	openedPortsLock sync.Mutex
)

func getPortFromCache(comName string) (SerialPort, bool) {
	openedPortsLock.Lock()
	defer openedPortsLock.Unlock()
	port, found := openedPorts[comName]
	return port, found
}

func addPortToCache(comName string, port SerialPort) {
	openedPortsLock.Lock()
	defer openedPortsLock.Unlock()
	openedPorts[comName] = port
}

// TopazAdapterSettings настройки адаптера
type TopazAdapterSettings struct {
	PortName    string
	BaudRate    int
	DataBits    int
	Parity      serial.Parity
	StopBits    serial.StopBits
	ReadTimeout int
}

// TopazAdapter структура адаптера
type TopazAdapter struct {
	port                 SerialPort
	TopazAdapterSettings TopazAdapterSettings
	mutex                sync.Mutex
}

// TRKConfig структура для десериализации конфигурации
type TRKConfig struct {
	TRK map[string]struct {
		IsUSB       bool   `json:"IsUSB"`
		COMPort     int    `json:"COMPort"`
		BaudRate    int    `json:"BaudRate"`
		StopBits    int    `json:"StopBits"`
		DataBits    int    `json:"DataBits"`
		Parity      string `json:"Parity"`
		ReadTimeout int    `json:"ReadTimeout"`
	} `json:"TRK"`
}

// NewTopazAdapter создает адаптеры
func NewTopazAdapter(portFactory func(string, *serial.Mode) (SerialPort, error)) (map[string]*TopazAdapter, error) {
	if portFactory == nil {
		portFactory = func(_ string, _ *serial.Mode) (SerialPort, error) {
			// Эмуляция порта по умолчанию
			return NewMockSerialPort([][]byte{
				{0x7F, 0x01, 0x03, 0x00, 0x00, 0x03}, // Пример ответа с корректным CRC
			}), nil
		}
	}

	topazInfo, err := getTRKConnectionInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get TRK connection info: %w", err)
	}

	topazAdapters := make(map[string]*TopazAdapter, len(topazInfo.TRK))
	for trkKey, trkConfig := range topazInfo.TRK {
		comName := getOSPortName(trkConfig.COMPort, trkConfig.IsUSB)
		var port SerialPort
		var portErr error

		cachedPort, found := getPortFromCache(comName)
		if found {
			port = cachedPort
		} else {
			port, portErr = portFactory(comName, &serial.Mode{
				BaudRate: trkConfig.BaudRate,
				DataBits: trkConfig.DataBits,
				Parity:   parity(trkConfig.Parity),
				StopBits: stopBits(trkConfig.StopBits),
			})
			if portErr != nil {
				return nil, fmt.Errorf("failed to open serial port %s for TRK %s: %w", comName, trkKey, portErr)
			}

			portErr = port.SetReadTimeout(time.Duration(trkConfig.ReadTimeout) * time.Second)
			if portErr != nil {
				port.Close()
				return nil, fmt.Errorf("failed to set read timeout for port %s for TRK %s: %w", comName, trkKey, portErr)
			}

			addPortToCache(comName, port)
		}

		topazAdapters[trkKey] = &TopazAdapter{
			port: port,
			TopazAdapterSettings: TopazAdapterSettings{
				PortName:    comName,
				BaudRate:    trkConfig.BaudRate,
				DataBits:    trkConfig.DataBits,
				Parity:      parity(trkConfig.Parity),
				StopBits:    stopBits(trkConfig.StopBits),
				ReadTimeout: trkConfig.ReadTimeout,
			},
			mutex: sync.Mutex{},
		}
	}
	return topazAdapters, nil
}

// IsOpen Проверка открытия COM порта
func (c *TopazAdapter) IsOpen() bool {
	return c.port != nil
}

// SendCommand отправляет команду и читает ответ
func (c *TopazAdapter) SendCommand(cmd []byte, resType string) ([]byte, error) {
	if len(cmd) == 0 || cmd[0] != StartByte {
		return nil, fmt.Errorf("%w: command must start with 0x7F and be non-empty", ErrInvalidSendCommand)
	}

	if !c.IsOpen() {
		return nil, fmt.Errorf("%w: port is not open", ErrOpenPort)
	}
	port := c.port

	_, err := port.Write(cmd)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrWriteCommand, err)
	}

	var response bytes.Buffer
	buf := make([]byte, 128)

	for {
		n, err := port.Read(buf)

		if n > 0 {
			response.Write(buf[:n])
			currentResponse := response.Bytes()

			if resType == "Short" {
				if len(currentResponse) == 2 {
					return currentResponse, nil
				}
				if len(currentResponse) > 2 {
					return currentResponse, fmt.Errorf("%w: expected 2 bytes, received %d", ErrUnexpectedResponse, len(currentResponse))
				}
			} else {
				if len(currentResponse) == 2 {
					if currentResponse[1] == ErrByte || currentResponse[1] == ErrByte2 {
						return currentResponse, nil
					}
				}

				if len(currentResponse) >= 3 {
					if currentResponse[len(currentResponse)-2] == ETX &&
						currentResponse[len(currentResponse)-3] == ETX {
						receivedCRC := currentResponse[len(currentResponse)-1]
						calculatedCRC := TOPAZCalculateChecksum(currentResponse[:len(currentResponse)-1])

						if receivedCRC == calculatedCRC {
							return currentResponse, nil
						} else {
							return currentResponse, fmt.Errorf("%w: received %x, calculated %x", ErrCrcMismatch, receivedCRC, calculatedCRC)
						}
					}
				}
			}
		}

		if err != nil {
			if errors.Is(err, ErrReadTimeout) {
				if response.Len() > 0 {
					return response.Bytes(), fmt.Errorf("%w: Timed out after receiving partial data", ErrIncompleteResponse)
				} else {
					return nil, fmt.Errorf("%w: Timed out waiting for response", ErrReadTimeout)
				}
			}
			return response.Bytes(), fmt.Errorf("%w: Read error: %v", ErrReadCommand, err)
		}
	}
}

// Reopen Метод для повторного открытия порта
func (c *TopazAdapter) Reopen() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.IsOpen() {
		err := c.port.Close()
		if err != nil {
			return fmt.Errorf("close port error: %s", err)
		}
		c.port = nil
	}

	// Используем мок для эмуляции
	newPort := NewMockSerialPort([][]byte{
		{0x7F, 0x01, 0x03, 0x00, 0x00, 0x03}, // Пример ответа с корректным CRC
	})

	c.port = newPort

	err := newPort.SetReadTimeout(time.Duration(c.TopazAdapterSettings.ReadTimeout) * time.Second)
	if err != nil {
		newPort.Close()
		c.port = nil
		return fmt.Errorf("set read timeout error: %s", err)
	}

	return nil
}
