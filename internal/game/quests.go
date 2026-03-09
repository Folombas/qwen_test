package game

import (
	"fmt"
	"time"
)

// DailyQuest представляет ежедневный квест
type DailyQuest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Goal        int    `json:"goal"`
	Progress    int    `json:"progress"`
	Reward      int    `json:"reward"`
	Completed   bool   `json:"completed"`
	Claimed     bool   `json:"claimed"`
}

// QuestSystem управляет квестами игрока
type QuestSystem struct {
	UserID        string        `json:"user_id"`
	Quests        []*DailyQuest `json:"quests"`
	Day           int           `json:"day"`
	Streak        int           `json:"streak"`
	TotalCompleted int          `json:"total_completed"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// DailyQuestsTemplates шаблоны ежедневных квестов
var DailyQuestsTemplates = []struct {
	ID          string
	Name        string
	Description string
	Goal        int
	Reward      int
}{
	{"study_30", "30 минут Go", "Изучайте Go в течение 30 минут", 30, 2},
	{"temptation_fighter", "Борец с искушениями", "Победите 3 искушения", 3, 3},
	{"code_practice", "Практика кода", "Напишите 50 строк кода", 50, 3},
	{"morning_ritual", "Утренний ритуал", "Начните день с обучения", 1, 1},
	{"digital_detox", "Цифровой детокс", "4 часа без соцсетей", 240, 2},
}

// NewQuestSystem создаёт новую систему квестов
func NewQuestSystem(userID string) *QuestSystem {
	qs := &QuestSystem{
		UserID:    userID,
		Quests:    make([]*DailyQuest, 0, 5),
		Day:       1,
		Streak:    0,
		UpdatedAt: time.Now(),
	}
	qs.GenerateDailyQuests()
	return qs
}

// GenerateDailyQuests генерирует 5 ежедневных квестов
func (qs *QuestSystem) GenerateDailyQuests() {
	qs.Quests = make([]*DailyQuest, 0, 5)
	for _, tmpl := range DailyQuestsTemplates {
		qs.Quests = append(qs.Quests, &DailyQuest{
			ID:          tmpl.ID,
			Name:        tmpl.Name,
			Description: tmpl.Description,
			Goal:        tmpl.Goal,
			Progress:    0,
			Reward:      tmpl.Reward,
			Completed:   false,
			Claimed:     false,
		})
	}
}

// UpdateProgress обновляет прогресс квеста
func (qs *QuestSystem) UpdateProgress(questID string, progress int) {
	for _, quest := range qs.Quests {
		if quest.ID == questID && !quest.Completed {
			if progress < 0 {
				progress = 0
			}
			quest.Progress += progress
			if quest.Progress >= quest.Goal {
				quest.Progress = quest.Goal
				quest.Completed = true
			}
			break
		}
	}
}

// GetCompletedCount возвращает количество выполненных квестов
func (qs *QuestSystem) GetCompletedCount() int {
	count := 0
	for _, quest := range qs.Quests {
		if quest.Completed && !quest.Claimed {
			count++
		}
	}
	return count
}

// ClaimRewards забирает награды за выполненные квесты
func (qs *QuestSystem) ClaimRewards() int {
	totalReward := 0
	for _, quest := range qs.Quests {
		if quest.Completed && !quest.Claimed {
			totalReward += quest.Reward
			quest.Claimed = true
		}
	}
	if totalReward > 0 {
		qs.TotalCompleted += qs.GetCompletedCount()
	}
	return totalReward
}

// CheckDayStreak проверяет серию дней
func (qs *QuestSystem) CheckDayStreak(allCompleted bool) {
	if allCompleted {
		qs.Streak++
	} else {
		qs.Streak = 0
	}
}

// Display возвращает строковое представление квестов
func (qs *QuestSystem) Display() string {
	result := "📋 ЕЖЕДНЕВНЫЕ КВЕСТЫ\n"
	result += "━━━━━━━━━━━━━━━━━━━━\n\n"

	for _, quest := range qs.Quests {
		status := "⏳"
		if quest.Completed {
			if quest.Claimed {
				status = "✅"
			} else {
				status = "🎁"
			}
		}

		bar := questProgressBar(quest.Progress, quest.Goal)
		result += fmt.Sprintf("%s %s\n", status, quest.Name)
		result += fmt.Sprintf("   %s\n", quest.Description)
		result += fmt.Sprintf("   Прогресс: %s %d/%d\n", bar, quest.Progress, quest.Goal)
		if quest.Completed && !quest.Claimed {
			result += fmt.Sprintf("   🎁 Награда: %d очков навыков\n", quest.Reward)
		}
		result += "\n"
	}

	result += fmt.Sprintf("🔥 Серия дней: %d\n", qs.Streak)
	result += fmt.Sprintf("📊 Всего выполнено: %d\n", qs.TotalCompleted)

	return result
}

func questProgressBar(progress, goal int) string {
	if goal == 0 {
		return "[░░░░░]"
	}
	percent := progress * 100 / goal
	filled := percent / 20
	empty := 5 - filled
	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := 0; i < empty; i++ {
		bar += "░"
	}
	return "[" + bar + "]"
}
