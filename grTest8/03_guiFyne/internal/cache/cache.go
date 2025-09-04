package cache

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// WordCache реализует асинхронный кэш для подсчёта слов
type WordCache struct {
	cache    map[string]int // Кэш: строка -> количество слов
	mutex    sync.RWMutex   // Мьютекс для безопасного доступа
	requests chan request   // Канал запросов
	logger   *logrus.Logger
}

type request struct {
	input    string
	response chan response
}

type response struct {
	count int
	err   error
}

// NewWordCache создаёт новый кэш
func NewWordCache(logger *logrus.Logger) *WordCache {
	return &WordCache{
		cache:    make(map[string]int),
		mutex:    sync.RWMutex{},
		requests: make(chan request),
		logger:   logger,
	}
}

// Run запускает воркер для обработки запросов
func (wc *WordCache) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			wc.logger.Info("WordCache worker stopped")
			return
		case req := <-wc.requests:
			wc.processRequest(req)
		}
	}
}

// processRequest обрабатывает запрос на подсчёт слов
func (wc *WordCache) processRequest(req request) {
	// Проверяем кэш
	wc.mutex.RLock()
	count, exists := wc.cache[req.input]
	wc.mutex.RUnlock()

	if exists {
		wc.logger.Info("Cache hit", "input", req.input, "count", count)
		req.response <- response{count: count, err: nil}
		return
	}

	// Проверяем пустую строку
	if req.input == "" {
		wc.logger.Warn("Empty input string")
		req.response <- response{count: 0, err: errors.New("empty input string")}
		return
	}

	// Подсчёт слов
	count = len(strings.Fields(req.input))
	wc.logger.Info("Calculated word count", "input", req.input, "count", count)

	// Сохраняем в кэш
	wc.mutex.Lock()
	wc.cache[req.input] = count
	wc.mutex.Unlock()

	req.response <- response{count: count, err: nil}
}

// CountWords отправляет запрос на подсчёт слов
func (wc *WordCache) CountWords(input string) (int, error) {
	respChan := make(chan response)
	wc.requests <- request{input: input, response: respChan}
	resp := <-respChan
	return resp.count, resp.err
}
