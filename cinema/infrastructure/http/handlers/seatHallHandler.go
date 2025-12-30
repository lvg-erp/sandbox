package handlers

import (
	"cinema/domain/ports"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func GetHallSeats(repo ports.SeatsRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//hallIDStr := mux.Vars(r)["id"]
		vars := mux.Vars(r)
		log.Printf("mux.Vars: %+v", vars)
		hallIDStr := vars["id"]
		if hallIDStr == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		hallID, err := strconv.Atoi(hallIDStr)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		log.Printf("Requested hall ID: %d", hallID)

		seats, err := repo.ListSeatsByHall(hallID)
		if err != nil {
			log.Printf("Error fetching seats: %v", err)
			http.Error(w, "failed to fetch seats", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(seats)
	}
}
