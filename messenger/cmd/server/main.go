package main

import (
	"context"
	"errors"
	"log"
	"messanger/internal/auth"
	"messanger/internal/domain/service"
	"messanger/internal/handler"
	"messanger/internal/middleware"
	"messanger/internal/repository"
	postgresRepo "messanger/internal/repository/postgres"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	tokenRepo := postgresRepo.NewTokenRepository(db.DB)

	// Инициализация JWT
	jwtConfig := auth.NewJWTConfig(jwtSecret())

	// Инициализация сервисов
	userService := service.NewUserService(userRepo)
	chatService := service.NewChatService(chatRepo, userRepo, messageRepo)
	messageService := service.NewMessageService(messageRepo, chatRepo, userRepo)
	authService := service.NewAuthService(userRepo, tokenRepo, jwtConfig.RefreshExpiry)

	// Инициализация хендлеров
	wsHandler := handler.NewWebSocketHandler(userService, chatService, messageService, jwtConfig)
	authHandler := handler.NewAuthHandler(jwtConfig, userService, authService)

	// Настройка маршрутов
	mux := http.NewServeMux()

	// Статические файлы
	fileServer := http.FileServer(http.Dir("./web"))
	// Отдаём статику без кэширования, чтобы правки фронтенда подхватывались сразу
	noCache := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			next.ServeHTTP(w, r)
		})
	}
	mux.Handle("/", noCache(fileServer))
	mux.Handle("/styles.css", noCache(fileServer))
	mux.Handle("/app.js", noCache(fileServer))

	// API маршруты
	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)
	mux.HandleFunc("/api/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("/api/auth/logout", authHandler.Logout)

	// WebSocket
	mux.HandleFunc("/ws", wsHandler.HandleWebSocket)

	// Защищенные API эндпоинты (пример)
	mux.Handle("/api/protected", middleware.JWTAuth(jwtConfig)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.GetUserFromContext(r.Context())
		w.Write([]byte("Hello " + claims.Username + "!"))
	})))

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// ReadTimeout/WriteTimeout не ставим: они ломают долгоживущие WebSocket-соединения
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Graceful shutdown по Ctrl+C / SIGTERM
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		log.Println("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Shutdown error: %v", err)
		}
	}()

	log.Printf("Server starting on :%s", port)
	log.Printf("Login endpoint: http://localhost:%s/api/auth/login", port)
	log.Printf("WebSocket endpoint: ws://localhost:%s/ws?token=<jwt_token>", port)

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
	log.Println("Server stopped")
}

// jwtSecret читает секрет из окружения. Без JWT_SECRET работает небезопасный
// дефолт — только для локальной разработки.
func jwtSecret() string {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret
	}
	log.Println("⚠️  JWT_SECRET is not set, using insecure default. Set JWT_SECRET in .env")
	return "your-secret-key-change-in-production"
}
