package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qwen_test/internal/game"
	"qwen_test/internal/ratelimit"
)

func TestNewQuizHandler(t *testing.T) {
	questions := []Question{
		{ID: 1, Question: "Test?", Options: []string{"A", "B"}, Correct: 0, Exp: 10},
	}
	playersCache := make(map[string]*game.Player)
	skillTrees := make(map[string]*game.SkillTree)
	questSystems := make(map[string]*game.QuestSystem)
	achievements := make(map[string]*game.AchievementSystem)

	h := NewQuizHandler(questions, playersCache, skillTrees, questSystems, achievements)

	if h == nil {
		t.Fatal("Ожидался не-nil QuizHandler")
	}
	if len(h.questions) != 1 {
		t.Errorf("Ожидаемое количество вопросов 1, получено %d", len(h.questions))
	}
	if h.rateLimiter == nil {
		t.Error("Ожидался не-nil rateLimiter")
	}
}

func TestQuizHandler_Success(t *testing.T) {
	questions := []Question{
		{ID: 1, Question: "Test?", Options: []string{"A", "B"}, Correct: 0, Exp: 10},
	}
	playersCache := make(map[string]*game.Player)
	skillTrees := make(map[string]*game.SkillTree)
	questSystems := make(map[string]*game.QuestSystem)
	achievements := make(map[string]*game.AchievementSystem)

	h := NewQuizHandler(questions, playersCache, skillTrees, questSystems, achievements)

	req := httptest.NewRequest(http.MethodGet, "/api/quiz?user_id=test123", nil)
	rr := httptest.NewRecorder()

	h.QuizHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Ожидаемый статус %d, получен %d", http.StatusOK, rr.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Ошибка декодирования ответа: %v", err)
	}

	if response["question"] == nil {
		t.Error("Ожидался вопрос в ответе")
	}
	if response["total"].(float64) != 1 {
		t.Errorf("Ожидаемое total 1, получено %v", response["total"])
	}
}

func TestQuizHandler_WrongMethod(t *testing.T) {
	h := &QuizHandler{
		questions:    []Question{},
		playersCache: make(map[string]*game.Player),
	}

	req := httptest.NewRequest(http.MethodPost, "/api/quiz", nil)
	rr := httptest.NewRecorder()

	h.QuizHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Ожидаемый статус %d, получен %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestAnswerHandler_Correct(t *testing.T) {
	questions := []Question{
		{ID: 1, Question: "Test?", Options: []string{"A", "B"}, Correct: 0, Exp: 10},
	}
	h := &QuizHandler{
		questions:    questions,
		playersCache: make(map[string]*game.Player),
		rateLimiter:  ratelimit.NewRateLimiter(100, time.Second),
	}

	body := `{"question_id": 1, "answer": 0, "user_id": "test123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/answer", nil)
	req.Body = http.MaxBytesReader(nil, req.Body, int64(len(body)))
	req.Body = nil
	req.Body = io.NopCloser(strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.AnswerHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Ожидаемый статус %d, получен %d", http.StatusOK, rr.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&response)

	if response["correct"] != true {
		t.Error("Ожидался правильный ответ")
	}
	if response["exp"].(float64) != 10 {
		t.Errorf("Ожидаемый exp 10, получено %v", response["exp"])
	}
}

func TestAnswerHandler_Wrong(t *testing.T) {
	questions := []Question{
		{ID: 1, Question: "Test?", Options: []string{"A", "B"}, Correct: 0, Exp: 10},
	}
	h := &QuizHandler{
		questions:    questions,
		playersCache: make(map[string]*game.Player),
		rateLimiter:  ratelimit.NewRateLimiter(100, time.Second),
	}

	body := `{"question_id": 1, "answer": 1, "user_id": "test123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/answer", nil)
	req.Body = io.NopCloser(strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.AnswerHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Ожидаемый статус %d, получен %d", http.StatusOK, rr.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&response)

	if response["correct"] != false {
		t.Error("Ожидался неправильный ответ")
	}
	if response["exp"].(float64) != 0 {
		t.Errorf("Ожидаемый exp 0, получено %v", response["exp"])
	}
}

func TestAnswerHandler_NotFound(t *testing.T) {
	questions := []Question{
		{ID: 1, Question: "Test?", Options: []string{"A", "B"}, Correct: 0, Exp: 10},
	}
	h := &QuizHandler{
		questions:    questions,
		playersCache: make(map[string]*game.Player),
		rateLimiter:  ratelimit.NewRateLimiter(100, time.Second),
	}

	body := `{"question_id": 999, "answer": 0}`
	req := httptest.NewRequest(http.MethodPost, "/api/answer", nil)
	req.Body = io.NopCloser(strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.AnswerHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Ожидаемый статус %d, получен %d", http.StatusNotFound, rr.Code)
	}
}

func TestStatsHandler_Success(t *testing.T) {
	playersCache := make(map[string]*game.Player)
	h := &QuizHandler{
		questions:    []Question{},
		playersCache: playersCache,
		rateLimiter:  ratelimit.NewRateLimiter(100, time.Second),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/stats?user_id=test123", nil)
	rr := httptest.NewRecorder()

	h.StatsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Ожидаемый статус %d, получен %d", http.StatusOK, rr.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&response)

	if response["player"] == nil {
		t.Error("Ожидался игрок в ответе")
	}
}

func TestLeaderboardHandler_Success(t *testing.T) {
	playersCache := make(map[string]*game.Player)
	// Добавляем тестовых игроков
	playersCache["user1"] = game.NewPlayer("user1", "Player1")
	playersCache["user2"] = game.NewPlayer("user2", "Player2")

	h := &QuizHandler{
		questions:    []Question{},
		playersCache: playersCache,
		rateLimiter:  ratelimit.NewRateLimiter(100, time.Second),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/leaderboard", nil)
	rr := httptest.NewRecorder()

	h.LeaderboardHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Ожидаемый статус %d, получен %d", http.StatusOK, rr.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&response)

	entries, ok := response["entries"].([]interface{})
	if !ok {
		t.Fatal("Ожидался массив entries")
	}

	if len(entries) != 2 {
		t.Errorf("Ожидаемое количество игроков 2, получено %d", len(entries))
	}
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	HealthHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Ожидаемый статус %d, получен %d", http.StatusOK, rr.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&response)

	if response["status"] != "healthy" {
		t.Errorf("Ожидаемый статус 'healthy', получен '%v'", response["status"])
	}
}

func TestReadyHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rr := httptest.NewRecorder()

	ReadyHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Ожидаемый статус %d, получен %d", http.StatusOK, rr.Code)
	}

	var response map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&response)

	if response["status"] != "ready" {
		t.Errorf("Ожидаемый статус 'ready', получен '%v'", response["status"])
	}
}

func TestCORS_Middleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := CORS(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Ожидался заголовок Access-Control-Allow-Origin: *")
	}
	if rr.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Ожидался заголовок Access-Control-Allow-Methods")
	}
}

func TestSecurityHeaders_Middleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := SecurityHeaders(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("Ожидался заголовок X-Content-Type-Options: nosniff")
	}
	if rr.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("Ожидался заголовок X-Frame-Options: DENY")
	}
}
