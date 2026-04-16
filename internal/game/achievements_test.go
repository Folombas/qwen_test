package game

import (
	"testing"
)

func TestNewAchievementSystem(t *testing.T) {
	as := NewAchievementSystem("user123")

	if as.UserID != "user123" {
		t.Errorf("Ожидаемый UserID 'user123', получен '%s'", as.UserID)
	}
	if len(as.Achievements) != 23 {
		t.Errorf("Ожидаемое количество достижений 23, получено %d", len(as.Achievements))
	}

	// Проверяем, что все достижения изначально закрыты
	for id, achievement := range as.Achievements {
		if achievement.Unlocked {
			t.Errorf("Достижение %s должно быть закрыто", id)
		}
		if achievement.ID != id {
			t.Errorf("ID достижения не совпадает: ожидаемый %s, получен %s", id, achievement.ID)
		}
	}
}

func TestAchievementTemplates_Count(t *testing.T) {
	if len(AchievementTemplates) != 23 {
		t.Errorf("Ожидаемое количество шаблонов 23, получено %d", len(AchievementTemplates))
	}
}

func TestAchievementTemplates_Categories(t *testing.T) {
	categories := make(map[string]int)

	for _, tmpl := range AchievementTemplates {
		categories[tmpl.Category]++
	}

	expectedCategories := map[string]int{
		"levels":      5,
		"quests":      3,
		"streaks":     2,
		"temptations": 3,
		"bosses":      2,
		"knowledge":   2,
		"skills":      2,
		"special":     4,
	}

	for category, expectedCount := range expectedCategories {
		if count, exists := categories[category]; !exists {
			t.Errorf("Категория %s не найдена", category)
		} else if count != expectedCount {
			t.Errorf("Категория %s: ожидаемое количество %d, получено %d", category, expectedCount, count)
		}
	}
}

func TestAchievementSystem_Unlock_Success(t *testing.T) {
	as := NewAchievementSystem("user1")

	result := as.Unlock("level_2")

	if !result {
		t.Error("Ожидалась успешная разблокировка достижения")
	}
	if !as.Achievements["level_2"].Unlocked {
		t.Error("Достижение должно быть разблокировано")
	}
	if as.Achievements["level_2"].UnlockedAt.IsZero() {
		t.Error("Ожидалась установленная дата разблокировки")
	}
}

func TestAchievementSystem_Unlock_AlreadyUnlocked(t *testing.T) {
	as := NewAchievementSystem("user1")

	// Разблокируем достижение
	as.Unlock("level_2")

	// Пытаемся разблокировать снова
	result := as.Unlock("level_2")

	if result {
		t.Error("Ожидалась неудачная разблокировка (уже разблокировано)")
	}
}

func TestAchievementSystem_Unlock_NonExistent(t *testing.T) {
	as := NewAchievementSystem("user1")

	result := as.Unlock("nonexistent_achievement")

	if result {
		t.Error("Ожидалась неудачная разблокировка (достижение не существует)")
	}
}

func TestGetUnlockedCount(t *testing.T) {
	as := NewAchievementSystem("user1")

	// Разблокируем 3 достижения
	as.Unlock("level_2")
	as.Unlock("level_5")
	as.Unlock("quest_10")

	count := as.GetUnlockedCount()

	if count != 3 {
		t.Errorf("Ожидаемое количество разблокированных 3, получено %d", count)
	}
}

func TestGetUnlockedCount_Zero(t *testing.T) {
	as := NewAchievementSystem("user1")

	count := as.GetUnlockedCount()

	if count != 0 {
		t.Errorf("Ожидаемое количество 0, получено %d", count)
	}
}

func TestGetTotalCount(t *testing.T) {
	as := NewAchievementSystem("user1")

	total := as.GetTotalCount()

	if total != 23 {
		t.Errorf("Ожидаемое общее количество 23, получено %d", total)
	}
}

func TestCheckAchievements_LevelAchievements(t *testing.T) {
	as := NewAchievementSystem("user1")
	tree := NewSkillTree("user1")
	questSystem := NewQuestSystem("user1")

	tests := []struct {
		level           int
		expectedAchieve string
	}{
		{2, "level_2"},
		{5, "level_5"},
		{10, "level_10"},
		{20, "level_20"},
		{30, "level_30"},
	}

	for _, tt := range tests {
		player := NewPlayer("user1", "Test")
		player.Level = tt.level

		unlocked := as.CheckAchievements(player, tree, questSystem)

		found := false
		for _, id := range unlocked {
			if id == tt.expectedAchieve {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Уровень %d: ожидаемое достижение %s не разблокировано", tt.level, tt.expectedAchieve)
		}
	}
}

func TestCheckAchievements_KnowledgeAchievements(t *testing.T) {
	as := NewAchievementSystem("user1")
	tree := NewSkillTree("user1")
	questSystem := NewQuestSystem("user1")

	// Тест для 50% знаний
	player := NewPlayer("user1", "Test")
	player.GoKnowledge = 50

	unlocked := as.CheckAchievements(player, tree, questSystem)

	found := false
	for _, id := range unlocked {
		if id == "go_50" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Ожидалась разблокировка достижения 'Студент Go' (50% знаний)")
	}

	// Тест для 100% знаний
	as2 := NewAchievementSystem("user2")
	player2 := NewPlayer("user2", "Test2")
	player2.GoKnowledge = 100

	unlocked2 := as2.CheckAchievements(player2, tree, questSystem)

	found50 := false
	found100 := false
	for _, id := range unlocked2 {
		if id == "go_50" {
			found50 = true
		}
		if id == "go_100" {
			found100 = true
		}
	}

	if !found50 {
		t.Error("Ожидалась разблокировка достижения 'Студент Go' (100% знаний)")
	}
	if !found100 {
		t.Error("Ожидалась разблокировка достижения 'Учёный Go' (100% знаний)")
	}
}

func TestCheckAchievements_SkillAchievements(t *testing.T) {
	as := NewAchievementSystem("user1")
	tree := NewSkillTree("user1")
	questSystem := NewQuestSystem("user1")

	// Улучшаем навык до максимума
	tree.EarnSkillPoints(20)
	for i := 0; i < 4; i++ {
		tree.UpgradeSkill("go_basics")
	}

	player := NewPlayer("user1", "Test")

	unlocked := as.CheckAchievements(player, tree, questSystem)

	found := false
	for _, id := range unlocked {
		if id == "skill_max" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Ожидалась разблокировка достижения 'Мастер навыков'")
	}
}
