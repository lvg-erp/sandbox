package main

import (
	"database/sql"
	"log"

	"fuelstation/internal/gui"
	"fuelstation/internal/processor"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func applyMigrations(dsn string) error {
	log.Println("applyMigrations: Применение миграций из file://migrations с DSN:", dsn)
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	log.Println("applyMigrations: Миграции успешно применены")
	return nil
}

func main() {
	log.Println("main: Начало работы приложения")

	// Подключение к базе данных
	dsn := "postgres://postgres:password@localhost:5454/fuelstation?sslmode=disable"
	log.Println("main: Попытка подключения к базе данных")
	dbConn, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("main: Ошибка подключения к базе данных: %v", err)
	}
	defer dbConn.Close()
	log.Println("main: Проверка подключения к базе данных")
	if err := dbConn.Ping(); err != nil {
		log.Fatalf("main: Не удалось подключиться к базе данных: %v", err)
	}
	log.Println("main: Подключение к базе данных успешно")

	// Применение миграций
	log.Println("main: Применение миграций")
	if err := applyMigrations(dsn); err != nil {
		log.Fatalf("main: Ошибка применения миграций: %v", err)
	}
	log.Println("main: Миграции применены")

	// Создание приложения Fyne
	log.Println("main: Создание приложения Fyne")
	app := gui.NewFyneApp()

	// Создание объекта GUI
	log.Println("main: Создание объекта Gui")
	g := gui.NewGui()

	// Установка ProcessorFunc
	g.SetProcessorFunc(processor.ProcessJSONFile)

	// Создание канала ready
	ready := make(chan struct{})

	// Запуск GUI
	log.Println("main: Запуск GUI")
	if err := g.RunGui(app, ready, dbConn, processor.ProcessJSONFile); err != nil {
		log.Fatalf("main: Ошибка запуска GUI: %v", err)
	}
}
