package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"net/http"
	"strconv"
)

type Seat struct {
	ID       int  `json:"id"`
	Row      int  `json:"row"`
	Number   int  `json:"number"`
	Reserved bool `json:"reserved"`
}

func fetchSeats(sessionID int) ([]Seat, error) {
	resp, err := client.Get(fmt.Sprintf("http://localhost:8080/seats?session=%d", sessionID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var seats []Seat
	if err := json.NewDecoder(resp.Body).Decode(&seats); err != nil {
		return nil, err
	}
	return seats, nil
}

func bookSeats(sessionID int, seatIDs []int) error {
	payload := map[string]interface{}{
		"session_id": sessionID,
		"seat_ids":   seatIDs,
	}
	body, _ := json.Marshal(payload)
	resp, _ := client.Post("http://localhost:8080/book", "application/json", bytes.NewBuffer(body))
	if resp.StatusCode != 200 {
		return fmt.Errorf("booking failed: %d", resp.StatusCode)
	}
	return nil
}

func createSeatButton(s Seat, selected map[int]bool, selectedIDs *[]int, refresh func()) *widget.Button {
	btn := widget.NewButton(fmt.Sprintf("%d-%d", s.Row, s.Number), func() {
		if s.Reserved {
			return
		}
		selected[s.ID] = !selected[s.ID]
		ids := *selectedIDs
		if selected[s.ID] {
			*selectedIDs = append(ids, s.ID)
		} else {
			for i, id := range ids {
				if id == s.ID {
					*selectedIDs = append(ids[:i], ids[i+1:]...)
					break
				}
			}
		}
		refresh()
	})

	switch {
	case s.Reserved:
		btn.Importance = widget.DangerImportance
	case selected[s.ID]:
		btn.Importance = widget.HighImportance
	default:
		btn.Importance = widget.MediumImportance
	}
	return btn
}

func ShowHall(sessionID int) {
	seats, err := fetchSeats(sessionID)
	if err != nil {
		dialog.ShowError(err, window)
		return
	}

	selected := make(map[int]bool)
	var selectedIDs []int

	maxRow, maxCol := 0, 0
	for _, s := range seats {
		if s.Row > maxRow {
			maxRow = s.Row
		}
		if s.Number > maxCol {
			maxCol = s.Number
		}
	}

	grid := container.NewGridWithColumns(maxCol)

	var refresh func()
	refresh = func() {
		grid.Objects = nil
		for _, s := range seats {
			grid.Add(createSeatButton(s, selected, &selectedIDs, refresh))
		}
		grid.Refresh()
	}

	refresh() // Вызов

	bookBtn := widget.NewButton("Забронировать", func() {
		if len(selectedIDs) == 0 {
			dialog.ShowInformation("Ошибка", "Выберите места", window)
			return
		}
		if err := bookSeats(sessionID, selectedIDs); err != nil {
			dialog.ShowError(err, window)
			return
		}
		dialog.ShowInformation("Успех", "Места забронированы!", window)
		showScreen(showFilmsList)
	})

	backBtn := widget.NewButton("Назад", goBack)

	window.SetContent(container.NewBorder(
		container.NewHBox(widget.NewLabel(fmt.Sprintf("Сеанс %d", sessionID)), layout.NewSpacer(), bookBtn),
		backBtn,
		nil, nil,
		container.NewScroll(grid),
	))
}

func ShowHallByHallID(hallID int) {
	seats, err := fetchSeatsByHall(hallID)
	if err != nil {
		dialog.ShowError(err, window)
		return
	}

	// Объявляем переменные для выбранных мест
	selected := make(map[int]bool)
	selectedIDs := []int{}

	maxRow, maxCol := 0, 0
	for _, s := range seats {
		if s.Row > maxRow {
			maxRow = s.Row
		}
		if s.Number > maxCol {
			maxCol = s.Number
		}
	}

	// Создаем контейнер с сеткой
	grid := &fyne.Container{}
	grid.Layout = layout.NewGridLayout(maxCol)

	for row := 1; row <= maxRow; row++ {
		for col := 1; col <= maxCol; col++ {
			var btn fyne.CanvasObject
			found := false
			for _, s := range seats {
				if s.Row == row && s.Number == col {
					btn = createSeatButton(s, selected, &selectedIDs, func() {
						ShowHallByHallID(hallID)
					})
					found = true
					break
				}
			}
			if !found {
				btn = widget.NewLabel("") // или используйте другой виджет-заполнитель
			}
			grid.AddObject(btn)
		}
	}

	// Создаем вертикальный контейнер:
	// - Заголовок
	// - Максимальный контейнер с сеткой (растягнется по высоте)
	// - Spacer, чтобы заполнить оставшееся пространство
	// - Кнопка "Назад" внизу
	contentContainer := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Зал %d", hallID)),
		// Оборачиваем сетку в Max, чтобы она растягивалась по доступной высоте
		container.NewMax(grid),
		layout.NewSpacer(), // заполнит все свободное пространство
		widget.NewButton("Назад", goBack),
	)

	// Обновляем UI
	window.SetContent(contentContainer)
	window.Resize(fyne.NewSize(800, 600))
}

func showCreateHallForm() {
	// Получаем кинотеатры
	cinemasResp, _ := client.Get("http://localhost:8080/admin/cinemas/list")
	var cinemas []map[string]interface{}
	json.NewDecoder(cinemasResp.Body).Decode(&cinemas)
	cinemasResp.Body.Close()

	cinemaOptions := []string{}
	cinemaIDs := []int{}
	for _, c := range cinemas {
		name := c["name"].(string)
		id := int(c["id"].(float64))
		cinemaOptions = append(cinemaOptions, name)
		cinemaIDs = append(cinemaIDs, id)
	}
	cinemaSelect := widget.NewSelect(cinemaOptions, nil)

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Название зала")

	rowsEntry := widget.NewEntry()
	rowsEntry.SetPlaceHolder("Рядов (например, 10)")

	seatsEntry := widget.NewEntry()
	seatsEntry.SetPlaceHolder("Мест в ряду (например, 5)")

	previewContainer := container.NewVBox(widget.NewLabel("Создайте зал для просмотра схемы"))

	createBtn := widget.NewButton("Создать зал", func() {
		if cinemaSelect.Selected == "" || nameEntry.Text == "" || rowsEntry.Text == "" || seatsEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("заполните все поля"), window)
			return
		}
		cinemaID := cinemaIDs[cinemaSelect.SelectedIndex()]
		rows, _ := strconv.Atoi(rowsEntry.Text)
		seatsPerRow, _ := strconv.Atoi(seatsEntry.Text)

		payload := map[string]interface{}{
			"cinema_id":     cinemaID,
			"name":          nameEntry.Text,
			"rows":          rows,
			"seats_per_row": seatsPerRow,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "http://localhost:8080/admin/halls/create", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := client.Do(req)
		if resp.StatusCode != 201 {
			dialog.ShowError(fmt.Errorf("ошибка создания зала"), window)
			return
		}

		// Получаем ID нового зала (предполагаем, что сервер возвращает { "id": X })
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		hallID := int(result["id"].(float64))

		dialog.ShowInformation("Успех", "Зал создан с местами!", window)
		ShowHallByHallID(hallID) // Показываем схему сразу после создания
	})

	backBtn := widget.NewButton("Назад", goBack)

	window.SetContent(container.NewBorder(
		nil, backBtn, nil, nil,
		container.NewVBox(
			widget.NewLabel("Создать зал"),
			widget.NewLabel("Кинотеатр:"), cinemaSelect,
			widget.NewLabel("Название зала:"), nameEntry,
			widget.NewLabel("Рядов:"), rowsEntry,
			widget.NewLabel("Мест в ряду:"), seatsEntry,
			createBtn,
			previewContainer, // Можно оставить для будущего превью
		),
	))
}
