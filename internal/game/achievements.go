package game

import (
	"fmt"
	"time"
)

// Achievement представляет достижение
type Achievement struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Category    string    `json:"category"`
	Unlocked    bool      `json:"unlocked"`
	UnlockedAt  time.Time `json:"unlocked_at"`
}

// AchievementSystem управляет достижениями игрока
type AchievementSystem struct {
	UserID       string                   `json:"user_id"`
	Achievements map[string]*Achievement  `json:"achievements"`
	UpdatedAt    time.Time                `json:"updated_at"`
}

// AchievementTemplates шаблоны всех достижений
var AchievementTemplates = []struct {
	ID          string
	Name        string
	Description string
	Icon        string
	Category    string
}{
	// 📈 За уровни (5)
	{"level_2", "Первые шаги", "Достичь уровня 2", "👶", "levels"},
	{"level_5", "Go-Новичок", "Достичь уровня 5", "🌱", "levels"},
	{"level_10", "Go-Разработчик", "Достичь уровня 10", "💻", "levels"},
	{"level_20", "Go-Мастер", "Достичь уровня 20", "🎓", "levels"},
	{"level_30", "Go-Легенда", "Достичь уровня 30", "🏆", "levels"},

	// 📋 За квесты (3)
	{"quest_10", "Начинающий квестер", "Выполнить 10 квестов", "📋", "quests"},
	{"quest_50", "Охотник за квестами", "Выполнить 50 квестов", "🎯", "quests"},
	{"quest_100", "Мастер квестов", "Выполнить 100 квестов", "👑", "quests"},

	// 🔥 За серию дней (2)
	{"streak_7", "Недельная серия", "7 дней подряд", "🔥", "streaks"},
	{"streak_30", "Месячная серия", "30 дней подряд", "🔥🔥", "streaks"},

	// 🛡️ За искушения (3)
	{"temptation_10", "Борец с искушениями", "10 побед над искушениями", "🛡️", "temptations"},
	{"temptation_50", "Воин искушений", "50 побед над искушениями", "⚔️", "temptations"},
	{"temptation_100", "Легенда сопротивления", "100 побед над искушениями", "🏅", "temptations"},

	// ⚔️ За боссов (2)
	{"boss_10", "Убийца боссов", "10 побед над боссами", "🗡️", "bosses"},
	{"boss_50", "Легенда боссов", "50 побед над боссами", "🐉", "bosses"},

	// 📚 За изучение Go (2)
	{"go_50", "Студент Go", "50% знаний Go", "📖", "knowledge"},
	{"go_100", "Учёный Go", "100% знаний Go", "🎓", "knowledge"},

	// ⭐ За навыки (2)
	{"skill_max", "Мастер навыков", "Макс. уровень навыка", "⭐", "skills"},
	{"skill_all", "Коллекционер навыков", "Все навыки разблокированы", "🌟", "skills"},

	// 💎 Специальные (4)
	{"perfectionist", "Перфекционист", "Все квесты за день", "💎", "special"},
	{"marathon", "Марафонец", "1000 минут в игре", "🏃", "special"},
	{"early_bird", "Ранняя пташка", "Начать день до 9 утра", "🌅", "special"},
	{"night_owl", "Ночная сова", "Завершить день после 23", "🦉", "special"},
}

// NewAchievementSystem создаёт новую систему достижений
func NewAchievementSystem(userID string) *AchievementSystem {
	as := &AchievementSystem{
		UserID:       userID,
		Achievements: make(map[string]*Achievement),
		UpdatedAt:    time.Now(),
	}

	// Инициализируем все достижения
	for _, tmpl := range AchievementTemplates {
		as.Achievements[tmpl.ID] = &Achievement{
			ID:          tmpl.ID,
			Name:        tmpl.Name,
			Description: tmpl.Description,
			Icon:        tmpl.Icon,
			Category:    tmpl.Category,
			Unlocked:    false,
		}
	}

	return as
}

// Unlock разблокирует достижение
func (as *AchievementSystem) Unlock(achievementID string) bool {
	achievement, exists := as.Achievements[achievementID]
	if !exists || achievement.Unlocked {
		return false
	}

	achievement.Unlocked = true
	achievement.UnlockedAt = time.Now()
	as.UpdatedAt = time.Now()
	return true
}

// CheckAchievements проверяет и разблокирует достижения
func (as *AchievementSystem) CheckAchievements(player *Player, tree *SkillTree, questSystem *QuestSystem) []string {
	unlocked := make([]string, 0)

	// За уровни
	levelAchievements := map[int]string{
		2: "level_2", 5: "level_5", 10: "level_10", 20: "level_20", 30: "level_30",
	}
	if achievementID, exists := levelAchievements[player.Level]; exists {
		if as.Unlock(achievementID) {
			unlocked = append(unlocked, achievementID)
		}
	}

	// За квесты
	if questSystem.TotalCompleted >= 10 {
		if as.Unlock("quest_10") {
			unlocked = append(unlocked, "quest_10")
		}
	}
	if questSystem.TotalCompleted >= 50 {
		if as.Unlock("quest_50") {
			unlocked = append(unlocked, "quest_50")
		}
	}
	if questSystem.TotalCompleted >= 100 {
		if as.Unlock("quest_100") {
			unlocked = append(unlocked, "quest_100")
		}
	}

	// За серию дней
	if questSystem.Streak >= 7 {
		if as.Unlock("streak_7") {
			unlocked = append(unlocked, "streak_7")
		}
	}
	if questSystem.Streak >= 30 {
		if as.Unlock("streak_30") {
			unlocked = append(unlocked, "streak_30")
		}
	}

	// За изучение Go
	if player.GoKnowledge >= 50 {
		if as.Unlock("go_50") {
			unlocked = append(unlocked, "go_50")
		}
	}
	if player.GoKnowledge >= 100 {
		if as.Unlock("go_100") {
			unlocked = append(unlocked, "go_100")
		}
	}

	// За навыки
	for _, skill := range tree.Skills {
		if skill.Level >= skill.MaxLevel {
			if as.Unlock("skill_max") {
				unlocked = append(unlocked, "skill_max")
			}
			break
		}
	}

	// Все навыки разблокированы
	allUnlocked := true
	for _, skill := range tree.Skills {
		if !skill.Unlocked {
			allUnlocked = false
			break
		}
	}
	if allUnlocked {
		if as.Unlock("skill_all") {
			unlocked = append(unlocked, "skill_all")
		}
	}

	// Специальные
	if player.PlayTime >= 1000 {
		if as.Unlock("marathon") {
			unlocked = append(unlocked, "marathon")
		}
	}

	return unlocked
}

// GetUnlockedCount возвращает количество разблокированных достижений
func (as *AchievementSystem) GetUnlockedCount() int {
	count := 0
	for _, achievement := range as.Achievements {
		if achievement.Unlocked {
			count++
		}
	}
	return count
}

// GetTotalCount возвращает общее количество достижений
func (as *AchievementSystem) GetTotalCount() int {
	return len(as.Achievements)
}

// Display возвращает строковое представление достижений
func (as *AchievementSystem) Display() string {
	unlockedCount := as.GetUnlockedCount()
	totalCount := as.GetTotalCount()

	result := "🏆 ДОСТИЖЕНИЯ\n"
	result += "━━━━━━━━━━━━━━━━━━━━\n\n"
	result += fmt.Sprintf("Всего разблокировано: %d/%d\n\n", unlockedCount, totalCount)

	// Разблокированные
	result += "🔓 Разблокированные:\n"
	hasUnlocked := false
	for _, achievement := range as.Achievements {
		if achievement.Unlocked {
			result += fmt.Sprintf("%s %s — %s\n", achievement.Icon, achievement.Name, achievement.Description)
			hasUnlocked = true
		}
	}
	if !hasUnlocked {
		result += "  Пока нет разблокированных достижений\n"
	}

	// Заблокированные
	result += "\n🔒 Заблокированные:\n"
	for _, achievement := range as.Achievements {
		if !achievement.Unlocked {
			result += fmt.Sprintf("🔒 %s — %s\n", achievement.Name, achievement.Description)
		}
	}

	return result
}
