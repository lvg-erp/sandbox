package main

import (
	"database/sql"
	"fuelstation/internal/gui"
	"fuelstation/internal/processor"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	_ "github.com/lib/pq" // Драйвер PostgreSQL
	"image/color"
	"log"
	"time"
)

func main() {
	// Инициализация приложения Fyne
	a := app.New()
	g := gui.NewGui(a)
	p := processor.NewProcessor(g)

	// Подключение к PostgreSQL
	connStr := "user=postgres password=password dbname=fuelstation sslmode=disable host=localhost port=5454"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	// Проверка подключения
	err = db.Ping()
	if err != nil {
		log.Fatalf("Ошибка проверки подключения к базе данных: %v", err)
	}
	log.Printf("Успешное подключение к базе данных на %s", time.Now().Format("15:04:05"))

	// Объявление кнопок
	var fuelGiveButton1, fuelGetButton1, fuelGiveButton2, fuelGetButton2 *widget.Button

	// Кнопки для левой колонки (jarNumber=1)
	fuelGiveButton1 = widget.NewButton("Заправка", func() {
		if !p.IsJarActive("1") {
			err := p.EmulateQRScan("FuelGive", "Petrol", 50.0, "1")
			if err != nil {
				log.Printf("Ошибка эмуляции QR-кода (Заправка 1): %v", err)
			}
			fuelGiveButton1.Disable()
			fuelGetButton1.Disable()
			fuelGiveButton2.Disable()
			fuelGetButton2.Disable()
		}
	})
	fuelGetButton1 = widget.NewButton("Слив", func() {
		if !p.IsJarActive("1") {
			err := p.EmulateQRScan("FuelGet", "Diesel", 30.0, "1")
			if err != nil {
				log.Printf("Ошибка эмуляции QR-кода (Слив 1): %v", err)
			}
			fuelGiveButton1.Disable()
			fuelGetButton1.Disable()
			fuelGiveButton2.Disable()
			fuelGetButton2.Disable()
		}
	})
	//leftButtonContainer := container.NewHBox(fuelGiveButton1, fuelGetButton1)
	//leftButtonContainer := container.NewHBox(fuelGiveButton1, fuelGetButton1, layout.NewSpacer())
	leftButtonContainer := container.NewHBox(layout.NewSpacer(), fuelGiveButton1, fuelGetButton1)
	//leftButtonContainer := container.New(layout.NewGridLayout(1), fuelGiveButton1, fuelGetButton1)
	// Кнопки для правой колонки (jarNumber=2)
	fuelGiveButton2 = widget.NewButton("Заправка", func() {
		if !p.IsJarActive("2") {
			err := p.EmulateQRScan("FuelGive", "Petrol", 50.0, "2")
			if err != nil {
				log.Printf("Ошибка эмуляции QR-кода (Заправка 2): %v", err)
			}
			fuelGiveButton1.Disable()
			fuelGetButton1.Disable()
			fuelGiveButton2.Disable()
			fuelGetButton2.Disable()
		}
	})
	fuelGetButton2 = widget.NewButton("Слив", func() {
		if !p.IsJarActive("2") {
			err := p.EmulateQRScan("FuelGet", "Diesel", 30.0, "2")
			if err != nil {
				log.Printf("Ошибка эмуляции QR-кода (Слив 2): %v", err)
			}
			fuelGiveButton1.Disable()
			fuelGetButton1.Disable()
			fuelGiveButton2.Disable()
			fuelGetButton2.Disable()
		}
	})
	rightButtonContainer := container.NewHBox(fuelGiveButton2, fuelGetButton2)
	//rightButtonContainer := container.New(layout.NewGridLayout(1), fuelGiveButton2, fuelGetButton2)
	// Создание разделительной линии
	separator := canvas.NewLine(color.White)
	separator.StrokeWidth = 1
	separator.Position1 = fyne.NewPos(0, 0)
	separator.Position2 = fyne.NewPos(0, 900) // Вертикальная линия от 0 до 800 по Y
	separator.Refresh()                       // Обновление отображения

	background := canvas.NewRectangle(color.Gray{Y: 128})
	background.Resize(fyne.NewSize(800, 800))
	// Обновление BottomSection с равной шириной колонок и разделителем
	leftColumn := container.NewVBox(leftButtonContainer, g.LeftSection.Content)
	rightColumn := container.NewVBox(rightButtonContainer, g.RightSection.Content)
	g.BottomSection = container.NewHBox(
		layout.NewSpacer(),
		leftColumn,
		container.NewWithoutLayout(separator), // Прямое размещение линии
		rightColumn,
		layout.NewSpacer(),
	)

	// Установка основного содержимого
	g.MainContent = container.NewVBox(g.TopSection.Content, g.BottomSection)
	//g.MainContent = container.NewWithoutLayout(background, container.NewVBox(g.TopSection.Content, g.BottomSection))
	// Добавление функции обновления состояния кнопок
	updateButtons := func() {
		fyne.Do(func() {
			fuelGiveButton1.Enable()
			fuelGetButton1.Enable()
			fuelGiveButton2.Enable()
			fuelGetButton2.Enable()
		})
	}

	// Подписка на изменение состояния активности
	go func() {
		for {
			if !p.IsJarActive("1") && !p.IsJarActive("2") {
				updateButtons()
			}
			time.Sleep(1 * time.Second) // Проверка каждую секунду
		}
	}()

	// Установка размера окна 800x800
	g.Window.Resize(fyne.NewSize(800, 800))

	// Запуск GUI
	g.RunGui()
}

//
//import (
//	"database/sql"
//	"log"
//
//	"fuelstation/internal/gui"
//	"fuelstation/internal/processor"
//
//	"github.com/golang-migrate/migrate/v4"
//	_ "github.com/golang-migrate/migrate/v4/database/postgres"
//	_ "github.com/golang-migrate/migrate/v4/source/file"
//	_ "github.com/lib/pq"
//)
//
//func applyMigrations(dsn string) error {
//	log.Println("applyMigrations: Применение миграций из file://migrations с DSN:", dsn)
//	m, err := migrate.New("file://migrations", dsn)
//	if err != nil {
//		return err
//	}
//	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
//		return err
//	}
//	log.Println("applyMigrations: Миграции успешно применены")
//	return nil
//}
//
//func main() {
//	log.Println("main: Начало работы приложения")
//
//	// Подключение к базе данных
//	dsn := "postgres://postgres:password@localhost:5454/fuelstation?sslmode=disable"
//	log.Println("main: Попытка подключения к базе данных")
//	dbConn, err := sql.Open("postgres", dsn)
//	if err != nil {
//		log.Fatalf("main: Ошибка подключения к базе данных: %v", err)
//	}
//	defer dbConn.Close()
//	log.Println("main: Проверка подключения к базе данных")
//	if err := dbConn.Ping(); err != nil {
//		log.Fatalf("main: Не удалось подключиться к базе данных: %v", err)
//	}
//	log.Println("main: Подключение к базе данных успешно")
//
//	// Применение миграций
//	log.Println("main: Применение миграций")
//	if err := applyMigrations(dsn); err != nil {
//		log.Fatalf("main: Ошибка применения миграций: %v", err)
//	}
//	log.Println("main: Миграции применены")
//
//	// Создание приложения Fyne
//	log.Println("main: Создание приложения Fyne")
//	app := gui.NewFyneApp()
//
//	// Создание объекта GUI
//	log.Println("main: Создание объекта Gui")
//	g := gui.NewGui()
//
//	// Установка ProcessorFunc
//	g.SetProcessorFunc(processor.ProcessJSONFile)
//
//	// Создание канала ready
//	ready := make(chan struct{})
//
//	// Запуск GUI
//	log.Println("main: Запуск GUI")
//	if err := g.RunGui(app, ready, dbConn, processor.ProcessJSONFile); err != nil {
//		log.Fatalf("main: Ошибка запуска GUI: %v", err)
//	}
//}
