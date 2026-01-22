package handlers

//func GetSessions(repo ports.SessionRepository) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		if r.Method != http.MethodGet {
//			http.Error(w, "GET only", http.StatusMethodNotAllowed)
//			return
//		}
//		cinemaIDStr := r.URL.Query().Get("cinema_id")
//		if cinemaIDStr == "" {
//			http.Error(w, "cinema_id required", http.StatusBadRequest)
//			return
//		}
//		cinemaID, _ := strconv.Atoi(cinemaIDStr)
//
//		uc := &usecases.SessionUseCase{Repo: repo}
//		input := &usecases.ListFilmSessionsForCinemaInput{CinemaID: cinemaID}
//		sessions, err := uc.ExecuteListSessionsForCinema(input)
//		if err != nil {
//			http.Error(w, err.Error(), http.StatusInternalServerError)
//			return
//		}
//		err = json.NewEncoder(w).Encode(sessions)
//		if err != nil {
//			return
//		}
//	}
//
//}
