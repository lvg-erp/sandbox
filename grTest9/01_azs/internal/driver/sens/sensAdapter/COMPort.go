package sensAdapter

import (
	"bytes"
	_ "encoding/json"
	"fmt"
	"fuelazs/internal/driver/sens"
	"go.bug.st/serial"
	_ "os"
	_ "path/filepath"
	"strconv"
	"sync"
	"time"
)

// SyncByte константа для синхронизации
const SyncByte = 0xB5

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

// SensAdapterSettings настройки адаптера
type SensAdapterSettings struct {
	PortName    string
	BaudRate    int
	DataBits    int
	Parity      serial.Parity
	StopBits    serial.StopBits
	ReadTimeout int
}

// SensAdapter структура адаптера
type SensAdapter struct {
	Port                SerialPort
	SensAdapterSettings SensAdapterSettings
	BK                  map[string]struct{}
	LC                  map[string]struct{}
	mutex               sync.Mutex
}

// getSENSConnectionInfo читает SENSmaping.json
//func getSENSConnectionInfo() (*SENSMaping, error) {
//	filePath := getFilePath()
//	data, err := os.ReadFile(filePath)
//	if err != nil {
//		return nil, fmt.Errorf("failed to read SENSmaping.json: %w", err)
//	}
//
//	var sensMaping SENSMaping
//	err = json.Unmarshal(data, &sensMaping)
//	if err != nil {
//		return nil, fmt.Errorf("failed to unmarshal SENSmaping.json: %w", err)
//	}
//	return &sensMaping, nil
//}

// NewSensAdapter создает адаптеры
func NewSensAdapter(portFactory func(string, *serial.Mode) (SerialPort, error)) (map[string]*SensAdapter, error) {
	if portFactory == nil {
		portFactory = func(name string, mode *serial.Mode) (SerialPort, error) {
			return serial.Open(name, mode)
		}
	}

	sensInfo, err := getSENSConnectionInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get SENS connection info: %w", err)
	}

	sensAdapters := make(map[string]*SensAdapter, len(sensInfo.LinAdapter))

	for sensKey, sensConfig := range sensInfo.LinAdapter {
		comName := getOSPortName(sensConfig.COMPort, sensInfo.LinAdapter["1"].IsUSB)
		var port SerialPort
		var portErr error

		cachedPort, found := getPortFromCache(comName)
		if found {
			port = cachedPort
		} else {
			port, portErr = portFactory(comName, &serial.Mode{
				BaudRate: sensConfig.BaudRate,
				DataBits: sensConfig.DataBits,
				Parity:   parity(sensConfig.Parity),
				StopBits: stopBits(sensConfig.StopBits),
			})
			if portErr != nil {
				return nil, fmt.Errorf("failed to open serial port %s for LIN %s: %w", comName, sensKey, portErr)
			}

			portErr = port.SetReadTimeout(time.Duration(sensConfig.ReadTimeout) * time.Second)
			if portErr != nil {
				port.Close()
				return nil, fmt.Errorf("failed to set read timeout for port %s for LIN %s: %w", comName, sensKey, portErr)
			}

			addPortToCache(comName, port)
		}

		bkList := make(map[string]struct{}, len(sensConfig.Bk))
		for _, bkNumber := range sensConfig.Bk {
			bkNumberString := strconv.Itoa(bkNumber)
			bkList[bkNumberString] = struct{}{}
		}

		lcList := make(map[string]struct{}, len(sensConfig.LC))
		for _, lcNumber := range sensConfig.LC {
			lcNumberString := strconv.Itoa(lcNumber)
			lcList[lcNumberString] = struct{}{}
		}

		sensAdapters[sensKey] = &SensAdapter{
			Port: port,
			SensAdapterSettings: SensAdapterSettings{
				PortName:    comName,
				BaudRate:    sensConfig.BaudRate,
				DataBits:    sensConfig.DataBits,
				Parity:      parity(sensConfig.Parity),
				StopBits:    stopBits(sensConfig.StopBits),
				ReadTimeout: sensConfig.ReadTimeout,
			},
			BK: bkList,
			LC: lcList,
		}
	}

	return sensAdapters, nil
}

// IsOpen Метод для проверки открытия COM порта
func (c *SensAdapter) IsOpen() bool {
	return c.Port != nil
}

// SendCommand Метод для отправки команды в COM порт
func (c *SensAdapter) SendCommand(cmd []byte) ([]byte, error) {
	if len(cmd) == 0 || cmd[0] != SyncByte {
		return nil, fmt.Errorf("%w: command must start with 0xB5 and be non-empty", ErrInvalidSendCommand)
	}

	// --- Проверка порта ---
	if !c.IsOpen() {
		return nil, fmt.Errorf("%w: port is not open", ErrOpenPort)
	}
	portIO := c.Port

	// --- 1. Отправка команды ---
	_, err := portIO.Write(cmd)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrWriteCommand, err)
	}

	// --- 2. Чтение ответа ---
	var response bytes.Buffer
	buf := make([]byte, 128)
	expectedTotalLength := -1

	for {
		n, err := portIO.Read(buf)

		if n > 0 {
			response.Write(buf[:n])

			currentResponse := response.Bytes()

			if expectedTotalLength == -1 && len(currentResponse) >= 3 {
				if currentResponse[0] != 0xB5 {
					return currentResponse, fmt.Errorf("%w: response does not start with 0xB5 (got %x)", ErrInvalidResponse, currentResponse[0])
				}
				dataLength := int(currentResponse[2])
				expectedTotalLength = 4 + dataLength + 1
			}

			if expectedTotalLength != -1 && len(currentResponse) >= expectedTotalLength {

				responseToCheck := currentResponse[:expectedTotalLength]

				receivedCRC := responseToCheck[expectedTotalLength-1]
				calculatedCRC := sens.SENSCalculateCRC(responseToCheck[:expectedTotalLength-1])

				if receivedCRC == calculatedCRC {
					return responseToCheck, nil
				} else {
					return currentResponse, fmt.Errorf("%w: received 0x%X, calculated 0x%X", ErrReadCRC, receivedCRC, calculatedCRC)
				}
			}
		}

		// --- 3. Обработка ошибок чтения ---
		if err != nil {
			return response.Bytes(), fmt.Errorf("%w: %w", ErrReadCommand, err)
		}

	}
}

// SendCommandWithoutRead Метод для отправки команды в COM порт
func (c *SensAdapter) SendCommandWithoutRead(cmd []byte) error {
	if len(cmd) == 0 || cmd[0] != SyncByte {
		return fmt.Errorf("%w: command must start with 0xB5 and be non-empty", ErrInvalidSendCommand)
	}

	// --- Проверка порта ---
	if !c.IsOpen() {
		return fmt.Errorf("%w: port is not open", ErrOpenPort)
	}
	portIO := c.Port

	// --- 1. Отправка команды ---
	_, err := portIO.Write(cmd)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWriteCommand, err)
	}

	return nil
}

// Clear Метод очистки COM порта

func (c *SensAdapter) Clear() error {
	err := c.Port.Drain()
	if err != nil {
		return fmt.Errorf("failed to drain serial port: %w", err)
	}
	err = c.Port.ResetInputBuffer()
	if err != nil {
		return fmt.Errorf("failed to reset input buffer: %w", err)
	}

	err = c.Port.ResetOutputBuffer()
	if err != nil {
		return fmt.Errorf("failed to reset output buffer: %w", err)
	}

	return nil
}

func (c *SensAdapter) Reopen() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.IsOpen() {
		err := c.Port.Close()
		if err != nil {
			return fmt.Errorf("close port error: %s", err)
		}
		c.Port = nil
	}

	newPort, err := serial.Open(c.SensAdapterSettings.PortName, &serial.Mode{
		BaudRate: c.SensAdapterSettings.BaudRate,
		DataBits: c.SensAdapterSettings.DataBits,
		Parity:   c.SensAdapterSettings.Parity,
		StopBits: c.SensAdapterSettings.StopBits,
	})
	if err != nil {
		return fmt.Errorf("open port error: %s", err)
	}

	c.Port = newPort

	err = newPort.SetReadTimeout(time.Duration(c.SensAdapterSettings.ReadTimeout) * time.Second)
	if err != nil {
		newPort.Close()
		c.Port = nil
		return fmt.Errorf("set read timeout error: %s", err)
	}

	return nil
}
