package usecase

import (
	"database/sql"
	"errors"
	"fmt"
	//"gl.iteco.com/technology/go_general/errproc"
	"fuelazs/config"
	"fuelazs/internal/driver"
	"fuelazs/internal/driver/topaz/trk"
	"fuelazs/internal/gui"
	"fuelazs/internal/integration"
	"fuelazs/internal/logger"
	"fuelazs/internal/repository/postgres"
	"fuelazs/internal/usecase/models"
	"strconv"
	"sync"
	"time"
)

type Processing struct {
	kazsActivation      bool
	startProgramm       bool
	oneJarActiveProcess bool
	twoJarActiveProcess bool
	qrActive            bool
	mu                  sync.RWMutex
	kazsOperator        *integration.KazsOperator
	driver              *driver.Drivers
	kazsConfig          *config.KazsConfig
	trkSettings         map[string]trk.TRKResponse
	driverConfig        *config.DriverConfig
	appConfig           *config.AppConfig
	repository          *postgres.Registry
	logger              *logger.Logger
	//errproc             *errproc.BaseErrProc

	appGui *gui.Gui

	qrStopChan   chan struct{}
	qrDataChan   chan models.ScannerResponse
	qrActiveChan chan bool

	telemetryBuffer []byte

	lastTRKTelemetry      map[string]TRKStatus
	lastTRKTelemetryMutex sync.Mutex

	lastSENSTelemetry      map[string]SENSStatus
	lastSENSTelemetryMutex sync.Mutex

	lastTempTelemetry      map[string]TempStatus
	lastTempTelemetryMutex sync.Mutex

	lastControllerDinTelemetry      map[string]RelayStatus
	lastControllerDinTelemetryMutex sync.Mutex

	lastControllerDoutTelemetry      map[string]RelayStatus
	lastControllerDoutTelemetryMutex sync.Mutex
}

func (p *Processing) UpdateLastControllerDinTelemetry(idx int, state string) {
	p.lastControllerDinTelemetryMutex.Lock()
	defer p.lastControllerDinTelemetryMutex.Unlock()

	idxStr := strconv.Itoa(idx)

	p.lastControllerDinTelemetry[idxStr] = RelayStatus{
		timeAS: time.Now().Unix(),
		status: state,
	}
}

func (p *Processing) UpdateLastControllerDoutTelemetry(idx int, state string) {
	p.lastControllerDoutTelemetryMutex.Lock()
	defer p.lastControllerDoutTelemetryMutex.Unlock()

	idxStr := strconv.Itoa(idx)

	p.lastControllerDoutTelemetry[idxStr] = RelayStatus{
		timeAS: time.Now().Unix(),
		status: state,
	}
}

func (p *Processing) UpdateLastTRKTelemetry(jarNumber string, status int) {
	p.lastTRKTelemetryMutex.Lock()
	defer p.lastTRKTelemetryMutex.Unlock()

	p.lastTRKTelemetry[jarNumber] = TRKStatus{
		timeAS: time.Now().Unix(),
		status: status,
	}
}

func (p *Processing) UpdateLastSENSTelemetry(jarNumber string, status SENSStatus) {
	p.lastSENSTelemetryMutex.Lock()
	defer p.lastSENSTelemetryMutex.Unlock()

	p.lastSENSTelemetry[jarNumber] = status
}

func (p *Processing) UpdateLastTempTelemetry(jarNumber string, status TempStatus) {
	p.lastTempTelemetryMutex.Lock()
	defer p.lastTempTelemetryMutex.Unlock()

	p.lastTempTelemetry[jarNumber] = status
}

type RelayStatus struct {
	timeAS int64
	status string
}

type TRKStatus struct {
	timeAS int64
	status int
}

type SENSStatus struct {
	timeAS int64
	H      float32
	T      float32
	Pr     float32
	U      float32
	G      float32
	R      float32
	U1     float32
	H2     float32
	Ut     float32
	Rt     float32
	Ri     float32
	Tr     float32
	U2     float32
	Dg     float32
	Ts     float32
}

type TempStatus struct {
	timeAS int64
	nt     string
}

type ProcessingDep struct {
	KazsOperator *integration.KazsOperator
	AppGui       *gui.Gui
	Driver       *driver.Drivers
	KazsConfig   *config.KazsConfig
	AppConfig    *config.AppConfig
	DriverConfig *config.DriverConfig
	Repository   *postgres.Registry
	Logger       *logger.Logger
	//ErrProc      *errproc.ErrProc
	TRKSettings map[string]trk.TRKResponse
}

func NewProcessing(pd ProcessingDep) *Processing {
	//baseErProc, err := pd.ErrProc.NewBaseErrProc(pd.KazsOperator.KazsID)
	//if err != nil {
	//	pd.Logger.Info("error init base err proc", "err", err)
	//}

	p := &Processing{
		kazsActivation: false,
		startProgramm:  false,
		qrStopChan:     make(chan struct{}),
		qrDataChan:     make(chan models.ScannerResponse, 1),
		qrActiveChan:   make(chan bool, 1),
		kazsOperator:   pd.KazsOperator,
		appGui:         pd.AppGui,
		driver:         pd.Driver,
		kazsConfig:     pd.KazsConfig,
		trkSettings:    pd.TRKSettings,
		appConfig:      pd.AppConfig,
		driverConfig:   pd.DriverConfig,
		mu:             sync.RWMutex{},
		repository:     pd.Repository,
		//errproc:        baseErProc,
		logger: pd.Logger,

		telemetryBuffer: make([]byte, 0, 1024),

		lastSENSTelemetry:      make(map[string]SENSStatus),
		lastSENSTelemetryMutex: sync.Mutex{},

		lastTempTelemetry:      make(map[string]TempStatus),
		lastTempTelemetryMutex: sync.Mutex{},

		lastControllerDinTelemetry:      make(map[string]RelayStatus),
		lastControllerDinTelemetryMutex: sync.Mutex{},

		lastControllerDoutTelemetry:      make(map[string]RelayStatus),
		lastControllerDoutTelemetryMutex: sync.Mutex{},

		lastTRKTelemetry:      make(map[string]TRKStatus),
		lastTRKTelemetryMutex: sync.Mutex{},
	}
	return p

}

func (p *Processing) MainProcess() {
	// Проверка активации КАЗС
	activation, err := p.repository.Activation.GetActivation()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		p.CaptureError(fmt.Errorf("activation checking error: %s", err.Error()), CaptureErrorDep{})
		p.logger.Error("не удалось получить данные активации из БД.", "err", err)
		return
	}

	if activation == nil {
		p.logger.Info("КАЗС деактивирована.")
		p.appGui.CreateActivation()
	}

	if activation != nil {
		p.logger.WithKazsNumber(activation.KazsNumber)
		p.logger.Info("КАЗС активирована.")
		p.kazsActivation = true
		p.appGui.SetUI(activation.Logo, activation.SupportNumber, activation.KazsNumber, activation.KazsTimezone)
		p.appGui.CreateHeader()
		p.appGui.CreateDefaultScreen("1")
		p.appGui.CreateDefaultScreen("2")
		err = p.kazsOperator.SetConfig(activation.URL, activation.KazsAPIKey, activation.KazsID, activation.KazsNumber, activation.ConfigHash)
		if err != nil {
			p.CaptureError(fmt.Errorf("error setting kazs config: %s", err.Error()), CaptureErrorDep{})
			p.logger.Error("не удалось установить настройки.", "err", err)
			return
		}

		err = p.StartProgram(activation.NTPServer)
		if err != nil {
			p.CaptureError(fmt.Errorf("error starting program"), CaptureErrorDep{})
			p.logger.Error("не удалось выполнить старт программы.", "err", err)
		}

		p.appGui.CreateDefaultScreen("1")
		p.appGui.CreateDefaultScreen("2")
	}

	go p.driver.QRDriver.Adapter.Read(p.qrDataChan, p.qrStopChan, p.qrActiveChan)
	p.UpdateQRActive(true)

	for data := range p.qrDataChan {
		p.mu.RLock()
		isActive := p.qrActive
		p.mu.RUnlock()
		if !isActive {
			continue
		}

		if data.TYPE == 1 && !p.kazsActivation {
			go func() {
				_ = p.Activation(data)
			}()
		}

		if data.TYPE == 2 && p.kazsActivation {
			go func() {
				_ = p.FuelGet(data)
			}()
		}
		if data.TYPE == 3 && p.kazsActivation {
			go func() {
				_ = p.FuelGive(data)
			}()
		}
	}
}

func (p *Processing) UpdateQRActive(active bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	var status string

	p.qrActive = active
	if active {
		status = "enabled"
	} else {
		status = "disabled"
	}
	p.logger.Info("QR изменил статус.", "qr_status", status)

}

func (p *Processing) UpdateJarStatus(jarNumber string, active bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if jarNumber == "1" {
		p.oneJarActiveProcess = active
	} else if jarNumber == "2" {
		p.twoJarActiveProcess = active
	}
	p.logger.Info("поток изменил статус.", "jarNumber", jarNumber, "jar_status", active)
}
