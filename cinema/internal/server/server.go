// internal/server/server.go
package server

import (
	"cinema/internal/db"
	"cinema/internal/handlers"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"
)

func waitForDB(dbURL string) {
	for i := 0; i < 15; i++ {
		db, err := sql.Open("postgres", dbURL)
		if err != nil {
			log.Printf("DB connect attempt %d failed: %v", i+1, err)
			time.Sleep(1 * time.Second)
			continue
		}
		if err := db.Ping(); err != nil {
			db.Close()
			log.Printf("DB ping failed: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		db.Close()
		log.Println("Подключено к БД!")
		return
	}
	log.Fatal("Не удалось подключиться к БД после 15 попыток")
}

func Start(dbURL string) {
	// ← ЖДЁМ БД
	waitForDB(dbURL)

	dbConn := db.Connect(dbURL)
	repo := db.NewRepo(dbConn)
	mux := http.NewServeMux()

	// ← ПУБЛИЧНЫЙ МАРШРУТ: РЕГИСТРАЦИЯ
	mux.HandleFunc("/register", handlers.Register(repo))

	// Остальные маршруты
	mux.HandleFunc("/login", handlers.LoginHandler(repo))
	mux.HandleFunc("/logout", handlers.Logout(repo))

	mux.HandleFunc("/films", AuthMiddleware(repo, handlers.GetFilms(repo)).ServeHTTP)
	mux.HandleFunc("/book", AuthMiddleware(repo, handlers.Book(repo)).ServeHTTP)

	mux.HandleFunc("/admin/films", AuthMiddleware(repo, RoleMiddleware("admin", handlers.AddFilm(repo)).ServeHTTP))
	mux.HandleFunc("/admin/cinemas", AuthMiddleware(repo, RoleMiddleware("admin", handlers.AddCinema(repo)).ServeHTTP))
	mux.HandleFunc("/admin/cinemas/list", AuthMiddleware(repo, RoleMiddleware("admin", handlers.ListCinemas(repo))))

	mux.HandleFunc("/seats", AuthMiddleware(repo, handlers.GetSeats(repo)).ServeHTTP)

	mux.HandleFunc("/protected", AuthMiddleware(repo, handlers.Protected(repo)).ServeHTTP)

	mux.HandleFunc("/admin/sessions",
		AuthMiddleware(repo,
			RoleMiddleware("admin", handlers.CreateFilmSession(repo)).ServeHTTP))

	//mux.HandleFunc("/admin/cinemas/list", AuthMiddleware(repo, RoleMiddleware("admin", handlers.ListCinemas(repo))))

	mux.HandleFunc("/sessions", AuthMiddleware(repo, handlers.GetSessions(repo)))

	//mux.HandleFunc("POST /admin/halls/generate",
	//	AuthMiddleware(repo, RoleMiddleware("admin", handlers.GenerateHall(repo))))

	//mux.HandleFunc("/ticket", handlers.DownloadTicket(repo))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Сервер запущен на :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal("Server error: ", err)
	}
}
