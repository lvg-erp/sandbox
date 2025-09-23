package application

import (
	"fuelazs/internal/driver"
	"sync"
)

func NewDrivers() (*driver.Drivers, error) {
	sensDriver, err := driver.NewSensDriver()
	if err != nil {
		return nil, err
	}

	topazDriver, err := driver.NewTopazDriver()
	if err != nil {
		return nil, err
	}

	controllerDriver, err := driver.NewControllerDriver()
	if err != nil {
		return nil, err
	}

	qrDriver, err := driver.NewQRDriver()
	if err != nil {
		return nil, err
	}

	return &driver.Drivers{
		SensDriver:       sensDriver,
		MuSENS:           sync.Mutex{},
		TopazDriver:      topazDriver,
		MuTRK:            sync.Mutex{},
		ControllerDriver: controllerDriver,
		MuController:     sync.Mutex{},
		QRDriver:         qrDriver,
		MuQR:             sync.Mutex{},
	}, nil
}
