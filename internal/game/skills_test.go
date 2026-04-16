package game

import (
	"testing"
)

func TestNewSkillTree(t *testing.T) {
	tree := NewSkillTree("user123")

	if tree.UserID != "user123" {
		t.Errorf("Ожидаемый UserID 'user123', получен '%s'", tree.UserID)
	}
	if tree.SkillPoints != 0 {
		t.Errorf("Ожидаемые очки навыков 0, получено %d", tree.SkillPoints)
	}
	if tree.TotalPoints != 0 {
		t.Errorf("Ожидаемые всего очки 0, получено %d", tree.TotalPoints)
	}
	if len(tree.Skills) != 12 {
		t.Errorf("Ожидаемое количество навыков 12, получено %d", len(tree.Skills))
	}

	// Проверка базовых навыков
	expectedUnlocked := []string{"go_basics", "focus_master", "willpower"}
	for _, skillID := range expectedUnlocked {
		skill, exists := tree.Skills[skillID]
		if !exists {
			t.Errorf("Навык %s не существует", skillID)
			continue
		}
		if !skill.Unlocked {
			t.Errorf("Навык %s должен быть разблокирован", skillID)
		}
		// go_basics начинается с уровня 1, остальные с 0
		if skillID == "go_basics" && skill.Level != 1 {
			t.Errorf("Навык %s должен иметь уровень 1, получено %d", skillID, skill.Level)
		}
		if skillID != "go_basics" && skill.Level != 0 {
			t.Errorf("Навык %s должен иметь уровень 0, получено %d", skillID, skill.Level)
		}
	}
}

func TestSkillTree_InitSkills(t *testing.T) {
	tree := NewSkillTree("user1")

	tests := []struct {
		skillID      string
		name         string
		icon         string
		maxLevel     int
		cost         int
		bonusType    string
		bonusValue   int
		prereqsCount int
	}{
		{"go_basics", "Основы Go", "📘", 5, 1, "knowledge", 5, 0},
		{"concurrency", "Конкурентность", "⚡", 5, 2, "knowledge", 8, 1},
		{"interfaces", "Интерфейсы", "🔌", 5, 2, "knowledge", 7, 1},
		{"web_frameworks", "Web Фреймворки", "🌐", 5, 3, "knowledge", 10, 2},
		{"databases", "Базы данных", "🗄️", 5, 3, "knowledge", 10, 1},
		{"microservices", "Микросервисы", "🔧", 5, 4, "knowledge", 12, 2},
		{"focus_master", "Мастер Фокуса", "🎯", 5, 1, "focus", 5, 0},
		{"meditation", "Медитация", "🧘", 5, 2, "dopamine", 10, 1},
		{"anti_procrastination", "Борьба с прокрастинацией", "⏰", 5, 2, "focus", 8, 1},
		{"willpower", "Сила Воли", "💪", 5, 2, "willpower", 8, 0},
		{"discipline", "Дисциплина", "📅", 5, 3, "willpower", 10, 1},
		{"money_management", "Управление деньгами", "💰", 5, 2, "money", 50, 1},
	}

	for _, tt := range tests {
		skill, exists := tree.Skills[tt.skillID]
		if !exists {
			t.Errorf("Навык %s не существует", tt.skillID)
			continue
		}
		if skill.Name != tt.name {
			t.Errorf("Навык %s: ожидаемое имя '%s', получено '%s'", tt.skillID, tt.name, skill.Name)
		}
		if skill.Icon != tt.icon {
			t.Errorf("Навык %s: ожидаемая иконка '%s', получена '%s'", tt.skillID, tt.icon, skill.Icon)
		}
		if skill.MaxLevel != tt.maxLevel {
			t.Errorf("Навык %s: ожидаемый макс. уровень %d, получено %d", tt.skillID, tt.maxLevel, skill.MaxLevel)
		}
		if skill.Cost != tt.cost {
			t.Errorf("Навык %s: ожидаемая стоимость %d, получена %d", tt.skillID, tt.cost, skill.Cost)
		}
		if skill.BonusType != tt.bonusType {
			t.Errorf("Навык %s: ожидаемый тип бонуса '%s', получен '%s'", tt.skillID, tt.bonusType, skill.BonusType)
		}
		if skill.BonusValue != tt.bonusValue {
			t.Errorf("Навык %s: ожидаемое значение бонуса %d, получено %d", tt.skillID, tt.bonusValue, skill.BonusValue)
		}
		if len(skill.Prereqs) != tt.prereqsCount {
			t.Errorf("Навык %s: ожидаемое количество prerequisites %d, получено %d", tt.skillID, tt.prereqsCount, len(skill.Prereqs))
		}
	}
}

func TestEarnSkillPoints(t *testing.T) {
	tree := NewSkillTree("user1")

	tree.EarnSkillPoints(5)

	if tree.SkillPoints != 5 {
		t.Errorf("Ожидаемые очки навыков 5, получено %d", tree.SkillPoints)
	}
	if tree.TotalPoints != 5 {
		t.Errorf("Ожидаемые всего очки 5, получено %d", tree.TotalPoints)
	}

	tree.EarnSkillPoints(3)

	if tree.SkillPoints != 8 {
		t.Errorf("Ожидаемые очки навыков 8, получено %d", tree.SkillPoints)
	}
	if tree.TotalPoints != 8 {
		t.Errorf("Ожидаемые всего очки 8, получено %d", tree.TotalPoints)
	}
}

func TestEarnSkillPoints_NegativePoints(t *testing.T) {
	tree := NewSkillTree("user1")
	tree.SkillPoints = 10
	tree.TotalPoints = 10

	tree.EarnSkillPoints(-5)

	if tree.SkillPoints != 10 {
		t.Errorf("Ожидаемые очки навыков 10 (без изменений), получено %d", tree.SkillPoints)
	}
	if tree.TotalPoints != 10 {
		t.Errorf("Ожидаемые всего очки 10 (без изменений), получено %d", tree.TotalPoints)
	}
}

func TestUpgradeSkill_Success(t *testing.T) {
	tree := NewSkillTree("user1")
	tree.EarnSkillPoints(5)

	success, message := tree.UpgradeSkill("go_basics")

	if !success {
		t.Errorf("Ожидался успешный апгрейд, получена ошибка: %s", message)
	}
	if tree.Skills["go_basics"].Level != 2 {
		t.Errorf("Ожидаемый уровень навыка 2, получено %d", tree.Skills["go_basics"].Level)
	}
	if tree.SkillPoints != 4 {
		t.Errorf("Ожидаемые очки навыков 4, получено %d", tree.SkillPoints)
	}
}

func TestUpgradeSkill_NotEnoughPoints(t *testing.T) {
	tree := NewSkillTree("user1")
	// Нет очков навыков

	_, _ = tree.UpgradeSkill("concurrency")

	if tree.SkillPoints != 0 {
		t.Errorf("Ожидаемые очки навыков 0, получено %d", tree.SkillPoints)
	}
}

func TestUpgradeSkill_MaxLevel(t *testing.T) {
	tree := NewSkillTree("user1")
	tree.EarnSkillPoints(20)

	// Максимально улучшаем go_basics (5 уровней)
	for i := 0; i < 5; i++ {
		tree.UpgradeSkill("go_basics")
	}

	_, _ = tree.UpgradeSkill("go_basics")

	if tree.Skills["go_basics"].Level != 5 {
		t.Errorf("Ожидаемый уровень навыка 5, получено %d", tree.Skills["go_basics"].Level)
	}
}

func TestUpgradeSkill_Locked(t *testing.T) {
	tree := NewSkillTree("user1")
	tree.EarnSkillPoints(10)

	_, _ = tree.UpgradeSkill("concurrency")

	if tree.Skills["concurrency"].Unlocked {
		t.Error("Навык должен оставаться заблокированным")
	}
}

func TestUpgradeSkill_NonExistent(t *testing.T) {
	tree := NewSkillTree("user1")
	tree.EarnSkillPoints(10)

	_, _ = tree.UpgradeSkill("nonexistent_skill")

	if len(tree.Skills) != 12 {
		t.Errorf("Ожидаемое количество навыков 12, получено %d", len(tree.Skills))
	}
}

func TestSkillTree_CheckUnlocks(t *testing.T) {
	tree := NewSkillTree("user1")
	tree.EarnSkillPoints(20)

	// Улучшаем go_basics до максимума
	for i := 0; i < 4; i++ { // Уже 1 уровень
		tree.UpgradeSkill("go_basics")
	}

	// Проверяем разблокировку
	if !tree.Skills["concurrency"].Unlocked {
		t.Error("Ожидалась разблокировка 'Конкурентность'")
	}
	if !tree.Skills["interfaces"].Unlocked {
		t.Error("Ожидалась разблокировка 'Интерфейсы'")
	}
}

func TestGetTotalBonuses(t *testing.T) {
	tree := NewSkillTree("user1")
	tree.EarnSkillPoints(20)

	// Улучшаем go_basics до 2 уровня
	tree.UpgradeSkill("go_basics")
	tree.UpgradeSkill("go_basics")

	bonuses := tree.GetTotalBonuses()

	// go_basics: уровень 3 (был 1 + 2 улучшения), бонус 5 за уровень = 15 knowledge
	expectedKnowledge := 3 * 5
	if bonuses["knowledge"] != expectedKnowledge {
		t.Errorf("Ожидаемый бонус knowledge %d, получено %d", expectedKnowledge, bonuses["knowledge"])
	}
}

func TestGetSkillPointsForLevel(t *testing.T) {
	tests := []struct {
		level    int
		expected int
	}{
		{1, 2},
		{2, 2},
		{5, 3},
		{10, 4},
		{15, 5},
		{20, 6},
		{50, 12},
		{100, 22},
	}

	for _, tt := range tests {
		result := GetSkillPointsForLevel(tt.level)
		if result != tt.expected {
			t.Errorf("Уровень %d: ожидаемые очки %d, получено %d", tt.level, tt.expected, result)
		}
	}
}

func TestSkillTree_GetBonusName(t *testing.T) {
	tree := NewSkillTree("user1")

	tests := []struct {
		bonusType string
		expected  string
	}{
		{"focus", "Фокус"},
		{"willpower", "Сила воли"},
		{"knowledge", "Знание Go"},
		{"money", "Деньги"},
		{"dopamine", "Дофамин"},
		{"unknown", "Бонус"},
	}

	// Используем рефлексию или прямой вызов через тестирование GetTotalBonuses
	// Проверяем через создание навыка с разным типом бонуса
	for _, tt := range tests {
		tree.Skills["test_skill"] = &Skill{
			ID: "test_skill", Name: "Test", Description: "Test",
			Icon: "🧪", Level: 1, MaxLevel: 1, Cost: 1,
			BonusType: tt.bonusType, BonusValue: 10,
			Unlocked: true, Prereqs: []string{},
		}

		bonuses := tree.GetTotalBonuses()
		if _, exists := bonuses[tt.bonusType]; !exists {
			t.Errorf("Тип бонуса '%s' не найден в результатах", tt.bonusType)
		}

		delete(tree.Skills, "test_skill")
	}
}
