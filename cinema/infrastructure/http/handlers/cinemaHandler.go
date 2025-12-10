package handlers

import (
	"cinema/application/usecases"
	"cinema/domain/ports"
	"encoding/json"
	"net/http"
)

func AddCinema(repo ports.CinemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var input usecases.CreateCinemaInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if input.Name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}
		uc := &usecases.CinemaUseCase{Repo: repo}
		output, err := uc.ExecuteCreateCinema(&input)
		if err != nil {
			http.Error(w, "failed to save", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(output)
	}
}

func ListCinemas(repo ports.CinemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "GET only", http.StatusMethodNotAllowed)
			return
		}

		uc := &usecases.CinemaUseCase{Repo: repo}
		cinemas, err := uc.ExecuteListCinema()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(cinemas)
		if err != nil {
			return
		}
	}
}

func GetCinema(repo ports.CinemaRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "GET only", http.StatusMethodNotAllowed)
			return
		}

		uc := &usecases.CinemaUseCase{Repo: repo}
		cinemas, err := uc.ExecuteListCinema()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(cinemas)
		if err != nil {
			return
		}
	}
}
