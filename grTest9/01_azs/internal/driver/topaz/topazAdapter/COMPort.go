package topazAdapter

import (
	"bytes"
	"errors"
	"fmt"
	"go.bug.st/serial"
	"sync"
	"time"
)

const (
	StartByte = 0x7F
	ErrByte   = 0x15
	ErrByte2  = 0x18

	STX = 0x02
	ETX = 0x03
)

var (
	openedPorts     = make(map[string]serial.Port)
	openedPortsLock sync.Mutex
)

func getPortFromCache(comName string) (serial.Port, bool) {
	openedPortsLock.Lock()
	defer openedPortsLock.Unlock()
	port, found := openedPorts[comName]
	return port, found
}

func addPortToCache(comName string, port serial.Port) {
	openedPortsLock.Lock()
	defer openedPortsLock.Unlock()
	openedPorts[comName] = port
}

type TopazAdapterSettings struct {
	PortName    string
	BaudRate    int
	DataBits    int
	Parity      serial.Parity
	StopBits    serial.StopBits
	ReadTimeout int
}

type TopazAdapter struct {
	port                 serial.Port
	TopazAdapterSettings TopazAdapterSettings
	mutex                sync.Mutex
}

func NewTopazAdapter() (map[string]*TopazAdapter, error) {
	topazInfo, err := getTRKConnectionInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get TRK connection info: %w", err)
	}

	topazAdapters := make(map[string]*TopazAdapter, len(topazInfo.TRK))
	for trkKey, trkConfig := range topazInfo.TRK {
		comName := getOSPortName(trkConfig.COMPort, trkConfig.IsUSB)
		var port serial.Port
		var portErr error

		cachedPort, found := getPortFromCache(comName)
		if found {
			port = cachedPort
		} else {
			port, portErr = serial.Open(comName, &serial.Mode{
				BaudRate: trkConfig.BaudRate,
				DataBits: trkConfig.DataBits,
				Parity:   parity(trkConfig.Parity),
				StopBits: stopBits(trkConfig.StopBits),
			})
			if portErr != nil {
				return nil, fmt.Errorf("failedo to open serial port %s for TRK %s: %w", comName, trkKey, portErr)
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

// // Close Закрытие COM порта
//
//	func (c *COMPort) Close() error {
//		if c.port == nil {
//			return nil
//		}
//
//		err := c.port.Close()
//		c.port = nil
//		if err != nil {
//			return fmt.Errorf("failed to close serial port: %w", err)
//		}
//
//		return nil
//	}
//
// IsOpen Проверка открытия COM порта
func (c *TopazAdapter) IsOpen() bool {
	return c.port != nil
}

// SendCommand отправляет команду и читает ответ
func (c *TopazAdapter) SendCommand(cmd []byte, resType string) ([]byte, error) {

	// TODO: Переделай алгоритм отправки команды в COM порт. Убрать Short, сделать один метод
	// --- Проверка отправляемой команды ---
	if len(cmd) == 0 || cmd[0] != StartByte {
		return nil, fmt.Errorf("%w: command must start with 0x7F and be non-empty", ErrInvalidSendCommand)
	}

	// --- Проверка порта ---
	if !c.IsOpen() {
		return nil, fmt.Errorf("%w: port is not open", ErrOpenPort) // Включаем текст ошибки
	}
	port := c.port

	// --- Запись команды ---
	_, err := port.Write(cmd)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrWriteCommand, err) // Возвращаем оригинальную ошибку записи
	}

	// 2. Чтение ответа с таймаутом
	var response bytes.Buffer
	buf := make([]byte, 128)

	// Цикл чтения: Читаем до таймаута или до определения конца ответа
	for {
		n, err := port.Read(buf)

		if n > 0 {
			response.Write(buf[:n])

			// 3. Проверка завершения ответа после добавления данных
			currentResponse := response.Bytes()

			if resType == "Short" {
				// Для "Short" ждем ровно 2 байта
				if len(currentResponse) == 2 {
					return currentResponse, nil // Получен короткий ответ
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
						calculatedCRC := TOPAZCalculateChecksum(currentResponse[:len(currentResponse)-1]) // Предполагается в пакете driver

						if receivedCRC == calculatedCRC {
							return currentResponse, nil
						} else {
							return currentResponse, fmt.Errorf("%w: received %x, calculated %x", ErrCrcMismatch, receivedCRC, calculatedCRC)
						}
					}
				}
			}
		}

		// 4. Обработка ошибок чтения
		if err != nil {
			// Проверяем, является ли ошибка таймаутом
			if errors.Is(err, ErrReadTimeout) {
				// Таймаут произошел. Если уже получили какие-то данные, это неполный ответ.
				if response.Len() > 0 {
					return response.Bytes(), fmt.Errorf("%w: Timed out after receiving partial data", ErrIncompleteResponse)
				} else {
					// Таймаут до получения каких-либо данных - это ошибка чтения/таймаута
					return nil, fmt.Errorf("%w: Timed out waiting for response", ErrReadTimeout)
				}
			}
			// Обработка других ошибок чтения
			return response.Bytes(), fmt.Errorf("%w: Read error: %v", ErrReadCommand, err)
		}
	}
}

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

	newPort, err := serial.Open(c.TopazAdapterSettings.PortName, &serial.Mode{
		BaudRate: c.TopazAdapterSettings.BaudRate,
		DataBits: c.TopazAdapterSettings.DataBits,
		Parity:   c.TopazAdapterSettings.Parity,
		StopBits: c.TopazAdapterSettings.StopBits,
	})
	if err != nil {
		return fmt.Errorf("open port error: %s", err)
	}

	c.port = newPort

	err = newPort.SetReadTimeout(time.Duration(c.TopazAdapterSettings.ReadTimeout) * time.Second)
	if err != nil {
		newPort.Close()
		c.port = nil
		return fmt.Errorf("set read timeout error: %s", err)
	}

	return nil
}
