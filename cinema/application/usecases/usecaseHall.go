package usecases

import (
	_ "cinema/domain/entities"
	"cinema/domain/ports"
)

type CreateHallInput struct {
	CinemaID    int
	Name        string
	Rows        int
	SeatsPerRow int
}

type CreateHallOutput struct {
	ID int
}

type CreateHallUseCase struct {
	Repo ports.HallRepository
}

func (uc *CreateHallUseCase) ExecuteCreateHall(input *CreateHallInput) (*CreateHallOutput, error) {
	hallID, err := uc.Repo.CreateHall(uint(input.CinemaID), input.Name, input.Rows, input.SeatsPerRow)
	if err != nil {
		return nil, err
	}

	if err := uc.Repo.GenerateSeats(hallID, input.Rows, input.SeatsPerRow); err != nil {
		return nil, err
	}

	return &CreateHallOutput{ID: int(hallID)}, nil
}
