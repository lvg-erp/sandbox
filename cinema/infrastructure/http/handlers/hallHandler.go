package handlers

import (
	"cinema/application/usecases"
	"cinema/domain/ports"
	"encoding/json"
	_ "github.com/gorilla/mux"
	"log"
	"net/http"
	_ "strconv"
)

func ListHalls(repo ports.HallRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		halls, err := repo.ListHalls()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(halls)
	}
}
func CreateHall(repo ports.HallRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			CinemaID    int    `json:"cinema_id"`
			Name        string `json:"name"`
			Rows        int    `json:"rows"`
			SeatsPerRow int    `json:"seats_per_row"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			log.Printf("CreateHall: invalid json: %v", err)
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		log.Printf("CreateHall input: %+v", input)

		if input.Name == "" || input.Rows <= 0 || input.SeatsPerRow <= 0 {
			log.Printf("CreateHall: validation failed")
			http.Error(w, "invalid data", http.StatusBadRequest)
			return
		}

		uc := &usecases.CreateHallUseCase{Repo: repo}
		output, err := uc.ExecuteCreateHall(&usecases.CreateHallInput{
			CinemaID:    input.CinemaID,
			Name:        input.Name,
			Rows:        input.Rows,
			SeatsPerRow: input.SeatsPerRow,
		})
		if err != nil {
			log.Printf("CreateHall error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(output)
	}
}

//import (
//	"cinema/internal/db"
//	"encoding/json"
//	"fmt"
//	"net/http"
//)
//
//func GenerateHall(repo db.Repository) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		if r.Method != http.MethodPost {
//			http.Error(w, "POST only", http.StatusMethodNotAllowed)
//			return
//		}
//
//		var input struct {
//			CinemaID uint `json:"cinema_id"`
//		}
//		json.NewDecoder(r.Body).Decode(&input)
//		if input.CinemaID == 0 {
//			http.Error(w, "cinema_id required", http.StatusBadRequest)
//			return
//		}
//
//		hallName := "Зал 1"
//		repo.db.QueryRow("SELECT COUNT(*) FROM halls WHERE cinema_id = $1", input.CinemaID).
//			Scan(&input) // переиспользуем переменную как счётчик
//		if count > 0 {
//			hallName = fmt.Sprintf("Зал %d", count+1)
//		}
//
//		hallID, err := repo.CreateHall(input.CinemaID, hallName, 10, 15) // 150 мест
//		if err != nil {
//			http.Error(w, err.Error(), http.StatusInternalServerError)
//			return
//		}
//
//		for row := 1; row <= 10; row++ {
//			for num := 1; num <= 15; num++ {
//				repo.CreateSeat(hallID, row, num)
//			}
//		}
//
//		w.WriteHeader(http.StatusCreated)
//		json.NewEncoder(w).Encode(map[string]any{
//			"hall_id": hallID,
//			"seats":   150,
//			"name":    hallName,
//		})
//	}
//}
