package application

import (
	"fmt"
	"fuelazs/internal/integration"
)

type Integration struct {
	KazsOperator *integration.KazsOperator
}

func NewIntegration() (*Integration, error) {
	// Инициализация интеграции с КАЗС
	kazsOperator, err := integration.NewKazsOperator()

	if err != nil {
		return nil, fmt.Errorf("error creating KazsOperator: %v", err)
	}

	return &Integration{
		KazsOperator: kazsOperator,
	}, nil
}
