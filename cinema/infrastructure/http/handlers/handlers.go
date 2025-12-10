package handlers

//func AddCinema(repo domain.Repository) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		var c entities.Cinema
//		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
//			http.Error(w, "invalid json", http.StatusBadRequest)
//			return
//		}
//		if c.Name == "" {
//			http.Error(w, "name required", http.StatusBadRequest)
//			return
//		}
//		if err := repo.CreateCinema(c); err != nil {
//			http.Error(w, "failed to save", http.StatusInternalServerError)
//			return
//		}
//		w.WriteHeader(http.StatusCreated)
//		json.NewEncoder(w).Encode(map[string]string{"status": "created"})
//	}
//}

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

//// Бронирование
//func Book(repo domain.Repository) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		user, ok := r.Context().Value("user").(entities.User)
//		if !ok {
//			http.Error(w, "unauthorized", http.StatusUnauthorized)
//			return
//		}
//
//		var b entities.Booking
//		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
//			http.Error(w, "invalid json", http.StatusBadRequest)
//			return
//		}
//
//		b.UserID = user.ID
//		if err := repo.CreateBooking(b); err != nil {
//			http.Error(w, err.Error(), http.StatusConflict)
//			return
//		}
//
//		w.WriteHeader(http.StatusCreated)
//		json.NewEncoder(w).Encode(map[string]string{"status": "booked"})
//	}
//}
//
//func CreateFilmSession(repo domain.Repository) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		var input struct {
//			FilmID   int    `json:"film_id"`
//			CinemaID int    `json:"cinema_id"`
//			Start    string `json:"start_time"`
//		}
//		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
//			http.Error(w, "invalid json", http.StatusBadRequest)
//			return
//		}
//		start, err := time.Parse("2006-01-02 15:04", input.Start)
//		if err != nil {
//			http.Error(w, "invalid time", http.StatusBadRequest)
//			return
//		}
//
//		sessionID, err := repo.CreateFilmSession(input.FilmID, input.CinemaID, start)
//		if err != nil {
//			http.Error(w, "failed to create session", http.StatusInternalServerError)
//			return
//		}
//		json.NewEncoder(w).Encode(map[string]int{"session_id": sessionID})
//	}
//}
//
//func GetSeats(repo domain.Repository) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		sessionID, _ := strconv.Atoi(r.URL.Query().Get("session"))
//		seats, err := repo.GetSeats(sessionID)
//		if err != nil {
//			http.Error(w, "not found", http.StatusNotFound)
//			return
//		}
//		json.NewEncoder(w).Encode(seats)
//	}
//}
//

//func GetSessions(repo db.Repository) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		sessions, err := repo.GetAllSessions() // ← МЕТОД РЕПОЗИТОРИЯ
//		if err != nil {
//			http.Error(w, "db error", http.StatusInternalServerError)
//			return
//		}
//		json.NewEncoder(w).Encode(sessions)
//	}
//}
