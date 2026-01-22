package ui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"time"
)

var (
	client  *http.Client
	window  fyne.Window
	history []func()
)

const cookieFile = "./cookies.json"

// Приложение запускается кастомно
// надо хранить куки например на диске
// куки
func saveCookies() {
	if client == nil || client.Jar == nil {
		return
	}
	u, _ := url.Parse("http://localhost:8080")
	cookies := client.Jar.Cookies(u)
	data, _ := json.Marshal(cookies)
	err := os.WriteFile(cookieFile, data, 0600)
	if err != nil {
		return
	}
	log.Printf("КУКИ СОХРАНЕНЫ: %d шт.", len(cookies))
}

// Загрузить куки
func loadCookies() {
	data, err := os.ReadFile(cookieFile)

	if err != nil {
		return
	}
	var cookies []*http.Cookie
	if json.Unmarshal(data, &cookies) != nil {
		return
	}
	u, _ := url.Parse("http://localhost:8080")
	client.Jar.SetCookies(u, cookies)
	log.Printf("КУКИ ЗАГРУЖЕНЫ: %d шт.", len(cookies))
}

func showScreen(f func()) {
	//Новый экран
	f()

	//
	content := window.Content()
	history = append(history, func() {
		window.SetContent(content)
		log.Printf("ВОССТАНОВЛЕНО: экран из истории")
	})

	log.Printf("showScreen: добавлено в историю, всего: %d", len(history))
}

func goBack() {
	log.Printf("goBack: нажато, history size = %d", len(history))
	if len(history) > 1 {
		history = history[:len(history)-1]
		prev := history[len(history)-1]
		prev()
		log.Printf("goBack: возвращено, осталось: %d", len(history))
	} else {
		log.Printf("goBack: нет куда возвращаться")
	}
}

func Start(w fyne.Window) {
	window = w

	//
	jar, _ := cookiejar.New(nil)
	client = &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}

	// Грузим куки
	loadCookies()

	time.Sleep(500 * time.Millisecond)
	history = nil

	if tryAutoLogin() {
		showScreen(showMainScreen)
	} else {
		showScreen(showLoginForm)
	}
}

func tryAutoLogin() bool {
	if client == nil {
		return false
	}

	req, _ := http.NewRequest("GET", "http://localhost:8080/protected", nil)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return false
	}
	resp.Body.Close()
	return true
}

// === ЭКРАНЫ ===

func showLoginForm() {
	email := widget.NewEntry()
	email.SetPlaceHolder("Email")
	pass := widget.NewPasswordEntry()
	pass.SetPlaceHolder("Пароль")

	loginBtn := widget.NewButton("Войти", func() {
		payload := map[string]string{"email": email.Text, "password": pass.Text}
		body, _ := json.Marshal(payload)
		resp, err := client.Post("http://localhost:8080/login", "application/json", bytes.NewBuffer(body))
		if err != nil || resp.StatusCode != 200 {
			dialog.ShowError(err, window)
			return
		}
		saveCookies() // сохраним куки
		history = nil
		showScreen(showMainScreen)
	})

	regBtn := widget.NewButton("Регистрация", func() { showScreen(showRegisterForm) })

	window.SetContent(container.NewVBox(
		widget.NewLabel("Вход"),
		email, pass, loginBtn, regBtn,
	))
}

func showRegisterForm() {
	email := widget.NewEntry()
	pass := widget.NewPasswordEntry()

	regBtn := widget.NewButton("Зарегистрироваться", func() {
		payload := map[string]string{"email": email.Text, "password": pass.Text}
		body, _ := json.Marshal(payload)
		resp, err := client.Post("http://localhost:8080/register", "application/json", bytes.NewBuffer(body))
		if err != nil || resp.StatusCode != 200 {
			dialog.ShowError(err, window)
			return
		}
		dialog.ShowInformation("Успех", "Аккаунт создан!", window)
		goBack()
	})

	backBtn := widget.NewButton("Назад", goBack)

	window.SetContent(container.NewBorder(
		nil, backBtn, nil, nil,
		container.NewVBox(
			widget.NewLabel("Регистрация"),
			email, pass, regBtn,
		),
	))
}

func showMainScreen() {
	filmsBtn := widget.NewButton("Фильмы", func() { showScreen(showFilmsList) })
	cinemasBtn := widget.NewButton("Кинотеатры", func() { showScreen(showCinemasList) })
	sessionsBtn := widget.NewButton("Создать сеанс", func() { showScreen(showCreateFilmSessionForm) })
	hallsBtn := widget.NewButton("Залы и места", func() { showScreen(showHallsList) })
	createHallBtn := widget.NewButton("Создать зал", func() { showScreen(showCreateHallForm) })
	adminBtn := widget.NewButton("Админка", func() { showScreen(showAdminPanel) })

	logoutBtn := widget.NewButton("Выйти", func() {
		_, err := client.Post("http://localhost:8080/logout", "application/json", nil)
		if err != nil {
			return
		}
		err = os.Remove(cookieFile)
		if err != nil {
			return
		} // из базы удаляем сессии и удаляем куки с диска
		history = nil
		showScreen(showLoginForm)
	})

	window.SetContent(container.NewVBox(
		widget.NewLabel("Главное меню"),
		filmsBtn,
		cinemasBtn,
		sessionsBtn,
		hallsBtn,
		createHallBtn,
		adminBtn,
		logoutBtn,
	))
}

//func showFilmsList() {
//	resp, _ := client.Get("http://localhost:8080/sessions")
//	var sessions []map[string]interface{}
//	json.NewDecoder(resp.Body).Decode(&sessions)
//	resp.Body.Close()
//
//	list := widget.NewList(
//		func() int { return len(sessions) },
//		func() fyne.CanvasObject { return widget.NewLabel("") },
//		func(i widget.ListItemID, o fyne.CanvasObject) {
//			title := sessions[i]["film_title"].(string)
//			time := sessions[i]["start_time"].(string)
//			o.(*widget.Label).SetText(fmt.Sprintf("%s — %s", title, time))
//		},
//	)
//
//	list.OnSelected = func(id widget.ListItemID) {
//		sess := sessions[id]
//		sessionID := int(sess["id"].(float64))
//		showScreen(func() { showHall(sessionID) })
//	}
//
//	backBtn := widget.NewButton("Назад", goBack)
//	window.SetContent(container.NewBorder(nil, backBtn, nil, nil, list))
//}

func showAdminPanel() {

	filmBtn := widget.NewButton("Добавить фильм", func() { showScreen(showCreateFilmForm) })
	cinemaBtn := widget.NewButton("Создать кинотеатр", func() { showScreen(showCreateCinemaForm) })
	sessionFilmBtn := widget.NewButton("Привязать фильм к кинотеатру", func() { showScreen(showCreateFilmSessionForm) })
	backBtn := widget.NewButton("Назад", goBack)

	window.SetContent(container.NewBorder(
		nil, backBtn, nil, nil,
		container.NewVBox(
			widget.NewLabel("Админка"),
			filmBtn,
			cinemaBtn,
			sessionFilmBtn,
		),
	))
}

func showHall(sessionID int) {
	resp, _ := client.Get(fmt.Sprintf("http://localhost:8080/seats?session=%d", sessionID))
	var seats []struct {
		ID     int  `json:"id"`
		Row    int  `json:"row"`
		Col    int  `json:"col"`
		Booked bool `json:"booked"`
	}
	err := json.NewDecoder(resp.Body).Decode(&seats)
	if err != nil {
		return
	}
	err = resp.Body.Close()
	if err != nil {
		return
	}

	selected := make(map[int]bool)
	var selectedIDs []int

	grid := container.NewGridWithColumns(10) // 10 мест в ряду

	for _, s := range seats {
		btn := widget.NewButton("", func(id int) func() {
			return func() {
				if s.Booked {
					return
				}
				selected[id] = !selected[id]
				selectedIDs = nil
				for id := range selected {
					if selected[id] {
						selectedIDs = append(selectedIDs, id)
					}
				}
				showHall(sessionID)
			}
		}(s.ID))

		if s.Booked {
			btn.Importance = widget.DangerImportance // красный
		} else if selected[s.ID] {
			btn.Importance = widget.HighImportance // синий
		} else {
			btn.Importance = widget.SuccessImportance // зелёный
		}

		btn.SetText(fmt.Sprintf("%d-%d", s.Row, s.Col)) // ← Row-Col
		grid.Add(btn)
	}

	bookBtn := widget.NewButton("Забронировать", func() {
		if len(selectedIDs) == 0 {
			dialog.ShowInformation("Ошибка", "Выберите места", window)
			return
		}
		payload := map[string]interface{}{
			"session_id": sessionID,
			"seat_ids":   selectedIDs,
		}
		body, _ := json.Marshal(payload)
		resp, _ := client.Post("http://localhost:8080/bookings", "application/json", bytes.NewBuffer(body))
		if resp.StatusCode == 200 {
			dialog.ShowInformation("Успех", "Места забронированы!", window)
			showScreen(showFilmsList)
		} else {
			dialog.ShowError(fmt.Errorf("ошибка бронирования"), window)
		}
	})

	backBtn := widget.NewButton("Назад", goBack)

	window.SetContent(container.NewBorder(
		container.NewHBox(
			widget.NewLabel(fmt.Sprintf("Сеанс ID: %d", sessionID)),
			layout.NewSpacer(),
			bookBtn,
		),
		backBtn,
		nil, nil,
		container.NewScroll(grid),
	))
}

// Методы создания сущностей
func showCreateCinemaForm() {
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

	createBtn := widget.NewButton("Создать", func() {
		totalSeats, err := strconv.Atoi(totalSeatsEntry.Text)
		fmt.Println("Count mest: ", totalSeats)
		if err != nil {
			dialog.ShowError(fmt.Errorf("некорректное количество мест"), window)
			return
		}

		payload := map[string]interface{}{
			"name":        nameEntry.Text,
			"address":     addressEntry.Text,
			"city":        cityEntry.Text,
			"phone":       phoneEntry.Text,
			"total_seats": totalSeats,
			"poster":      posterEntry.Text,
		}

		body, _ := json.Marshal(payload)
		log.Printf("%s", string(body))
		req, _ := http.NewRequest("POST", "http://localhost:8080/admin/cinemas", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			dialog.ShowError(fmt.Errorf("ошибка сети: %v", err), window)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 201 {
			dialog.ShowError(fmt.Errorf("ошибка сервера: %d", resp.StatusCode), window)
			return
		}

		// Безопасное чтение id
		var created map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
			dialog.ShowError(fmt.Errorf("неверный ответ сервера"), window)
			return
		}

		_, ok := created["ID"]
		if !ok {
			dialog.ShowError(fmt.Errorf("в ответе нет поля id"), window)
			return
		}

		//var cinemaID int
		//switch v := idVal.(type) {
		//case float64:
		//	cinemaID = int(v)
		//case int:
		//	cinemaID = v
		//case int64:
		//	cinemaID = int(v)
		//default:
		//	dialog.ShowError(fmt.Errorf("id имеет неверный тип"), window)
		//	return
		//}

		//// Генерация зала
		//hallPayload := map[string]interface{}{"cinema_id": cinemaID}
		//hallBody, _ := json.Marshal(hallPayload)
		//hallReq, _ := http.NewRequest("POST", "http://localhost:8080/admin/halls/create", bytes.NewBuffer(hallBody))
		//hallReq.Header.Set("Content-Type", "application/json")
		//
		//hallResp, hallErr := client.Do(hallReq)
		//if hallErr != nil || hallResp.StatusCode != 201 {
		//	log.Printf("Ошибка генерации зала: %v %d", hallErr, hallResp.StatusCode)
		//	// можно показать предупреждение, но не обязательно прерывать
		//}
		//
		//if hallResp != nil {
		//	hallResp.Body.Close()
		//}

		dialog.ShowInformation("Успех", "Кинотеатр создан!", window)
		goBack()
	})

	backBtn := widget.NewButton("Назад", goBack)

	window.SetContent(container.NewBorder(
		nil, backBtn, nil, nil,
		container.NewVBox(
			widget.NewLabel("Создать кинотеатр"),
			widget.NewLabel("Название:"), nameEntry,
			widget.NewLabel("Адрес:"), addressEntry,
			widget.NewLabel("Город:"), cityEntry,
			widget.NewLabel("Телефон:"), phoneEntry,
			widget.NewLabel("Мест всего:"), totalSeatsEntry,
			widget.NewLabel("Постер:"), posterEntry,
			createBtn,
		),
	))
}

func showCreateFilmForm() {
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Название фильма")

	posterEntry := widget.NewEntry()
	posterEntry.SetPlaceHolder("URL постера")

	descEntry := widget.NewMultiLineEntry()
	descEntry.SetPlaceHolder("Описание")

	durationEntry := widget.NewEntry()
	durationEntry.SetPlaceHolder("Продолжительность (мин)")

	createBtn := widget.NewButton("Создать", func() {
		duration, _ := strconv.Atoi(durationEntry.Text)
		payload := map[string]interface{}{
			"title":       titleEntry.Text,
			"poster":      posterEntry.Text,
			"description": descEntry.Text,
			"duration":    duration,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", "http://localhost:8080/admin/films", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 201 {
			dialog.ShowError(fmt.Errorf("ошибка: %v", err), window)
			return
		}
		dialog.ShowInformation("Успех", "Фильм создан!", window)
		goBack()
	})

	backBtn := widget.NewButton("Назад", goBack)

	window.SetContent(container.NewBorder(
		nil, backBtn, nil, nil,
		container.NewVBox(
			widget.NewLabel("Создать фильм"),
			widget.NewLabel("Название:"), titleEntry,
			widget.NewLabel("Постер:"), posterEntry,
			widget.NewLabel("Описание:"), descEntry,
			widget.NewLabel("Длительность:"), durationEntry,
			createBtn,
		),
	))
}

type TimePair struct {
	Date    *widget.DateEntry
	Time    *widget.Entry
	EndTime *widget.Entry
}

func showCreateFilmSessionForm() {
	filmsResp, _ := client.Get("http://localhost:8080/films")
	var films []map[string]interface{}
	json.NewDecoder(filmsResp.Body).Decode(&films)
	filmsResp.Body.Close()

	cinemasResp, _ := client.Get("http://localhost:8080/admin/cinemas/list")
	var cinemas []map[string]interface{}
	json.NewDecoder(cinemasResp.Body).Decode(&cinemas)
	cinemasResp.Body.Close()

	filmOptions := []string{}
	filmIDs := []int{}
	for _, f := range films {
		title := f["title"].(string)
		id := int(f["id"].(float64))
		filmOptions = append(filmOptions, title)
		filmIDs = append(filmIDs, id)
	}
	filmSelect := widget.NewSelect(filmOptions, nil)

	cinemaOptions := []string{}
	cinemaIDs := []int{}
	for _, c := range cinemas {
		name := c["name"].(string)
		id := int(c["id"].(float64))
		cinemaOptions = append(cinemaOptions, name)
		cinemaIDs = append(cinemaIDs, id)
	}

	multiCinemaSelect := widget.NewCheckGroup(cinemaOptions, nil)

	timesContainer := container.NewVBox()
	var timePairs []TimePair // предполагается, что TimePair имеет Date *widget.DateEntry, Time, EndTime *widget.Entry

	updateTimes := func() {
		timesContainer.Objects = nil
		timePairs = nil
		selected := multiCinemaSelect.Selected
		for _, name := range selected {
			dateEntry := widget.NewDateEntry()
			timeStart := widget.NewEntry()
			timeStart.SetPlaceHolder("Начало (HH:MM)")
			timeStart.OnChanged = func(s string) {
				var formatted string
				for i, c := range s {
					if i == 2 {
						formatted += ":"
					}
					if c >= '0' && c <= '9' {
						formatted += string(c)
					}
					if len(formatted) == 5 {
						break
					}
				}
				timeStart.SetText(formatted)
			}

			timeEnd := widget.NewEntry()
			timeEnd.SetPlaceHolder("Окончание (HH:MM)")
			timeEnd.OnChanged = func(s string) {
				var formatted string
				for i, c := range s {
					if i == 2 {
						formatted += ":"
					}
					if c >= '0' && c <= '9' {
						formatted += string(c)
					}
					if len(formatted) == 5 {
						break
					}
				}
				timeEnd.SetText(formatted)
			}

			timeHBox := container.NewGridWithColumns(3, dateEntry, timeStart, timeEnd)
			timesContainer.Add(widget.NewLabel(name + ":"))
			timesContainer.Add(timeHBox)

			timePairs = append(timePairs, TimePair{
				Date:    dateEntry,
				Time:    timeStart,
				EndTime: timeEnd,
			})
		}
		timesContainer.Refresh()
	}

	multiCinemaSelect.OnChanged = func(selected []string) {
		updateTimes()
	}

	createBtn := widget.NewButton("Создать сеансы", func() {
		if filmSelect.Selected == "" {
			dialog.ShowError(errors.New("выберите фильм"), window)
			return
		}
		filmID := filmIDs[filmSelect.SelectedIndex()]

		selected := multiCinemaSelect.Selected
		if len(selected) == 0 {
			dialog.ShowError(errors.New("выберите кинотеатры"), window)
			return
		}

		for i, name := range selected {
			if i >= len(timePairs) {
				dialog.ShowError(fmt.Errorf("укажите время для %s", name), window)
				return
			}

			pair := timePairs[i]

			// 3-й путь — используем .Date() из виджета
			selectedDate := pair.Date // time.Time
			if selectedDate.Date.IsZero() {
				dialog.ShowError(fmt.Errorf("укажите дату для %s", name), window)
				return
			}

			startStr := pair.Time.Text
			if startStr == "" {
				dialog.ShowError(fmt.Errorf("укажите время начала для %s", name), window)
				return
			}
			startTime, err := time.Parse("15:04", startStr)
			if err != nil {
				dialog.ShowError(fmt.Errorf("неверное время начала для %s", name), window)
				return
			}

			endStr := pair.EndTime.Text
			if endStr == "" {
				dialog.ShowError(fmt.Errorf("укажите время окончания для %s", name), window)
				return
			}
			endTime, err := time.Parse("15:04", endStr)
			if err != nil {
				dialog.ShowError(fmt.Errorf("неверное время окончания для %s", name), window)
				return
			}

			if endTime.Before(startTime) {
				dialog.ShowError(fmt.Errorf("время окончания раньше начала для %s", name), window)
				return
			}

			cinemaIdx := -1
			for j, opt := range cinemaOptions {
				if opt == name {
					cinemaIdx = j
					break
				}
			}
			if cinemaIdx == -1 {
				continue
			}
			cinemaID := cinemaIDs[cinemaIdx]

			// Формируем полное время начала
			tStart := time.Date(
				selectedDate.Date.Year(),
				selectedDate.Date.Month(),
				selectedDate.Date.Day(),
				startTime.Hour(),
				startTime.Minute(),
				0, 0,
				time.Local, // или UTC — зависит от вашей логики
			)

			tEnd := time.Date(
				selectedDate.Date.Year(),
				selectedDate.Date.Month(),
				selectedDate.Date.Day(),
				endTime.Hour(),
				endTime.Minute(),
				0, 0,
				time.Local,
			)

			payload := map[string]interface{}{
				"film_id":    filmID,
				"cinema_id":  cinemaID,
				"start_time": tStart.Format("2006-01-02 15:04:05"),
				"end_time":   tEnd.Format("2006-01-02 15:04:05"),
			}

			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest("POST", "http://localhost:8080/admin/film/sessions", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil || resp.StatusCode != http.StatusOK {
				log.Printf("Ошибка создания сеанса для %s: %v %d", name, err, resp.StatusCode)
			}
			if resp != nil {
				resp.Body.Close()
			}
		}

		dialog.ShowInformation("Успех", "Сеансы созданы!", window)
		goBack()
	})

	backBtn := widget.NewButton("Назад", goBack)

	window.SetContent(container.NewBorder(
		nil, backBtn, nil, nil,
		container.NewVBox(
			widget.NewLabel("Создать сеансы"),
			widget.NewLabel("Фильм:"), filmSelect,
			widget.NewLabel("Кинотеатры:"), multiCinemaSelect,
			widget.NewLabel("Времена:"), timesContainer,
			createBtn,
		),
	))

	updateTimes()
}

//	func showFilmsList() {
//		resp, _ := client.Get("http://localhost:8080/films")
//		var sessions []map[string]interface{}
//		err := json.NewDecoder(resp.Body).Decode(&sessions)
//		if err != nil {
//			return
//		}
//		fmt.Println(sessions)
//		err = resp.Body.Close()
//		if err != nil {
//			return
//		}
//
//		//TODO: Заменить все
//		//now := time.Now()
//		upcoming := []map[string]interface{}{}
//		//for _, s := range sessions {
//		//	t, _ := time.Parse("2006-01-02 15:04:05", s["start_time"].(string))
//		//	if t.After(now) {
//		//		upcoming = append(upcoming, s)
//		//	}
//		//}
//
//		list := widget.NewList(
//			func() int { return len(upcoming) },
//			func() fyne.CanvasObject { return widget.NewLabel("") },
//			func(i widget.ListItemID, o fyne.CanvasObject) {
//				title := upcoming[i]["film_title"].(string)
//				//timeStr := upcoming[i]["start_time"].(string)
//				//o.(*widget.Label).SetText(fmt.Sprintf("%s — %s", title, timeStr))
//				o.(*widget.Label).SetText(fmt.Sprintf("%s", title))
//			},
//		)
//
//		list.OnSelected = func(id widget.ListItemID) {
//			sessionID := int(upcoming[id]["id"].(float64))
//			showScreen(func() { showHall(sessionID) })
//		}
//
//		backBtn := widget.NewButton("Назад", goBack)
//		window.SetContent(container.NewBorder(nil, backBtn, nil, nil, list))
//	}
//
// Все фильмы без учета кинотеатра.
func showFilmsList() {
	resp, _ := client.Get("http://localhost:8080/films")
	var films []map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&films)
	if err != nil {
		return
	}

	err = resp.Body.Close()
	if err != nil {
		return
	}

	list := widget.NewList(
		func() int { return len(films) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			title := films[i]["title"].(string)
			o.(*widget.Label).SetText(fmt.Sprintf("%s", title))
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		sessionID := int(films[id]["id"].(float64))
		showScreen(func() { showHall(sessionID) })
	}

	backBtn := widget.NewButton("Назад", goBack)
	window.SetContent(container.NewBorder(nil, backBtn, nil, nil, list))
}

// Добавь функцию для сеансов фильма
func showSessionsForFilm(filmID int) {
	resp, _ := client.Get(fmt.Sprintf("http://localhost:8080/sessions?film_id=%d", filmID))
	var sessions []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&sessions)
	resp.Body.Close()

	list := widget.NewList(
		func() int { return len(sessions) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			time := sessions[i]["start_time"].(string)
			o.(*widget.Label).SetText(time)
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		sessionID := int(sessions[id]["id"].(float64))
		showScreen(func() { showHall(sessionID) })
	}

	backBtn := widget.NewButton("Назад", goBack)
	window.SetContent(container.NewBorder(nil, backBtn, nil, nil, list))
}

// Кинотеатры

func showCinemasList() {
	resp, _ := client.Get("http://localhost:8080/admin/cinemas/list")

	//
	//body, _ := io.ReadAll(resp.Body)
	//log.Printf("JSON body: %s", body)
	//
	var cinemas []map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&cinemas)

	if err != nil {
		//log.Printf("JSON decode error: %v", err)
		return
	}

	err = resp.Body.Close()

	if err != nil {
		return
	}

	log.Printf("Cinemas count: %d", len(cinemas))

	list := widget.NewList(
		func() int { return len(cinemas) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			name := cinemas[i]["name"].(string)
			var count int
			if sessionCount, ok := cinemas[i]["session_count"].(float64); ok {
				count = int(sessionCount)
			} else {
				count = 0
			}
			o.(*widget.Label).SetText(fmt.Sprintf("%s (%d сеансов)", name, count))
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		cinemaID := int(cinemas[id]["id"].(float64))
		showScreen(func() { showSessionsForCinema(cinemaID, cinemas[id]["name"].(string)) })
	}

	backBtn := widget.NewButton("Назад", goBack)
	window.SetContent(container.NewBorder(nil, backBtn, nil, nil, list))
}

// Подфункция для сеансов кинотеатра
func showSessionsForCinema(cinemaID int, cinemaName string) {
	resp, _ := client.Get(fmt.Sprintf("http://localhost:8080/sessions?cinema_id=%d", cinemaID))
	var sessions []map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&sessions)
	if err != nil {
		return
	}
	err = resp.Body.Close()
	if err != nil {
		return
	}

	list := widget.NewList(
		func() int { return len(sessions) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			title := sessions[i]["film_title"].(string)
			duration := int(sessions[i]["duration"].(float64))
			o.(*widget.Label).SetText(fmt.Sprintf("%s (%d мин)", title, duration))
		},
	)

	backBtn := widget.NewButton("Назад", goBack)
	window.SetContent(container.NewBorder(
		widget.NewLabel(fmt.Sprintf("Фильмы в %s", cinemaName)),
		backBtn, nil, nil, list,
	))
}

// sessionsByFilmBtn := widget.NewButton("Сеансы по фильму", func() { showScreen(showSessionsByFilmForm) })
//
// // В container.NewVBox добавь: sessionsByFilmBtn,
//
// // Новая функция
//
//	func showSessionsByFilmForm() {
//		// Получаем фильмы
//		filmsResp, _ := client.Get("http://localhost:8080/films")
//		var films []map[string]interface{}
//		json.NewDecoder(filmsResp.Body).Decode(&films)
//		filmsResp.Body.Close()
//
//		filmSelect := widget.NewSelect([]string{}, nil)
//		filmIDs := []int{}
//		for _, f := range films {
//			title := f["title"].(string)
//			id := int(f["id"].(float64))
//			filmSelect.Options = append(filmSelect.Options, title)
//			filmIDs = append(filmIDs, id)
//		}
//
//		var sessionsList *widget.List
//		var sessions []map[string]interface{}
//
//		updateList := func() {
//			if filmSelect.Selected == "" {
//				return
//			}
//			filmID := filmIDs[filmSelect.SelectedIndex()]
//			resp, _ := client.Get(fmt.Sprintf("http://localhost:8080/sessions?film_id=%d", filmID))
//			json.NewDecoder(resp.Body).Decode(&sessions)
//			resp.Body.Close()
//
//			sessionsList.Length = func() int { return len(sessions) }
//			sessionsList.UpdateItem = func(i widget.ListItemID, o fyne.CanvasObject) {
//				s := sessions[i]
//				cinema := s["cinema_name"].(string) // предполагаем поле
//				time := s["start_time"].(string)
//				o.(*widget.Label).SetText(fmt.Sprintf("%s — %s", cinema, time))
//			}
//			sessionsList.Refresh()
//		}
//
//		filmSelect.OnChanged = func(s string) { updateList() }
//
//		sessionsList = widget.NewList(
//			func() int { return 0 },
//			func() fyne.CanvasObject { return widget.NewLabel("") },
//			func(i widget.ListItemID, o fyne.CanvasObject) {},
//		)
//
//		backBtn := widget.NewButton("Назад", goBack)
//
//		window.SetContent(container.NewBorder(
//			nil, backBtn, nil, nil,
//			container.NewVBox(
//				widget.NewLabel("Сеансы по фильму"),
//				widget.NewLabel("Выберите фильм:"), filmSelect,
//				sessionsList,
//			),
//		))
//	}
func showHallsList() {
	// Загрузка залов
	resp, err := client.Get("http://localhost:8080/admin/halls")
	if err != nil || resp.StatusCode != 200 {
		dialog.ShowError(fmt.Errorf("ошибка загрузки залов"), window)
		return
	}
	defer resp.Body.Close()

	var halls []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&halls); err != nil {
		dialog.ShowError(err, window)
		return
	}

	if len(halls) == 0 {
		window.SetContent(container.NewVBox(
			widget.NewLabel("Залы и места"),
			widget.NewLabel("Нет залов"),
			widget.NewButton("Назад", goBack),
		))
		return
	}

	hallOptions := []string{}
	hallIDs := []int{}
	for _, h := range halls {
		name := h["name"].(string)
		id := int(h["id"].(float64))
		hallOptions = append(hallOptions, name)
		hallIDs = append(hallIDs, id)
	}

	hallSelect := widget.NewSelect(hallOptions, nil)

	seatsContainer := container.NewVBox(widget.NewLabel("Выберите зал"))
	var loadSeats func()
	loadSeats = func() {
		if hallSelect.SelectedIndex() < 0 {
			seatsContainer.Objects = []fyne.CanvasObject{widget.NewLabel("Выберите зал")}
			seatsContainer.Refresh()
			return
		}
		hallID := hallIDs[hallSelect.SelectedIndex()]

		seats, err := fetchSeatsByHall(hallID)
		log.Printf("Seats loaded for hall %d: %d items", hallID, len(seats))
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		if len(seats) == 0 {
			seatsContainer.Objects = []fyne.CanvasObject{widget.NewLabel("Мест нет")}
			seatsContainer.Refresh()
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

		grid := container.NewGridWithRows(maxRow)

		for row := 1; row <= maxRow; row++ {
			rowGrid := container.NewGridWithColumns(maxCol)
			for col := 1; col <= maxCol; col++ {
				found := false
				for _, s := range seats {
					if s.Row == row && s.Number == col {
						btn := createSeatButton(s, selected, &selectedIDs, loadSeats)
						rowGrid.Add(btn)
						found = true
						break
					}
				}
				if !found {
					rowGrid.Add(layout.NewSpacer())
				}
			}
			grid.Add(container.NewMax(rowGrid))
		}

		// Без скролла — прямое содержимое
		seatsContainer.Objects = []fyne.CanvasObject{container.NewMax(grid)}
		seatsContainer.Refresh()
		log.Printf("Grid objects after refresh: %d", len(grid.Objects))
	}

	hallSelect.OnChanged = func(_ string) { loadSeats() }

	backBtn := widget.NewButton("Назад", goBack)

	window.SetContent(container.NewBorder(
		nil, backBtn, nil, nil,
		container.NewVBox(
			widget.NewLabel("Залы и места"),
			widget.NewLabel("Выберите зал:"), hallSelect,
			seatsContainer,
		),
	))

	loadSeats()
}

// Новая функция получения мест по залу
func fetchSeatsByHall(hallID int) ([]Seat, error) {
	log.Printf("Fetching seats for hall %d", hallID)
	resp, err := client.Get(fmt.Sprintf("http://localhost:8080/halls/%d/seats", hallID))
	log.Printf("Status: %d", resp.StatusCode)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server error: %d", resp.StatusCode)
	}

	var seats []Seat
	if err := json.NewDecoder(resp.Body).Decode(&seats); err != nil {
		return nil, err
	}
	return seats, nil
}
