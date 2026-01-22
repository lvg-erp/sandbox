package ports

import (
	"cinema/domain/entities"
	"time"
)

type SessionFilmRepository interface {
	GetAllFilmsSessions() ([]entities.SessionFilm, error)
	GetFilmSessions(filmID, cinemaID int) ([]entities.SessionFilm, error)
	CreateFilmSession(filmID, cinemaID int, start, end time.Time) (int, error)
	ListFilmsForCinema(cinemaID int) ([]map[string]interface{}, error)
	ListCinemasWithSessions() ([]map[string]interface{}, error)
}

//func (s SessionFilmRepository) CreateFilm(film entities.Film) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (s SessionFilmRepository) ListFilms() ([]entities.Film, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (s SessionFilmRepository) GetFilm(id int) (entities.Film, error) {
//	//TODO implement me
//	panic("implement me")
//}
