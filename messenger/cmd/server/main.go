package main

import (
	"log"
	"messanger/internal/auth"
	"messanger/internal/domain/service"
	"messanger/internal/handler"
	"messanger/internal/middleware"
	"messanger/internal/repository"
	postgresRepo "messanger/internal/repository/postgres"

	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Подключение к БД
	db, err := repository.NewDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to database")

	// Инициализация репозиториев
	userRepo := postgresRepo.NewUserRepository(db.DB)
	chatRepo := postgresRepo.NewChatRepository(db.DB)
	messageRepo := postgresRepo.NewMessageRepository(db.DB)

	// Инициализация сервисов
	userService := service.NewUserService(userRepo)
	chatService := service.NewChatService(chatRepo, userRepo, messageRepo)
	messageService := service.NewMessageService(messageRepo, chatRepo, userRepo)

	// Инициализация JWT
	jwtConfig := auth.NewJWTConfig("your-secret-key-change-in-production")

	// Инициализация хендлеров
	wsHandler := handler.NewWebSocketHandler(userService, chatService, messageService, jwtConfig)
	authHandler := handler.NewAuthHandler(jwtConfig, userService)

	// Настройка маршрутов
	http.HandleFunc("/ws", wsHandler.HandleWebSocket)
	http.HandleFunc("/", serveIndex)

	// REST API - АВТОРИЗАЦИЯ
	http.HandleFunc("/api/auth/register", authHandler.Register) // ДОБАВЬТЕ ЭТУ СТРОКУ!
	http.HandleFunc("/api/auth/login", authHandler.Login)
	http.HandleFunc("/api/auth/refresh", authHandler.Refresh)

	// Защищенные API эндпоинты (пример)
	http.Handle("/api/protected", middleware.JWTAuth(jwtConfig)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.GetUserFromContext(r.Context())
		w.Write([]byte("Hello " + claims.Username + "!"))
	})))

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	log.Printf("Login endpoint: http://localhost:%s/api/auth/login", port)
	log.Printf("WebSocket endpoint: ws://localhost:%s/ws?token=<jwt_token>", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "./web/index.html")
}
