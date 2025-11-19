package handlers

//import (
//	"cinema/internal/db"
//	"cinema/internal/pdf"
//	"net/http"
//)
//
//// internal/handlers/ticket.go
//func DownloadTicket(repo db.Repository) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		// Пример данных — замени на реальные из БД!
//		ticketData := pdf.TicketData{
//			BookingID:  42,
//			FilmTitle:  "Дюна 2",
//			HallNumber: 1,
//			Row:        7,
//			Seat:       12,
//			Date:       "2025-11-15",
//			Time:       "19:30",
//			QRContent:  "TICKET:42",
//		}
//
//		pdfBytes, err := pdf.GenerateTicket(ticketData)
//		if err != nil {
//			http.Error(w, "Ошибка генерации PDF", 500)
//			return
//		}
//
//		w.Header().Set("Content-Type", "application/pdf")
//		w.Header().Set("Content-Disposition", `attachment; filename="ticket_42.pdf"`)
//		w.Write(pdfBytes)
//	}
//}
