package controller

import (
	"bytes"
	"fmt"
	"go.bug.st/serial"
	"sync"
	"time"
)

var (
	PingCmd = "ping"
	PongRes = "pong"
	DinCmd  = "din"
	DoutCmd = "dout"
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
		return 0, fmt.Errorf("length of cmd is zero")
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

type ControllerAdapter struct {
	Port                      SerialPort
	ControllerAdapterSettings ControllerAdapterSettings
	mutex                     sync.Mutex
	Maping                    *ControllerMaping
}

type ControllerAdapterSettings struct {
	PortName    string
	BaudRate    int
	DataBits    int
	Parity      serial.Parity
	StopBits    serial.StopBits
	ReadTimeout int
	Token       string
}

// NewControllerAdapter создает адаптер контроллера
func NewControllerAdapter(portFactory func(string, *serial.Mode) (SerialPort, error)) (*ControllerAdapter, error) {
	if portFactory == nil {
		portFactory = func(_ string, _ *serial.Mode) (SerialPort, error) {
			// Эмуляция порта по умолчанию
			return NewMockSerialPort([][]byte{
				[]byte("ok"),   // Ответ на команду токена
				[]byte("pong"), // Ответ на ping
				[]byte("ok"),   // Ответ на другие команды
			}), nil
		}
	}
	controllerInfo, err := getControllerConnectionInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get controller connection info: %w", err)
	}
	fmt.Println("This controller info")
	fmt.Println(controllerInfo)
	comName := getOSPortName(controllerInfo.Controller.Settings.COMPort, controllerInfo.Controller.Settings.IsUSB)

	port, err := portFactory(comName, &serial.Mode{
		BaudRate: controllerInfo.Controller.Settings.BaudRate,
		DataBits: controllerInfo.Controller.Settings.DataBits,
		Parity:   parity(controllerInfo.Controller.Settings.Parity),
		StopBits: stopBits(controllerInfo.Controller.Settings.StopBits),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to open serial port: %w", err)
	}

	err = port.SetReadTimeout(time.Duration(controllerInfo.Controller.Settings.ReadTimeout) * time.Second)
	if err != nil {
		port.Close()
		return nil, fmt.Errorf("failed to set read timeout: %w", err)
	}

	adapter := &ControllerAdapter{
		Port: port,
		ControllerAdapterSettings: ControllerAdapterSettings{
			PortName:    comName,
			BaudRate:    controllerInfo.Controller.Settings.BaudRate,
			DataBits:    controllerInfo.Controller.Settings.DataBits,
			Parity:      parity(controllerInfo.Controller.Settings.Parity),
			StopBits:    stopBits(controllerInfo.Controller.Settings.StopBits),
			ReadTimeout: controllerInfo.Controller.Settings.ReadTimeout,
			Token:       controllerInfo.Controller.Settings.Token,
		},
		Maping: controllerInfo,
		mutex:  sync.Mutex{},
	}

	// Проверка токена
	err = adapter.Verify()
	if err != nil {
		port.Close()
		return nil, fmt.Errorf("failed to verify controller: %w", err)
	}

	return adapter, nil
}

func (c *ControllerAdapter) IsOpen() bool {
	return c.Port != nil
}

func (c *ControllerAdapter) SendCommand(cmd []byte) ([]byte, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if len(cmd) == 0 {
		return nil, fmt.Errorf("length of cmd is zero")
	}

	if !c.IsOpen() {
		return nil, fmt.Errorf("port is not open")
	}
	portIO := c.Port

	_, err := portIO.Write(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	// Чтение ответа
	var response bytes.Buffer
	buf := make([]byte, 128)
	for {
		n, err := portIO.Read(buf)
		if n > 0 {
			response.Write(buf[:n])
			return response.Bytes(), nil
		}
		if err != nil {
			return response.Bytes(), fmt.Errorf("failed to read response: %w", err)
		}
	}
}

func (c *ControllerAdapter) Reopen() error {
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
		[]byte("pong"), // Ответ на ping
		[]byte("ok"),   // Ответ на другие команды
	})

	err := newPort.SetReadTimeout(time.Duration(c.ControllerAdapterSettings.ReadTimeout) * time.Second)
	if err != nil {
		newPort.Close()
		return fmt.Errorf("set read timeout error: %s", err)
	}

	c.Port = newPort

	_, err = c.Port.Write([]byte(c.ControllerAdapterSettings.Token))
	if err != nil {
		return fmt.Errorf("failed to verify controller: %w", err)
	}

	return nil
}

func (c *ControllerAdapter) Verify() error {
	resp, err := c.SendCommand([]byte(c.ControllerAdapterSettings.Token))
	if err != nil {
		return fmt.Errorf("failed to verify token: %w", err)
	}
	if string(resp) != "ok" {
		return fmt.Errorf("invalid token response: %s", resp)
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (c *ControllerAdapter) Ping() error {
	resp, err := c.SendCommand([]byte(PingCmd))
	if err != nil {
		return fmt.Errorf("failed to ping: %w", err)
	}
	if string(resp) != PongRes {
		return fmt.Errorf("invalid ping response: %s", resp)
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (c *ControllerAdapter) SetDefaultValues() error {
	resp, err := c.SendCommand([]byte("dout_set 1:0,2:0,3:0,4:0,5:0,6:0,7:0,8:0"))
	if err != nil {
		return fmt.Errorf("failed to set default values: %w", err)
	}
	if string(resp) != "ok" {
		return fmt.Errorf("invalid set default values response: %s", resp)
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (c *ControllerAdapter) OpenLock(jarNumber string) error {
	cmd := fmt.Sprintf("dout_set %s:%s", c.Maping.Controller.Lock[jarNumber].Number, c.Maping.Controller.Lock[jarNumber].Open)

	resp, err := c.SendCommand([]byte(cmd))
	if err != nil {
		return fmt.Errorf("failed to open lock: %w", err)
	}
	if string(resp) != "ok" {
		return fmt.Errorf("invalid open lock response: %s", resp)
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (c *ControllerAdapter) CloseLock(jarNumber string) error {
	cmd := fmt.Sprintf("dout_set %s:%s", c.Maping.Controller.Lock[jarNumber].Number, c.Maping.Controller.Lock[jarNumber].Close)

	resp, err := c.SendCommand([]byte(cmd))
	if err != nil {
		return fmt.Errorf("failed to close lock: %w", err)
	}
	if string(resp) != "ok" {
		return fmt.Errorf("invalid close lock response: %s", resp)
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (c *ControllerAdapter) EnablePump(jarNumber string) error {
	cmd := fmt.Sprintf("dout_set %s:%s", c.Maping.Controller.Pump[jarNumber].Number, c.Maping.Controller.Pump[jarNumber].Enable)

	resp, err := c.SendCommand([]byte(cmd))
	if err != nil {
		return fmt.Errorf("failed to enable pump: %w", err)
	}
	if string(resp) != "ok" {
		return fmt.Errorf("invalid enable pump response: %s", resp)
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (c *ControllerAdapter) DisablePump(jarNumber string) error {
	cmd := fmt.Sprintf("dout_set %s:%s", c.Maping.Controller.Pump[jarNumber].Number, c.Maping.Controller.Pump[jarNumber].Disable)

	resp, err := c.SendCommand([]byte(cmd))
	if err != nil {
		return fmt.Errorf("failed to disable pump: %w", err)
	}
	if string(resp) != "ok" {
		return fmt.Errorf("invalid disable pump response: %s", resp)
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (c *ControllerAdapter) GetDin() error {
	resp, err := c.SendCommand([]byte("din"))
	if err != nil {
		return fmt.Errorf("failed to get din: %w", err)
	}
	if string(resp) != "ok" {
		return fmt.Errorf("invalid get din response: %s", resp)
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (c *ControllerAdapter) GetDout() error {
	resp, err := c.SendCommand([]byte("dout"))
	if err != nil {
		return fmt.Errorf("failed to get dout: %w", err)
	}
	if string(resp) != "ok" {
		return fmt.Errorf("invalid get dout response: %s", resp)
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}
