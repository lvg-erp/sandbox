package usecases

import (
	"cinema/domain/entities"
	"cinema/domain/ports"
)

type CreateCinemaInput struct {
	Name       string
	Address    string
	City       string
	Phone      string
	TotalSeats int
	Poster     string
}

type CreateCinemaOutput struct {
	ID int
}

type ListCinemasOutput []entities.Cinema

type GetCinemaInput struct {
	ID int
}

type GetCinemaOutput struct {
	Cinema entities.Cinema
}

type CinemaUseCase struct {
	Repo ports.CinemaRepository
}

func (uc *CinemaUseCase) ExecuteCreateCinema(input *CreateCinemaInput) (*CreateCinemaOutput, error) {
	cinema := entities.Cinema{
		Name:       input.Name,
		Address:    input.Address,
		City:       input.City,
		Phone:      input.Phone,
		TotalSeats: input.TotalSeats,
		Poster:     input.Poster,
	}
	if err := uc.Repo.CreateCinema(cinema); err != nil {
		return nil, err
	}
	return &CreateCinemaOutput{ID: cinema.ID}, nil
}

func (uc *CinemaUseCase) ExecuteListCinema() (ListCinemasOutput, error) {
	cinemas, err := uc.Repo.ListCinemas()
	if err != nil {
		return nil, err
	}
	return cinemas, nil
}

func (uc *CinemaUseCase) ExecuteGetCinema(input *GetCinemaInput) (*GetCinemaOutput, error) {
	cinema, err := uc.Repo.GetCinema(input.ID)
	if err != nil {
		return nil, err
	}
	return &GetCinemaOutput{Cinema: cinema}, nil
}
