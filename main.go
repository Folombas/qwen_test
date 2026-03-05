package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

// Question представляет вопрос викторины
type Question struct {
	ID       int      `json:"id"`
	Question string   `json:"question"`
	Options  []string `json:"options"`
	Correct  int      `json:"correct"`
	Exp      int      `json:"exp"`
}

// UserData хранит статистику игрока
type UserData struct {
	ID             string `json:"id"`
	TotalEXP       int    `json:"total_exp"`
	CorrectAnswers int    `json:"correct_answers"`
	WrongAnswers   int    `json:"wrong_answers"`
	Level          int    `json:"level"`
	AskedQuestions []int  `json:"asked_questions"`
	CreatedAt      time.Time `json:"created_at"`
}

// Session хранит активную сессию игрока
type Session struct {
	UserID      string
	CurrentQ    *Question
	StartTime   time.Time
}

// API Request/Response types
type AnswerRequest struct {
	QuestionID int `json:"question_id"`
	OptionIndex int `json:"option_index"`
}

type AnswerResponse struct {
	Correct       bool   `json:"correct"`
	Exp           int    `json:"exp"`
	CorrectOption int    `json:"correct_option"`
	Message       string `json:"message"`
	NewExp        int    `json:"new_exp"`
	NewLevel      int    `json:"new_level"`
}

type QuizResponse struct {
	Question *Question `json:"question"`
	Total    int       `json:"total"`
	Answered int       `json:"answered"`
}

type StatsResponse struct {
	User           *UserData `json:"user"`
	TotalQuestions int       `json:"total_questions"`
	Progress       float64   `json:"progress"`
}

type LeaderboardEntry struct {
	ID        string `json:"id"`
	Level     int    `json:"level"`
	TotalEXP  int    `json:"total_exp"`
	Correct   int    `json:"correct"`
}

type LeaderboardResponse struct {
	Entries []LeaderboardEntry `json:"entries"`
}

// глобальные переменные
var (
	questions     []Question
	users         = make(map[string]*UserData)
	sessions      = make(map[string]*Session)
	usersMu       sync.RWMutex
	sessionsMu    sync.RWMutex
	questionsFile = "questions.json"
	dataFile      = "users.json"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Загружаем вопросы
	if err := loadQuestions(); err != nil {
		log.Fatal("Ошибка загрузки вопросов:", err)
	}
	log.Printf("Загружено %d вопросов", len(questions))

	// Загружаем сохранённые данные пользователей
	loadUserData()

	// Автосохранение каждые 5 минут
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			saveUserData()
		}
	}()

	// HTTP handlers
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/quiz", quizHandler)
	http.HandleFunc("/api/answer", answerHandler)
	http.HandleFunc("/api/stats", statsHandler)
	http.HandleFunc("/api/leaderboard", leaderboardHandler)
	http.HandleFunc("/api/reset", resetHandler)

	port := ":8080"
	fmt.Printf("🚀 Go Quiz Web Server starting on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// --- Загрузка/сохранение данных ---

func loadQuestions() error {
	data, err := os.ReadFile(questionsFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &questions)
}

func loadUserData() {
	data, err := os.ReadFile(dataFile)
	if err != nil {
		log.Println("Файл пользователей не найден, начинаем с пустой статистикой")
		return
	}
	err = json.Unmarshal(data, &users)
	if err != nil {
		log.Println("Ошибка парсинга users.json, используем пустую статистику")
	}
}

func saveUserData() {
	usersMu.RLock()
	defer usersMu.RUnlock()
	
	data, _ := json.MarshalIndent(users, "", "  ")
	if err := os.WriteFile(dataFile, data, 0644); err != nil {
		log.Println("Ошибка сохранения данных пользователей:", err)
	} else {
		log.Println("Данные пользователей сохранены")
	}
}

// --- Работа с пользователями ---

func getUser(userID string) *UserData {
	usersMu.Lock()
	defer usersMu.Unlock()
	
	if _, ok := users[userID]; !ok {
		users[userID] = &UserData{
			ID:             userID,
			TotalEXP:       0,
			CorrectAnswers: 0,
			WrongAnswers:   0,
			Level:          1,
			AskedQuestions: []int{},
			CreatedAt:      time.Now(),
		}
	}
	return users[userID]
}

func updateLevel(user *UserData) {
	newLevel := int(math.Floor(float64(user.TotalEXP)/100)) + 1
	if newLevel > user.Level {
		user.Level = newLevel
	}
}

// --- Handlers ---

var tmpl = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Quiz - Викторина по языку Go</title>
    <link href="https://fonts.googleapis.com/css2?family=Montserrat:wght@400;500;600;700&family=Fira+Code:wght@400;600&display=swap" rel="stylesheet">
    <style>
        /* Material Design Color Palette */
        :root {
            /* Dark Theme Colors */
            --md-primary: #00d9ff;
            --md-primary-variant: #00b8d9;
            --md-secondary: #00ff88;
            --md-secondary-variant: #00e676;
            --md-background: #1a1a2e;
            --md-background-variant: #16213e;
            --md-surface: #242442;
            --md-surface-variant: #2d2d4a;
            --md-error: #ff5252;
            --md-success: #00e676;
            --md-warning: #ffb74d;
            --md-text-primary: #ffffff;
            --md-text-secondary: rgba(255, 255, 255, 0.7);
            --md-text-disabled: rgba(255, 255, 255, 0.5);
            --md-divider: rgba(255, 255, 255, 0.12);
            
            /* Elevation Shadows */
            --elevation-1: 0 1px 3px rgba(0,0,0,0.2), 0 1px 1px rgba(0,0,0,0.14), 0 2px 1px -1px rgba(0,0,0,0.12);
            --elevation-2: 0 1px 5px rgba(0,0,0,0.2), 0 2px 2px rgba(0,0,0,0.14), 0 3px 1px -2px rgba(0,0,0,0.12);
            --elevation-3: 0 1px 8px rgba(0,0,0,0.2), 0 3px 4px rgba(0,0,0,0.14), 0 3px 3px -2px rgba(0,0,0,0.12);
            --elevation-4: 0 2px 4px rgba(0,0,0,0.2), 0 4px 5px rgba(0,0,0,0.14), 0 1px 10px rgba(0,0,0,0.12);
            --elevation-8: 0 5px 5px rgba(0,0,0,0.2), 0 8px 10px rgba(0,0,0,0.14), 0 3px 14px rgba(0,0,0,0.12);
            
            /* Animation */
            --transition-fast: 150ms cubic-bezier(0.4, 0, 0.2, 1);
            --transition-standard: 250ms cubic-bezier(0.4, 0, 0.2, 1);
            --transition-slow: 350ms cubic-bezier(0.4, 0, 0.2, 1);
        }
        
        body.light-theme {
            --md-primary: #1976d2;
            --md-primary-variant: #1565c0;
            --md-secondary: #388e3c;
            --md-secondary-variant: #2e7d32;
            --md-background: #fafafa;
            --md-background-variant: #f5f5f5;
            --md-surface: #ffffff;
            --md-surface-variant: #f8f9fa;
            --md-error: #d32f2f;
            --md-success: #388e3c;
            --md-warning: #f57c00;
            --md-text-primary: #212121;
            --md-text-secondary: rgba(0, 0, 0, 0.6);
            --md-text-disabled: rgba(0, 0, 0, 0.38);
            --md-divider: rgba(0, 0, 0, 0.12);
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Montserrat', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            min-height: 100vh;
            background: linear-gradient(135deg, var(--md-background) 0%, var(--md-background-variant) 50%, var(--md-surface) 100%);
            color: var(--md-text-primary);
            transition: background var(--transition-slow), color var(--transition-standard);
            -webkit-font-smoothing: antialiased;
            -moz-osx-font-smoothing: grayscale;
        }
        .container {
            max-width: 960px;
            margin: 0 auto;
            padding: 16px;
        }
        @media (min-width: 600px) {
            .container { padding: 24px; }
        }
        @media (min-width: 960px) {
            .container { padding: 32px; }
        }
        header {
            display: flex;
            flex-wrap: wrap;
            justify-content: space-between;
            align-items: center;
            padding: 16px 0;
            margin-bottom: 24px;
        }
        @media (min-width: 600px) {
            header {
                padding: 20px 0;
                margin-bottom: 32px;
            }
        }
        .logo {
            font-size: 1.5rem;
            font-weight: 700;
            background: linear-gradient(135deg, var(--md-primary), var(--md-secondary));
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            letter-spacing: -0.5px;
        }
        @media (min-width: 600px) {
            .logo { font-size: 1.8rem; }
        }
        .header-actions {
            display: flex;
            gap: 8px;
            align-items: center;
            flex-wrap: wrap;
        }
        .theme-toggle, .nav-btn {
            background: transparent;
            border: none;
            border-radius: 50%;
            width: 40px;
            height: 40px;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: all var(--transition-fast);
            font-size: 1.2rem;
            color: var(--md-text-primary);
            position: relative;
            overflow: hidden;
        }
        .theme-toggle:hover, .nav-btn:hover {
            background: rgba(128, 128, 128, 0.1);
        }
        .nav-btn {
            border-radius: 20px;
            width: auto;
            padding: 0 16px;
            font-size: 0.875rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.75px;
            height: 36px;
        }
        @media (min-width: 600px) {
            .nav-btn {
                font-size: 0.9rem;
                padding: 0 20px;
                height: 40px;
            }
        }
        .nav-btn.active {
            background: linear-gradient(135deg, var(--md-primary), var(--md-secondary));
            color: #fff;
            box-shadow: var(--elevation-2);
        }
        .nav-btn.active:hover {
            box-shadow: var(--elevation-4);
        }
        /* Main Content */
        .main-content {
            min-height: 60vh;
        }
        .page {
            display: none;
            animation: fadeIn 0.3s ease;
        }
        .page.active {
            display: block;
        }
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(20px); }
            to { opacity: 1; transform: translateY(0); }
        }
        /* Home Page */
        .hero {
            text-align: center;
            padding: 60px 20px;
        }
        .hero h1 {
            font-size: 3rem;
            margin-bottom: 20px;
            background: linear-gradient(135deg, #00d9ff, #00ff88);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        body.light-theme .hero h1 {
            background: linear-gradient(135deg, #1a1a2e, #0f3460);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        .hero p {
            font-size: 1.2rem;
            color: #aaa;
            margin-bottom: 40px;
        }
        body.light-theme .hero p {
            color: var(--md-text-secondary);
        }
        .start-btn {
            background: linear-gradient(135deg, var(--md-primary), var(--md-secondary));
            border: none;
            border-radius: 28px;
            padding: 16px 40px;
            font-size: 1.1rem;
            font-weight: 600;
            color: #fff;
            cursor: pointer;
            transition: all var(--transition-standard);
            box-shadow: var(--elevation-4);
            text-transform: uppercase;
            letter-spacing: 0.75px;
        }
        @media (min-width: 600px) {
            .start-btn {
                padding: 18px 50px;
                font-size: 1.2rem;
            }
        }
        .start-btn:hover {
            transform: translateY(-4px);
            box-shadow: var(--elevation-8);
        }
        .features {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 30px;
            margin-top: 48px;
        }
        @media (min-width: 600px) {
            .features { gap: 24px; }
        }
        .feature-card {
            background: var(--md-surface);
            border-radius: 16px;
            padding: 24px;
            text-align: center;
            border: 1px solid var(--md-divider);
            box-shadow: var(--elevation-2);
            transition: all var(--transition-standard);
        }
        @media (min-width: 600px) {
            .feature-card { padding: 32px; }
        }
        .feature-card:hover {
            box-shadow: var(--elevation-8);
            transform: translateY(-4px);
        }
        .feature-icon {
            font-size: 3rem;
            margin-bottom: 15px;
        }
        .feature-card h3 {
            margin-bottom: 10px;
        }
        /* Quiz Page */
        .quiz-container {
            background: var(--md-surface);
            border-radius: 16px;
            padding: 24px;
            border: 1px solid var(--md-divider);
            box-shadow: var(--elevation-4);
        }
        @media (min-width: 600px) {
            .quiz-container { padding: 32px; }
        }
        .quiz-header {
            display: flex;
            justify-content: space-between;
            margin-bottom: 24px;
            font-size: 0.875rem;
            color: var(--md-text-secondary);
            font-weight: 500;
        }
        .progress-bar {
            width: 100%;
            height: 6px;
            background: var(--md-divider);
            border-radius: 10px;
            overflow: hidden;
            margin-bottom: 24px;
        }
        .progress-fill {
            height: 100%;
            background: linear-gradient(135deg, var(--md-primary), var(--md-secondary));
            transition: width var(--transition-standard);
        }
        .question-text {
            font-size: 1.2rem;
            margin-bottom: 24px;
            line-height: 1.6;
            font-weight: 500;
        }
        @media (min-width: 600px) {
            .question-text { font-size: 1.4rem; }
        }
        .options {
            display: flex;
            flex-direction: column;
            gap: 15px;
        }
        .option-btn {
            background: var(--md-surface);
            border: 2px solid var(--md-divider);
            border-radius: 12px;
            padding: 16px 20px;
            text-align: left;
            cursor: pointer;
            transition: all var(--transition-fast);
            font-size: 0.95rem;
            color: var(--md-text-primary);
            font-weight: 500;
            box-shadow: var(--elevation-1);
        }
        .option-btn:hover {
            border-color: var(--md-primary);
            background: rgba(0, 217, 255, 0.08);
            box-shadow: var(--elevation-3);
            transform: translateX(4px);
        }
        .option-btn.correct {
            background: rgba(0, 230, 118, 0.15);
            border-color: var(--md-success);
            box-shadow: 0 0 20px rgba(0, 230, 118, 0.4), inset 0 0 10px rgba(0, 230, 118, 0.1);
            animation: correctPulse 0.5s ease-out;
        }
        .option-btn.wrong {
            background: rgba(255, 82, 82, 0.15);
            border-color: var(--md-error);
            box-shadow: 0 0 20px rgba(255, 82, 82, 0.4), inset 0 0 10px rgba(255, 82, 82, 0.1);
            animation: wrongShake 0.4s ease-out;
        }
        @keyframes correctPulse {
            0% { transform: scale(1); box-shadow: 0 0 0 rgba(0,255,136,0); }
            50% { transform: scale(1.05); box-shadow: 0 0 35px rgba(0,255,136,0.7); }
            100% { transform: scale(1.02); box-shadow: 0 0 25px rgba(0,255,136,0.5); }
        }
        @keyframes wrongShake {
            0%, 100% { transform: translateX(0); }
            20% { transform: translateX(-10px); }
            40% { transform: translateX(10px); }
            60% { transform: translateX(-10px); }
            80% { transform: translateX(10px); }
        }
        .option-btn.disabled {
            pointer-events: none;
            opacity: 0.6;
        }
        .quiz-footer {
            margin-top: 30px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .exp-badge {
            background: linear-gradient(135deg, var(--md-warning), #ff9800);
            color: #fff;
            padding: 8px 16px;
            border-radius: 16px;
            font-weight: 600;
            font-size: 0.875rem;
            box-shadow: var(--elevation-1);
        }
        .next-btn {
            background: linear-gradient(135deg, var(--md-primary), var(--md-secondary));
            border: none;
            border-radius: 20px;
            padding: 12px 28px;
            font-size: 0.875rem;
            font-weight: 600;
            color: #fff;
            cursor: pointer;
            transition: all var(--transition-fast);
            display: none;
            text-transform: uppercase;
            letter-spacing: 0.75px;
            box-shadow: var(--elevation-2);
        }
        .next-btn:hover {
            transform: translateY(-2px);
            box-shadow: var(--elevation-4);
        }
        .next-btn.visible {
            display: block;
        }
        /* Stats Page */
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 40px;
        }
        .stat-card {
            background: var(--md-surface);
            border-radius: 16px;
            padding: 20px;
            text-align: center;
            border: 1px solid var(--md-divider);
            box-shadow: var(--elevation-2);
            transition: all var(--transition-standard);
        }
        @media (min-width: 600px) {
            .stat-card { padding: 28px; }
        }
        .stat-card:hover {
            box-shadow: var(--elevation-4);
            transform: translateY(-2px);
        }
        .stat-value {
            font-size: 2rem;
            font-weight: 700;
            background: linear-gradient(135deg, var(--md-primary), var(--md-secondary));
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        @media (min-width: 600px) {
            .stat-value { font-size: 2.5rem; }
        }
        .stat-label {
            color: var(--md-text-secondary);
            margin-top: 10px;
            font-size: 0.875rem;
        }
        /* Leaderboard */
        .leaderboard-table {
            width: 100%;
            border-collapse: collapse;
            background: var(--md-surface);
            border-radius: 16px;
            overflow: hidden;
            border: 1px solid var(--md-divider);
            box-shadow: var(--elevation-2);
        }
        .leaderboard-table th,
        .leaderboard-table td {
            padding: 16px 20px;
            text-align: left;
        }
        @media (min-width: 600px) {
            .leaderboard-table th,
            .leaderboard-table td { padding: 18px 25px; }
        }
        .leaderboard-table th {
            background: rgba(128, 128, 128, 0.1);
            font-weight: 600;
            text-transform: uppercase;
            font-size: 0.75rem;
            letter-spacing: 1px;
            color: var(--md-text-secondary);
        }
        .leaderboard-table tr:not(:last-child) {
            border-bottom: 1px solid var(--md-divider);
        }
        .rank-1 { background: linear-gradient(90deg, rgba(255,215,0,0.2), transparent); }
        .rank-2 { background: linear-gradient(90deg, rgba(192,192,192,0.2), transparent); }
        .rank-3 { background: linear-gradient(90deg, rgba(205,127,50,0.2), transparent); }
        .rank-badge {
            display: inline-block;
            width: 35px;
            height: 35px;
            line-height: 35px;
            text-align: center;
            border-radius: 50%;
            background: rgba(255,255,255,0.1);
            font-weight: 700;
        }
        .rank-1 .rank-badge { background: linear-gradient(135deg, #ffd700, #ffaa00); color: #1a1a2e; }
        .rank-2 .rank-badge { background: linear-gradient(135deg, #c0c0c0, #a0a0a0); color: #1a1a2e; }
        .rank-3 .rank-badge { background: linear-gradient(135deg, #cd7f32, #b87333); color: #fff; }
        /* Reset button */
        .reset-section {
            text-align: center;
            margin-top: 40px;
            padding-top: 30px;
            border-top: 1px solid rgba(255,255,255,0.1);
        }
        body.light-theme .reset-section {
            border-top: 1px solid var(--md-divider);
        }
        .reset-btn {
            background: transparent;
            border: 2px solid var(--md-error);
            color: var(--md-error);
            border-radius: 20px;
            padding: 10px 28px;
            cursor: pointer;
            transition: all var(--transition-fast);
            font-size: 0.875rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.75px;
        }
        .reset-btn:hover {
            background: rgba(255, 82, 82, 0.1);
            box-shadow: var(--elevation-2);
            transform: translateY(-2px);
        }
        /* Mobile */
        @media (max-width: 600px) {
            .hero h1 { font-size: 1.75rem; }
            .quiz-container { padding: 20px; }
            .question-text { font-size: 1rem; }
            .option-btn { padding: 14px 16px; font-size: 0.9rem; }
            header { gap: 12px; }
            .nav-btn { font-size: 0.75rem; padding: 0 12px; height: 32px; }
            .leaderboard-table th,
            .leaderboard-table td { padding: 12px 14px; font-size: 0.8rem; }
        }

        /* Interview Prep Page */
        .interview-container {
            max-width: 700px;
            margin: 0 auto;
            padding: 20px 0;
        }
        .card-stack {
            position: relative;
            height: 500px;
            perspective: 1000px;
        }
        @media (max-width: 600px) {
            .card-stack { height: 450px; }
        }
        .interview-card {
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: var(--md-surface);
            border-radius: 20px;
            border: 1px solid var(--md-divider);
            box-shadow: var(--elevation-8);
            padding: 32px;
            display: flex;
            flex-direction: column;
            transition: all var(--transition-slow);
            transform-origin: center center;
            backface-visibility: hidden;
        }
        .interview-card.swipe-left {
            transform: translateX(-150%) rotate(-30deg);
            opacity: 0;
        }
        .interview-card.swipe-right {
            transform: translateX(150%) rotate(30deg);
            opacity: 0;
        }
        .interview-card.fade-out {
            transform: scale(0.9);
            opacity: 0;
        }
        .interview-card.fade-in {
            animation: cardFadeIn 0.4s ease-out forwards;
        }
        @keyframes cardFadeIn {
            from {
                opacity: 0;
                transform: scale(1.1) translateY(20px);
            }
            to {
                opacity: 1;
                transform: scale(1) translateY(0);
            }
        }
        .card-question {
            font-size: 1.3rem;
            font-weight: 700;
            color: var(--md-text-primary);
            line-height: 1.5;
            margin-bottom: 24px;
            display: flex;
            align-items: flex-start;
            gap: 12px;
        }
        .card-question::before {
            content: '💡';
            font-size: 1.5rem;
            flex-shrink: 0;
        }
        .card-answer {
            flex: 1;
            overflow-y: auto;
            color: var(--md-text-secondary);
            line-height: 1.8;
            font-size: 0.95rem;
            opacity: 0;
            transform: translateY(20px);
            transition: all var(--transition-standard);
            max-height: 0;
        }
        .card-answer.visible {
            opacity: 1;
            transform: translateY(0);
            max-height: 500px;
        }
        .card-answer p, .card-answer ul, .card-answer ol {
            margin: 0 0 16px;
        }
        .card-answer ul, .card-answer ol {
            padding-left: 24px;
        }
        .card-answer li {
            margin-bottom: 8px;
        }
        .card-answer code {
            font-family: 'Fira Code', monospace;
            background: rgba(128, 128, 128, 0.15);
            padding: 2px 8px;
            border-radius: 4px;
            font-size: 0.85rem;
        }
        .card-answer pre {
            background: var(--md-background);
            border-radius: 8px;
            padding: 16px;
            overflow-x: auto;
            margin: 16px 0;
        }
        .card-answer pre code {
            background: transparent;
            padding: 0;
        }
        .interview-controls {
            display: flex;
            justify-content: center;
            gap: 16px;
            margin-top: 24px;
        }
        .control-btn {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 14px 28px;
            border: none;
            border-radius: 28px;
            font-size: 0.95rem;
            font-weight: 600;
            cursor: pointer;
            transition: all var(--transition-fast);
            text-transform: uppercase;
            letter-spacing: 0.75px;
        }
        .skip-btn {
            background: transparent;
            border: 2px solid var(--md-error);
            color: var(--md-error);
        }
        .skip-btn:hover {
            background: rgba(255, 82, 82, 0.1);
            transform: translateY(-2px);
            box-shadow: var(--elevation-2);
        }
        .show-answer-btn {
            background: linear-gradient(135deg, var(--md-primary), var(--md-secondary));
            color: #fff;
        }
        .show-answer-btn:hover {
            transform: translateY(-2px);
            box-shadow: var(--elevation-4);
        }
        .progress-indicator {
            text-align: center;
            margin-top: 20px;
            color: var(--md-text-secondary);
            font-size: 0.9rem;
            font-weight: 600;
        }
        .completed-message {
            position: absolute;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            text-align: center;
            color: var(--md-text-secondary);
        }
        .completed-message h3 {
            font-size: 1.5rem;
            margin-bottom: 12px;
            color: var(--md-success);
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <div class="logo">🧠 Go Quiz</div>
            <div class="header-actions">
                <button class="nav-btn" onclick="showPage('home')">🏠 Главная</button>
                <button class="nav-btn" onclick="showPage('quiz')">🎯 Викторина</button>
                <button class="nav-btn" onclick="showPage('interview')">💼 Собеседование</button>
                <button class="nav-btn" onclick="showPage('stats')">📊 Статистика</button>
                <button class="nav-btn" onclick="showPage('leaderboard')">🏆 Лидеры</button>
                <button class="theme-toggle" onclick="toggleTheme()" title="Переключить тему">
                    <span id="theme-icon">☀️</span>
                </button>
            </div>
        </header>

        <main class="main-content">
            <!-- Home Page -->
            <div id="home" class="page active">
                <div class="hero">
                    <h1>Проверь свои знания Go</h1>
                    <p>Интерактивная викторина по языку программирования Go.<br>
                    Отвечай на вопросы, получай EXP и соревнуйся с другими!</p>
                    <button class="start-btn" onclick="startQuiz()">🚀 Начать викторину</button>
                </div>
                <div class="features">
                    <div class="feature-card">
                        <div class="feature-icon">📚</div>
                        <h3>{{.TotalQuestions}} вопросов</h3>
                        <p>Разные темы и уровни сложности</p>
                    </div>
                    <div class="feature-card">
                        <div class="feature-icon">🎮</div>
                        <h3>Геймификация</h3>
                        <p>EXP, уровни и таблица лидеров</p>
                    </div>
                    <div class="feature-card">
                        <div class="feature-icon">💾</div>
                        <h3>Сохранение</h3>
                        <p>Прогресс сохраняется автоматически</p>
                    </div>
                </div>
            </div>

            <!-- Quiz Page -->
            <div id="quiz" class="page">
                <div class="quiz-container">
                    <div class="quiz-header">
                        <span id="question-counter">Вопрос 1 из {{.TotalQuestions}}</span>
                        <span id="level-display">Уровень 1</span>
                    </div>
                    <div class="progress-bar">
                        <div class="progress-fill" id="progress-fill" style="width: 0%"></div>
                    </div>
                    <div class="question-text" id="question-text">Загрузка вопроса...</div>
                    <div class="options" id="options-container">
                        <!-- Options will be inserted here -->
                    </div>
                    <div class="quiz-footer">
                        <div class="exp-badge" id="exp-display">EXP: 0</div>
                        <button class="next-btn" id="next-btn" onclick="nextQuestion()">Следующий вопрос →</button>
                    </div>
                </div>
            </div>

            <!-- Stats Page -->
            <div id="stats" class="page">
                <h2 style="margin-bottom: 30px; text-align: center;">📊 Твоя статистика</h2>
                <div class="stats-grid" id="stats-grid">
                    <div class="stat-card">
                        <div class="stat-value" id="stat-level">-</div>
                        <div class="stat-label">Уровень</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value" id="stat-exp">-</div>
                        <div class="stat-label">Всего EXP</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value" id="stat-correct">-</div>
                        <div class="stat-label">Правильных</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value" id="stat-wrong">-</div>
                        <div class="stat-label">Неправильных</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value" id="stat-progress">-</div>
                        <div class="stat-label">Прогресс</div>
                    </div>
                </div>
                <div class="reset-section">
                    <button class="reset-btn" onclick="resetProgress()">🔄 Сбросить прогресс</button>
                </div>
            </div>

            <!-- Leaderboard Page -->
            <div id="leaderboard" class="page">
                <h2 style="margin-bottom: 30px; text-align: center;">🏆 Таблица лидеров</h2>
                <table class="leaderboard-table">
                    <thead>
                        <tr>
                            <th>#</th>
                            <th>Игрок</th>
                            <th>Уровень</th>
                            <th>EXP</th>
                            <th>Правильных</th>
                        </tr>
                    </thead>
                    <tbody id="leaderboard-body">
                        <!-- Leaderboard entries will be inserted here -->
                    </tbody>
                </table>
            </div>

            <!-- Interview Prep Page -->
            <div id="interview" class="page">
                <h2 style="margin-bottom: 10px; text-align: center;">💼 Gopher, Go Offer</h2>
                <p style="text-align: center; color: var(--md-text-secondary); margin-bottom: 30px;">Вопросы с собеседований для подготовки</p>
                
                <div class="interview-container">
                    <div class="card-stack" id="card-stack">
                        <!-- Cards will be inserted here by JS -->
                    </div>
                    
                    <div class="interview-controls">
                        <button class="control-btn skip-btn" onclick="skipCard()" title="Пропустить вопрос">
                            <span>✕</span>
                            <span>Пропустить</span>
                        </button>
                        <button class="control-btn show-answer-btn" onclick="showAnswer()" title="Показать ответ">
                            <span>👁</span>
                            <span>Ответ</span>
                        </button>
                    </div>
                    
                    <div class="progress-indicator" id="progress-indicator">
                        <span id="current-card-num">1</span> / <span id="total-cards">26</span>
                    </div>
                </div>
            </div>
        </main>
    </div>

    <script>
        // User ID stored in localStorage
        let userId = localStorage.getItem('goquiz_user_id');
        if (!userId) {
            userId = 'user_' + Math.random().toString(36).substr(2, 9);
            localStorage.setItem('goquiz_user_id', userId);
        }

        // Set cookie for server to read
        document.cookie = "user_id=" + userId + "; path=/; max-age=31536000";

        // Helper to get headers with user ID
        function getHeaders() {
            return {
                'Content-Type': 'application/json',
                'X-User-ID': userId
            };
        }

        let currentQuestion = null;
        let answered = false;

        // Theme toggle
        const savedTheme = localStorage.getItem('goquiz_theme');
        if (savedTheme === 'light') {
            document.body.classList.add('light-theme');
            document.getElementById('theme-icon').textContent = '🌙';
        }

        function toggleTheme() {
            document.body.classList.toggle('light-theme');
            const icon = document.getElementById('theme-icon');
            if (document.body.classList.contains('light-theme')) {
                icon.textContent = '🌙';
                localStorage.setItem('goquiz_theme', 'light');
            } else {
                icon.textContent = '☀️';
                localStorage.setItem('goquiz_theme', 'dark');
            }
        }

        // Toggle interview Q&A answer
        function toggleAnswer(element) {
            const card = element.closest('.qa-card');
            card.classList.toggle('active');
        }

        // Interview cards data
        const interviewCards = [
            {
                question: "Как реализовано ООП в Go?",
                answer: "<p>В Go нет классической реализации ООП, так как он не объектно-ориентированный язык. При этом в Go есть свои приближения:</p><ul><li><strong>Структуры</strong> — типы, в которые можно включать другие типы</li><li><strong>Методы</strong> — функции с получателем</li><li><strong>Интерфейсы</strong> — контракты на поведение</li></ul>"
            },
            {
                question: "Как реализовано наследование в Go?",
                answer: "<p>Как такового наследования в Go нет, но есть <strong>встраивание (embedding) структур</strong>. Мы можем включать одни структуры в другие. При этом методы встраиваемой структуры становятся доступны в родительской.</p><p><strong>Если в родительской и дочерней структуре есть методы с одинаковым названием</strong> — реализация родительского метода будет переписана реализацией дочернего метода.</p>"
            },
            {
                question: "Как реализована инкапсуляция в Go?",
                answer: "<p>Инкапсуляция в Go реализована через <strong>регистр первой буквы</strong> названия:</p><ul><li><strong>Верхний регистр</strong> (Exported) — доступно за рамками пакета</li><li><strong>Нижний регистр</strong> (Unexported) — доступно только в рамках пакета</li></ul>"
            },
            {
                question: "Как реализован полиморфизм в Go?",
                answer: "<p>Полиморфизм в Go реализован через <strong>интерфейсы</strong>. Интерфейс — это контракт на определённое поведение. Типы реализуют методы, удовлетворяющие интерфейсу, и мы можем работать со всеми такими типами как с единым интерфейсным типом.</p><p>Реализация интерфейса в Go <strong>неявная</strong> — не нужно объявлять implements.</p>"
            },
            {
                question: "Как работает append для слайсов?",
                answer: "<p>Функция <code>append</code> принимает слайс и переменное количество элементов. Она расширяет слайс за пределы его <code>len</code>, возвращая новый слайс:</p><ul><li>Если количество элементов не превышает <code>cap</code> — возвращается слайс, ссылающийся на тот же базовый массив</li><li>Если превышает — создаётся новый массив, и возвращается слайс с ссылкой на него</li></ul>"
            },
            {
                question: "Какой размер выделяется при расширении слайса?",
                answer: "<p>При расширении слайса:</p><ul><li>Если требуемая <code>cap</code> больше чем вдвое исходной — новая <code>cap</code> равна требуемой</li><li>Если <code>len</code> текущего слайса меньше 1024 — новая <code>cap</code> в два раза больше</li><li>Иначе — емкость увеличивается в цикле на 25% пока не будет обработано переполнение</li></ul>"
            },
            {
                question: "Как реализована map в Go?",
                answer: "<p>Map в Go — структура, реализующая операции хеширования. Ссылается на <strong>bucket</strong> (ведра), каждый содержит:</p><ul><li>8 экстра-бит для доступа к значениям</li><li>Ссылку на следующий коллизионный bucket</li><li>8 пар ключ-значение в массиве</li></ul>"
            },
            {
                question: "Почему нельзя брать ссылку на значение map по ключу?",
                answer: "<p>Map поддерживает <strong>процедуру эвакуации</strong> — перенос значений из одной области памяти в другую. Значения могут перемещаться, поэтому ссылка на них будет недействительной.</p>"
            },
            {
                question: "Что такое эвакуация в map?",
                answer: "<p>Эвакуация — процесс переноса значений map из одной области памяти в другую. Происходит когда среднее количество значений в bucket достигает ~6.5 (максимум 8). Начинается расширение map, старые и новые данные связаны на время процесса.</p>"
            },
            {
                question: "Как происходит поиск по ключу в map?",
                answer: "<ol><li>Вычисляется хэш от ключа</li><li>По хэшу и размеру bucket вычисляется bucket</li><li>Вычисляется дополнительный хэш (первые 8 бит)</li><li>В bucket сравниваются 8 дополнительных хэшей</li><li>При совпадении — возвращается значение</li><li>При несовпадении — переход в следующий bucket</li><li>Если не найдено — дефолтное значение</li></ol>"
            },
            {
                question: "Что такое пустой интерфейс?",
                answer: "<p><code>interface{}</code> (или <code>any</code> в Go 1.18+) — интерфейс без методов. Ему соответствует <strong>абсолютно любой тип</strong>, так как не нужно реализовывать ни одного метода.</p>"
            },
            {
                question: "Что такое nil интерфейс?",
                answer: "<p>Интерфейс в Go — структура с ссылкой на значение и itab (служебная информация). <strong>nil интерфейс</strong> не ссылается на значение, но содержит itab. Булево сравнение <code>nil == интерфейс</code> всегда ложное.</p>"
            },
            {
                question: "Как преобразовать интерфейс к типу?",
                answer: "<p>Используется <strong>type assertion</strong>: <code>value, ok := interfaceValue.(Type)</code></p><ul><li><code>ok == true</code> — преобразование успешно</li><li><code>ok == false</code> — возврат дефолтного значения типа</li></ul>"
            },
            {
                question: "Зачем используется defer?",
                answer: "<p><code>defer</code> используется для <strong>отложенного вызова функции</strong>. Функция с defer выполняется перед выходом из внешней функции. Аргументы оцениваются немедленно в момент объявления defer.</p>"
            },
            {
                question: "Порядок выполнения нескольких defer?",
                answer: "<p>defer добавляет функцию в <strong>стек</strong>. При возврате вызовы выполняются в порядке <strong>LIFO</strong> (last in first out) — от последнего к первому.</p>"
            },
            {
                question: "Как работает сборщик мусора в Go?",
                answer: "<p>GC в Go использует подход <strong>\"пометка и освобождение\"</strong> (mark-and-sweep). Автоматически отслеживает использование памяти и освобождает неиспользуемые ресурсы.</p>"
            },
            {
                question: "Сколько памяти занимает горутина?",
                answer: "<p>Горутина по умолчанию занимает <strong>2KB стековой памяти</strong>. Стек может расти или уменьшаться динамически по мере необходимости.</p>"
            },
            {
                question: "Способы синхронизации в Go?",
                answer: "<ul><li><strong>Каналы (channels)</strong> — обмен данными между горутинами</li><li><strong>Мьютексы (mutexes)</strong> — защита доступа к общим данным</li><li><strong>Wait Groups</strong> — координация завершения горутин</li></ul>"
            },
            {
                question: "Сложность поиска по срезу и map?",
                answer: "<ul><li><strong>Срез (slice)</strong>: O(n) — линейный поиск</li><li><strong>Map</strong>: O(1) — поиск по ключу за константное время</li></ul>"
            },
            {
                question: "Что делает default в select?",
                answer: "<p><code>default</code> в <code>select</code> выполняется когда <strong>ни один из каналов не готов</strong> для обмена сообщениями. Позволяет избежать блокировки.</p>"
            },
            {
                question: "Что такое nil канал?",
                answer: "<p>Nil канал — неинициализированный канал (не создан через <code>make</code>).</p><ul><li><strong>Чтение из nil канала</strong> — блокировка навсегда</li><li><strong>Запись в nil канал</strong> — блокировка навсегда</li></ul>"
            },
            {
                question: "Безопасен ли слайс для параллелизма?",
                answer: "<p><strong>Нет, слайс не потокобезопасен</strong>. Внутренняя структура может привести к гонкам данных при доступе из разных горутин. Используйте мьютексы или каналы для синхронизации.</p>"
            },
            {
                question: "Правила выделения переменной в горутине?",
                answer: "<p>Необходимо <strong>передавать переменные как параметры</strong> при создании горутины, чтобы избежать состояния гонки. Каждая горутина имеет собственный стек, но доступ к общей памяти может быть проблематичным.</p>"
            },
            {
                question: "Паттерн Singleton в Go?",
                answer: "<p>Singleton гарантирует один экземпляр класса и глобальную точку доступа. В Go реализуется через <code>sync.Once</code>:</p><pre><code>var once sync.Once\nvar instance *Singleton\n\nfunc GetInstance() *Singleton {\n    once.Do(func() {\n        instance = &Singleton{}\n    })\n    return instance\n}</code></pre>"
            },
            {
                question: "Как устроены контексты в Go?",
                answer: "<p><code>context.Context</code> передаётся между горутинами для:</p><ul><li>Управления отменой операций</li><li>Сигнализации об ошибках</li><li>Передачи значений</li><li>Управления временем жизни запросов</li></ul>"
            },
            {
                question: "Как обработать панику с defer и recovery?",
                answer: "<p><code>recover()</code> вызывается в отложенной функции для обработки паники:</p><pre><code>func safe() {\n    defer func() {\n        if r := recover(); r != nil {\n            fmt.Println(\"Recovered:\", r)\n        }\n    }()\n    panic(\"error\")\n}</code></pre>"
            }
        ];

        let currentCardIndex = 0;

        // Render interview card
        function renderCard(index) {
            const stack = document.getElementById('card-stack');
            if (index >= interviewCards.length) {
                stack.innerHTML = '<div class="completed-message"><h3>🎉 Все вопросы пройдены!</h3><p>Вы просмотрели все вопросы для подготовки к собеседованию.</p></div>';
                document.getElementById('current-card-num').textContent = interviewCards.length;
                return;
            }

            const card = interviewCards[index];
            stack.innerHTML = '<div class="interview-card fade-in">' +
                '<div class="card-question">' + card.question + '</div>' +
                '<div class="card-answer" id="current-answer">' + card.answer + '</div>' +
                '</div>';

            document.getElementById('current-card-num').textContent = index + 1;
            document.getElementById('total-cards').textContent = interviewCards.length;
        }

        // Show answer
        function showAnswer() {
            const answer = document.getElementById('current-answer');
            if (answer) {
                answer.classList.add('visible');
            }
        }

        // Skip card with swipe animation
        function skipCard() {
            const card = document.querySelector('.interview-card');
            if (card) {
                card.classList.add('swipe-left');
                setTimeout(() => {
                    currentCardIndex++;
                    renderCard(currentCardIndex);
                }, 300);
            }
        }

        // Page navigation
        function showPage(pageId) {
            document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
            document.getElementById(pageId).classList.add('active');

            document.querySelectorAll('.nav-btn').forEach(b => b.classList.remove('active'));
            if (event && event.target) {
                event.target.classList.add('active');
            }

            if (pageId === 'stats') loadStats();
            if (pageId === 'leaderboard') loadLeaderboard();
            if (pageId === 'interview') {
                currentCardIndex = 0;
                renderCard(currentCardIndex);
            }
        }

        // Start quiz
        async function startQuiz() {
            showPage('quiz');
            document.querySelector('[onclick="showPage(\'quiz\')"]').classList.add('active');
            await loadQuestion();
        }

        // Load question
        async function loadQuestion() {
            answered = false;
            document.getElementById('next-btn').classList.remove('visible');
            document.getElementById('options-container').innerHTML = '<p style="text-align:center;color:#888;">Загрузка...</p>';

            try {
                const res = await fetch('/api/quiz', { headers: getHeaders() });
                const data = await res.json();
                
                if (!data.question) {
                    document.getElementById('question-text').textContent = '🎉 Вы ответили на все вопросы!';
                    document.getElementById('options-container').innerHTML = '';
                    document.getElementById('next-btn').classList.add('visible');
                    document.getElementById('next-btn').onclick = () => { showPage('home'); };
                    return;
                }

                currentQuestion = data.question;
                document.getElementById('question-text').textContent = currentQuestion.question;
                document.getElementById('question-counter').textContent =
                    'Вопрос ' + (data.answered + 1) + ' из ' + data.total;
                document.getElementById('progress-fill').style.width =
                    ((data.answered / data.total) * 100) + '%';

                // Render options
                const container = document.getElementById('options-container');
                container.innerHTML = '';
                currentQuestion.options.forEach((opt, idx) => {
                    const btn = document.createElement('button');
                    btn.className = 'option-btn';
                    btn.textContent = opt;
                    btn.onclick = () => selectAnswer(idx, btn);
                    container.appendChild(btn);
                });

                // Update level display
                await updateStatsDisplay();
            } catch (err) {
                console.error('Error loading question:', err);
                document.getElementById('question-text').textContent = 'Ошибка загрузки вопроса';
            }
        }

        // Select answer
        async function selectAnswer(optionIndex, btn) {
            if (answered) return;
            answered = true;

            // Disable all buttons
            document.querySelectorAll('.option-btn').forEach(b => b.classList.add('disabled'));

            try {
                const res = await fetch('/api/answer', {
                    method: 'POST',
                    headers: getHeaders(),
                    body: JSON.stringify({
                        question_id: currentQuestion.id,
                        option_index: optionIndex
                    })
                });
                const data = await res.json();

                // Highlight correct/wrong
                if (data.correct) {
                    btn.classList.add('correct');
                } else {
                    btn.classList.add('wrong');
                    // Highlight correct answer
                    document.querySelectorAll('.option-btn')[data.correct_option].classList.add('correct');
                }

                // Update EXP display
                document.getElementById('exp-display').textContent = 'EXP: ' + data.new_exp;
                document.getElementById('level-display').textContent = 'Уровень ' + data.new_level;

                // Show next button
                document.getElementById('next-btn').classList.add('visible');
            } catch (err) {
                console.error('Error submitting answer:', err);
            }
        }

        // Next question
        function nextQuestion() {
            loadQuestion();
        }

        // Update stats display
        async function updateStatsDisplay() {
            try {
                const res = await fetch('/api/stats', { headers: getHeaders() });
                const data = await res.json();
                if (data.user) {
                    document.getElementById('level-display').textContent = 'Уровень ' + data.user.level;
                    document.getElementById('exp-display').textContent = 'EXP: ' + data.user.total_exp;
                }
            } catch (err) {
                console.error('Error updating stats:', err);
            }
        }

        // Load stats
        async function loadStats() {
            try {
                const res = await fetch('/api/stats', { headers: getHeaders() });
                const data = await res.json();
                
                if (data.user) {
                    document.getElementById('stat-level').textContent = data.user.level;
                    document.getElementById('stat-exp').textContent = data.user.total_exp;
                    document.getElementById('stat-correct').textContent = data.user.correct_answers;
                    document.getElementById('stat-wrong').textContent = data.user.wrong_answers;
                    document.getElementById('stat-progress').textContent = 
                        Math.round(data.progress) + '%';
                }
            } catch (err) {
                console.error('Error loading stats:', err);
            }
        }

        // Load leaderboard
        async function loadLeaderboard() {
            try {
                const res = await fetch('/api/leaderboard', { headers: getHeaders() });
                const data = await res.json();
                
                const tbody = document.getElementById('leaderboard-body');
                tbody.innerHTML = '';
                
                if (data.entries.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="5" style="text-align:center;padding:40px;">Пока нет игроков. Будь первым!</td></tr>';
                    return;
                }

                data.entries.forEach((entry, idx) => {
                    const tr = document.createElement('tr');
                    tr.className = idx < 3 ? 'rank-' + (idx + 1) : '';
                    tr.innerHTML = 
                        '<td><span class="rank-badge">' + (idx + 1) + '</span></td>' +
                        '<td>' + entry.id + '</td>' +
                        '<td>' + entry.level + '</td>' +
                        '<td>' + entry.total_exp + '</td>' +
                        '<td>' + entry.correct + '</td>';
                    tbody.appendChild(tr);
                });
            } catch (err) {
                console.error('Error loading leaderboard:', err);
            }
        }

        // Reset progress
        async function resetProgress() {
            if (!confirm('Вы уверены? Весь прогресс будет сброшен.')) return;

            try {
                await fetch('/api/reset', { method: 'POST', headers: getHeaders() });
                alert('Прогресс сброшен!');
                loadStats();
            } catch (err) {
                console.error('Error resetting progress:', err);
            }
        }

        // Initial stats update
        updateStatsDisplay();
    </script>
</body>
</html>
`))

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	
	data := struct {
		TotalQuestions int
	}{
		TotalQuestions: len(questions),
	}
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getOrCreateUserID(w http.ResponseWriter, r *http.Request) string {
	// Try cookie first
	cookie, err := r.Cookie("user_id")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Create new user ID and set cookie
	var userID = fmt.Sprintf("user_%d", time.Now().UnixNano())
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    userID,
		Path:     "/",
		MaxAge:   86400 * 365,
		HttpOnly: false,
	})
	return userID
}

func quizHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get or create user ID (sets cookie if new)
	userID := getOrCreateUserID(w, r)
	user := getUser(userID)

	// Find available questions
	askedMap := make(map[int]bool)
	for _, id := range user.AskedQuestions {
		askedMap[id] = true
	}

	var available []Question
	for _, q := range questions {
		if !askedMap[q.ID] {
			available = append(available, q)
		}
	}

	response := QuizResponse{
		Total:    len(questions),
		Answered: len(user.AskedQuestions),
	}

	if len(available) > 0 {
		q := available[rand.Intn(len(available))]
		response.Question = &q

		// Store current question in session
		sessionsMu.Lock()
		sessions[userID] = &Session{
			UserID:    userID,
			CurrentQ:  &q,
			StartTime: time.Now(),
		}
		sessionsMu.Unlock()
	}

	w.Header().Set("Content-Type", "application/json")
	// Cookie is already set by getOrCreateUserID
	json.NewEncoder(w).Encode(response)
}

func answerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := getOrCreateUserID(w, r)
	user := getUser(userID)

	// Find question
	var q *Question
	for i := range questions {
		if questions[i].ID == req.QuestionID {
			q = &questions[i]
			break
		}
	}
	
	if q == nil {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	response := AnswerResponse{
		CorrectOption: q.Correct,
	}

	if req.OptionIndex == q.Correct {
		response.Correct = true
		response.Exp = q.Exp
		user.TotalEXP += q.Exp
		user.CorrectAnswers++
		response.Message = "✅ Правильно!"
	} else {
		response.Correct = false
		user.WrongAnswers++
		response.Message = "❌ Неправильно"
	}

	updateLevel(user)
	response.NewExp = user.TotalEXP
	response.NewLevel = user.Level

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getOrCreateUserID(w, r)
	user := getUser(userID)
	
	response := StatsResponse{
		User:           user,
		TotalQuestions: len(questions),
		Progress:       float64(len(user.AskedQuestions)) / float64(len(questions)) * 100,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func leaderboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	usersMu.RLock()
	defer usersMu.RUnlock()

	type kv struct {
		ID   string
		User *UserData
	}
	var list []kv
	for id, u := range users {
		list = append(list, kv{id, u})
	}

	// Sort by EXP descending
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].User.TotalEXP > list[i].User.TotalEXP {
				list[i], list[j] = list[j], list[i]
			}
		}
	}

	limit := 10
	if len(list) < limit {
		limit = len(list)
	}

	entries := make([]LeaderboardEntry, limit)
	for i := 0; i < limit; i++ {
		entries[i] = LeaderboardEntry{
			ID:        list[i].ID,
			Level:     list[i].User.Level,
			TotalEXP:  list[i].User.TotalEXP,
			Correct:   list[i].User.CorrectAnswers,
		}
	}

	response := LeaderboardResponse{
		Entries: entries,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getOrCreateUserID(w, r)
	user := getUser(userID)
	user.AskedQuestions = []int{}

	response := map[string]string{
		"status": "ok",
		"message": "Прогресс сброшен",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
