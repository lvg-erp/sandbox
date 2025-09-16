package usecase

import "fuelstation/internal/gui"

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
