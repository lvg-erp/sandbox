package handlers

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
