package ports

import "cinema/domain/entities"

// Фильмы
type FilmRepository interface {
	CreateFilm(entities.Film) error
	ListFilms() ([]entities.Film, error)
	GetFilm(id int) (entities.Film, error)
}
