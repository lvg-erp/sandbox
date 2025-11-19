package main

import (
	"cinema/internal/server"
	"cinema/internal/ui"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"log"
	"os"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:pass@localhost:5464/cinema?sslmode=disable" // ← localhost!
	}

	go server.Start(dbURL)

	log.Println("GUI запущен...")

	a := app.New()
	w := a.NewWindow("Кинотеатр")
	w.Resize(fyne.NewSize(400, 600))

	ui.Start(w)
	w.ShowAndRun()
}

//func waitForDB(dbURL string) {
//	for i := 0; i < 15; i++ {
//		db, err := sql.Open("postgres", dbURL)
//		if err == nil {
//			if err := db.Ping(); err == nil {
//				db.Close()
//				return
//			}
//		}
//		time.Sleep(1 * time.Second)
//	}
//	log.Fatal("Не удалось подключиться к БД")
//}

//func main() {
//	//dbURL := os.Getenv("DB_URL")
//	//dbURL := "postgres://user:pass@localhost:5464/cinema?sslmode=disable"
//	////if err := db.Migrate(dbURL); err != nil {
//	////	panic(err)
//	////}
//
//	dbURL := os.Getenv("DATABASE_URL")
//	if dbURL == "" {
//		dbURL = "postgres://user:pass@db:5432/cinema?sslmode=disable"
//	}
//
//	go server.Start(dbURL)
//
//	a := app.New()
//	w := a.NewWindow("Кинотеатр")
//	w.Resize(fyne.NewSize(400, 600))
//
//	ui.Start(w)
//	w.ShowAndRun()
//}

//func main() {
//	log.Println("UI запущен — сервер: http://localhost:8080")
//
//	a := app.New()
//	w := a.NewWindow("Кинотеатр")
//	w.Resize(fyne.NewSize(400, 600))
//
//	ui.Start(w)
//	w.ShowAndRun()
//}
