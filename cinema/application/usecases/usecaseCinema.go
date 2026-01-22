package usecases

import (
	"cinema/domain/entities"
	"cinema/domain/ports"
	"log"
)

type CreateCinemaInput struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	City       string `json:"city"`
	Phone      string `json:"phone"`
	TotalSeats int    `json:"total_seats"`
	Poster     string `json:"poster"`
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
	log.Printf("Input TotalSeats: %d", input.TotalSeats)
	cinema := entities.Cinema{
		Name:       input.Name,
		Address:    input.Address,
		City:       input.City,
		Phone:      input.Phone,
		TotalSeats: input.TotalSeats,
		Poster:     input.Poster,
	}
	log.Printf("Cinema.TotalSeats перед repo: %d", cinema.TotalSeats)
	id, err := uc.Repo.CreateCinema(cinema)
	if err != nil {
		return nil, err
	}

	return &CreateCinemaOutput{ID: id}, nil
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
