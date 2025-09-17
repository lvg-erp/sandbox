package usecase

import (
	"fuelstation/internal/gui"
	"log"
	"time"
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

func (p *Processing) getAvailableSection() *gui.Section {
	if !p.oneJarActive {
		return p.gui.GetSection("1")
	}
	if !p.twoJarActive {
		return p.gui.GetSection("2")
	}
	return nil
}

func (p *Processing) UpdateJarStatus(jarNumber string, status bool) {
	if jarNumber == "1" {
		p.oneJarActive = status
	} else {
		p.twoJarActive = status
	}
	log.Printf("UpdateJarStatus: jarNumber=%s, status=%v", jarNumber, status)
}

func (p *Processing) TRKRequest(action, jarNumber string, liters float32) struct {
	ValueStr string
	Err      error
} {
	// Эмуляция: возвращаем NozzleLifted в течение 20 секунд после начала
	if action == "GetTRKStatus" {
		if time.Since(time.Now().Add(-20*time.Second)).Seconds() > 0 { // Эмуляция в течение 20 секунд
			return struct {
				ValueStr string
				Err      error
			}{ValueStr: "NozzleLifted", Err: nil}
		}
		return struct {
			ValueStr string
			Err      error
		}{ValueStr: "NozzleDown", Err: nil} // После 20 секунд
	}
	return struct {
		ValueStr string
		Err      error
	}{ValueStr: "", Err: nil}
}
