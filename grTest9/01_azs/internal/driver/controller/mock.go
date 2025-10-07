package controller

//import (
//	"fmt"
//	"sync"
//	"time"
//)
//
//// MockSerialPort заглушка для эмуляции COM-порта
//type MockSerialPort struct {
//	responseQueue [][]byte
//	readIndex     int
//	mutex         sync.Mutex
//}
//
//func NewMockSerialPort(responses [][]byte) *MockSerialPort {
//	return &MockSerialPort{
//		responseQueue: responses,
//		readIndex:     0,
//	}
//}
//
//func (m *MockSerialPort) Write(data []byte) (int, error) {
//	if len(data) == 0 {
//		return 0, fmt.Errorf("length of cmd is zero")
//	}
//	return len(data), nil
//}
//
//func (m *MockSerialPort) Read(buf []byte) (int, error) {
//	m.mutex.Lock()
//	defer m.mutex.Unlock()
//
//	if m.readIndex >= len(m.responseQueue) {
//		return 0, fmt.Errorf("no more responses")
//	}
//
//	response := m.responseQueue[m.readIndex]
//	copy(buf, response)
//	m.readIndex++
//	return len(response), nil
//}
//
//func (m *MockSerialPort) Close() error {
//	return nil
//}
//
//func (m *MockSerialPort) Drain() error {
//	return nil
//}
//
//func (m *MockSerialPort) ResetInputBuffer() error {
//	return nil
//}
//
//func (m *MockSerialPort) ResetOutputBuffer() error {
//	return nil
//}
//
//func (m *MockSerialPort) SetReadTimeout(_ time.Duration) error {
//	return nil
//}

//
//func NewMockControllerAdapter() *ControllerAdapter {
//	responses := [][]byte{
//		[]byte("ok\n"),
//		[]byte("pongok\n"),
//		[]byte(`{"din":[{"i":"1","s":"0"}],"dout":[{"i":"1","s":"0"}]}\n`),
//		[]byte(`{"din_set":[{"i":"1","s":"1"}],"dout_set":[{"i":"1","s":"1"}]}\n`),
//		[]byte("ok\n"),
//	}
//	return &ControllerAdapter{
//		Port: NewMockSerialPort(responses),
//		ControllerAdapterSettings: ControllerAdapterSettings{
//			Token: "test_token",
//		},
//		Maping: &ControllerMaping{
//			Controller: Controller{
//				Settings: Settings{
//					IsUSB:       true,
//					COMPort:     1,
//					BaudRate:    9600,
//					StopBits:    1.0,
//					DataBits:    8,
//					Parity:      "none",
//					ReadTimeout: 5,
//					Token:       "test_token",
//				},
//				Pump: map[string]Pump{
//					"1": {Number: "1", Relay: "1", Enable: "1", Disable: "0"},
//					"2": {Number: "2", Relay: "2", Enable: "1", Disable: "0"},
//				},
//				Lock: map[string]Lock{
//					"1": {Number: "1", Relay: "1", Open: "1", Close: "0"},
//					"2": {Number: "2", Relay: "2", Open: "1", Close: "0"},
//				},
//				Doors: map[string]Door{
//					"1": {Number: "1", Relay: "1", Open: "1", Close: "0"},
//					"2": {Number: "2", Relay: "2", Open: "1", Close: "0"},
//				},
//			},
//		},
//	}
//}
