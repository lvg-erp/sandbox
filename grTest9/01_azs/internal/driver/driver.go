package driver

import (
	"fuelazs/internal/driver/controller"
	"fuelazs/internal/driver/qr"
	"fuelazs/internal/driver/sens/bk/sens_BK_2P"
	"fuelazs/internal/driver/sens/lc/sens_PMP_118_Modbus"
	"fuelazs/internal/driver/sens/lin/sens_LIN_RS_USB_LAN"
	"fuelazs/internal/driver/sens/sensAdapter"
	"fuelazs/internal/driver/topaz/topazAdapter"
	"fuelazs/internal/driver/topaz/trk"
	"go.bug.st/serial"
	"sync"
)

type Drivers struct {
	SensDriver       *SensDriver
	MuSENS           sync.Mutex
	TopazDriver      *TopazDriver
	MuTRK            sync.Mutex
	ControllerDriver *ControllerDriver
	MuController     sync.Mutex
	QRDriver         *QRDriver
	MuQR             sync.Mutex
}

type SensDriver struct {
	Adapter   map[string]*sensAdapter.SensAdapter
	LinDriver *sens_LIN_RS_USB_LAN.LinDriver
	BKDriver  *sens_BK_2P.BK
	LCDriver  *sens_PMP_118_Modbus.LCDriver
}

type TopazDriver struct {
	Adapter     map[string]*topazAdapter.TopazAdapter
	TopazDriver *trk.TRK
}

type ControllerDriver struct {
	Adapter *controller.ControllerAdapter
}

type QRDriver struct {
	Adapter *qr.QRAdapter
}

func NewSensDriver() (*SensDriver, error) {
	// Используем MockSerialPort для эмуляции порта
	portFactory := func(_ string, _ *serial.Mode) (sensAdapter.SerialPort, error) {
		return sensAdapter.NewMockSerialPort([][]byte{
			{0xB5, 0x01, 0x03, 0x00, 0x00, 0x00, 0xB5}, // Пример ответа с корректным CRC
		}), nil
	}

	adapter, err := sensAdapter.NewSensAdapter(portFactory)
	if err != nil {
		return nil, err
	}

	return &SensDriver{
		Adapter:   adapter,
		LinDriver: sens_LIN_RS_USB_LAN.NewLinDriver(&adapter),
		BKDriver:  sens_BK_2P.NewBK(&adapter),
		LCDriver:  sens_PMP_118_Modbus.NewLCDriver(&adapter),
	}, nil
}

func NewTopazDriver() (*TopazDriver, error) {
	adapter, err := topazAdapter.NewTopazAdapter(nil)
	if err != nil {
		return nil, err
	}

	return &TopazDriver{
		Adapter:     adapter,
		TopazDriver: trk.NewTRK(&adapter),
	}, nil
}

func NewControllerDriver() (*ControllerDriver, error) {
	adapter, err := controller.NewControllerAdapter(nil)
	if err != nil {
		return nil, err
	}

	return &ControllerDriver{
		Adapter: adapter,
	}, nil
}

func NewQRDriver() (*QRDriver, error) {
	adapter, err := qr.NewQRAdapter(nil)
	if err != nil {
		return nil, err
	}

	return &QRDriver{
		Adapter: adapter,
	}, nil
}
