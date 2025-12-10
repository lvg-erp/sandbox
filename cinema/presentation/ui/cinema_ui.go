package ui

import (
	"bytes"
	"cinema/domain/entities"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	_ "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"net/http"
	"strconv"
)

func CreateCinemaForm(w fyne.Window) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Название кинотеатра")

	addressEntry := widget.NewEntry()
	addressEntry.SetPlaceHolder("Адрес")

	cityEntry := widget.NewEntry()
	cityEntry.SetPlaceHolder("Город")

	phoneEntry := widget.NewEntry()
	phoneEntry.SetPlaceHolder("Телефон")

	totalSeatsEntry := widget.NewEntry()
	totalSeatsEntry.SetPlaceHolder("Общее количество мест")

	posterEntry := widget.NewEntry()
	posterEntry.SetPlaceHolder("URL постера")

	form := container.NewVBox(
		widget.NewLabel("Создать кинотеатр"),
		widget.NewLabel("Название:"), nameEntry,
		widget.NewLabel("Адрес:"), addressEntry,
		widget.NewLabel("Город:"), cityEntry,
		widget.NewLabel("Телефон:"), phoneEntry,
		widget.NewLabel("Мест всего:"), totalSeatsEntry,
		widget.NewLabel("Постер:"), posterEntry,
		widget.NewButton("Создать", func() {
			_ = entities.Cinema{
				Name:       nameEntry.Text,
				Address:    addressEntry.Text,
				City:       cityEntry.Text,
				Phone:      phoneEntry.Text,
				TotalSeats: parseInt(totalSeatsEntry.Text),
				Poster:     posterEntry.Text,
			}

			totalSeats, _ := strconv.Atoi(totalSeatsEntry.Text)
			payload := map[string]interface{}{
				"name":        nameEntry.Text,
				"address":     addressEntry.Text,
				"city":        cityEntry.Text,
				"phone":       phoneEntry.Text,
				"total_seats": totalSeats,
				"poster":      posterEntry.Text,
			}
			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", "http://localhost:8080/admin/cinemas", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req) // client из пакета ui

			if err != nil || resp.StatusCode != 201 {
				dialog.ShowError(fmt.Errorf("ошибка: %v", err), w)
				return
			}
			dialog.ShowInformation("Успех", "Кинотеатр создан!", w)
		}),
	)

	w.SetContent(container.NewScroll(form))
}

// Вспомогательная функция
func parseInt(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}
