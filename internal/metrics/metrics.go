// Package metrics предоставляет метрики и мониторинг для Go Quiz приложения
package metrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Metrics хранит все метрики приложения
type Metrics struct {
	mu sync.RWMutex

	// Счётчики
	requestsTotal   int64
	requestsSuccess int64
	requestsError   int64
	errorsTotal     int64

	// Время начала работы
	startTime time.Time

	// Метрики по endpoint'ам
	endpoints map[string]*EndpointMetrics

	// Метрики базы данных
	dbConnections      int32
	dbQueriesTotal     int64
	dbQueriesSlow      int64
	dbLastBackupTime   time.Time

	// Метрики игроков
	totalPlayers    int32
	activePlayers   int32
	newPlayersToday int32

	// Метрики викторины
	questionsAnswered int64
	correctAnswers    int64
	wrongAnswers      int64

	// Метрики rate limiting
	rateLimitHits int64
}

// EndpointMetrics метрики для конкретного endpoint'а
type EndpointMetrics struct {
	Path        string
	Method      string
	Count       int64
	AvgDuration time.Duration
	MinDuration time.Duration
	MaxDuration time.Duration
	Errors      int64
}

// NewMetrics создаёт новый экземпляр метрик
func NewMetrics() *Metrics {
	return &Metrics{
		startTime: time.Now(),
		endpoints: make(map[string]*EndpointMetrics),
	}
}

// IncRequests увеличивает счётчик запросов
func (m *Metrics) IncRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestsTotal++
}

// IncSuccess увеличивает счётчик успешных запросов
func (m *Metrics) IncSuccess() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestsSuccess++
}

// IncError увеличивает счётчик ошибок
func (m *Metrics) IncError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestsError++
	m.errorsTotal++
}

// ObserveEndpoint фиксирует метрики endpoint'а
func (m *Metrics) ObserveEndpoint(path, method string, duration time.Duration, isError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := method + ":" + path
	endpoint, exists := m.endpoints[key]
	if !exists {
		endpoint = &EndpointMetrics{
			Path:   path,
			Method: method,
		}
		m.endpoints[key] = endpoint
	}

	endpoint.Count++
	endpoint.AvgDuration = (endpoint.AvgDuration*time.Duration(endpoint.Count-1) + duration) / time.Duration(endpoint.Count)

	if endpoint.MinDuration == 0 || duration < endpoint.MinDuration {
		endpoint.MinDuration = duration
	}
	if duration > endpoint.MaxDuration {
		endpoint.MaxDuration = duration
	}

	if isError {
		endpoint.Errors++
	}
}

// IncRateLimitHit фиксирует срабатывание rate limiter
func (m *Metrics) IncRateLimitHit() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rateLimitHits++
}

// IncQuestionAnswered фиксирует ответ на вопрос
func (m *Metrics) IncQuestionAnswered(correct bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.questionsAnswered++
	if correct {
		m.correctAnswers++
	} else {
		m.wrongAnswers++
	}
}

// SetDBMetrics устанавливает метрики базы данных
func (m *Metrics) SetDBMetrics(connections int32, queriesTotal, queriesSlow int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dbConnections = connections
	m.dbQueriesTotal = queriesTotal
	m.dbQueriesSlow = queriesSlow
}

// SetDBBackupTime устанавливает время последнего бэкапа
func (m *Metrics) SetDBBackupTime(t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dbLastBackupTime = t
}

// SetPlayerMetrics устанавливает метрики игроков
func (m *Metrics) SetPlayerMetrics(total, active, newToday int32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalPlayers = total
	m.activePlayers = active
	m.newPlayersToday = newToday
}

// GetStats возвращает текущую статистику
func (m *Metrics) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uptime := time.Since(m.startTime)

	stats := map[string]interface{}{
		"uptime_seconds":       int(uptime.Seconds()),
		"uptime_human":         uptime.String(),
		"start_time":           m.startTime.Format(time.RFC3339),
		"requests_total":       m.requestsTotal,
		"requests_success":     m.requestsSuccess,
		"requests_error":       m.requestsError,
		"errors_total":         m.errorsTotal,
		"rate_limit_hits":      m.rateLimitHits,
		"questions_answered":   m.questionsAnswered,
		"correct_answers":      m.correctAnswers,
		"wrong_answers":        m.wrongAnswers,
		"accuracy_percent":     0.0,
		"db_connections":       m.dbConnections,
		"db_queries_total":     m.dbQueriesTotal,
		"db_queries_slow":      m.dbQueriesSlow,
		"db_last_backup":       m.dbLastBackupTime.Format(time.RFC3339),
		"total_players":        m.totalPlayers,
		"active_players":       m.activePlayers,
		"new_players_today":    m.newPlayersToday,
		"endpoints_count":      len(m.endpoints),
	}

	if m.questionsAnswered > 0 {
		stats["accuracy_percent"] = float64(m.correctAnswers) / float64(m.questionsAnswered) * 100
	}

	if m.requestsTotal > 0 {
		stats["success_rate_percent"] = float64(m.requestsSuccess) / float64(m.requestsTotal) * 100
		stats["error_rate_percent"] = float64(m.requestsError) / float64(m.requestsTotal) * 100
	}

	return stats
}

// GetEndpointStats возвращает статистику по endpoint'ам
func (m *Metrics) GetEndpointStats() []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]map[string]interface{}, 0, len(m.endpoints))
	for _, ep := range m.endpoints {
		result = append(result, map[string]interface{}{
			"method":         ep.Method,
			"path":           ep.Path,
			"count":          ep.Count,
			"avg_duration_ms": float64(ep.AvgDuration.Milliseconds()),
			"min_duration_ms": float64(ep.MinDuration.Milliseconds()),
			"max_duration_ms": float64(ep.MaxDuration.Milliseconds()),
			"errors":         ep.Errors,
			"error_rate":     float64(ep.Errors) / float64(ep.Count) * 100,
		})
	}

	return result
}

// MetricsHandler HTTP handler для получения метрик
func (m *Metrics) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	stats := m.GetStats()
	endpointStats := m.GetEndpointStats()

	response := map[string]interface{}{
		"application": "Go Quiz",
		"version":     "1.0.0",
		"metrics":     stats,
		"endpoints":   endpointStats,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)
}

// PrometheusMetrics возвращает метрики в формате Prometheus
func (m *Metrics) PrometheusMetrics() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sb string

	sb += "# HELP goquiz_uptime_seconds Uptime in seconds\n"
	sb += "# TYPE goquiz_uptime_seconds counter\n"
	sb += fmt.Sprintf("goquiz_uptime_seconds %.0f\n", time.Since(m.startTime).Seconds())

	sb += "# HELP goquiz_requests_total Total number of requests\n"
	sb += "# TYPE goquiz_requests_total counter\n"
	sb += fmt.Sprintf("goquiz_requests_total %d\n", m.requestsTotal)

	sb += "# HELP goquiz_requests_success Total number of successful requests\n"
	sb += "# TYPE goquiz_requests_success counter\n"
	sb += fmt.Sprintf("goquiz_requests_success %d\n", m.requestsSuccess)

	sb += "# HELP goquiz_requests_error Total number of error requests\n"
	sb += "# TYPE goquiz_requests_error counter\n"
	sb += fmt.Sprintf("goquiz_requests_error %d\n", m.errorsTotal)

	sb += "# HELP goquiz_rate_limit_hits Total number of rate limit hits\n"
	sb += "# TYPE goquiz_rate_limit_hits counter\n"
	sb += fmt.Sprintf("goquiz_rate_limit_hits %d\n", m.rateLimitHits)

	sb += "# HELP goquiz_questions_answered Total number of questions answered\n"
	sb += "# TYPE goquiz_questions_answered counter\n"
	sb += fmt.Sprintf("goquiz_questions_answered %d\n", m.questionsAnswered)

	sb += "# HELP goquiz_total_players Total number of players\n"
	sb += "# TYPE goquiz_total_players gauge\n"
	sb += fmt.Sprintf("goquiz_total_players %d\n", m.totalPlayers)

	sb += "# HELP goquiz_active_players Number of active players\n"
	sb += "# TYPE goquiz_active_players gauge\n"
	sb += fmt.Sprintf("goquiz_active_players %d\n", m.activePlayers)

	return sb
}
