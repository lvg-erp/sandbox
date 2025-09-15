package usecase

import (
	"context"
	"fmt"
	"fuelstation/internal/gui"
	"fuelstation/internal/models"
	"fyne.io/fyne/v2"
	"log"
	"os"
	"time"
)

const (
	StatusIdle              = "30"
	StatusNozzleLifted      = "31"
	StatusFuelingComplete   = "34"
	StatusFuelingInProgress = "33"
	StatusAuthorized        = "32"
)

type Processing struct {
	gui            gui.SectionInterface
	oneJarActive   bool
	twoJarActive   bool
	fuelGiveConfig struct {
		FuelGiveStartScreenTimeout int
		FuelGiveTimeout            int
	}
}

func NewProcessing(g gui.SectionInterface) *Processing {
	return &Processing{
		gui: g,
		fuelGiveConfig: struct {
			FuelGiveStartScreenTimeout int
			FuelGiveTimeout            int
		}{
			FuelGiveStartScreenTimeout: 30,
			FuelGiveTimeout:            300,
		},
	}
}

func (p *Processing) FuelGive(qrInfo models.ScannerResponse) error {
	logFile, err := os.OpenFile("fuelstation.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Ошибка открытия файла логов: %v", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "USECASE: ", log.LstdFlags)
	logger.Printf("FuelGive: TID=%s, oneJarActive=%v, twoJarActive=%v", qrInfo.TID, p.oneJarActive, p.twoJarActive)

	if p.oneJarActive && p.twoJarActive {
		logger.Println("Все емкости заняты")
		section := p.getAvailableSection()
		if section != nil {
			var jarNumber string
			if section == p.gui.GetSection("1") {
				jarNumber = "1"
			} else {
				jarNumber = "2"
			}
			fyne.Do(func() {
				p.gui.ShowSectionDialog(section.Content, "Ошибка", "Все емкости заняты", 10, func() {
					p.gui.CreateDefaultScreen(jarNumber)
				})
			})
		}
		return nil
	}

	jarNumber := "1"
	if p.oneJarActive {
		jarNumber = "2"
	}
	logger.Printf("Selected jarNumber=%s", jarNumber)

	p.UpdateJarStatus(jarNumber, true)
	p.gui.CreateDownloadScreen(jarNumber)

	fuelType := "Petrol"
	liters := float32(50)
	maxLiters := float32(50)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.fuelGiveConfig.FuelGiveStartScreenTimeout)*time.Second)
	go p.setFuelGiveStartScreen(ctx, setFuelGiveStartScreenDep{
		jarNumber:      jarNumber,
		fuelType:       fuelType,
		expectedAmount: liters,
	}, logger)
	go p.monitoringFuelGiveStart(ctx, cancel, monitoringFuelGiveStartDep{
		fuelGiveID:    qrInfo.TID,
		jarNumber:     jarNumber,
		fuelType:      fuelType,
		expectedLiter: maxLiters,
	}, logger)

	return nil
}

type setFuelGiveStartScreenDep struct {
	jarNumber      string
	fuelType       string
	expectedAmount float32
}

func (p *Processing) setFuelGiveStartScreen(ctx context.Context, dep setFuelGiveStartScreenDep, logger *log.Logger) {
	logger.Printf("setFuelGiveStartScreen: jarNumber=%s", dep.jarNumber)
	deadline, ok := ctx.Deadline()
	timeout := p.fuelGiveConfig.FuelGiveStartScreenTimeout
	if ok {
		timeout = int(time.Until(deadline).Seconds())
	}
	p.gui.CreateFuelGiveStartScreen(dep.jarNumber, dep.expectedAmount, dep.fuelType, timeout)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for i := timeout - 1; i >= 0; i-- {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fyne.Do(func() {
				p.gui.CreateFuelGiveStartScreen(dep.jarNumber, dep.expectedAmount, dep.fuelType, i)
			})
		}
	}
}

type monitoringFuelGiveStartDep struct {
	fuelGiveID    string
	jarNumber     string
	fuelType      string
	expectedLiter float32
}

func (p *Processing) monitoringFuelGiveStart(ctx context.Context, cancel context.CancelFunc, dep monitoringFuelGiveStartDep, logger *log.Logger) {
	logger.Printf("monitoringFuelGiveStart: jarNumber=%s", dep.jarNumber)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			logger.Println("Timeout: Пистолет не снят")
			fyne.Do(func() {
				p.gui.CreateDefaultScreen(dep.jarNumber)
				p.UpdateJarStatus(dep.jarNumber, false)
			})
			return
		default:
			trkStatus := p.TRKRequest("GetTRKStatus", dep.jarNumber, 0)
			if trkStatus.Err != nil {
				logger.Printf("TRKRequest error: %v", trkStatus.Err)
				fyne.Do(func() {
					section := p.gui.GetSection(dep.jarNumber)
					p.gui.ShowSectionDialog(section.Content, "Ошибка", "ТРК не отвечает", 10, func() {
						p.gui.CreateDefaultScreen(dep.jarNumber)
						p.UpdateJarStatus(dep.jarNumber, false)
					})
				})
				return
			}
			if trkStatus.ValueStr == StatusNozzleLifted {
				logger.Println("Пистолет снят")
				p.TRKRequest("SetFuelGive", dep.jarNumber, dep.expectedLiter)
				p.TRKRequest("ApprovalTRK", dep.jarNumber, 0)
				fyne.Do(func() {
					p.gui.CreateFuelGiveInProgressScreen(dep.jarNumber, dep.fuelType, 0, dep.expectedLiter)
				})
				go p.startFuelGiveInProgress(startFuelGiveInProgressDep{
					jarNumber:     dep.jarNumber,
					fuelType:      dep.fuelType,
					expectedLiter: dep.expectedLiter,
				}, logger)
				return
			}
			time.Sleep(time.Second)
		}
	}
}

type startFuelGiveInProgressDep struct {
	jarNumber     string
	fuelType      string
	expectedLiter float32
}

func (p *Processing) startFuelGiveInProgress(dep startFuelGiveInProgressDep, logger *log.Logger) {
	logger.Printf("startFuelGiveInProgress: jarNumber=%s", dep.jarNumber)
	liters := float32(0)
	for i := 0; i < 20; i++ {
		liters += dep.expectedLiter / 20
		fyne.Do(func() {
			p.gui.CreateFuelGiveInProgressScreen(dep.jarNumber, dep.fuelType, liters, dep.expectedLiter)
		})
		time.Sleep(time.Second)
	}
	fyne.Do(func() {
		p.gui.CreateDefaultScreen(dep.jarNumber)
		p.UpdateJarStatus(dep.jarNumber, false)
	})
	logger.Println("Заправка завершена")
}

func (p *Processing) UpdateJarStatus(jarNumber string, status bool) {
	if jarNumber == "1" {
		p.oneJarActive = status
	} else {
		p.twoJarActive = status
	}
	log.Printf("UpdateJarStatus: jarNumber=%s, status=%v", jarNumber, status)
}

func (p *Processing) getAvailableSection() *gui.Section {
	if !p.oneJarActive {
		return p.gui.GetSection("1")
	}
	if !p.twoJarActive {
		return p.gui.GetSection("2")
	}
	return nil
}

func (p *Processing) TRKRequest(requestType, jarNumber string, value float32) models.TRKResponse {
	switch requestType {
	case "GetTRKStatus":
		return models.TRKResponse{ValueStr: StatusNozzleLifted}
	case "SetFuelGive", "ApprovalTRK":
		return models.TRKResponse{}
	default:
		return models.TRKResponse{Err: fmt.Errorf("unknown request type: %s", requestType)}
	}
}
