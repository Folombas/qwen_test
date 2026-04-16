// Package middleware предоставляет HTTP middleware для Go Quiz приложения
package middleware

import (
	"net/http"
	"time"

	"qwen_test/internal/ratelimit"
)

// RateLimitConfigs стандартные конфигурации rate limiting
var (
	// Strict: 5 запросов в минуту (для auth endpoints)
	StrictRateLimiter = ratelimit.NewRateLimiter(5, time.Minute)

	// Normal: 30 запросов в минуту (для обычных endpoints)
	NormalRateLimiter = ratelimit.NewRateLimiter(30, time.Minute)

	// Relaxed: 60 запросов в минуту (для quiz и stats)
	RelaxedRateLimiter = ratelimit.NewRateLimiter(60, time.Minute)
)

// WithRateLimit применяет rate limiting с заданным limiter
func WithRateLimit(limiter *ratelimit.RateLimiter, next http.Handler) http.Handler {
	return limiter.Middleware(next)
}

// WithAuthRateLimit применяет строгий rate limiting для auth endpoints
func WithAuthRateLimit(next http.Handler) http.Handler {
	return WithRateLimit(StrictRateLimiter, next)
}

// WithNormalRateLimit применяет стандартный rate limiting
func WithNormalRateLimit(next http.Handler) http.Handler {
	return WithRateLimit(NormalRateLimiter, next)
}

// WithRelaxedRateLimit применяет мягкий rate limiting для quiz/stats
func WithRelaxedRateLimit(next http.Handler) http.Handler {
	return WithRateLimit(RelaxedRateLimiter, next)
}

// Chain объединяет несколько middleware в цепочку
func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
