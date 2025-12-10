package handlers

import (
	"cinema/domain/entities"
	"cinema/domain/ports"
	"encoding/json"
	"net/http"
)

func GetFilms(repoFilm ports.FilmRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		films, err := repoFilm.ListFilms()
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(films)
	}
}

func AddFilm(repoFilm ports.FilmRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var f entities.Film
		if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if f.Title == "" {
			http.Error(w, "title required", http.StatusBadRequest)
			return
		}
		if err := repoFilm.CreateFilm(f); err != nil {
			http.Error(w, "failed to save", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "added"})
	}
}
