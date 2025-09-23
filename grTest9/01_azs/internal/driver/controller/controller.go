package controller

import (
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

type ControllerAdapter struct {
	Port                      serial.Port
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

func NewControllerAdapter() (*ControllerAdapter, error) {
	controllerInfo, err := getControllerConnectionInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get controller connection info: %w", err)
	}

	comName := getOSPortName(controllerInfo.Controller.Settings.COMPort, controllerInfo.Controller.Settings.IsUSB)

	Port, err := serial.Open(comName, &serial.Mode{
		BaudRate: controllerInfo.Controller.Settings.BaudRate,
		DataBits: controllerInfo.Controller.Settings.DataBits,
		Parity:   parity(controllerInfo.Controller.Settings.Parity),
		StopBits: stopBits(controllerInfo.Controller.Settings.StopBits),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to open serial Port: %w", err)
	}

	err = Port.SetReadTimeout(time.Duration(controllerInfo.Controller.Settings.ReadTimeout) * time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to set read timeout: %w", err)
	}

	return &ControllerAdapter{
		Port: Port,
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
	}, nil

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
		return nil, fmt.Errorf("Port is not open")
	}
	PortIO := c.Port

	// 1. Отправка команды
	_, err := PortIO.Write(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	return nil, nil
}

func (c *ControllerAdapter) Reopen() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.Port.Close()
	if err != nil {
		return fmt.Errorf("close port error: %s", err)
	}

	newPort, err := serial.Open(c.ControllerAdapterSettings.PortName, &serial.Mode{
		BaudRate: c.ControllerAdapterSettings.BaudRate,
		DataBits: c.ControllerAdapterSettings.DataBits,
		Parity:   c.ControllerAdapterSettings.Parity,
		StopBits: c.ControllerAdapterSettings.StopBits,
	})
	if err != nil {
		return fmt.Errorf("open port error: %s", err)
	}

	err = newPort.SetReadTimeout(time.Duration(c.ControllerAdapterSettings.ReadTimeout) * time.Second)
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
	_, err := c.SendCommand([]byte(c.ControllerAdapterSettings.Token))
	if err != nil {
		return fmt.Errorf("failed to verify token: %w", err)
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (c *ControllerAdapter) Ping() error {
	_, err := c.SendCommand([]byte(PingCmd))
	if err != nil {
		return fmt.Errorf("failed to ping: %w", err)
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (c *ControllerAdapter) SetDefaultValues() error {
	_, err := c.SendCommand([]byte("dout_set 1:0,2:0,3:0,4:0,5:0,6:0,7:0,8:0"))
	if err != nil {
		return fmt.Errorf("failed to set default values: %w", err)
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (c *ControllerAdapter) OpenLock(jarNumber string) error {
	cmd := fmt.Sprintf("dout_set %s:%s", c.Maping.Controller.Lock[jarNumber].Number, c.Maping.Controller.Lock[jarNumber].Open)

	_, err := c.SendCommand([]byte(cmd))
	if err != nil {
		return fmt.Errorf("failed to open lock: %w", err)
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (c *ControllerAdapter) CloseLock(jarNumber string) error {
	cmd := fmt.Sprintf("dout_set %s:%s", c.Maping.Controller.Lock[jarNumber].Number, c.Maping.Controller.Lock[jarNumber].Close)

	_, err := c.SendCommand([]byte(cmd))
	if err != nil {
		return fmt.Errorf("failed to close lock: %w", err)
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (c *ControllerAdapter) EnablePump(jarNumber string) error {
	cmd := fmt.Sprintf("dout_set %s:%s", c.Maping.Controller.Pump[jarNumber].Number, c.Maping.Controller.Pump[jarNumber].Enable)

	_, err := c.SendCommand([]byte(cmd))
	if err != nil {
		return fmt.Errorf("failed to enable pump: %w", err)
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (c *ControllerAdapter) DisablePump(jarNumber string) error {
	cmd := fmt.Sprintf("dout_set %s:%s", c.Maping.Controller.Pump[jarNumber].Number, c.Maping.Controller.Pump[jarNumber].Disable)

	_, err := c.SendCommand([]byte(cmd))
	if err != nil {
		return fmt.Errorf("failed to disable pump: %w", err)
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (c *ControllerAdapter) GetDin() error {
	_, err := c.SendCommand([]byte("din"))
	if err != nil {
		return fmt.Errorf("failed to get din: %w", err)
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (c *ControllerAdapter) GetDout() error {
	_, err := c.SendCommand([]byte("dout"))
	if err != nil {
		return fmt.Errorf("failed to get dout: %w", err)
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}
