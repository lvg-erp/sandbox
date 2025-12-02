package sensAdapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fuelazs/internal/driver/sens"
	"go.bug.st/serial"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	SyncByte = 0xB5
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

// FilePort имитирует serial.Port для чтения из JSON
type FilePort struct {
	data       [][]byte
	currentIdx int
	filename   string
	resetCount int
	maxResets  int
	deviceNum  string
	config     DeviceConfig
}

func (fp *FilePort) Read(b []byte) (int, error) {
	if fp.currentIdx >= len(fp.data) {
		if fp.resetCount >= fp.maxResets {
			fmt.Printf("FilePort: max reset attempts reached for %s\n", fp.filename)
			return 0, fmt.Errorf("max reset attempts reached for %s", fp.filename)
		}
		fmt.Printf("FilePort: reached end of data in %s, resetting (attempt %d/%d)\n", fp.filename, fp.resetCount+1, fp.maxResets)
		fp.currentIdx = 0
		fp.resetCount++
	}
	if len(fp.data) == 0 {
		fmt.Printf("FilePort: no data available in %s\n", fp.filename)
		return 0, fmt.Errorf("no data available in %s", fp.filename)
	}
	data := fp.data[fp.currentIdx]
	fmt.Printf("FilePort: reading data from %s for device %s: %X\n", fp.filename, fp.deviceNum, data)
	n := copy(b, data)
	fp.currentIdx++
	return n, nil
}

func (fp *FilePort) Write(b []byte) (int, error) {
	fmt.Printf("FilePort: processing write to %s: %X\n", fp.filename, b)
	// Извлекаем команду и формируем соответствующий ответ
	if len(b) < 3 || b[0] != SyncByte {
		fmt.Printf("FilePort: invalid command format: %X\n", b)
		return len(b), nil
	}
	cmd := b[2]
	// Формируем ответ в зависимости от команды
	var response []byte
	switch cmd {
	case 0x0F: // Команда проверки (cmd=15)
		addrInt, _ := strconv.Atoi(fp.deviceNum)
		if addrInt == 1 {
			response = []byte{0xB5, 0x01, 0x04, 0x15, 0x30, 0x00, 0x00, 0x00} // LIN 1: ID=0x3000
			crc := sens.SENSCalculateCRC(response)
			response = append(response, crc)
			fmt.Printf("FilePort: LIN 1 response ID=0x3000, calculated CRC: %02X, full response: %X\n", crc, response)
		} else if addrInt == 2 {
			response = []byte{0xB5, 0x02, 0x04, 0x15, 0x30, 0x02, 0x00, 0x00} // LIN 2: ID=0x3002
			crc := sens.SENSCalculateCRC(response)
			response = append(response, crc)
			fmt.Printf("FilePort: LIN 2 response ID=0x3002, calculated CRC: %02X, full response: %X\n", crc, response)
		} else {
			response = []byte{0xB5, byte(addrInt), 0x04, 0x15, 0x00, 0x00, 0x00, 0x00}
			crc := sens.SENSCalculateCRC(response)
			response = append(response, crc)
			fmt.Printf("FilePort: default response for addr %d, calculated CRC: %02X, full response: %X\n", addrInt, crc, response)
		}
	case 0x41: // Команда чтения объёма
		addrInt, _ := strconv.Atoi(fp.deviceNum)
		if addrInt == 1 {
			response = []byte{0xB5, 0x01, 0x04, 0x41, 0x03, 0xE8, 0x00, 0x00} // LIN 1: 1000 литров (0x03E8)
		} else if addrInt == 2 {
			response = []byte{0xB5, 0x02, 0x04, 0x41, 0x05, 0xDC, 0x00, 0x00} // LIN 2: 1500 литров (0x05DC)
		} else {
			response = []byte{0xB5, byte(addrInt), 0x04, 0x41, 0x00, 0x00, 0x00, 0x00}
		}
		crc := sens.SENSCalculateCRC(response)
		response = append(response, crc)
		fmt.Printf("FilePort: calculated CRC for response %X: %02X, full response: %X\n", response[:len(response)-1], crc, response)
	default:
		// Для других команд возвращаем данные из JSON
		for _, param := range fp.config.SettingsParameters {
			if fmt.Sprintf("0x%02X", cmd) == fmt.Sprintf("0x%02X", cmd) {
				addrInt, _ := strconv.Atoi(fp.deviceNum)
				value := int(param.Value * 10)
				hexStr := fmt.Sprintf("B5%02X%04X", addrInt, value)
				data, err := hexToBytes(hexStr)
				if err != nil {
					fmt.Printf("FilePort: failed to convert %s to hex: %v\n", hexStr, err)
					return len(b), nil
				}
				crc := sens.SENSCalculateCRC(data)
				response = append(data, crc)
				fmt.Printf("FilePort: calculated CRC for response %X: %02X, full response: %X\n", data, crc, response)
				break
			}
		}
	}
	if len(response) > 0 {
		fp.data = [][]byte{response}
		fp.currentIdx = 0
		fmt.Printf("FilePort: prepared response for %s: %X\n", fp.filename, response)
	}
	return len(b), nil
}

func (fp *FilePort) Close() error {
	fmt.Printf("FilePort: closing %s\n", fp.filename)
	return nil
}

func (fp *FilePort) SetReadTimeout(t time.Duration) error {
	return nil
}

func (fp *FilePort) Drain() error {
	return nil
}

func (fp *FilePort) ResetInputBuffer() error {
	fmt.Printf("FilePort: resetting data index for %s\n", fp.filename)
	fp.currentIdx = 0
	return nil
}

func (fp *FilePort) ResetOutputBuffer() error {
	return nil
}

func (fp *FilePort) Break(t time.Duration) error {
	fmt.Printf("FilePort: break signal for %s\n", fp.filename)
	return nil
}

func (fp *FilePort) GetModemStatusBits() (*serial.ModemStatusBits, error) {
	fmt.Printf("FilePort: getting modem status bits for %s\n", fp.filename)
	return &serial.ModemStatusBits{}, nil
}

func (fp *FilePort) SetDTR(state bool) error {
	fmt.Printf("FilePort: setting DTR to %v for %s\n", state, fp.filename)
	return nil
}

func (fp *FilePort) SetMode(mode *serial.Mode) error {
	fmt.Printf("FilePort: setting mode for %s\n", fp.filename)
	return nil
}

func (fp *FilePort) SetRTS(state bool) error {
	fmt.Printf("FilePort: setting RTS to %v for %s\n", state, fp.filename)
	return nil
}

// hexToBytes преобразует строку в формате hex в байты
func hexToBytes(hexStr string) ([]byte, error) {
	hexStr = strings.TrimSpace(hexStr)
	if len(hexStr) == 0 {
		return nil, nil
	}
	hexStr = strings.ReplaceAll(hexStr, " ", "")
	hexStr = strings.ReplaceAll(hexStr, `\x`, "")
	for _, c := range hexStr {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')) {
			return nil, fmt.Errorf("invalid hex char in string: %s", hexStr)
		}
	}
	if len(hexStr)%2 != 0 {
		return nil, fmt.Errorf("invalid hex string length: %d", len(hexStr))
	}
	data := make([]byte, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		b, err := strconv.ParseUint(hexStr[i:i+2], 16, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid hex char at position %d: %w", i, err)
		}
		data[i/2] = byte(b)
	}
	return data, nil
}

type SensAdapterSettings struct {
	PortName    string
	BaudRate    int
	DataBits    int
	Parity      serial.Parity
	StopBits    serial.StopBits
	ReadTimeout int
}

type SensAdapter struct {
	Port                serial.Port
	SensAdapterSettings SensAdapterSettings
	BK                  map[string]struct{}
	LC                  map[string]struct{}
	mutex               sync.Mutex
}

type DeviceConfig struct {
	SettingsParameters map[string]struct {
		Comment string  `json:"Comment"`
		Value   float64 `json:"Value"`
	} `json:"SettingsParameters"`
	DriverName    string `json:"DriverName"`
	DevicePNumber string `json:"DevicePNumber"`
}

func loadDeviceConfigs(fileName string) ([]DeviceConfig, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file %s: %w", fileName, err)
	}
	var configs []DeviceConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON from %s: %w", fileName, err)
	}
	return configs, nil
}

func NewSensAdapter() (map[string]*SensAdapter, error) {
	sensInfo, err := getSENSConnectionInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get SENS connection info: %w", err)
	}

	configs, err := loadDeviceConfigs("./internal/driver/sens/sensAdapter/device_config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load device configs: %w", err)
	}

	sensAdapters := make(map[string]*SensAdapter, len(sensInfo.LinAdapter))

	for sensKey, sensConfig := range sensInfo.LinAdapter {
		comName := getOSPortName(sensConfig.COMPort, sensInfo.LinAdapter["1"].IsUSB)
		fmt.Printf("NewSensAdapter: opening port for LIN %s, comName: %s\n", sensKey, comName)
		var port serial.Port
		var portErr error

		cachedPort, found := getPortFromCache(comName)
		//spew.Dump(sensConfig)
		if found {
			fmt.Printf("NewSensAdapter: using cached port for %s\n", comName)
			port = cachedPort
		} else {
			port, portErr = serial.Open(comName, &serial.Mode{
				BaudRate: sensConfig.BaudRate,
				DataBits: sensConfig.DataBits,
				Parity:   parity(sensConfig.Parity),
				StopBits: stopBits(sensConfig.StopBits),
			})
			if portErr != nil {
				var deviceData [][]byte
				var config DeviceConfig
				for _, cfg := range configs {
					if cfg.DevicePNumber == sensKey {
						config = cfg
						for _, param := range cfg.SettingsParameters {
							addrInt, _ := strconv.Atoi(sensKey)
							hexStr := fmt.Sprintf("B5%02X%04X", addrInt, int(param.Value*10))
							data, err := hexToBytes(hexStr)
							if err != nil {
								return nil, fmt.Errorf("failed to convert %s to hex for LIN %s: %w", hexStr, sensKey, err)
							}
							data = append(data, sens.SENSCalculateCRC(data))
							deviceData = append(deviceData, data)
						}
						break
					}
				}
				if len(deviceData) == 0 {
					return nil, fmt.Errorf("no config found for LIN %s", sensKey)
				}
				absFileName, err := filepath.Abs(fmt.Sprintf("device_config_%s.json", sensKey))
				if err != nil {
					return nil, fmt.Errorf("failed to get absolute path for LIN %s: %w", sensKey, err)
				}
				fmt.Printf("NewSensAdapter: failed to open serial port %s, using JSON config for LIN %s (file: %s)\n", comName, sensKey, absFileName)
				port = &FilePort{
					data:      deviceData,
					filename:  absFileName,
					maxResets: 3,
					deviceNum: sensKey,
					config:    config,
				}
			} else {
				portErr = port.SetReadTimeout(time.Duration(sensConfig.ReadTimeout) * time.Second)
				if portErr != nil {
					port.Close()
					return nil, fmt.Errorf("failed to set read timeout for port %s for LIN %s: %w", comName, sensKey, portErr)
				}
				addPortToCache(comName, port)
			}
		}

		bkList := make(map[string]struct{}, len(sensConfig.Bk))
		for _, bkNumber := range sensConfig.Bk {
			bkNumberString := strconv.Itoa(bkNumber)
			bkList[bkNumberString] = struct{}{}
		}
		//spew.Dump(bkList)
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
			BK:    bkList,
			LC:    lcList,
			mutex: sync.Mutex{},
		}
	}

	//spew.Dump(sensAdapters)

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

	if !c.IsOpen() {
		return nil, fmt.Errorf("%w: port is not open", ErrOpenPort)
	}
	portIO := c.Port

	fmt.Printf("SendCommand: sending command: %X\n", cmd)
	_, err := portIO.Write(cmd)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrWriteCommand, err)
	}

	var response bytes.Buffer
	buf := make([]byte, 128)
	expectedTotalLength := -1
	timeout := time.After(time.Duration(c.SensAdapterSettings.ReadTimeout+5) * time.Second)

	for {
		select {
		case <-timeout:
			fmt.Printf("SendCommand: timeout waiting for response, current response: %X\n", response.Bytes())
			return response.Bytes(), fmt.Errorf("%w: timeout waiting for response", ErrReadCommand)
		default:
			n, err := portIO.Read(buf)
			if n > 0 {
				response.Write(buf[:n])
				currentResponse := response.Bytes()
				fmt.Printf("SendCommand: received %d bytes: %X\n", n, currentResponse)
				if len(currentResponse) == 0 {
					fmt.Printf("SendCommand: empty response received\n")
					continue
				}
				if expectedTotalLength == -1 && len(currentResponse) >= 3 {
					if currentResponse[0] != 0xB5 {
						fmt.Printf("SendCommand: invalid response format: does not start with 0xB5 (got %X)\n", currentResponse[0])
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
						fmt.Printf("SendCommand: valid response received: %X\n", responseToCheck)
						return responseToCheck, nil
					} else {
						fmt.Printf("SendCommand: CRC mismatch, received: %X, calculated: %X\n", receivedCRC, calculatedCRC)
						return responseToCheck, fmt.Errorf("%w: received 0x%X, calculated 0x%X", ErrReadCRC, receivedCRC, calculatedCRC)
					}
				}
			}
			if err != nil {
				fmt.Printf("SendCommand: read error: %v\n", err)
				return response.Bytes(), fmt.Errorf("%w: %w", ErrReadCommand, err)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// SendCommandWithoutRead Метод для отправки команды в COM порт
func (c *SensAdapter) SendCommandWithoutRead(cmd []byte) error {
	if len(cmd) == 0 || cmd[0] != SyncByte {
		return fmt.Errorf("%w: command must start with 0xB5 and be non-empty", ErrInvalidSendCommand)
	}
	if !c.IsOpen() {
		return fmt.Errorf("%w: port is not open", ErrOpenPort)
	}
	portIO := c.Port
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

// Reopen Метод для повторного открытия порта
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
		configs, err := loadDeviceConfigs("./internal/driver/sens/sensAdapter/device_config.json")
		if err != nil {
			return fmt.Errorf("failed to load device configs: %w", err)
		}
		var deviceData [][]byte
		var config DeviceConfig
		for _, cfg := range configs {
			if cfg.DevicePNumber == strings.TrimPrefix(c.SensAdapterSettings.PortName, "/dev/ttyS") {
				config = cfg
				for _, param := range cfg.SettingsParameters {
					addrInt, _ := strconv.Atoi(strings.TrimPrefix(c.SensAdapterSettings.PortName, "/dev/ttyS"))
					hexStr := fmt.Sprintf("B5%02X%04X", addrInt, int(param.Value*10))
					data, err := hexToBytes(hexStr)
					if err != nil {
						return fmt.Errorf("failed to convert %s to hex: %w", hexStr, err)
					}
					data = append(data, sens.SENSCalculateCRC(data))
					deviceData = append(deviceData, data)
				}
				break
			}
		}
		if len(deviceData) == 0 {
			return fmt.Errorf("no config found for port %s", c.SensAdapterSettings.PortName)
		}
		absFileName, err := filepath.Abs(fmt.Sprintf("device_config_%s.json", strings.TrimPrefix(c.SensAdapterSettings.PortName, "/dev/")))
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}
		fmt.Printf("Reopen: failed to open serial port %s, using JSON config (file: %s)\n", c.SensAdapterSettings.PortName, absFileName)
		newPort = &FilePort{
			data:      deviceData,
			filename:  absFileName,
			maxResets: 3,
			deviceNum: strings.TrimPrefix(c.SensAdapterSettings.PortName, "/dev/ttyS"),
			config:    config,
		}
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
