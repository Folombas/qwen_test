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

	"qwen_test/internal/admin"
	"qwen_test/internal/auth"
	"qwen_test/internal/database"
	"qwen_test/internal/game"
	"qwen_test/internal/social"

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

	// Запускаем миграции для аутентификации
	if err := database.RunAuthMigrations(); err != nil {
		log.Fatal("Ошибка миграций auth:", err)
	}

	// Запускаем миграции для админ-панели
	if err := database.RunAdminMigrations(); err != nil {
		log.Fatal("Ошибка миграций admin:", err)
	}

	// Запускаем миграции для социальных функций
	if err := database.RunSocialMigrations(); err != nil {
		log.Fatal("Ошибка миграций social:", err)
	}

	// Инициализируем JWT сервис
	jwtService := auth.NewJWTService(
		"your-super-secret-key-change-me-in-production-min-32-chars",
		15*time.Minute,  // Access token
		7*24*time.Hour,  // Refresh token (7 days)
	)

	// Инициализируем auth сервис и handlers
	authService := auth.NewAuthService(database.DB, jwtService)
	authHandler := auth.NewAuthHandler(authService)
	authMiddleware := auth.NewAuthMiddleware(jwtService)

	// Инициализируем admin сервис и handlers
	adminService := admin.NewAdminService(database.DB)
	adminHandler := admin.NewAdminHandler(adminService)

	// Инициализируем social сервис и handlers
	socialService := social.NewSocialService(database.DB)
	socialHandler := social.NewSocialHandler(socialService)

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
	
	// Public routes
	http.HandleFunc("/", homeHandler)
	
	// Auth routes (public)
	http.HandleFunc("/api/auth/register", authHandler.Register)
	http.HandleFunc("/api/auth/login", authHandler.Login)
	http.HandleFunc("/api/auth/forgot-password", authHandler.ForgotPassword)
	http.HandleFunc("/api/auth/reset-password", authHandler.ResetPassword)
	http.HandleFunc("/api/auth/verify-email", authHandler.VerifyEmail)
	
	// Auth routes (protected)
	http.Handle("/api/auth/logout", authMiddleware.Middleware(http.HandlerFunc(authHandler.Logout)))
	http.Handle("/api/auth/refresh", authMiddleware.Middleware(http.HandlerFunc(authHandler.Refresh)))
	http.Handle("/api/auth/me", authMiddleware.Middleware(http.HandlerFunc(authHandler.Me)))
	http.Handle("/api/auth/change-password", authMiddleware.Middleware(http.HandlerFunc(authHandler.ChangePassword)))

	// Admin routes (требуется роль admin)
	adminRouter := http.NewServeMux()
	adminRouter.HandleFunc("/dashboard", adminHandler.Dashboard)
	adminRouter.HandleFunc("/users", adminHandler.GetUsers)
	adminRouter.HandleFunc("/user", adminHandler.GetUser)
	adminRouter.HandleFunc("/user/update", adminHandler.UpdateUser)
	adminRouter.HandleFunc("/user/ban", adminHandler.BanUser)
	adminRouter.HandleFunc("/user/unban", adminHandler.UnbanUser)
	adminRouter.HandleFunc("/user/delete", adminHandler.DeleteUser)
	adminRouter.HandleFunc("/activity", adminHandler.GetActivity)
	
	http.Handle("/api/admin/", authMiddleware.Middleware(
		authMiddleware.RequireRole("admin", "moderator")(adminRouter),
	))

	// Social routes (требуется аутентификация)
	socialRouter := http.NewServeMux()
	socialRouter.HandleFunc("/friends/requests/send", socialHandler.SendFriendRequest)
	socialRouter.HandleFunc("/friends/requests/accept", socialHandler.AcceptFriendRequest)
	socialRouter.HandleFunc("/friends/requests/reject", socialHandler.RejectFriendRequest)
	socialRouter.HandleFunc("/friends/requests", socialHandler.GetFriendRequests)
	socialRouter.HandleFunc("/friends", socialHandler.GetFriends)
	socialRouter.HandleFunc("/friends/remove", socialHandler.RemoveFriend)
	socialRouter.HandleFunc("/messages/send", socialHandler.SendMessage)
	socialRouter.HandleFunc("/messages", socialHandler.GetMessages)
	socialRouter.HandleFunc("/messages/unread", socialHandler.GetUnreadCount)
	socialRouter.HandleFunc("/challenges/send", socialHandler.SendChallenge)
	socialRouter.HandleFunc("/challenges", socialHandler.GetChallenges)
	socialRouter.HandleFunc("/activity", socialHandler.GetActivityFeed)
	
	http.Handle("/api/social/", authMiddleware.Middleware(socialRouter))

	// Game routes
	http.HandleFunc("/api/quiz", quizHandler)
	http.HandleFunc("/api/answer", answerHandler)
	http.Handle("/api/stats", authMiddleware.OptionalMiddleware(http.HandlerFunc(statsHandler)))
	http.HandleFunc("/api/leaderboard", leaderboardHandler)
	http.Handle("/api/reset", authMiddleware.Middleware(http.HandlerFunc(resetHandler)))
	http.Handle("/api/skills", authMiddleware.OptionalMiddleware(http.HandlerFunc(skillsHandler)))
	http.Handle("/api/skills/upgrade", authMiddleware.Middleware(http.HandlerFunc(upgradeSkillHandler)))
	http.Handle("/api/quests", authMiddleware.OptionalMiddleware(http.HandlerFunc(questsHandler)))
	http.Handle("/api/achievements", authMiddleware.OptionalMiddleware(http.HandlerFunc(achievementsHandler)))
	http.Handle("/api/study", authMiddleware.OptionalMiddleware(http.HandlerFunc(studyGoHandler)))
	http.Handle("/api/rest", authMiddleware.OptionalMiddleware(http.HandlerFunc(restHandler)))
	http.Handle("/api/backup", authMiddleware.Middleware(http.HandlerFunc(backupHandler)))
	http.Handle("/api/game", authMiddleware.OptionalMiddleware(http.HandlerFunc(gameHandler)))

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
    <link rel="stylesheet" href="/static/auth-styles.css">
    <link rel="stylesheet" href="/static/admin-styles.css">
    <link rel="stylesheet" href="/static/tutorial-styles.css">
    <link rel="stylesheet" href="/static/sound-styles.css">
    <link rel="stylesheet" href="/static/social-styles.css">
    <!-- Vue.js 3 CDN -->
    <script src="https://unpkg.com/vue@3/dist/vue.global.js"></script>
</head>
<body>
    <div class="container" id="app">
        <header>
            <div class="logo">🧠 Go Quiz</div>
            <div class="header-actions">
                <transition name="nav-fade" appear>
                    <button class="nav-btn" :class="{ active: currentPage === 'home' }" @click="navigate('home')">🏠</button>
                </transition>
                <transition name="nav-fade" appear>
                    <button class="nav-btn" :class="{ active: currentPage === 'quiz' }" @click="navigate('quiz')">🎯</button>
                </transition>
                <transition name="nav-fade" appear>
                    <button class="nav-btn" :class="{ active: currentPage === 'study' }" @click="navigate('study')">📚</button>
                </transition>
                <transition name="nav-fade" appear>
                    <button class="nav-btn" :class="{ active: currentPage === 'skills' }" @click="navigate('skills')">🌳</button>
                </transition>
                <transition name="nav-fade" appear>
                    <button class="nav-btn" :class="{ active: currentPage === 'quests' }" @click="navigate('quests')">📋</button>
                </transition>
                <transition name="nav-fade" appear>
                    <button class="nav-btn" :class="{ active: currentPage === 'achievements' }" @click="navigate('achievements')">🏆</button>
                </transition>
                <transition name="nav-fade" appear>
                    <button class="nav-btn" :class="{ active: currentPage === 'stats' }" @click="navigate('stats')">📊</button>
                </transition>
                <transition name="nav-fade" appear>
                    <button class="nav-btn" :class="{ active: currentPage === 'leaderboard' }" @click="navigate('leaderboard')">👑</button>
                </transition>
                <transition name="nav-fade" appear>
                    <button class="nav-btn" :class="{ active: currentPage === 'game' }" @click="navigate('game')">🎮</button>
                </transition>
                <transition name="nav-fade" appear>
                    <button class="nav-btn" :class="{ active: currentPage === 'friends' }" @click="navigate('friends')">👥</button>
                </transition>
                <transition name="nav-fade" appear>
                    <button class="nav-btn" :class="{ active: currentPage === 'chat' }" @click="navigate('chat')">💬</button>
                </transition>
                <transition name="nav-fade" appear>
                    <button class="nav-btn" :class="{ active: currentPage === 'activity' }" @click="navigate('activity')">📜</button>
                </transition>
                
                <!-- Auth Button -->
                <transition name="nav-fade" appear>
                    <button v-if="!window.AuthStore || !window.AuthStore.isAuthenticated()" class="nav-btn" @click="navigate('login')">🔐</button>
                    <button v-else class="nav-btn profile-btn" @click="navigate('profile')">
                        <span v-text="window.AuthStore && window.AuthStore.getUser() ? window.AuthStore.getUser().name : 'Profile'"></span>
                    </button>
                </transition>
                
                <button class="theme-toggle" @click="toggleTheme()">
                    <span v-text="theme === 'dark' ? '☀️' : '🌙'"></span>
                </button>
                <tutorial-help-button></tutorial-help-button>
                <sound-settings></sound-settings>
            </div>
        </header>

        <main class="main-content">
            <!-- Loading -->
            <transition name="fade">
                <div v-if="isLoading" class="loading-overlay">
                    <div class="spinner"></div>
                    <p>Загрузка...</p>
                </div>
            </transition>

            <!-- Home -->
            <transition name="page">
                <div v-show="currentPage === 'home'" class="page">
                    <div class="hero">
                        <h1 class="gradient-text">Прокачай знания Go</h1>
                        <p>Викторина + RPG элементы: уровни, навыки, достижения!</p>
                        <button class="start-btn pulse" @click="startQuiz()">🚀 Начать</button>
                    </div>
                    <div class="features">
                        <transition-group name="stagger">
                            <div class="feature-card hover-lift" key="f1">
                                <div class="feature-icon bounce">📚</div>
                                <h3 v-text="quizTotal + ' вопросов'"></h3>
                                <p>Разные темы и сложности</p>
                            </div>
                            <div class="feature-card hover-lift" key="f2">
                                <div class="feature-icon bounce">🌳</div>
                                <h3>Навыки</h3>
                                <p>12 навыков в 4 категориях</p>
                            </div>
                            <div class="feature-card hover-lift" key="f3">
                                <div class="feature-icon bounce">🏆</div>
                                <h3>Достижения</h3>
                                <p>23 достижения для коллекции</p>
                            </div>
                        </transition-group>
                    </div>
                </div>
            </transition>

            <!-- Quiz -->
            <transition name="page">
                <div v-show="currentPage === 'quiz'" class="page" :class="{ 'shake': isShaking }">
                    <div class="quiz-container">
                        <div class="quiz-header">
                            <span>Вопрос {{ quizAnswered + 1 }} из {{ quizTotal }}</span>
                            <span class="level-badge">Уровень {{ player.level }}</span>
                        </div>
                        <div class="progress-bar">
                            <div class="progress-fill animated" :style="{ width: ((quizAnswered + 1) / quizTotal) * 100 + '%' }"></div>
                        </div>
                        <div class="question-text">{{ currentQuestion?.Question || 'Загрузка...' }}</div>
                        <div class="options">
                            <transition-group name="option">
                                <button
                                    v-for="(option, idx) in currentQuestion?.Options"
                                    :key="idx"
                                    class="option-btn"
                                    :class="{
                                        disabled: answered,
                                        correct: answered && idx === currentQuestion?.Correct,
                                        wrong: answered && idx === selectedOption
                                    }"
                                    @click="answerQuestion(idx, $event)"
                                >
                                    <span class="option-text">{{ option }}</span>
                                    <span class="option-indicator" v-if="answered && idx === currentQuestion?.Correct">✅</span>
                                    <span class="option-indicator" v-if="answered && idx === selectedOption && idx !== currentQuestion?.Correct">❌</span>
                                </button>
                            </transition-group>
                        </div>
                        <div class="quiz-footer">
                            <div class="exp-badge glow">{{ player.experience }} EXP</div>
                            <transition name="slide">
                                <button v-if="answered" class="next-btn pulse" @click="nextQuestion()">Далее →</button>
                            </transition>
                        </div>
                        
                        <!-- Combo Counter -->
                        <transition name="combo">
                            <div v-if="showCombo && combo >= 2" class="combo-counter">
                                <span class="combo-text">🔥 COMBO x{{ combo }}</span>
                            </div>
                        </transition>
                    </div>
                </div>
            </transition>

            <!-- Study -->
            <transition name="page">
                <div v-show="currentPage === 'study'" class="page">
                    <h2 style="margin-bottom: 30px; text-align: center;">📚 Обучение и отдых</h2>
                    <div class="action-cards">
                        <transition-group name="stagger">
                            <div class="action-card hover-lift" @click="studyGo(30)" key="s1">
                                <div class="action-icon bounce">📖</div>
                                <div class="action-title">Изучить Go (30 мин)</div>
                                <div class="action-desc">+15 EXP, +6 Знание Go, +10 Дофамин</div>
                                <div class="action-reward">🎯 Квест: 30 минут Go</div>
                            </div>
                            <div class="action-card hover-lift" @click="studyGo(60)" key="s2">
                                <div class="action-icon bounce">📖</div>
                                <div class="action-title">Изучить Go (60 мин)</div>
                                <div class="action-desc">+30 EXP, +12 Знание Go, +20 Дофамин</div>
                                <div class="action-reward">🎯 Квест: 30 минут Go</div>
                            </div>
                            <div class="action-card hover-lift" @click="rest(15)" key="s3">
                                <div class="action-icon bounce">💤</div>
                                <div class="action-title">Отдохнуть (15 мин)</div>
                                <div class="action-desc">+7 Фокус, +5 Дофамин</div>
                                <div class="action-reward">😌 Восстановление</div>
                            </div>
                            <div class="action-card hover-lift" @click="rest(30)" key="s4">
                                <div class="action-icon bounce">💤</div>
                                <div class="action-title">Отдохнуть (30 мин)</div>
                                <div class="action-desc">+15 Фокус, +10 Дофамин</div>
                                <div class="action-reward">😌 Восстановление</div>
                            </div>
                        </transition-group>
                    </div>
                    <div style="text-align: center;">
                        <button class="backup-btn pulse" @click="createBackup()">💾 Создать бэкап</button>
                    </div>
                </div>
            </transition>

            <!-- Skills -->
            <transition name="page">
                <div v-show="currentPage === 'skills'" class="page">
                    <h2 style="margin-bottom: 20px; text-align: center;">🌳 Дерево навыков</h2>
                    <div class="skill-points-display glow">
                        ✨ Очки навыков: {{ skillTree.skill_points }} (всего: {{ skillTree.total_points }})
                    </div>
                    <div class="skills-container">
                        <div v-for="(catName, catIdx) in Object.keys(skillCategories)" :key="catIdx" class="skill-category">
                            <h3 class="category-title">{{ catName }}</h3>
                            <transition-group name="skill">
                                <div v-for="skillId in skillCategories[catName]" :key="skillId" class="skill-item hover-lift">
                                    <div class="skill-header">
                                        <div class="skill-name">
                                            <span class="skill-icon">{{ skillTree.skills[skillId]?.icon }}</span>
                                            <span>{{ skillTree.skills[skillId]?.name }}</span>
                                        </div>
                                        <div class="skill-level">Ур. {{ skillTree.skills[skillId]?.level }}/{{ skillTree.skills[skillId]?.max_level }}</div>
                                    </div>
                                    <div class="skill-bar">
                                        <div class="skill-bar-fill animated" :style="{ width: getSkillProgress(skillTree.skills[skillId]) + '%' }"></div>
                                    </div>
                                    <div class="skill-description">{{ skillTree.skills[skillId]?.description }}</div>
                                    <div class="skill-bonus">
                                        +{{ (skillTree.skills[skillId]?.bonus_value || 0) * (skillTree.skills[skillId]?.level || 0) }} к {{ getBonusName(skillTree.skills[skillId]?.bonus_type) }}
                                        (всего: +{{ bonuses[skillTree.skills[skillId]?.bonus_type] || 0 }})
                                    </div>
                                    <button
                                        class="upgrade-btn pulse-small"
                                        @click="upgradeSkill(skillId, $event)"
                                        :disabled="!skillTree.skills[skillId]?.unlocked || skillTree.skills[skillId]?.level >= skillTree.skills[skillId]?.max_level"
                                    >
                                        ⬆️ Улучшить ({{ skillTree.skills[skillId]?.cost }} очк.)
                                    </button>
                                </div>
                            </transition-group>
                        </div>
                    </div>
                </div>
            </transition>

            <!-- Quests -->
            <transition name="page">
                <div v-show="currentPage === 'quests'" class="page">
                    <h2 style="margin-bottom: 20px; text-align: center;">📋 Ежедневные квесты</h2>
                    <div class="quests-container">
                        <transition-group name="quest">
                            <div v-for="quest in questSystem.quests" :key="quest.id" class="quest-item hover-lift">
                                <div class="quest-header">
                                    <div class="quest-name">
                                        <span v-if="quest.completed && quest.claimed" class="quest-icon">✅</span>
                                        <span v-else-if="quest.completed" class="quest-icon gift">🎁</span>
                                        <span v-else class="quest-icon">⏳</span>
                                        {{ quest.name }}
                                    </div>
                                    <div class="quest-status">{{ quest.progress }}/{{ quest.goal }}</div>
                                </div>
                                <div class="quest-progress-bar">
                                    <div class="quest-progress-fill animated" :style="{ width: getQuestProgress(quest) + '%' }"></div>
                                </div>
                                <div>{{ quest.description }}</div>
                                <transition name="fade">
                                    <button
                                        v-if="quest.completed && !quest.claimed"
                                        class="claim-btn pulse"
                                        @click="claimQuest(quest.id)"
                                    >
                                        🎁 Забрать ({{ quest.reward }} очк.)
                                    </button>
                                </transition>
                            </div>
                        </transition-group>
                        <div class="quest-stats">
                            <p>🔥 Серия дней: <span class="highlight">{{ questSystem.streak }}</span></p>
                            <p>📊 Всего выполнено: <span class="highlight">{{ questSystem.total_completed }}</span></p>
                        </div>
                    </div>
                </div>
            </transition>

            <!-- Achievements -->
            <transition name="page">
                <div v-show="currentPage === 'achievements'" class="page">
                    <h2 style="margin-bottom: 20px; text-align: center;">🏆 Достижения</h2>
                    <div class="achievements-container">
                        <p style="margin-bottom: 20px;" class="achievement-progress">
                            Всего разблокировано: <span class="highlight">{{ achievements.unlocked_count }}</span>/<span class="highlight">{{ achievements.total_count }}</span>
                        </p>
                        <transition-group name="achievement">
                            <div
                                v-for="ach in Object.values(achievements.system)"
                                :key="ach.id"
                                class="achievement-item"
                                :class="{ unlocked: ach.unlocked, 'hover-lift': ach.unlocked }"
                            >
                                <div class="achievement-icon">{{ ach.unlocked ? ach.icon : '🔒' }}</div>
                                <div class="achievement-info">
                                    <div class="achievement-name">{{ ach.name }}</div>
                                    <div class="achievement-description">{{ ach.description }}</div>
                                </div>
                            </div>
                        </transition-group>
                    </div>
                </div>
            </transition>

            <!-- Stats -->
            <transition name="page">
                <div v-show="currentPage === 'stats'" class="page">
                    <h2 style="margin-bottom: 30px; text-align: center;">📊 Статистика</h2>
                    <div class="stats-grid">
                        <transition-group name="stat">
                            <div class="stat-card hover-lift" key="s1">
                                <div class="stat-value">{{ player.level }}</div>
                                <div class="stat-label">Уровень</div>
                            </div>
                            <div class="stat-card hover-lift" key="s2">
                                <div class="stat-value">{{ player.experience }}</div>
                                <div class="stat-label">Всего EXP</div>
                            </div>
                            <div class="stat-card hover-lift" key="s3">
                                <div class="stat-value">{{ player.correct_answers }}</div>
                                <div class="stat-label">Правильных</div>
                            </div>
                            <div class="stat-card hover-lift" key="s4">
                                <div class="stat-value">{{ player.wrong_answers }}</div>
                                <div class="stat-label">Неправильных</div>
                            </div>
                            <div class="stat-card hover-lift" key="s5">
                                <div class="stat-value">{{ player.go_knowledge }}/100</div>
                                <div class="stat-label">Знание Go</div>
                            </div>
                            <div class="stat-card hover-lift" key="s6">
                                <div class="stat-value">{{ player.focus }}%</div>
                                <div class="stat-label">Фокус</div>
                            </div>
                            <div class="stat-card hover-lift" key="s7">
                                <div class="stat-value">{{ player.willpower }}%</div>
                                <div class="stat-label">Сила воли</div>
                            </div>
                            <div class="stat-card hover-lift" key="s8">
                                <div class="stat-value glow-text">{{ player.rating }}</div>
                                <div class="stat-label">Рейтинг</div>
                            </div>
                        </transition-group>
                    </div>
                    <div class="reset-section">
                        <button class="reset-btn" @click="resetProgress()">🔄 Сбросить прогресс</button>
                    </div>
                </div>
            </transition>

            <!-- Leaderboard -->
            <transition name="page">
                <div v-show="currentPage === 'leaderboard'" class="page">
                    <h2 style="margin-bottom: 30px; text-align: center;">👑 Таблица лидеров</h2>
                    <transition-group name="list" tag="table" class="leaderboard-table">
                        <thead key="head">
                            <tr>
                                <th>#</th>
                                <th>Игрок</th>
                                <th>Уровень</th>
                                <th>Рейтинг</th>
                                <th>Правильных</th>
                            </tr>
                        </thead>
                        <tbody key="body">
                            <tr v-for="(entry, idx) in leaderboard" :key="entry.id" :class="'rank-' + (idx + 1)">
                                <td><span class="rank-badge" :class="'rank-' + (idx + 1)">{{ idx + 1 }}</span></td>
                                <td>{{ entry.name }}</td>
                                <td>{{ entry.level }}</td>
                                <td class="rating">{{ entry.rating }}</td>
                                <td>{{ entry.correct }}</td>
                            </tr>
                        </tbody>
                    </transition-group>
                </div>
            </transition>

            <!-- Game -->
            <transition name="page">
                <div v-show="currentPage === 'game'" class="page">
                    <godot-game></godot-game>
                </div>
            </transition>
            
            <!-- Login -->
            <transition name="page">
                <div v-show="currentPage === 'login'" class="page">
                    <login-component></login-component>
                </div>
            </transition>
            
            <!-- Register -->
            <transition name="page">
                <div v-show="currentPage === 'register'" class="page">
                    <register-component></register-component>
                </div>
            </transition>
            
            <!-- Profile -->
            <transition name="page">
                <div v-show="currentPage === 'profile'" class="page">
                    <profile-component></profile-component>
                </div>
            </transition>
            
            <!-- Admin -->
            <transition name="page">
                <div v-show="currentPage === 'admin'" class="page admin-page">
                    <admin-layout></admin-layout>
                </div>
            </transition>
            
            <!-- Friends -->
            <transition name="page">
                <div v-show="currentPage === 'friends'" class="page">
                    <friends-component></friends-component>
                </div>
            </transition>
            
            <!-- Chat -->
            <transition name="page">
                <div v-show="currentPage === 'chat'" class="page">
                    <chat-component></chat-component>
                </div>
            </transition>
            
            <!-- Activity -->
            <transition name="page">
                <div v-show="currentPage === 'activity'" class="page">
                    <activity-feed-component></activity-feed-component>
                </div>
            </transition>
            
            <!-- Challenges -->
            <transition name="page">
                <div v-show="currentPage === 'challenges'" class="page">
                    <challenges-component></challenges-component>
                </div>
            </transition>
        </main>

        <!-- Toast Notifications -->
        <transition-group name="toast" tag="div" class="toast-container">
            <div v-for="t in toasts" :key="t.id" class="toast" :class="'toast-' + t.type">
                <span class="toast-message">{{ t.message }}</span>
            </div>
        </transition-group>

        <!-- Tutorial Overlay -->
        <tutorial-overlay></tutorial-overlay>

        <!-- Confetti Canvas -->
        <canvas v-if="isActive" class="confetti-canvas"></canvas>

        <!-- Level Up Overlay -->
        <transition name="levelup">
            <div v-if="isAnimating" class="levelup-overlay">
                <div class="levelup-content">
                    <div class="levelup-stars">
                        <span v-for="star in stars" :key="star.id" class="star" 
                              :style="{ left: star.x + '%', top: star.y + '%', transform: 'scale(' + star.scale + ') rotate(' + star.rotation + 'deg)' }">⭐</span>
                    </div>
                    <h1 class="levelup-title">🎉 LEVEL UP!</h1>
                    <p class="levelup-level">Уровень {{ level }}</p>
                </div>
            </div>
        </transition>

        <!-- Floating Text -->
        <transition-group name="float">
            <div v-for="text in texts" :key="text.id" class="floating-text"
                 :style="{ left: text.x + 'px', top: text.y + 'px', color: text.color, opacity: text.opacity, transform: 'translateY(' + text.yOffset + 'px)' }">
                {{ text.text }}
            </div>
        </transition-group>

        <!-- Particles -->
        <div v-for="p in particles" :key="p.id" class="particle"
             :style="{ left: p.x + 'px', top: p.y + 'px', width: p.size + 'px', height: p.size + 'px', backgroundColor: p.color, opacity: p.life }">
        </div>
    </div>

    <script src="/static/vue-effects.js"></script>
    <script>
        window.VueEffects = {};
    </script>
    <script src="/static/auth-store.js"></script>
    <script src="/static/sound-store.js"></script>
    <script src="/static/sound-settings.js"></script>
    <script src="/static/social-store.js"></script>
    <script src="/static/friends-component.js"></script>
    <script src="/static/chat-component.js"></script>
    <script src="/static/activity-component.js"></script>
    <script src="/static/tutorial-store.js"></script>
    <script src="/static/tutorial-overlay.js"></script>
    <script src="/static/tutorial-button.js"></script>
    <script src="/static/admin-store.js"></script>
    <script src="/static/admin-layout.js"></script>
    <script src="/static/admin-components.js"></script>
    <script src="/static/admin-dashboard.js"></script>
    <script src="/static/admin-users.js"></script>
    <script src="/static/login-component.js"></script>
    <script src="/static/register-component.js"></script>
    <script src="/static/profile-component.js"></script>
    <script src="/static/godot-bridge.js"></script>
    <script src="/static/vue-game.js"></script>
    <script src="/static/vue-app.js"></script>
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

// gameHandler возвращает данные для Godot игры
func gameHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	player := getPlayer(userID)
	tree := getSkillTree(userID)
	questSystem := getQuestSystem(userID)

	// Проверяем достижения
	achievements := getAchievements(userID)
	achievements.CheckAchievements(player, tree, questSystem)

	// Формируем ответ для Godot
	response := map[string]interface{}{
		"player": map[string]interface{}{
			"id":              player.ID,
			"level":           player.Level,
			"experience":      player.Experience,
			"go_knowledge":    player.GoKnowledge,
			"focus":           player.Focus,
			"willpower":       player.Willpower,
			"combo":           0,
			"max_combo":       0,
			"correct_answers": player.CorrectAnswers,
			"wrong_answers":   player.WrongAnswers,
			"rating":          player.GetRating(),
		},
		"skill_tree":  tree,
		"quests":      questSystem,
		"achievements": achievements,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
