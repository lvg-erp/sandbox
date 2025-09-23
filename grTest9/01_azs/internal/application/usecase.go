package application

import (
	"fmt"
	"fuelazs/internal/repository/postgres"

	//"gl.iteco.com/technology/go_general/errproc"
	"fuelazs/config"
	"fuelazs/internal/COMPort"
	drivers "fuelazs/internal/driver"
	"fuelazs/internal/driver/controller"
	"fuelazs/internal/driver/topaz/trk"
	"fuelazs/internal/gui"
	"fuelazs/internal/integration"
	"fuelazs/internal/logger"
	_ "fuelazs/internal/repository/postgres"
	"fuelazs/internal/usecase"
)

type UseCases struct {
	Processing *usecase.Processing
}

type UseCasesDep struct {
	Config       *config.Config
	Logger       *logger.Logger
	QRPort       *COMPort.COMPort
	KazsOperator *integration.KazsOperator
	AppGui       *gui.Gui
	Driver       *drivers.Drivers
	Repository   *postgres.Registry
	//ErrProc      *errproc.ErrProc
	TRKSettings map[string]trk.TRKResponse
	Controller  *controller.ControllerAdapter
}

func NewUseCases(uc UseCasesDep) (*UseCases, error) {
	if uc.Config == nil {
		return nil, fmt.Errorf("nil config.Config")
	}

	processing := usecase.NewProcessing(usecase.ProcessingDep{
		KazsOperator: uc.KazsOperator,
		AppGui:       uc.AppGui,
		Driver:       uc.Driver,
		KazsConfig:   &uc.Config.KazsConfig,
		AppConfig:    &uc.Config.AppConfig,
		DriverConfig: &uc.Config.DriverConfig,
		Repository:   uc.Repository,
		Logger:       uc.Logger,
		ErrProc:      uc.ErrProc,
		TRKSettings:  uc.TRKSettings,
	})

	return &UseCases{
		Processing: processing,
	}, nil

}
