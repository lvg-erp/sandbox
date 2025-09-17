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

func (p *Processing) FuelGive(qrInfo models.ScannerResponse) error {
	logFile, err := os.OpenFile("fuelstation.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Ошибка открытия файла логов: %v", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "USECASE: ", log.LstdFlags)
	logger.Printf("FuelGive: TID=%s, oneJarActive=%v, twoJarActive=%v", qrInfo.TID, p.oneJarActive, p.twoJarActive)

	if p.oneJarActive && p.twoJarActive {
		logger.Println("Все пистолеты заняты")
		section := p.getAvailableSection()
		if section != nil {
			var jarNumber string
			if section == p.gui.GetSection("1") {
				jarNumber = "1"
			} else {
				jarNumber = "2"
			}
			fyne.Do(func() {
				p.gui.ShowSectionDialog(section.Content, "Ошибка", "Все пистолеты заняты", 10, func() {
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.fuelGiveConfig.FuelGiveStartScreenTimeout)*time.Second)
	go p.setFuelGiveStartScreen(ctx, setFuelGiveStartScreenDep{
		jarNumber: jarNumber,
		fuelType:  fuelType,
		liters:    liters,
	}, logger)
	go p.monitoringFuelGiveStart(ctx, cancel, monitoringFuelGiveStartDep{
		fuelGiveID: qrInfo.TID,
		jarNumber:  jarNumber,
		fuelType:   fuelType,
		liters:     liters,
	}, logger)

	return nil
}

type setFuelGiveStartScreenDep struct {
	jarNumber string
	fuelType  string
	liters    float32
}

func (p *Processing) setFuelGiveStartScreen(ctx context.Context, dep setFuelGiveStartScreenDep, logger *log.Logger) {
	logger.Printf("setFuelGiveStartScreen: jarNumber=%s", dep.jarNumber)
	deadline, ok := ctx.Deadline()
	timeout := p.fuelGiveConfig.FuelGiveStartScreenTimeout
	if ok {
		timeout = int(time.Until(deadline).Seconds())
	}
	p.gui.CreateFuelGiveStartScreen(dep.jarNumber, dep.liters, dep.fuelType, timeout)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for i := timeout - 1; i >= 0; i-- {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fyne.Do(func() {
				p.gui.CreateFuelGiveStartScreen(dep.jarNumber, dep.liters, dep.fuelType, i)
			})
		}
	}
}

type monitoringFuelGiveStartDep struct {
	fuelGiveID string
	jarNumber  string
	fuelType   string
	liters     float32
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
			if trkStatus.ValueStr == "NozzleLifted" {
				logger.Println("Пистолет снят")
				p.TRKRequest("SetFuelGive", dep.jarNumber, dep.liters)
				p.TRKRequest("ApprovalTRK", dep.jarNumber, 0)
				fyne.Do(func() {
					p.gui.CreateFuelGiveInProgressScreen(dep.jarNumber, 0, dep.liters)
				})
				go p.startFuelGiveInProgress(startFuelGiveInProgressDep{
					jarNumber: dep.jarNumber,
					fuelType:  dep.fuelType,
					liters:    dep.liters,
				}, logger)
				return
			}
			time.Sleep(time.Second)
		}
	}
}

type startFuelGiveInProgressDep struct {
	jarNumber string
	fuelType  string
	liters    float32
}

func (p *Processing) startFuelGiveInProgress(dep startFuelGiveInProgressDep, logger *log.Logger) {
	logger.Printf("startFuelGiveInProgress: jarNumber=%s", dep.jarNumber)
	liters := float32(0)
	for i := 0; i < 20; i++ {
		liters += dep.liters / 20
		fyne.Do(func() {
			p.gui.CreateFuelGiveInProgressScreen(dep.jarNumber, liters, dep.liters)
		})
		time.Sleep(time.Second)
	}
	fyne.Do(func() {
		p.gui.CreateFuelGiveCompleteScreen(dep.jarNumber, dep.liters, dep.fuelType)
		p.UpdateJarStatus(dep.jarNumber, false)
	})
	logger.Println("Заправка завершена")
}
