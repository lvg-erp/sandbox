package main

import (
	"log"
	"messanger/internal/domain/service"
	"messanger/internal/handler"
	"messanger/internal/repository"
	postgresRepo "messanger/internal/repository/postgres"

	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Загрузка .env файла
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Подключение к БД
	db, err := repository.NewDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Инициализация репозиториев
	userRepo := postgresRepo.NewUserRepository(db.DB)
	chatRepo := postgresRepo.NewChatRepository(db.DB)
	messageRepo := postgresRepo.NewMessageRepository(db.DB)

	// Инициализация сервисов
	userService := service.NewUserService(userRepo)
	chatService := service.NewChatService(chatRepo, userRepo, messageRepo)
	messageService := service.NewMessageService(messageRepo, chatRepo, userRepo)

	// Инициализация WebSocket хендлера
	wsHandler := handler.NewWebSocketHandler(userService, chatService, messageService)

	// Настройка маршрутов
	http.HandleFunc("/ws", wsHandler.HandleWebSocket)
	http.HandleFunc("/", serveIndex)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
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
