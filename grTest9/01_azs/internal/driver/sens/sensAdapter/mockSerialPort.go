package sensAdapter

import (
	"fmt"
	"sync"
	"time"
)

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
	if len(data) == 0 || data[0] != SyncByte {
		return 0, fmt.Errorf("invalid command")
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
