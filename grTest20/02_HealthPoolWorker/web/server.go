package web

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
	"workerpool/metrics"
	"workerpool/workerPool"
)

// Server - структура веб-сервера
type Server struct {
	pool     *workerPool.WorkerPool
	metrics  *metrics.Metrics
	http     *http.Server
	mu       sync.RWMutex
	port     string
	handlers *Handlers
}

// Config - конфигурация сервера
type Config struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// DefaultConfig - возвращает конфигурацию по умолчанию
func DefaultConfig() Config {
	return Config{
		Port:            "8080",
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}
}

// NewServer - создает новый экземпляр сервера
func NewServer(pool *workerPool.WorkerPool, metrics *metrics.Metrics, config Config) *Server {
	if config.Port == "" {
		config.Port = "8080"
	}

	s := &Server{
		pool:    pool,
		metrics: metrics,
		port:    config.Port,
	}

	// Создаем обработчики
	s.handlers = NewHandlers(s)

	// Создаем HTTP сервер
	s.http = &http.Server{
		Addr:         ":" + config.Port,
		Handler:      s.setupRoutes(),
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	return s
}

// setupRoutes - настраивает маршруты
func (s *Server) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	// API эндпоинты
	mux.HandleFunc("/api/status", s.handlers.HandleStatus)
	mux.HandleFunc("/api/workers", s.handlers.HandleWorkers)
	mux.HandleFunc("/api/metrics", s.handlers.HandleMetrics)
	mux.HandleFunc("/api/queue", s.handlers.HandleQueue)
	mux.HandleFunc("/api/config", s.handlers.HandleConfig)
	mux.HandleFunc("/api/health", s.handlers.HandleHealth)
	mux.HandleFunc("/api/restart", s.handlers.HandleRestartWorker)

	// Главная страница
	mux.HandleFunc("/", s.handlers.HandleIndex)

	// Применяем middleware
	handler := s.loggingMiddleware(mux)
	handler = s.corsMiddleware(handler)
	handler = s.recoveryMiddleware(handler)

	return handler
}

// Start - запускает сервер
func (s *Server) Start() error {
	log.Printf("🌐 Web interface: http://localhost:%s", s.port)
	log.Printf("📊 API: http://localhost:%s/api/status", s.port)
	log.Printf("📋 Health: http://localhost:%s/api/health", s.port)

	log.Println("")
	log.Println("📟 Available API endpoints:")
	log.Println("  GET  /api/status   - Full system status")
	log.Println("  GET  /api/workers  - Worker information")
	log.Println("  POST /api/workers  - Set workers ({\"count\": N})")
	log.Println("  GET  /api/metrics  - Metrics")
	log.Println("  GET  /api/queue    - Queue status")
	log.Println("  GET  /api/config   - Configuration")
	log.Println("  POST /api/config   - Update config ({\"min\": N, \"max\": M})")
	log.Println("  GET  /api/health   - Health check")
	log.Println("  POST /api/restart  - Restart a worker")
	log.Println("")

	return s.http.ListenAndServe()
}

// Stop - останавливает сервер
func (s *Server) Stop() error {
	if s.http == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("🛑 Stopping web server...")
	return s.http.Shutdown(ctx)
}

// GetPool - возвращает пул воркеров
func (s *Server) GetPool() *workerPool.WorkerPool {
	return s.pool
}

// GetMetrics - возвращает метрики
func (s *Server) GetMetrics() *metrics.Metrics {
	return s.metrics
}

// GetPort - возвращает порт
func (s *Server) GetPort() string {
	return s.port
}

// ===== MIDDLEWARE =====

// loggingMiddleware - логирует все запросы
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		log.Printf("[%s] %s %s -> %d (%v)",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			wrapped.statusCode,
			duration.Round(time.Millisecond))
	})
}

// corsMiddleware - добавляет CORS заголовки
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// recoveryMiddleware - восстанавливает после паники
func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("🔥 Panic recovered: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// responseWriter - обертка для ResponseWriter
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
