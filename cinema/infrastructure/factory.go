package infrastructure

import (
	"cinema/domain/ports"
	"cinema/infrastructure/repository"
	"database/sql"
)

type Repositories struct {
	CinemaRepo      ports.CinemaRepository
	UserRepo        ports.UserRepository
	SessionRepo     ports.SessionRepository
	FilmRepo        ports.FilmRepository
	BookingRepo     ports.BookingRepository
	SeatsRepository ports.SeatsRepository
	HallRepository  ports.HallRepository
	//Repo       domain.Repository
	// Добавьте другие
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		CinemaRepo:      repository.NewCinemaRepo(db),
		UserRepo:        repository.NewUserRepo(db),
		SessionRepo:     repository.NewSessionRepo(db),
		FilmRepo:        repository.NewFilmRepo(db),
		BookingRepo:     repository.NewBookingRepo(db),
		SeatsRepository: repository.NewSeatsRepo(db),
		HallRepository:  repository.NewHallRepo(db),
		//Repo:       repository.NewRepo(db),
		// Добавьте другие
	}
}
