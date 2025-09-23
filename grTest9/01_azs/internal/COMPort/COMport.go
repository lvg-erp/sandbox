package COMPort

import (
	"fmt"
	"go.bug.st/serial"
	"sync"
	"time"
)

type COMPort struct {
	Port            serial.Port
	Mu              sync.Mutex
	COMPortSettings COMPortSettings
}

type COMPortSettings struct {
	PortName    string
	BaudRate    int
	DataBits    int
	Parity      string
	StopBits    float32
	ReadTimeout int
}

type COMPortDep struct {
	PortName    string
	BaudRate    int
	DataBits    int
	Parity      string
	StopBits    float32
	ReadTimeout int
}

func NewCOMPort(dep *COMPortDep) (*COMPort, error) {

	port, err := serial.Open(dep.PortName, &serial.Mode{
		BaudRate: dep.BaudRate,
		DataBits: dep.DataBits,
		Parity:   parity(dep.Parity),
		StopBits: stopBits(dep.StopBits),
	})

	if err != nil {
		return nil, fmt.Errorf("open port error: %s", err)
	}

	err = port.SetReadTimeout(time.Duration(dep.ReadTimeout) * time.Second)
	if err != nil {
		port.Close()
		return nil, fmt.Errorf("set read timeout error: %s", err)
	}

	return &COMPort{
		Port: port,
		Mu:   sync.Mutex{},
		COMPortSettings: COMPortSettings{
			BaudRate:    dep.BaudRate,
			DataBits:    dep.DataBits,
			Parity:      dep.Parity,
			StopBits:    dep.StopBits,
			ReadTimeout: dep.ReadTimeout,
			PortName:    dep.PortName,
		},
	}, nil
}

// Reopen Метод для переоткрытия COM порта
func (port *COMPort) Reopen() error {
	port.Mu.Lock()
	defer port.Mu.Unlock()

	err := port.Port.ResetInputBuffer()
	if err != nil {
		return fmt.Errorf("reset input buffer error: %s", err)
	}
	err = port.Port.ResetOutputBuffer()
	if err != nil {
		return fmt.Errorf("reset output buffer error: %s", err)
	}

	err = port.Port.Close()
	if err != nil {
		return fmt.Errorf("close port error: %s", err)
	}

	newPort, err := serial.Open(port.COMPortSettings.PortName, &serial.Mode{
		BaudRate: port.COMPortSettings.BaudRate,
		DataBits: port.COMPortSettings.DataBits,
		Parity:   parity(port.COMPortSettings.Parity),
		StopBits: stopBits(port.COMPortSettings.StopBits),
	})
	if err != nil {
		return fmt.Errorf("open port error: %s", err)
	}
	err = newPort.SetReadTimeout(time.Duration(port.COMPortSettings.ReadTimeout) * time.Second)
	if err != nil {
		newPort.Close()
		return fmt.Errorf("set read timeout error: %s", err)
	}
	port.Port = newPort

	return nil
}

// Clear Метод для очистки COM порта
func (port *COMPort) Clear() error {
	port.Mu.Lock()
	defer port.Mu.Unlock()

	err := port.Port.ResetInputBuffer()
	if err != nil {
		return fmt.Errorf("reset input buffer error: %s", err)
	}

	err = port.Port.ResetOutputBuffer()
	if err != nil {
		return fmt.Errorf("reset output buffer error: %s", err)
	}

	return nil
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

// stopBits Метод для конвертации стоп битов
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
