package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"qwen_test/internal/database"
	"qwen_test/internal/game"

	_ "github.com/mattn/go-sqlite3"
)

// Question представляет вопрос викторины
type Question struct {
	ID       int      `json:"id"`
	Question string   `json:"question"`
	Options  []string `json:"options"`
	Correct  int      `json:"correct"`
	Exp      int      `json:"exp"`
}

// API Response types
type AnswerResponse struct {
	Correct       bool   `json:"correct"`
	Exp           int    `json:"exp"`
	CorrectOption int    `json:"correct_option"`
	Message       string `json:"message"`
	NewExp        int    `json:"new_exp"`
	NewLevel      int    `json:"new_level"`
	LevelUp       bool   `json:"level_up"`
}

type QuizResponse struct {
	Question *Question `json:"question"`
	Total    int       `json:"total"`
	Answered int       `json:"answered"`
}

type StatsResponse struct {
	Player         *game.Player `json:"player"`
	TotalQuestions int          `json:"total_questions"`
	Progress       float64      `json:"progress"`
	SkillPoints    int          `json:"skill_points"`
}

type LeaderboardEntry struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Level     int    `json:"level"`
	TotalEXP  int    `json:"total_exp"`
	Correct   int    `json:"correct"`
	Rating    int    `json:"rating"`
}

type LeaderboardResponse struct {
	Entries []LeaderboardEntry `json:"entries"`
}

type SkillsResponse struct {
	Tree        *game.SkillTree `json:"tree"`
	Bonuses     map[string]int  `json:"bonuses"`
}

type QuestsResponse struct {
	System *game.QuestSystem `json:"system"`
}

type AchievementsResponse struct {
	System        *game.AchievementSystem `json:"system"`
	UnlockedCount int                     `json:"unlocked_count"`
	TotalCount    int                     `json:"total_count"`
}

type UpgradeSkillRequest struct {
	SkillID string `json:"skill_id"`
}

type StudyGoRequest struct {
	Minutes int `json:"minutes"`
}

type RestRequest struct {
	Minutes int `json:"minutes"`
}

// Глобальные переменные
var (
	questions     []Question
	questionsFile = "questions.json"
	dbPath        = "qwen_test.db"
)

// Кэш игроков в памяти
var (
	playersCache      = make(map[string]*game.Player)
	skillTreesCache   = make(map[string]*game.SkillTree)
	questSystemsCache = make(map[string]*game.QuestSystem)
	achievementsCache = make(map[string]*game.AchievementSystem)
	cacheMu           sync.RWMutex
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Загружаем вопросы
	if err := loadQuestions(); err != nil {
		log.Fatal("Ошибка загрузки вопросов:", err)
	}
	log.Printf("📚 Загружено %d вопросов", len(questions))

	// Инициализируем базу данных
	if err := database.InitDB(dbPath); err != nil {
		log.Fatal("Ошибка инициализации БД:", err)
	}
	defer database.CloseDB()
	log.Printf("💾 База данных инициализирована: %s", dbPath)

	// Загружаем игроков из БД
	loadPlayersCache()

	// Автосохранение каждые 5 минут
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			saveAllPlayers()
		}
	}()

	// HTTP handlers
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/quiz", quizHandler)
	http.HandleFunc("/api/answer", answerHandler)
	http.HandleFunc("/api/stats", statsHandler)
	http.HandleFunc("/api/leaderboard", leaderboardHandler)
	http.HandleFunc("/api/reset", resetHandler)
	http.HandleFunc("/api/skills", skillsHandler)
	http.HandleFunc("/api/skills/upgrade", upgradeSkillHandler)
	http.HandleFunc("/api/quests", questsHandler)
	http.HandleFunc("/api/achievements", achievementsHandler)
	http.HandleFunc("/api/study", studyGoHandler)
	http.HandleFunc("/api/rest", restHandler)
	http.HandleFunc("/api/backup", backupHandler)

	port := ":8080"
	fmt.Printf("🚀 Go Quiz Web Server starting on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// --- Загрузка вопросов ---

func loadQuestions() error {
	data, err := os.ReadFile(questionsFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &questions)
}

// --- Работа с кэшем игроков ---

func loadPlayersCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	rows, err := database.DB.Query("SELECT user_id, name, level, experience, go_knowledge, focus, willpower, money, dopamine, play_time, days_played, current_day, current_hour, correct_answers, wrong_answers, asked_questions, created_at FROM players")
	if err != nil {
		log.Println("⚠️  Не удалось загрузить игроков из БД:", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var player game.Player
		var askedQuestionsStr string
		var createdAt time.Time

		err := rows.Scan(
			&player.ID, &player.Name, &player.Level, &player.Experience,
			&player.GoKnowledge, &player.Focus, &player.Willpower,
			&player.Money, &player.Dopamine, &player.PlayTime,
			&player.DaysPlayed, &player.CurrentDay, &player.CurrentHour,
			&player.CorrectAnswers, &player.WrongAnswers,
			&askedQuestionsStr, &createdAt,
		)
		if err != nil {
			log.Println("⚠️  Ошибка сканирования игрока:", err)
			continue
		}

		json.Unmarshal([]byte(askedQuestionsStr), &player.AskedQuestions)
		player.CreatedAt = createdAt
		player.SkillBonuses = make(map[string]int)

		playersCache[player.ID] = &player
		count++
	}

	log.Printf("👥 Загружено %d игроков из БД", count)
}

func getPlayer(userID string) *game.Player {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	if player, ok := playersCache[userID]; ok {
		return player
	}

	// Создаём нового игрока
	player := game.NewPlayer(userID, "Player"+userID[len(userID)-4:])
	playersCache[userID] = player

	// Сохраняем в БД
	savePlayerToDB(player)

	// Создаём дерево навыков
	skillTree := game.NewSkillTree(userID)
	skillTreesCache[userID] = skillTree

	// Создаём систему квестов
	questSystem := game.NewQuestSystem(userID)
	questSystemsCache[userID] = questSystem

	// Создаём систему достижений
	achievements := game.NewAchievementSystem(userID)
	achievementsCache[userID] = achievements

	// Применяем бонусы от навыков
	player.ApplySkillBonuses(skillTree)

	return player
}

func getSkillTree(userID string) *game.SkillTree {
	cacheMu.RLock()
	if tree, ok := skillTreesCache[userID]; ok {
		cacheMu.RUnlock()
		return tree
	}
	cacheMu.RUnlock()

	// Загружаем из БД или создаём новое
	tree := loadSkillTreeFromDB(userID)
	if tree == nil {
		tree = game.NewSkillTree(userID)
	}

	cacheMu.Lock()
	skillTreesCache[userID] = tree
	cacheMu.Unlock()

	return tree
}

func getQuestSystem(userID string) *game.QuestSystem {
	cacheMu.RLock()
	if qs, ok := questSystemsCache[userID]; ok {
		cacheMu.RUnlock()
		return qs
	}
	cacheMu.RUnlock()

	qs := loadQuestSystemFromDB(userID)
	if qs == nil {
		qs = game.NewQuestSystem(userID)
	}

	cacheMu.Lock()
	questSystemsCache[userID] = qs
	cacheMu.Unlock()

	return qs
}

func getAchievements(userID string) *game.AchievementSystem {
	cacheMu.RLock()
	if as, ok := achievementsCache[userID]; ok {
		cacheMu.RUnlock()
		return as
	}
	cacheMu.RUnlock()

	as := loadAchievementsFromDB(userID)
	if as == nil {
		as = game.NewAchievementSystem(userID)
	}

	cacheMu.Lock()
	achievementsCache[userID] = as
	cacheMu.Unlock()

	return as
}

func saveAllPlayers() {
	cacheMu.RLock()
	defer cacheMu.RUnlock()

	for _, player := range playersCache {
		savePlayerToDB(player)
		saveSkillTreeToDB(player.ID, skillTreesCache[player.ID])
		saveQuestSystemToDB(player.ID, questSystemsCache[player.ID])
		saveAchievementsToDB(player.ID, achievementsCache[player.ID])
	}

	log.Println("💾 Все игроки сохранены в БД")
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
    <link rel="stylesheet" href="/static/modern.css">
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <header>
            <div class="logo">🧠 Go Quiz</div>
            <div class="header-actions">
                <button class="nav-btn" onclick="showPage('home')">🏠</button>
                <button class="nav-btn" onclick="showPage('quiz')">🎯</button>
                <button class="nav-btn" onclick="showPage('study')">📚</button>
                <button class="nav-btn" onclick="showPage('skills')">🌳</button>
                <button class="nav-btn" onclick="showPage('quests')">📋</button>
                <button class="nav-btn" onclick="showPage('achievements')">🏆</button>
                <button class="nav-btn" onclick="showPage('stats')">📊</button>
                <button class="nav-btn" onclick="showPage('leaderboard')">👑</button>
                <button class="theme-toggle" onclick="toggleTheme()">☀️</button>
            </div>
        </header>

        <main class="main-content">
            <!-- Home -->
            <div id="home" class="page active">
                <div class="hero">
                    <h1>Прокачай знания Go</h1>
                    <p>Викторина + RPG элементы: уровни, навыки, достижения!</p>
                    <button class="start-btn" onclick="startQuiz()">🚀 Начать</button>
                </div>
                <div class="features">
                    <div class="feature-card">
                        <div class="feature-icon">📚</div>
                        <h3>{{.TotalQuestions}} вопросов</h3>
                        <p>Разные темы и сложности</p>
                    </div>
                    <div class="feature-card">
                        <div class="feature-icon">🌳</div>
                        <h3>Навыки</h3>
                        <p>12 навыков в 4 категориях</p>
                    </div>
                    <div class="feature-card">
                        <div class="feature-icon">🏆</div>
                        <h3>Достижения</h3>
                        <p>23 достижения для коллекции</p>
                    </div>
                </div>
            </div>

            <!-- Quiz -->
            <div id="quiz" class="page">
                <div class="quiz-container">
                    <div class="quiz-header">
                        <span id="question-counter">Вопрос 1 из {{.TotalQuestions}}</span>
                        <span id="level-display">Уровень 1</span>
                    </div>
                    <div class="progress-bar">
                        <div class="progress-fill" id="progress-fill" style="width: 0%"></div>
                    </div>
                    <div class="question-text" id="question-text">Загрузка...</div>
                    <div class="options" id="options-container"></div>
                    <div class="quiz-footer">
                        <div class="exp-badge" id="exp-display">EXP: 0</div>
                        <button class="next-btn" id="next-btn" onclick="nextQuestion()">Далее →</button>
                    </div>
                </div>
            </div>

            <!-- Study -->
            <div id="study" class="page">
                <h2 style="margin-bottom: 30px; text-align: center;">📚 Обучение и отдых</h2>
                <div class="action-cards">
                    <div class="action-card" onclick="studyGo(30)">
                        <div class="action-icon">📖</div>
                        <div class="action-title">Изучить Go (30 мин)</div>
                        <div class="action-desc">+15 EXP, +6 Знание Go, +10 Дофамин</div>
                        <div class="action-reward">🎯 Квест: 30 минут Go</div>
                    </div>
                    <div class="action-card" onclick="studyGo(60)">
                        <div class="action-icon">📖</div>
                        <div class="action-title">Изучить Go (60 мин)</div>
                        <div class="action-desc">+30 EXP, +12 Знание Go, +20 Дофамин</div>
                        <div class="action-reward">🎯 Квест: 30 минут Go</div>
                    </div>
                    <div class="action-card" onclick="rest(15)">
                        <div class="action-icon">💤</div>
                        <div class="action-title">Отдохнуть (15 мин)</div>
                        <div class="action-desc">+7 Фокус, +5 Дофамин</div>
                        <div class="action-reward">😌 Восстановление</div>
                    </div>
                    <div class="action-card" onclick="rest(30)">
                        <div class="action-icon">💤</div>
                        <div class="action-title">Отдохнуть (30 мин)</div>
                        <div class="action-desc">+15 Фокус, +10 Дофамин</div>
                        <div class="action-reward">😌 Восстановление</div>
                    </div>
                </div>
                <div style="text-align: center;">
                    <button class="backup-btn" onclick="createBackup()">💾 Создать бэкап</button>
                </div>
            </div>

            <!-- Skills -->
            <div id="skills" class="page">
                <h2 style="margin-bottom: 20px; text-align: center;">🌳 Дерево навыков</h2>
                <div class="skill-points-display" id="skill-points-display">
                    ✨ Очки навыков: 0 (всего: 0)
                </div>
                <div class="skills-container" id="skills-container">
                    Загрузка...
                </div>
            </div>

            <!-- Quests -->
            <div id="quests" class="page">
                <h2 style="margin-bottom: 20px; text-align: center;">📋 Ежедневные квесты</h2>
                <div class="quests-container" id="quests-container">
                    Загрузка...
                </div>
            </div>

            <!-- Achievements -->
            <div id="achievements" class="page">
                <h2 style="margin-bottom: 20px; text-align: center;">🏆 Достижения</h2>
                <div class="achievements-container" id="achievements-container">
                    Загрузка...
                </div>
            </div>

            <!-- Stats -->
            <div id="stats" class="page">
                <h2 style="margin-bottom: 30px; text-align: center;">📊 Статистика</h2>
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
                        <div class="stat-value" id="stat-knowledge">-</div>
                        <div class="stat-label">Знание Go</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value" id="stat-focus">-</div>
                        <div class="stat-label">Фокус</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value" id="stat-willpower">-</div>
                        <div class="stat-label">Сила воли</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value" id="stat-rating">-</div>
                        <div class="stat-label">Рейтинг</div>
                    </div>
                </div>
                <div class="reset-section">
                    <button class="reset-btn" onclick="resetProgress()">🔄 Сбросить прогресс</button>
                </div>
            </div>

            <!-- Leaderboard -->
            <div id="leaderboard" class="page">
                <h2 style="margin-bottom: 30px; text-align: center;">👑 Таблица лидеров</h2>
                <table class="leaderboard-table">
                    <thead>
                        <tr>
                            <th>#</th>
                            <th>Игрок</th>
                            <th>Уровень</th>
                            <th>Рейтинг</th>
                            <th>Правильных</th>
                        </tr>
                    </thead>
                    <tbody id="leaderboard-body"></tbody>
                </table>
            </div>
        </main>
    </div>
    <script src="/static/app.js"></script>
</body>
</html>
`))

// --- Handlers ---

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data := struct{ TotalQuestions int }{TotalQuestions: len(questions)}
	tmpl.Execute(w, data)
}

func quizHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	player := getPlayer(userID)

	cacheMu.RLock()
	askedQ := player.AskedQuestions
	cacheMu.RUnlock()

	// Выбираем случайный вопрос, который ещё не задавали
	availableQ := make([]int, 0)
	for i, q := range questions {
		found := false
		for _, asked := range askedQ {
			if asked == q.ID {
				found = true
				break
			}
		}
		if !found {
			availableQ = append(availableQ, i)
		}
	}

	if len(availableQ) == 0 {
		// Все вопросы заданы, начинаем сначала
		availableQ = make([]int, len(questions))
		for i := range questions {
			availableQ[i] = i
		}
	}

	idx := availableQ[rand.Intn(len(availableQ))]
	question := questions[idx]

	cacheMu.Lock()
	player.AskedQuestions = append(player.AskedQuestions, question.ID)
	cacheMu.Unlock()

	resp := QuizResponse{
		Question: &question,
		Total:    len(questions),
		Answered: len(player.AskedQuestions) - 1,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func answerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		QuestionID  int `json:"question_id"`
		OptionIndex int `json:"option_index"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	player := getPlayer(userID)

	// Находим вопрос
	var question *Question
	for i := range questions {
		if questions[i].ID == req.QuestionID {
			question = &questions[i]
			break
		}
	}

	if question == nil {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	correct := req.OptionIndex == question.Correct
	expGain := 0
	levelUp := false

	if correct {
		expGain = question.Exp
		cacheMu.Lock()
		player.CorrectAnswers++
		cacheMu.Unlock()
	} else {
		cacheMu.Lock()
		player.WrongAnswers++
		cacheMu.Unlock()
	}

	cacheMu.Lock()
	oldLevel := player.Level
	player.AddExperience(expGain)
	levelUp = player.Level > oldLevel
	newExp := player.Experience
	newLevel := player.Level
	cacheMu.Unlock()

	// Обновляем квесты
	questSystem := getQuestSystem(userID)
	if correct {
		questSystem.UpdateProgress("study_30", expGain)
	}

	resp := AnswerResponse{
		Correct:       correct,
		Exp:           expGain,
		CorrectOption: question.Correct,
		Message:       map[bool]string{true: "✅ Правильно!", false: "❌ Неправильно"}[correct],
		NewExp:        newExp,
		NewLevel:      newLevel,
		LevelUp:       levelUp,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	player := getPlayer(userID)
	tree := getSkillTree(userID)
	questSystem := getQuestSystem(userID)

	// Проверяем достижения
	achievements := getAchievements(userID)
	unlocked := achievements.CheckAchievements(player, tree, questSystem)
	if len(unlocked) > 0 {
		log.Printf("🏆 Игрок %s разблокировал достижения: %v", userID, unlocked)
	}

	// Начисляем очки навыков за уровень
	skillPoints := game.GetSkillPointsForLevel(player.Level)
	cacheMu.Lock()
	tree.EarnSkillPoints(skillPoints)
	cacheMu.Unlock()

	resp := StatsResponse{
		Player:         player,
		TotalQuestions: len(questions),
		Progress:       float64(len(player.AskedQuestions)) / float64(len(questions)) * 100,
		SkillPoints:    tree.SkillPoints,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func leaderboardHandler(w http.ResponseWriter, r *http.Request) {
	cacheMu.RLock()
	defer cacheMu.RUnlock()

	entries := make([]LeaderboardEntry, 0, len(playersCache))
	for _, player := range playersCache {
		entries = append(entries, LeaderboardEntry{
			ID:        player.ID,
			Name:      player.Name,
			Level:     player.Level,
			TotalEXP:  player.Experience,
			Correct:   player.CorrectAnswers,
			Rating:    player.GetRating(),
		})
	}

	// Сортируем по рейтингу
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].Rating < entries[j].Rating {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Топ-10
	if len(entries) > 10 {
		entries = entries[:10]
	}

	resp := LeaderboardResponse{Entries: entries}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	cacheMu.Lock()
	if player, ok := playersCache[userID]; ok {
		player.Experience = 0
		player.Level = 1
		player.CorrectAnswers = 0
		player.WrongAnswers = 0
		player.AskedQuestions = []int{}
		player.GoKnowledge = 0
		player.Focus = 100
		player.Willpower = 100
		player.Dopamine = 100
	}
	if tree, ok := skillTreesCache[userID]; ok {
		*tree = *game.NewSkillTree(userID)
	}
	if qs, ok := questSystemsCache[userID]; ok {
		*qs = *game.NewQuestSystem(userID)
	}
	if as, ok := achievementsCache[userID]; ok {
		*as = *game.NewAchievementSystem(userID)
	}
	cacheMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func skillsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	tree := getSkillTree(userID)
	bonuses := tree.GetTotalBonuses()

	resp := SkillsResponse{
		Tree:    tree,
		Bonuses: bonuses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func upgradeSkillHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UpgradeSkillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	tree := getSkillTree(userID)
	player := getPlayer(userID)

	success, msg := tree.UpgradeSkill(req.SkillID)

	if success {
		// Применяем бонусы к игроку
		player.ApplySkillBonuses(tree)
	}

	resp := map[string]string{"message": msg}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func questsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	qs := getQuestSystem(userID)

	resp := QuestsResponse{System: qs}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func achievementsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	player := getPlayer(userID)
	tree := getSkillTree(userID)
	qs := getQuestSystem(userID)
	as := getAchievements(userID)

	// Проверяем достижения
	as.CheckAchievements(player, tree, qs)

	resp := AchievementsResponse{
		System:        as,
		UnlockedCount: as.GetUnlockedCount(),
		TotalCount:    as.GetTotalCount(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func studyGoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StudyGoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	player := getPlayer(userID)
	msg := player.StudyGo(req.Minutes)

	// Обновляем квест
	questSystem := getQuestSystem(userID)
	questSystem.UpdateProgress("study_30", req.Minutes)

	resp := map[string]string{"message": msg}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func restHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	player := getPlayer(userID)
	msg := player.Rest(req.Minutes)

	resp := map[string]string{"message": msg}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func backupHandler(w http.ResponseWriter, r *http.Request) {
	backupPath, err := database.CreateBackup(dbPath)
	if err != nil {
		resp := map[string]string{"message": "❌ Ошибка бэкапа: " + err.Error()}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Очищаем старые бэкапы
	database.CleanupOldBackups(7)

	resp := map[string]string{"message": "Бэкап создан: " + backupPath}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// --- Database save/load helpers ---

func savePlayerToDB(player *game.Player) {
	askedJSON, _ := json.Marshal(player.AskedQuestions)

	_, err := database.DB.Exec(`
		INSERT OR REPLACE INTO players 
		(user_id, name, level, experience, go_knowledge, focus, willpower, money, dopamine, 
		 play_time, days_played, current_day, current_hour, correct_answers, wrong_answers, 
		 asked_questions, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, player.ID, player.Name, player.Level, player.Experience, player.GoKnowledge,
		player.Focus, player.Willpower, player.Money, player.Dopamine,
		player.PlayTime, player.DaysPlayed, player.CurrentDay, player.CurrentHour,
		player.CorrectAnswers, player.WrongAnswers, string(askedJSON),
		player.CreatedAt, time.Now())

	if err != nil {
		log.Printf("⚠️  Ошибка сохранения игрока %s: %v", player.ID, err)
	}
}

func loadSkillTreeFromDB(userID string) *game.SkillTree {
	tree := game.NewSkillTree(userID)

	var skillPoints, totalPoints int
	err := database.DB.QueryRow(`
		SELECT skill_points, total_points FROM skill_trees WHERE user_id = ?
	`, userID).Scan(&skillPoints, &totalPoints)

	if err == nil {
		tree.SkillPoints = skillPoints
		tree.TotalPoints = totalPoints
	}

	// Загружаем навыки
	rows, err := database.DB.Query(`
		SELECT skill_id, level, unlocked FROM skills WHERE user_id = ?
	`, userID)
	if err != nil {
		return tree
	}
	defer rows.Close()

	for rows.Next() {
		var skillID string
		var level, unlocked int
		if err := rows.Scan(&skillID, &level, &unlocked); err != nil {
			continue
		}
		if skill, ok := tree.Skills[skillID]; ok {
			skill.Level = level
			skill.Unlocked = unlocked == 1
		}
	}

	return tree
}

func saveSkillTreeToDB(userID string, tree *game.SkillTree) {
	if tree == nil {
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	// Сохраняем дерево
	tx.Exec(`
		INSERT OR REPLACE INTO skill_trees (user_id, skill_points, total_points)
		VALUES (?, ?, ?)
	`, userID, tree.SkillPoints, tree.TotalPoints)

	// Сохраняем навыки
	for _, skill := range tree.Skills {
		tx.Exec(`
			INSERT OR REPLACE INTO skills (user_id, skill_id, level, unlocked)
			VALUES (?, ?, ?, ?)
		`, userID, skill.ID, skill.Level, boolToInt(skill.Unlocked))
	}

	tx.Commit()
}

func loadQuestSystemFromDB(userID string) *game.QuestSystem {
	qs := game.NewQuestSystem(userID)

	var day, streak, totalCompleted int
	err := database.DB.QueryRow(`
		SELECT 1, 0, 0
	`).Scan(&day, &streak, &totalCompleted)

	if err == nil {
		qs.Day = day
		qs.Streak = streak
		qs.TotalCompleted = totalCompleted
	}

	return qs
}

func saveQuestSystemToDB(userID string, qs *game.QuestSystem) {
	if qs == nil {
		return
	}
	// Заглушка для сохранения
}

func loadAchievementsFromDB(userID string) *game.AchievementSystem {
	as := game.NewAchievementSystem(userID)

	rows, err := database.DB.Query(`
		SELECT achievement_id, unlocked_at FROM achievements WHERE user_id = ?
	`, userID)
	if err != nil {
		return as
	}
	defer rows.Close()

	for rows.Next() {
		var achievementID string
		var unlockedAt time.Time
		if err := rows.Scan(&achievementID, &unlockedAt); err != nil {
			continue
		}
		if achievement, ok := as.Achievements[achievementID]; ok {
			achievement.Unlocked = true
			achievement.UnlockedAt = unlockedAt
		}
	}

	return as
}

func saveAchievementsToDB(userID string, as *game.AchievementSystem) {
	if as == nil {
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	for _, achievement := range as.Achievements {
		if achievement.Unlocked {
			tx.Exec(`
				INSERT OR REPLACE INTO achievements (user_id, achievement_id, unlocked_at)
				VALUES (?, ?, ?)
			`, userID, achievement.ID, achievement.UnlockedAt)
		}
	}

	tx.Commit()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
