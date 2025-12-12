package usecases

import (
	"cinema/domain/entities"
	"cinema/domain/ports"
)

type ListFilmsOutput []entities.Film

type GetFilmInput struct {
	ID int
}

type GetFilmOutput struct {
	Film entities.Film
}

type FilmUseCase struct {
	Repo ports.FilmRepository
}

//func (uc *FilmUseCase) ExecuteCreateCinema(input *CreateCinemaInput) (*CreateCinemaOutput, error) {
//	cinema := entities.Cinema{
//		Name:       input.Name,
//		Address:    input.Address,
//		City:       input.City,
//		Phone:      input.Phone,
//		TotalSeats: input.TotalSeats,
//		Poster:     input.Poster,
//	}
//	if err := uc.Repo.CreateCinema(cinema); err != nil {
//		return nil, err
//	}
//	return &CreateCinemaOutput{ID: cinema.ID}, nil
//}

func (uc *FilmUseCase) ExecuteListFilm() (ListFilmsOutput, error) {
	films, err := uc.Repo.ListFilms()
	if err != nil {
		return nil, err
	}
	return films, nil
}
