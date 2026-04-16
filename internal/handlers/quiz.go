// Package handlers предоставляет HTTP handlers для Go Quiz приложения
package handlers

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"qwen_test/internal/game"
	"qwen_test/internal/ratelimit"
)

// QuizHandler обрабатывает запросы викторины
type QuizHandler struct {
	questions     []Question
	playersCache  map[string]*game.Player
	skillTrees    map[string]*game.SkillTree
	questSystems  map[string]*game.QuestSystem
	achievements  map[string]*game.AchievementSystem
	rateLimiter   *ratelimit.RateLimiter
}

// Question представляет вопрос викторины
type Question struct {
	ID       int      `json:"id"`
	Question string   `json:"question"`
	Options  []string `json:"options"`
	Correct  int      `json:"correct"`
	Exp      int      `json:"exp"`
}

// NewQuizHandler создаёт новый QuizHandler
func NewQuizHandler(
	questions []Question,
	playersCache map[string]*game.Player,
	skillTrees map[string]*game.SkillTree,
	questSystems map[string]*game.QuestSystem,
	achievements map[string]*game.AchievementSystem,
) *QuizHandler {
	return &QuizHandler{
		questions:    questions,
		playersCache: playersCache,
		skillTrees:   skillTrees,
		questSystems: questSystems,
		achievements: achievements,
		// Лимит: 10 запросов в минуту для quiz endpoints
		rateLimiter: ratelimit.NewRateLimiter(10, time.Minute),
	}
}

// QuizHandler обрабатывает GET /api/quiz
func (h *QuizHandler) QuizHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Rate limiting
	key := getClientIP(r)
	if !h.rateLimiter.Allow(key) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Слишком много запросов. Попробуйте позже.",
		})
		return
	}

	// Получаем ID пользователя (из cookie или query параметра)
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "anonymous"
	}

	// Выбираем случайный вопрос
	question := getRandomQuestion(h.questions)

	response := map[string]interface{}{
		"question": question,
		"total":    len(h.questions),
		"answered": 0,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getRandomQuestion выбирает случайный вопрос
func getRandomQuestion(questions []Question) Question {
	if len(questions) == 0 {
		return Question{}
	}
	return questions[rand.Intn(len(questions))]
}

// getClientIP получает IP адрес клиента
func getClientIP(r *http.Request) string {
	// Проверяем X-Forwarded-For
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Проверяем X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Используем RemoteAddr
	return r.RemoteAddr
}

// AnswerRequest запрос ответа
type AnswerRequest struct {
	QuestionID int `json:"question_id"`
	Answer     int `json:"answer"`
	UserID     string `json:"user_id"`
}

// AnswerHandler обрабатывает POST /api/answer
func (h *QuizHandler) AnswerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Rate limiting
	key := getClientIP(r)
	if !h.rateLimiter.Allow(key) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Слишком много запросов. Попробуйте позже.",
		})
		return
	}

	var req AnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Находим вопрос
	var question *Question
	for i := range h.questions {
		if h.questions[i].ID == req.QuestionID {
			question = &h.questions[i]
			break
		}
	}

	if question == nil {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	// Проверяем ответ
	correct := req.Answer == question.Correct
	exp := 0
	if correct {
		exp = question.Exp
	}

	response := map[string]interface{}{
		"correct":        correct,
		"exp":            exp,
		"correct_option": question.Correct,
		"message":        getMessage(correct),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getMessage(correct bool) string {
	if correct {
		return "✅ Правильно!"
	}
	return "❌ Неправильно"
}

// StatsHandler обрабатывает GET /api/stats
func (h *QuizHandler) StatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Rate limiting
	key := getClientIP(r)
	if !h.rateLimiter.Allow(key) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Слишком много запросов. Попробуйте позже.",
		})
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "anonymous"
	}

	player := getOrCreatePlayer(h.playersCache, userID)

	response := map[string]interface{}{
		"player": player,
		"stats": map[string]interface{}{
			"level":          player.Level,
			"experience":     player.Experience,
			"go_knowledge":   player.GoKnowledge,
			"focus":          player.Focus,
			"willpower":      player.Willpower,
			"rating":         player.GetRating(),
			"correct_answers": player.CorrectAnswers,
			"wrong_answers":  player.WrongAnswers,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getOrCreatePlayer(playersCache map[string]*game.Player, userID string) *game.Player {
	if player, ok := playersCache[userID]; ok {
		return player
	}

	player := game.NewPlayer(userID, "Player")
	playersCache[userID] = player
	return player
}

// LeaderboardHandler обрабатывает GET /api/leaderboard
func (h *QuizHandler) LeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Rate limiting
	key := getClientIP(r)
	if !h.rateLimiter.Allow(key) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Слишком много запросов. Попробуйте позже.",
		})
		return
	}

	// Собираем leaderboard из кэша
	entries := make([]map[string]interface{}, 0, 10)
	for _, player := range h.playersCache {
		entries = append(entries, map[string]interface{}{
			"id":       player.ID,
			"name":     player.Name,
			"level":    player.Level,
			"rating":   player.GetRating(),
			"correct":  player.CorrectAnswers,
		})
	}

	// Сортируем по рейтингу (bubble sort для простоты)
	for i := 0; i < len(entries)-1; i++ {
		for j := 0; j < len(entries)-i-1; j++ {
			rating1 := entries[j]["rating"].(int)
			rating2 := entries[j+1]["rating"].(int)
			if rating1 < rating2 {
				entries[j], entries[j+1] = entries[j+1], entries[j]
			}
		}
	}

	// Берём топ-10
	if len(entries) > 10 {
		entries = entries[:10]
	}

	response := map[string]interface{}{
		"entries": entries,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HealthHandler обрабатывает GET /health
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ReadyHandler обрабатывает GET /ready
func ReadyHandler(w http.ResponseWriter, r *http.Request) {
	// Здесь можно проверить подключение к БД и другие зависимости
	response := map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// MetricsHandler обрабатывает GET /metrics (упрощённая версия)
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	// В будущем можно интегрировать Prometheus metrics
	response := map[string]interface{}{
		"uptime_seconds": int(time.Since(startTime).Seconds()),
		"requests_total": 0, // Нужно добавить счётчик
		"errors_total":   0,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

var startTime = time.Now()

// LogRequest middleware для логирования запросов
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("🌍 %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("✅ %s %s (%v)", r.Method, r.URL.Path, time.Since(start))
	})
}

// RecoverPanic middleware для восстановления после паник
func RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("❌ PANIC: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CORS middleware
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SecurityHeaders middleware для добавления заголовков безопасности
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}
