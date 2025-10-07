package sensAdapter

import (
	"bytes"
	"fmt"
	"sync"
	"time"
)

// MockSerialPort заглушка для эмуляции COM-порта
type MockSerialPort struct {
	responseQueue [][]byte
	readIndex     int
	mutex         sync.Mutex
	lastCommand   []byte
}

func NewMockSerialPort(responses [][]byte) *MockSerialPort {
	responseDevice1 := []byte{0xB5, 0x01, 0x0F, 0x00, 0x00, 0x12, 0x34, 0x00, 0x00, 0xA2} // Устройство №1, CRC=0xA2
	responseDevice2 := []byte{0xB5, 0x02, 0x0F, 0x00, 0x00, 0x12, 0x35, 0x00, 0x00, 0x3D} // Устройство №2, CRC=0x3D

	defaultResponses := [][]byte{
		responseDevice1,
		responseDevice2,
		responseDevice1,
		responseDevice2,
	}
	if len(responses) > 0 {
		defaultResponses = append(defaultResponses, responses...)
	}
	return &MockSerialPort{
		responseQueue: defaultResponses,
		readIndex:     0,
	}
}

func (m *MockSerialPort) Write(data []byte) (int, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.lastCommand = data
	fmt.Printf("Write: %v\n", data)
	return len(data), nil
}

func (m *MockSerialPort) Read(buf []byte) (int, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.responseQueue) == 0 || m.readIndex >= len(m.responseQueue) {
		return 0, fmt.Errorf("no more responses")
	}

	var response []byte
	if bytes.HasPrefix(m.lastCommand, []byte{0xB5, 0x01}) {
		response = m.responseQueue[0] // Устройство №1
	} else if bytes.HasPrefix(m.lastCommand, []byte{0xB5, 0x02}) {
		response = m.responseQueue[1] // Устройство №2
	} else {
		return 0, fmt.Errorf("unknown command address: %v", m.lastCommand)
	}

	fmt.Printf("MockSerialPort Read: %v\n", response)
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
