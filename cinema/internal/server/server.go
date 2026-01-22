// internal/server/server.go
package server

import (
	"cinema/application/usecases"
	"cinema/infrastructure"
	db2 "cinema/infrastructure/db"
	"cinema/infrastructure/http/handlers"
	"database/sql"
	"github.com/gorilla/mux"
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
			err := db.Close()
			if err != nil {
				return
			}
			log.Printf("DB ping failed: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		err = db.Close()
		if err != nil {
			return
		}
		log.Println("Подключено к БД!")
		return
	}
	log.Fatal("Не удалось подключиться к БД после 15 попыток")
}

func Start(dbURL string) {
	// ← ЖДЁМ БД
	waitForDB(dbURL)

	dbConn := db2.Connect(dbURL)
	repos := infrastructure.NewRepositories(dbConn)
	//mux := http.NewServeMux() --- не работает с параметрами
	muxRoute := mux.NewRouter()
	///

	// ← ПУБЛИЧНЫЙ МАРШРУТ: РЕГИСТРАЦИЯ
	muxRoute.HandleFunc("/register", handlers.Register(repos.UserRepo))

	// Остальные маршруты
	muxRoute.HandleFunc("/login", handlers.LoginHandler(repos.UserRepo, repos.SessionRepo))
	muxRoute.HandleFunc("/logout", handlers.Logout(repos.SessionRepo))

	//TODO
	//mux.HandleFunc("/book", AuthMiddleware(repos.UserRepo, handlers.Book(repos.BookingRepo)).ServeHTTP)

	authUC := &usecases.AuthUseCase{UserRepo: repos.UserRepo, SessionRepo: repos.SessionRepo}
	muxRoute.HandleFunc("/films", AuthMiddleware(authUC, handlers.GetListFilms(repos.FilmRepo)).ServeHTTP)

	muxRoute.HandleFunc("/admin/films", AuthMiddleware(authUC, RoleMiddleware("admin", handlers.AddFilm(repos.FilmRepo)).ServeHTTP))
	muxRoute.HandleFunc("/admin/cinemas", AuthMiddleware(authUC, RoleMiddleware("admin", handlers.AddCinema(repos.CinemaRepo)).ServeHTTP))
	//mux.HandleFunc("/admin/cinemas/list", AuthMiddleware(authUC, RoleMiddleware("admin", handlers.ListCinemas(repos.CinemaRepo))))
	//TODO
	//mux.HandleFunc("/seats", AuthMiddleware(repos.UserRepo, handlers.GetSeats(repos.SeatsRepository)).ServeHTTP)

	muxRoute.HandleFunc("/protected", AuthMiddleware(authUC, handlers.Protected().ServeHTTP)) //!!!!!!!!!!!!!!!!!!!!!

	//TODO
	muxRoute.HandleFunc("/admin/film/sessions",
		AuthMiddleware(authUC,
			RoleMiddleware("admin", handlers.CreateFilmSession(repos.SessionFilmRepository)).ServeHTTP))

	muxRoute.HandleFunc("/admin/cinemas/list", AuthMiddleware(authUC, RoleMiddleware("admin", handlers.ListCinemas(repos.CinemaRepo))))

	muxRoute.HandleFunc("/film/sessions", AuthMiddleware(authUC, handlers.GetFilmSessions(repos.SessionFilmRepository)))
	muxRoute.HandleFunc("/admin/halls/create", AuthMiddleware(authUC, RoleMiddleware("admin", handlers.CreateHall(repos.HallRepository)).ServeHTTP))
	muxRoute.HandleFunc("/admin/halls", AuthMiddleware(authUC, RoleMiddleware("admin", handlers.ListHalls(repos.HallRepository)).ServeHTTP))

	muxRoute.HandleFunc("/halls/{id}/seats", handlers.GetHallSeats(repos.SeatsRepository))
	//mux.HandleFunc("POST /admin/halls/generate",
	//	AuthMiddleware(repo, RoleMiddleware("admin", handlers.GenerateHall(repo))))

	//mux.HandleFunc("/ticket", handlers.DownloadTicket(repo))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Сервер запущен на :%s", port)
	if err := http.ListenAndServe(":"+port, muxRoute); err != nil {
		log.Fatal("Server error: ", err)
	}
}
