// internal/handlers/handlers.go
package handlers

import (
	"cinema/internal/auth"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strconv"
	"time"

	"cinema/internal/db"
)

func Register(repo db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u db.User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if u.Email == "" || u.Pass == "" {
			http.Error(w, "email and password required", http.StatusBadRequest)
			return
		}

		// Проверка на существование
		_, err := repo.GetUser(u.Email)
		if err == nil {
			http.Error(w, "email already exists", http.StatusConflict)
			return
		}

		// Генерим пассворд
		hashed, err := bcrypt.GenerateFromPassword([]byte(u.Pass), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "failed to hash password", http.StatusInternalServerError)
			return
		}
		u.Pass = string(hashed)

		//
		if err := repo.CreateUser(u); err != nil {
			log.Printf("CreateUser error: %v", err)
			http.Error(w, "failed to create user", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func LoginHandler(repo db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Email string `json:"email"`
			Pass  string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		user, err := repo.GetUser(input.Email)
		if err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		if !auth.CheckPassword(input.Pass, user.Pass) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		if err := repo.DeleteUserSessions(user.ID); err != nil {
			log.Printf("Login: failed to delete old sessions: %v", err)
		}

		token, err := repo.CreateSession(user.ID, time.Now().Add(30*24*time.Hour))
		if err != nil {
			http.Error(w, "failed to create session", http.StatusInternalServerError)
			return
		}

		// ← ТОЛЬКО КУКА В ОТВЕТЕ
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    token,
			Expires:  time.Now().Add(30 * 24 * time.Hour),
			HttpOnly: true,
			Path:     "/",
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})

		json.NewEncoder(w).Encode(map[string]string{
			"session_id": token,
		})
	}
}

func GetFilms(repo db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		films, err := repo.ListFilms()
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(films)
	}
}

func AddFilm(repo db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var f db.Film
		if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if f.Title == "" {
			http.Error(w, "title required", http.StatusBadRequest)
			return
		}
		if err := repo.CreateFilm(f); err != nil {
			http.Error(w, "failed to save", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "added"})
	}
}

func AddCinema(repo db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c db.Cinema
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if c.Name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}
		if err := repo.CreateCinema(c); err != nil {
			http.Error(w, "failed to save", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "created"})
	}
}

// Создание сеанса фильма
//func CreateFilmSession(repo db.Repository) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		var input struct {
//			FilmID    int       `json:"film_id"`
//			CinemaID  int       `json:"cinema_id"`
//			StartTime time.Time `json:"start_time"`
//			Price     float64   `json:"price"`
//		}
//		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
//			http.Error(w, "invalid json", http.StatusBadRequest)
//			return
//		}
//
//		id, err := repo.CreateFilmSession(input.FilmID, input.CinemaID, input.StartTime, input.Price)
//		if err != nil {
//			http.Error(w, err.Error(), http.StatusInternalServerError)
//			return
//		}
//		w.WriteHeader(http.StatusCreated)
//		json.NewEncoder(w).Encode(map[string]int{"id": id})
//	}
//}

// Бронирование
func Book(repo db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value("user").(db.User)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var b db.Booking
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		b.UserID = user.ID
		if err := repo.CreateBooking(b); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "booked"})
	}
}

func Logout(repo db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie("session_token")
		if cookie != nil {
			repo.DeleteSession(cookie.Value)
		}
		http.SetCookie(w, &http.Cookie{
			Name:   "session_token",
			Value:  "",
			MaxAge: -1,
			Path:   "/",
		})

	}
}

func Protected(repo db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(db.User)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"email":  user.Email,
			"role":   user.Role,
		})
	}
}

func CreateFilmSession(repo db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			FilmID   int    `json:"film_id"`
			CinemaID int    `json:"cinema_id"`
			Start    string `json:"start_time"` // ISO
			Rows     int    `json:"rows"`
			Cols     int    `json:"cols"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		start, err := time.Parse(time.RFC3339, input.Start)
		if err != nil {
			http.Error(w, "invalid time", http.StatusBadRequest)
			return
		}

		sessionID, err := repo.CreateFilmSession(input.FilmID, input.CinemaID, start, input.Rows, input.Cols)
		if err != nil {
			http.Error(w, "failed to create session", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]int{"session_id": sessionID})
	}
}

func GetSeats(repo db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, _ := strconv.Atoi(r.URL.Query().Get("session"))
		seats, err := repo.GetSeats(sessionID)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(seats)
	}
}

func GetSessions(repo db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessions, err := repo.GetAllSessions() // ← МЕТОД РЕПОЗИТОРИЯ
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(sessions)
	}
}
