package usecase

import (
	"context"
	_ "fmt"
	_ "fuelstation/internal/gui"
	"fuelstation/internal/models"
	"fyne.io/fyne/v2"
	"log"
	"os"
	"time"
)

func (p *Processing) FuelGet(qrInfo models.ScannerResponse) error {
	logFile, err := os.OpenFile("fuelstation.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Ошибка открытия файла логов: %v", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "USECASE: ", log.LstdFlags)
	logger.Printf("FuelGet: TID=%s, oneJarActive=%v, twoJarActive=%v", qrInfo.TID, p.oneJarActive, p.twoJarActive)

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

	fuelType := qrInfo.FuelType
	liters := float32(qrInfo.Liters)
	maxLiters := float32(qrInfo.Liters)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
	go p.setFuelGetStartScreen(ctx, setFuelGetStartScreenDep{
		jarNumber:      jarNumber,
		fuelType:       fuelType,
		expectedAmount: liters,
	}, logger)
	go p.monitoringFuelGetStart(ctx, cancel, monitoringFuelGetStartDep{
		fuelGetID:     qrInfo.TID,
		jarNumber:     jarNumber,
		fuelType:      fuelType,
		expectedLiter: maxLiters,
	}, logger)

	return nil
}

type setFuelGetStartScreenDep struct {
	jarNumber      string
	fuelType       string
	expectedAmount float32
}

func (p *Processing) setFuelGetStartScreen(ctx context.Context, dep setFuelGetStartScreenDep, logger *log.Logger) {
	logger.Printf("setFuelGetStartScreen: jarNumber=%s", dep.jarNumber)
	deadline, ok := ctx.Deadline()
	timeout := 30
	if ok {
		timeout = int(time.Until(deadline).Seconds())
	}
	p.gui.CreateFuelGetStartScreen(dep.jarNumber, dep.expectedAmount, 0, 0, dep.expectedAmount, timeout)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for i := timeout - 1; i >= 0; i-- {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fyne.Do(func() {
				p.gui.CreateFuelGetStartScreen(dep.jarNumber, dep.expectedAmount, 0, 0, dep.expectedAmount, i)
			})
		}
	}
}

type monitoringFuelGetStartDep struct {
	fuelGetID     string
	jarNumber     string
	fuelType      string
	expectedLiter float32
}

func (p *Processing) monitoringFuelGetStart(ctx context.Context, cancel context.CancelFunc, dep monitoringFuelGetStartDep, logger *log.Logger) {
	logger.Printf("monitoringFuelGetStart: jarNumber=%s", dep.jarNumber)
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
			if trkStatus.ValueStr == "NozzleLifted" {
				logger.Println("Пистолет снят")
				p.TRKRequest("SetFuelGet", dep.jarNumber, dep.expectedLiter)
				p.TRKRequest("ApprovalTRK", dep.jarNumber, 0)
				fyne.Do(func() {
					p.gui.CreateFuelGetInProgressScreen(dep.jarNumber, dep.expectedLiter, 0, 0, dep.expectedLiter, 30)
				})
				go p.startFuelGetInProgress(startFuelGetInProgressDep{
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

type startFuelGetInProgressDep struct {
	jarNumber     string
	fuelType      string
	expectedLiter float32
}

func (p *Processing) startFuelGetInProgress(dep startFuelGetInProgressDep, logger *log.Logger) {
	logger.Printf("startFuelGetInProgress: jarNumber=%s", dep.jarNumber)
	liters := float32(0)
	for i := 0; i < 20; i++ {
		liters += dep.expectedLiter / 20
		fyne.Do(func() {
			p.gui.CreateFuelGetInProgressScreen(dep.jarNumber, dep.expectedLiter, liters, liters, dep.expectedLiter, 30)
		})
		time.Sleep(time.Second)
	}
	fyne.Do(func() {
		p.gui.CreateDefaultScreen(dep.jarNumber)
		p.UpdateJarStatus(dep.jarNumber, false)
	})
	logger.Println("Слив завершен")
}
