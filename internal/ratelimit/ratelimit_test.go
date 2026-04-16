package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)

	if rl.limit != 10 {
		t.Errorf("Ожидаемый лимит 10, получено %d", rl.limit)
	}
	if rl.window != time.Minute {
		t.Errorf("Ожидаемое окно 1 минута, получено %v", rl.window)
	}
	if rl.requests == nil {
		t.Error("Ожидаемая не-nil карта requests")
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(3, time.Second)

	// Первые 3 запроса должны быть разрешены
	for i := 0; i < 3; i++ {
		if !rl.Allow("user1") {
			t.Errorf("Запрос %d должен быть разрешён", i+1)
		}
	}

	// 4-й запрос должен быть заблокирован
	if rl.Allow("user1") {
		t.Error("4-й запрос должен быть заблокирован")
	}

	// Ждём окончания окна
	time.Sleep(1100 * time.Millisecond)

	// Снова должен быть разрешён
	if !rl.Allow("user1") {
		t.Error("Запрос после окончания окна должен быть разрешён")
	}
}

func TestRateLimiter_Allow_DifferentKeys(t *testing.T) {
	rl := NewRateLimiter(2, time.Second)

	// user1 делает 2 запроса
	rl.Allow("user1")
	rl.Allow("user1")

	// user2 делает 1 запрос
	if !rl.Allow("user2") {
		t.Error("Запрос user2 должен быть разрешён")
	}

	// user1 должен быть заблокирован
	if rl.Allow("user1") {
		t.Error("Запрос user1 должен быть заблокирован")
	}

	// user2 ещё один запрос
	if !rl.Allow("user2") {
		t.Error("Второй запрос user2 должен быть разрешён")
	}

	// user2 должен быть заблокирован
	if rl.Allow("user2") {
		t.Error("Третий запрос user2 должен быть заблокирован")
	}
}

func TestRateLimiter_GetRemaining(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)

	remaining := rl.GetRemaining("user1")
	if remaining != 5 {
		t.Errorf("Ожидаемое количество 5, получено %d", remaining)
	}

	rl.Allow("user1")
	rl.Allow("user1")

	remaining = rl.GetRemaining("user1")
	if remaining != 3 {
		t.Errorf("Ожидаемое количество 3, получено %d", remaining)
	}

	rl.Allow("user1")
	rl.Allow("user1")
	rl.Allow("user1")

	remaining = rl.GetRemaining("user1")
	if remaining != 0 {
		t.Errorf("Ожидаемое количество 0, получено %d", remaining)
	}
}

func TestRateLimiter_Middleware(t *testing.T) {
	rl := NewRateLimiter(2, time.Second)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	wrappedHandler := rl.Middleware(handler)

	// Первые 2 запроса должны пройти
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		rr := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Запрос %d: ожидаемый статус %d, получен %d", i+1, http.StatusOK, rr.Code)
		}
	}

	// 3-й запрос должен быть заблокирован
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("3-й запрос: ожидаемый статус %d, получен %d", http.StatusTooManyRequests, rr.Code)
	}

	// Проверяем заголовок Retry-After
	if retryAfter := rr.Header().Get("Retry-After"); retryAfter == "" {
		t.Error("Ожидался заголовок Retry-After")
	}
}

func TestRateLimiter_Middleware_XForwardedFor(t *testing.T) {
	rl := NewRateLimiter(1, time.Second)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := rl.Middleware(handler)

	// Запрос с X-Forwarded-For
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Ожидаемый статус %d, получен %d", http.StatusOK, rr.Code)
	}

	// Второй запрос с тем же X-Forwarded-For должен быть заблокирован
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "127.0.0.1:12346"
	req2.Header.Set("X-Forwarded-For", "192.168.1.1")
	rr2 := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("Ожидаемый статус %d, получен %d", http.StatusTooManyRequests, rr2.Code)
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := NewRateLimiter(2, 100*time.Millisecond)

	// Делаем запрос
	rl.Allow("user1")
	rl.Allow("user1")

	// Проверяем что запросы записаны
	if rl.GetRemaining("user1") != 0 {
		t.Error("Ожидаемое количество 0")
	}

	// Ждём окончания окна + время на очистку
	time.Sleep(150 * time.Millisecond)

	// Записи должны быть очищены (или пересчитаны)
	// GetRemaining должен вернуть полный лимит
	remaining := rl.GetRemaining("user1")
	if remaining < 2 {
		t.Errorf("Ожидаемое количество >= 2 после очистки, получено %d", remaining)
	}
}
