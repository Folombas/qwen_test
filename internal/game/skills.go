package game

import (
	"fmt"
	"time"
)

// Skill представляет навык
type Skill struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Level       int    `json:"level"`
	MaxLevel    int    `json:"max_level"`
	Cost        int    `json:"cost"`
	BonusType   string `json:"bonus_type"`
	BonusValue  int    `json:"bonus_value"`
	Unlocked    bool   `json:"unlocked"`
	Prereqs     []string `json:"prereqs"`
}

// SkillTree представляет дерево навыков игрока
type SkillTree struct {
	UserID      string             `json:"user_id"`
	Skills      map[string]*Skill  `json:"skills"`
	SkillPoints int                `json:"skill_points"`
	TotalPoints int                `json:"total_points"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// NewSkillTree создаёт новое дерево навыков
func NewSkillTree(userID string) *SkillTree {
	tree := &SkillTree{
		UserID:      userID,
		Skills:      make(map[string]*Skill),
		SkillPoints: 0,
		TotalPoints: 0,
		UpdatedAt:   time.Now(),
	}

	// Инициализируем все навыки
	tree.initSkills()
	return tree
}

// initSkills инициализирует навыки
func (st *SkillTree) initSkills() {
	// 📚 GO-НАВЫКИ (6)
	st.Skills["go_basics"] = &Skill{
		ID: "go_basics", Name: "Основы Go", Description: "Синтаксис, типы данных, функции",
		Icon: "📘", Level: 1, MaxLevel: 5, Cost: 1, BonusType: "knowledge", BonusValue: 5,
		Unlocked: true, Prereqs: []string{},
	}
	st.Skills["concurrency"] = &Skill{
		ID: "concurrency", Name: "Конкурентность", Description: "Горутины, каналы, sync package",
		Icon: "⚡", Level: 0, MaxLevel: 5, Cost: 2, BonusType: "knowledge", BonusValue: 8,
		Unlocked: false, Prereqs: []string{"go_basics"},
	}
	st.Skills["interfaces"] = &Skill{
		ID: "interfaces", Name: "Интерфейсы", Description: "Интерфейсы, полиморфизм",
		Icon: "🔌", Level: 0, MaxLevel: 5, Cost: 2, BonusType: "knowledge", BonusValue: 7,
		Unlocked: false, Prereqs: []string{"go_basics"},
	}
	st.Skills["web_frameworks"] = &Skill{
		ID: "web_frameworks", Name: "Web Фреймворки", Description: "Gin, Echo, Fiber",
		Icon: "🌐", Level: 0, MaxLevel: 5, Cost: 3, BonusType: "knowledge", BonusValue: 10,
		Unlocked: false, Prereqs: []string{"concurrency", "interfaces"},
	}
	st.Skills["databases"] = &Skill{
		ID: "databases", Name: "Базы данных", Description: "SQL, NoSQL, ORM",
		Icon: "🗄️", Level: 0, MaxLevel: 5, Cost: 3, BonusType: "knowledge", BonusValue: 10,
		Unlocked: false, Prereqs: []string{"concurrency"},
	}
	st.Skills["microservices"] = &Skill{
		ID: "microservices", Name: "Микросервисы", Description: "Архитектура, gRPC, API Gateway",
		Icon: "🔧", Level: 0, MaxLevel: 5, Cost: 4, BonusType: "knowledge", BonusValue: 12,
		Unlocked: false, Prereqs: []string{"web_frameworks", "databases"},
	}

	// 🎯 ФОКУС (3)
	st.Skills["focus_master"] = &Skill{
		ID: "focus_master", Name: "Мастер Фокуса", Description: "Умение концентрироваться на задачах",
		Icon: "🎯", Level: 0, MaxLevel: 5, Cost: 1, BonusType: "focus", BonusValue: 5,
		Unlocked: true, Prereqs: []string{},
	}
	st.Skills["meditation"] = &Skill{
		ID: "meditation", Name: "Медитация", Description: "Восстановление ментальных ресурсов",
		Icon: "🧘", Level: 0, MaxLevel: 5, Cost: 2, BonusType: "dopamine", BonusValue: 10,
		Unlocked: false, Prereqs: []string{"focus_master"},
	}
	st.Skills["anti_procrastination"] = &Skill{
		ID: "anti_procrastination", Name: "Борьба с прокрастинацией", Description: "Техники продуктивности",
		Icon: "⏰", Level: 0, MaxLevel: 5, Cost: 2, BonusType: "focus", BonusValue: 8,
		Unlocked: false, Prereqs: []string{"focus_master"},
	}

	// 💪 СИЛА ВОЛИ (2)
	st.Skills["willpower"] = &Skill{
		ID: "willpower", Name: "Сила Воли", Description: "Самоконтроль и дисциплина",
		Icon: "💪", Level: 0, MaxLevel: 5, Cost: 2, BonusType: "willpower", BonusValue: 8,
		Unlocked: true, Prereqs: []string{},
	}
	st.Skills["discipline"] = &Skill{
		ID: "discipline", Name: "Дисциплина", Description: "Регулярные занятия и привычки",
		Icon: "📅", Level: 0, MaxLevel: 5, Cost: 3, BonusType: "willpower", BonusValue: 10,
		Unlocked: false, Prereqs: []string{"willpower"},
	}

	// 💰 ФИНАНСЫ (1)
	st.Skills["money_management"] = &Skill{
		ID: "money_management", Name: "Управление деньгами", Description: "Финансовая грамотность",
		Icon: "💰", Level: 0, MaxLevel: 5, Cost: 2, BonusType: "money", BonusValue: 50,
		Unlocked: false, Prereqs: []string{"willpower"},
	}
}

// EarnSkillPoints начисляет очки навыков
func (st *SkillTree) EarnSkillPoints(points int) {
	if points < 0 {
		points = 0
	}
	st.SkillPoints += points
	st.TotalPoints += points
}

// UpgradeSkill улучшает навык
func (st *SkillTree) UpgradeSkill(skillID string) (bool, string) {
	skill, exists := st.Skills[skillID]
	if !exists {
		return false, "❌ Навык не найден"
	}

	if !skill.Unlocked {
		return false, "🔒 Навык заблокирован"
	}

	if skill.Level >= skill.MaxLevel {
		return false, "⛔ Максимальный уровень"
	}

	if st.SkillPoints < skill.Cost {
		return false, fmt.Sprintf("❌ Недостаточно очков (нужно %d)", skill.Cost)
	}

	// Списываем очки и повышаем уровень
	st.SkillPoints -= skill.Cost
	skill.Level++
	st.UpdatedAt = time.Now()

	// Проверяем разблокировку следующих навыков
	st.checkUnlocks()

	return true, fmt.Sprintf("✅ Навык \"%s\" улучшен до уровня %d!\n+%d к %s", 
		skill.Name, skill.Level, skill.BonusValue, st.getBonusName(skill.BonusType))
}

// checkUnlocks проверяет разблокировку навыков
func (st *SkillTree) checkUnlocks() {
	for _, skill := range st.Skills {
		if skill.Unlocked {
			continue
		}

		// Проверяем prerequisites
		allPrereqsMet := true
		for _, prereqID := range skill.Prereqs {
			prereq, exists := st.Skills[prereqID]
			if !exists || prereq.Level == 0 {
				allPrereqsMet = false
				break
			}
		}

		if allPrereqsMet {
			skill.Unlocked = true
		}
	}
}

// GetTotalBonuses возвращает суммарные бонусы от всех навыков
func (st *SkillTree) GetTotalBonuses() map[string]int {
	bonuses := make(map[string]int)
	bonusTypes := []string{"focus", "willpower", "knowledge", "money", "dopamine"}

	for _, bt := range bonusTypes {
		bonuses[bt] = 0
	}

	for _, skill := range st.Skills {
		if skill.Level > 0 {
			bonuses[skill.BonusType] += skill.Level * skill.BonusValue
		}
	}

	return bonuses
}

// getBonusName возвращает название типа бонуса
func (st *SkillTree) getBonusName(bonusType string) string {
	switch bonusType {
	case "focus":
		return "Фокус"
	case "willpower":
		return "Сила воли"
	case "knowledge":
		return "Знание Go"
	case "money":
		return "Деньги"
	case "dopamine":
		return "Дофамин"
	default:
		return "Бонус"
	}
}

// Display возвращает строковое представление дерева навыков
func (st *SkillTree) Display() string {
	result := fmt.Sprintf("🌳 ДЕРЕВО НАВЫКОВ\n")
	result += "━━━━━━━━━━━━━━━━━━━━\n\n"
	result += fmt.Sprintf("✨ Очки навыков: %d (всего: %d)\n\n", st.SkillPoints, st.TotalPoints)

	categories := map[string][]string{
		"📚 GO-НАВЫКИ":    {"go_basics", "concurrency", "interfaces", "web_frameworks", "databases", "microservices"},
		"🎯 ФОКУС":        {"focus_master", "meditation", "anti_procrastination"},
		"💪 СИЛА ВОЛИ":    {"willpower", "discipline"},
		"💰 ФИНАНСЫ":      {"money_management"},
	}

	for catName, skillIDs := range categories {
		result += catName + "\n"
		result += "────────────────────\n"
		for _, id := range skillIDs {
			skill := st.Skills[id]
			status := "🔒"
			if skill.Unlocked {
				status = "✅"
			}
			bar := st.progressBar(skill.Level, skill.MaxLevel)
			result += fmt.Sprintf("  %s %s %-20s [%s] Ур.%d/%d [%d очк.]\n",
				status, skill.Icon, skill.Name, bar, skill.Level, skill.MaxLevel, skill.Cost)
			result += fmt.Sprintf("     %s\n", skill.Description)
		}
		result += "\n"
	}

	return result
}

// progressBar создаёт строку прогресса
func (st *SkillTree) progressBar(level, maxLevel int) string {
	filled := level
	empty := maxLevel - level
	return stringRepeat("█", filled) + stringRepeat("░", empty)
}

func stringRepeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// GetSkillPointsForLevel возвращает очки навыков за уровень
func GetSkillPointsForLevel(level int) int {
	return 2 + (level / 5)
}
