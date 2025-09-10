package main

import (
	"database/sql"
	"fuelstation/internal/db"
	"fuelstation/internal/gui"
	"fuelstation/internal/processor"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	// Настройка логирования в run.log и консоль
	logFile, err := os.OpenFile("run.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Ошибка открытия файла логов: %v", err)
		os.Exit(1)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.Println("main: Начало работы приложения")

	// Подключение к базе данных
	dsn := "postgres://postgres:password@localhost:5454/fuelstation?sslmode=disable"
	log.Println("main: Попытка подключения к базе данных")
	dbConn, err := db.ConnectDB(dsn)
	if err != nil {
		log.Printf("Ошибка подключения к базе данных: %v", err)
		os.Exit(1)
	}
	defer dbConn.Close()
	log.Println("main: Подключение к базе данных успешно")

	// Применение миграций
	log.Println("main: Применение миграций")
	if err := applyMigrations(dbConn, dsn); err != nil {
		log.Printf("Ошибка применения миграций: %v", err)
		os.Exit(1)
	}
	log.Println("main: Миграции применены")

	// Создание приложения Fyne
	log.Println("main: Создание приложения Fyne")
	app := gui.NewFyneApp()

	// Создание GUI
	log.Println("main: Создание объекта Gui")
	guiInstance := gui.NewGui()

	// Канал для синхронизации
	ready := make(chan struct{})

	// Запуск обработки JSON файла
	log.Println("main: Запуск обработки JSON файла")
	go processor.ProcessJSONFile(dbConn, guiInstance, "fuel_data.json", ready)

	// Запуск GUI с передачей БД
	log.Println("main: Запуск GUI")
	if err := guiInstance.RunGui(app, ready, dbConn); err != nil {
		log.Printf("main: Ошибка запуска GUI: %v", err)
		os.Exit(1)
	}
}

func applyMigrations(db *sql.DB, dsn string) error {
	log.Printf("applyMigrations: Применение миграций из file://migrations с DSN: %s", dsn)
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Printf("Ошибка создания драйвера миграций: %v", err)
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		log.Printf("Ошибка инициализации миграций: %v", err)
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Printf("Ошибка выполнения миграций: %v", err)
		return err
	}
	log.Println("applyMigrations: Миграции успешно применены")
	return nil
}
