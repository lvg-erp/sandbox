package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

// Handlers - структура для всех обработчиков
type Handlers struct {
	server *Server
	tmpl   *template.Template
}

// NewHandlers - создает новые обработчики
func NewHandlers(server *Server) *Handlers {
	var tmpl *template.Template

	// Пытаемся загрузить шаблон из встроенных файлов
	templateData, err := GetTemplate()
	if err != nil {
		log.Printf("⚠️ Error loading embedded template: %v", err)
		// Используем шаблон по умолчанию (запасной вариант)
		tmpl = template.Must(template.New("index").Parse(getFallbackTemplate()))
	} else {
		// Парсим шаблон из встроенного файла
		tmpl = template.Must(template.New("index").Parse(string(templateData)))
		log.Println("✅ Template loaded from embedded file")
	}

	return &Handlers{
		server: server,
		tmpl:   tmpl,
	}
}

// HandleIndex - главная страница
func (h *Handlers) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	data := map[string]interface{}{
		"Port":    h.server.GetPort(),
		"Time":    time.Now().Format("2006-01-02 15:04:05"),
		"Workers": h.server.GetPool().GetWorkerCount(),
		"Queue":   h.server.GetPool().QueueSize(),
	}

	if err := h.tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// getFallbackTemplate - запасной шаблон на случай, если не удалось загрузить файл
func getFallbackTemplate() string {
	return `<!DOCTYPE html>
<html>
<head><title>Worker Pool</title></head>
<body>
    <h1>⚡ Worker Pool Monitor</h1>
    <p>Workers: {{.Workers}}, Queue: {{.Queue}}</p>
    <p>Time: {{.Time}}</p>
    <p>API: /api/status, /api/workers</p>
</body>
</html>`
}

// ===== ОСНОВНЫЕ ОБРАБОТЧИКИ =====

// HandleStatus - возвращает полный статус системы
func (h *Handlers) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.errorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := h.getStatus()
	h.jsonResponse(w, status, http.StatusOK)
}

// HandleWorkers - управление воркерами
func (h *Handlers) HandleWorkers(w http.ResponseWriter, r *http.Request) {
	pool := h.server.GetPool()

	switch r.Method {
	case http.MethodGet:
		// Получить текущее состояние воркеров
		status := h.getStatus()
		h.jsonResponse(w, map[string]interface{}{
			"current": status.Workers.Current,
			"target":  status.Workers.Target,
			"min":     status.Workers.Min,
			"max":     status.Workers.Max,
		}, http.StatusOK)

	case http.MethodPost:
		// Установить количество воркеров
		var req struct {
			Count int `json:"count"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.errorResponse(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Count <= 0 {
			h.errorResponse(w, "Count must be positive", http.StatusBadRequest)
			return
		}

		if err := pool.SetWorkers(req.Count); err != nil {
			h.errorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}

		h.jsonResponse(w, map[string]interface{}{
			"status":  "ok",
			"message": fmt.Sprintf("Workers set to %d", req.Count),
			"current": pool.GetWorkerCount(),
		}, http.StatusOK)

	default:
		h.errorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleMetrics - возвращает метрики
func (h *Handlers) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.errorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics := h.server.GetMetrics()

	h.jsonResponse(w, map[string]interface{}{
		"submitted": metrics.TasksSubmitted.Load(),
		"completed": metrics.TasksCompleted.Load(),
		"failed":    metrics.TasksFailed.Load(),
		"panicked":  metrics.TasksPanicked.Load(),
		"restarted": metrics.WorkersRestarted.Load(),
		"uptime":    time.Since(metrics.StartTime).String(),
		"timestamp": time.Now().Unix(),
	}, http.StatusOK)
}

// HandleQueue - возвращает состояние очереди
func (h *Handlers) HandleQueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.errorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pool := h.server.GetPool()
	queueSize := pool.QueueSize()
	capacity := cap(pool.Tasks())

	var utilization float64
	if capacity > 0 {
		utilization = float64(queueSize) / float64(capacity) * 100
	}

	h.jsonResponse(w, map[string]interface{}{
		"size":        queueSize,
		"capacity":    capacity,
		"utilization": fmt.Sprintf("%.1f%%", utilization),
		"available":   capacity - queueSize,
		"timestamp":   time.Now().Unix(),
	}, http.StatusOK)
}

// HandleConfig - управление конфигурацией
func (h *Handlers) HandleConfig(w http.ResponseWriter, r *http.Request) {
	pool := h.server.GetPool()

	switch r.Method {
	case http.MethodGet:
		h.jsonResponse(w, map[string]interface{}{
			"min": pool.GetMinWorkers(),
			"max": pool.GetMaxWorkers(),
		}, http.StatusOK)

	case http.MethodPost:
		var req struct {
			Min int `json:"min"`
			Max int `json:"max"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.errorResponse(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Min > 0 {
			pool.SetMinWorkers(req.Min)
		}
		if req.Max > 0 {
			pool.SetMaxWorkers(req.Max)
		}

		h.jsonResponse(w, map[string]interface{}{
			"status": "ok",
			"min":    pool.GetMinWorkers(),
			"max":    pool.GetMaxWorkers(),
		}, http.StatusOK)

	default:
		h.errorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleHealth - проверка здоровья
func (h *Handlers) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.errorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pool := h.server.GetPool()
	metrics := h.server.GetMetrics()

	status := "healthy"
	statusCode := http.StatusOK

	// Проверяем состояние
	if pool.IsShuttingDown() {
		status = "shutting_down"
		statusCode = http.StatusServiceUnavailable
	}

	h.jsonResponse(w, map[string]interface{}{
		"status":    status,
		"workers":   pool.GetWorkerCount(),
		"queue":     pool.QueueSize(),
		"uptime":    time.Since(metrics.StartTime).String(),
		"timestamp": time.Now().Unix(),
	}, statusCode)
}

// HandleRestartWorker - перезапускает воркера (для демонстрации)
func (h *Handlers) HandleRestartWorker(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.errorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		WorkerID int `json:"worker_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Здесь можно добавить логику перезапуска конкретного воркера
	// Пока просто возвращаем успех

	h.jsonResponse(w, map[string]interface{}{
		"status":    "ok",
		"message":   "Worker restart requested",
		"worker_id": req.WorkerID,
	}, http.StatusOK)
}

// ===== ВСПОМОГАТЕЛЬНЫЕ МЕТОДЫ =====

// getStatus - собирает статус системы
func (h *Handlers) getStatus() StatusResponse {
	pool := h.server.GetPool()
	metrics := h.server.GetMetrics()

	var status StatusResponse

	status.Workers.Current = pool.GetWorkerCount()
	status.Workers.Target = pool.GetTargetWorkerCount()
	status.Workers.Min = pool.GetMinWorkers()
	status.Workers.Max = pool.GetMaxWorkers()

	status.Queue = pool.QueueSize()

	status.Metrics.Submitted = metrics.TasksSubmitted.Load()
	status.Metrics.Completed = metrics.TasksCompleted.Load()
	status.Metrics.Failed = metrics.TasksFailed.Load()
	status.Metrics.Panicked = metrics.TasksPanicked.Load()
	status.Metrics.Restarted = metrics.WorkersRestarted.Load()
	status.Metrics.Uptime = time.Since(metrics.StartTime).String()

	status.Timestamp = time.Now().Unix()

	return status
}

// jsonResponse - отправляет JSON ответ
func (h *Handlers) jsonResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// errorResponse - отправляет ошибку в JSON формате
func (h *Handlers) errorResponse(w http.ResponseWriter, message string, statusCode int) {
	h.jsonResponse(w, map[string]interface{}{
		"error":     true,
		"message":   message,
		"code":      statusCode,
		"timestamp": time.Now().Unix(),
	}, statusCode)
}
