package web

import "time"

// StatusResponse - структура полного статуса системы
type StatusResponse struct {
	Workers struct {
		Current int `json:"current"`
		Target  int `json:"target"`
		Min     int `json:"min"`
		Max     int `json:"max"`
	} `json:"workers"`

	Queue int `json:"queue"`

	Metrics struct {
		Submitted int64  `json:"submitted"`
		Completed int64  `json:"completed"`
		Failed    int64  `json:"failed"`
		Panicked  int64  `json:"panicked"`
		Restarted int64  `json:"restarted"`
		Uptime    string `json:"uptime"`
	} `json:"metrics"`

	Timestamp int64 `json:"timestamp"`
}

// WorkerResponse - информация о воркере
type WorkerResponse struct {
	ID      int64     `json:"id"`
	Status  string    `json:"status"`
	Started time.Time `json:"started"`
	Tasks   int64     `json:"tasks_processed"`
}

// MetricsResponse - метрики
type MetricsResponse struct {
	Submitted   int64   `json:"submitted"`
	Completed   int64   `json:"completed"`
	Failed      int64   `json:"failed"`
	Panicked    int64   `json:"panicked"`
	Restarted   int64   `json:"restarted"`
	Uptime      string  `json:"uptime"`
	SuccessRate float64 `json:"success_rate"`
}

// QueueResponse - состояние очереди
type QueueResponse struct {
	Size        int    `json:"size"`
	Capacity    int    `json:"capacity"`
	Utilization string `json:"utilization"`
	Available   int    `json:"available"`
	Timestamp   int64  `json:"timestamp"`
}

// ErrorResponse - ошибка
type ErrorResponse struct {
	Error     bool   `json:"error"`
	Message   string `json:"message"`
	Code      int    `json:"code"`
	Timestamp int64  `json:"timestamp"`
}

// ConfigResponse - конфигурация
type ConfigResponse struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// HealthResponse - проверка здоровья
type HealthResponse struct {
	Status    string `json:"status"`
	Workers   int    `json:"workers"`
	Queue     int    `json:"queue"`
	Uptime    string `json:"uptime"`
	Timestamp int64  `json:"timestamp"`
}
