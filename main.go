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
    <link href="https://fonts.googleapis.com/css2?family=Montserrat:wght@400;600;700&family=Fira+Code:wght@400;600&display=swap" rel="stylesheet">
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Montserrat', sans-serif;
            min-height: 100vh;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            color: #fff;
            transition: background 0.3s ease;
        }
        body.light-theme {
            background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 50%, #e8eaf6 100%);
            color: #333;
        }
        .container {
            max-width: 900px;
            margin: 0 auto;
            padding: 20px;
        }
        header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 20px 0;
            border-bottom: 1px solid rgba(255,255,255,0.1);
            margin-bottom: 30px;
        }
        body.light-theme header {
            border-bottom-color: rgba(0,0,0,0.1);
        }
        .logo {
            font-size: 1.8rem;
            font-weight: 700;
            background: linear-gradient(135deg, #00d9ff, #00ff88);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        body.light-theme .logo {
            background: linear-gradient(135deg, #1a1a2e, #0f3460);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        .header-actions {
            display: flex;
            gap: 15px;
            align-items: center;
        }
        .theme-toggle, .nav-btn {
            background: rgba(255,255,255,0.1);
            border: 2px solid #00d9ff;
            border-radius: 50%;
            width: 44px;
            height: 44px;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: all 0.3s ease;
            font-size: 1.2rem;
        }
        body.light-theme .theme-toggle,
        body.light-theme .nav-btn {
            background: rgba(0,0,0,0.1);
            border-color: #333;
        }
        .nav-btn {
            border-radius: 25px;
            width: auto;
            padding: 0 20px;
            font-size: 0.9rem;
            font-weight: 600;
        }
        .theme-toggle:hover, .nav-btn:hover {
            transform: scale(1.1);
            background: rgba(255,255,255,0.2);
        }
        body.light-theme .theme-toggle:hover,
        body.light-theme .nav-btn:hover {
            background: rgba(0,0,0,0.15);
        }
        .nav-btn.active {
            background: linear-gradient(135deg, #00d9ff, #00ff88);
            border-color: transparent;
            color: #1a1a2e;
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
            color: #555;
        }
        .start-btn {
            background: linear-gradient(135deg, #00d9ff, #00ff88);
            border: none;
            border-radius: 50px;
            padding: 18px 50px;
            font-size: 1.3rem;
            font-weight: 700;
            color: #1a1a2e;
            cursor: pointer;
            transition: all 0.3s ease;
            box-shadow: 0 10px 30px rgba(0,217,255,0.3);
        }
        .start-btn:hover {
            transform: translateY(-5px);
            box-shadow: 0 15px 40px rgba(0,217,255,0.5);
        }
        .features {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 30px;
            margin-top: 60px;
        }
        .feature-card {
            background: rgba(255,255,255,0.05);
            border-radius: 20px;
            padding: 30px;
            text-align: center;
            border: 1px solid rgba(255,255,255,0.1);
        }
        body.light-theme .feature-card {
            background: rgba(255,255,255,0.8);
            border-color: rgba(0,0,0,0.1);
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
            background: rgba(255,255,255,0.05);
            border-radius: 20px;
            padding: 40px;
            border: 1px solid rgba(255,255,255,0.1);
        }
        body.light-theme .quiz-container {
            background: rgba(255,255,255,0.8);
            border-color: rgba(0,0,0,0.1);
        }
        .quiz-header {
            display: flex;
            justify-content: space-between;
            margin-bottom: 30px;
            font-size: 0.9rem;
            color: #888;
        }
        body.light-theme .quiz-header {
            color: #555;
        }
        .progress-bar {
            width: 100%;
            height: 8px;
            background: rgba(255,255,255,0.1);
            border-radius: 10px;
            overflow: hidden;
            margin-bottom: 30px;
        }
        body.light-theme .progress-bar {
            background: rgba(0,0,0,0.1);
        }
        .progress-fill {
            height: 100%;
            background: linear-gradient(135deg, #00d9ff, #00ff88);
            transition: width 0.3s ease;
        }
        .question-text {
            font-size: 1.4rem;
            margin-bottom: 30px;
            line-height: 1.6;
        }
        .options {
            display: flex;
            flex-direction: column;
            gap: 15px;
        }
        .option-btn {
            background: rgba(255,255,255,0.05);
            border: 2px solid rgba(255,255,255,0.2);
            border-radius: 15px;
            padding: 18px 25px;
            text-align: left;
            cursor: pointer;
            transition: all 0.3s ease;
            font-size: 1rem;
            color: inherit;
        }
        body.light-theme .option-btn {
            background: rgba(0,0,0,0.05);
            border-color: rgba(0,0,0,0.2);
        }
        .option-btn:hover {
            border-color: #00d9ff;
            background: rgba(0,217,255,0.1);
            transform: translateX(10px);
        }
        .option-btn.correct {
            background: rgba(0,255,136,0.2);
            border-color: #00ff88;
        }
        .option-btn.wrong {
            background: rgba(255,71,87,0.2);
            border-color: #ff4757;
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
            background: linear-gradient(135deg, #ffd700, #ffaa00);
            color: #1a1a2e;
            padding: 8px 20px;
            border-radius: 25px;
            font-weight: 700;
            font-size: 0.9rem;
        }
        .next-btn {
            background: linear-gradient(135deg, #00d9ff, #00ff88);
            border: none;
            border-radius: 25px;
            padding: 12px 35px;
            font-size: 1rem;
            font-weight: 700;
            color: #1a1a2e;
            cursor: pointer;
            transition: all 0.3s ease;
            display: none;
        }
        .next-btn:hover {
            transform: scale(1.05);
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
            background: rgba(255,255,255,0.05);
            border-radius: 20px;
            padding: 25px;
            text-align: center;
            border: 1px solid rgba(255,255,255,0.1);
        }
        body.light-theme .stat-card {
            background: rgba(255,255,255,0.8);
            border-color: rgba(0,0,0,0.1);
        }
        .stat-value {
            font-size: 2.5rem;
            font-weight: 700;
            background: linear-gradient(135deg, #00d9ff, #00ff88);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        body.light-theme .stat-value {
            background: linear-gradient(135deg, #1a1a2e, #0f3460);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        .stat-label {
            color: #888;
            margin-top: 10px;
        }
        body.light-theme .stat-label {
            color: #555;
        }
        /* Leaderboard */
        .leaderboard-table {
            width: 100%;
            border-collapse: collapse;
            background: rgba(255,255,255,0.05);
            border-radius: 20px;
            overflow: hidden;
        }
        body.light-theme .leaderboard-table {
            background: rgba(255,255,255,0.8);
        }
        .leaderboard-table th,
        .leaderboard-table td {
            padding: 18px 25px;
            text-align: left;
        }
        .leaderboard-table th {
            background: rgba(0,217,255,0.2);
            font-weight: 600;
            text-transform: uppercase;
            font-size: 0.85rem;
            letter-spacing: 1px;
        }
        .leaderboard-table tr:not(:last-child) {
            border-bottom: 1px solid rgba(255,255,255,0.1);
        }
        body.light-theme .leaderboard-table tr:not(:last-child) {
            border-bottom-color: rgba(0,0,0,0.1);
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
            border-top-color: rgba(0,0,0,0.1);
        }
        .reset-btn {
            background: rgba(255,71,87,0.2);
            border: 2px solid #ff4757;
            color: #ff4757;
            border-radius: 25px;
            padding: 12px 35px;
            cursor: pointer;
            transition: all 0.3s ease;
            font-size: 0.9rem;
        }
        .reset-btn:hover {
            background: rgba(255,71,87,0.3);
            transform: scale(1.05);
        }
        /* Mobile */
        @media (max-width: 600px) {
            .hero h1 { font-size: 2rem; }
            .quiz-container { padding: 25px; }
            .question-text { font-size: 1.1rem; }
            .option-btn { padding: 15px 18px; }
            header { flex-direction: column; gap: 15px; }
            .leaderboard-table th,
            .leaderboard-table td { padding: 12px 15px; font-size: 0.85rem; }
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

        // Page navigation
        function showPage(pageId) {
            document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
            document.getElementById(pageId).classList.add('active');
            
            document.querySelectorAll('.nav-btn').forEach(b => b.classList.remove('active'));
            event.target.classList.add('active');

            if (pageId === 'stats') loadStats();
            if (pageId === 'leaderboard') loadLeaderboard();
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
	userID = fmt.Sprintf("user_%d", time.Now().UnixNano())
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    userID,
		Path:     "/",
		MaxAge:   86400 * 365,
		HttpOnly: false, // Allow JS to read
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
