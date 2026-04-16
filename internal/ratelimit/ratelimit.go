// Package ratelimit предоставляет middleware для ограничения частоты запросов
package ratelimit

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter ограничивает частоту запросов
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int           // Максимальное количество запросов
	window   time.Duration // Временное окно
}

// NewRateLimiter создаёт новый rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Запускаем goroutine для очистки старых записей
	go rl.cleanup()

	return rl
}

// Allow проверяет, разрешён ли запрос для данного ключа
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Получаем существующие запросы
	requests := rl.requests[key]

	// Фильтруем запросы в текущем окне
	validRequests := make([]time.Time, 0, len(requests))
	for _, t := range requests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}

	// Проверяем лимит
	if len(validRequests) >= rl.limit {
		rl.requests[key] = validRequests
		return false
	}

	// Добавляем новый запрос
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests

	return true
}

// cleanup периодически удаляет старые записи
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		windowStart := now.Add(-rl.window)

		for key, requests := range rl.requests {
			validRequests := make([]time.Time, 0, len(requests))
			for _, t := range requests {
				if t.After(windowStart) {
					validRequests = append(validRequests, t)
				}
			}

			if len(validRequests) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = validRequests
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware создаёт HTTP middleware для rate limiting
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Используем IP адрес как ключ
		key := r.RemoteAddr

		// Проверяем X-Forwarded-For для запросов за proxy
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			key = xff
		}

		if !rl.Allow(key) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "Слишком много запросов. Попробуйте позже."}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetRemaining возвращает оставшееся количество запросов для ключа
func (rl *RateLimiter) GetRemaining(key string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	requests := rl.requests[key]
	validCount := 0
	for _, t := range requests {
		if t.After(windowStart) {
			validCount++
		}
	}

	remaining := rl.limit - validCount
	if remaining < 0 {
		return 0
	}
	return remaining
}
