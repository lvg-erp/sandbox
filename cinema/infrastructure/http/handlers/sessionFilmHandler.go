package handlers

import (
	"cinema/application/usecases"
	"cinema/domain/ports"
	"encoding/json"
	"net/http"
	"strconv"
)

func GetFilmSessions(repo ports.SessionFilmRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method GET only", http.StatusMethodNotAllowed)
			return
		}

		//filmIDStr := r.URL.Query().Get("film_id")
		cinemaIDStr := r.URL.Query().Get("cinema_id")

		//if filmIDStr == "" || cinemaIDStr == "" {
		if cinemaIDStr == "" {
			http.Error(w, "film_id and cinema_id required", http.StatusBadRequest)
			return
		}

		//filmID, err1 := strconv.Atoi(filmIDStr)
		cinemaID, err := strconv.Atoi(cinemaIDStr)
		if err != nil {
			http.Error(w, "invalid id format", http.StatusBadRequest)
			return
		}

		uc := &usecases.SessionFilmUseCase{Repo: repo}

		sessions, err := uc.ExecuteListSessionsForCinema(&usecases.ListFilmSessionsForCinemaInput{
			//FilmID:   filmID,
			CinemaID: cinemaID,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sessions)
	}
}

func CreateFilmSession(repo ports.SessionFilmRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		cinemaIDStr := r.URL.Query().Get("cinema_id")
		if cinemaIDStr == "" {
			http.Error(w, "cinema_id required", http.StatusBadRequest)
			return
		}
		cinemaID, _ := strconv.Atoi(cinemaIDStr)

		uc := &usecases.SessionFilmUseCase{Repo: repo}
		input := &usecases.ListFilmSessionsForCinemaInput{CinemaID: cinemaID}
		sessions, err := uc.ExecuteListSessionsForCinema(input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = json.NewEncoder(w).Encode(sessions)
		if err != nil {
			return
		}
	}
}
