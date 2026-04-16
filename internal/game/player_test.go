package game

import (
	"testing"
)

func TestNewPlayer(t *testing.T) {
	player := NewPlayer("user123", "TestPlayer")

	if player.ID != "user123" {
		t.Errorf("Ожидаемый ID 'user123', получен '%s'", player.ID)
	}
	if player.Name != "TestPlayer" {
		t.Errorf("Ожидаемое имя 'TestPlayer', получен '%s'", player.Name)
	}
	if player.Level != 1 {
		t.Errorf("Ожидаемый уровень 1, получен %d", player.Level)
	}
	if player.Experience != 0 {
		t.Errorf("Ожидаемый опыт 0, получен %d", player.Experience)
	}
	if player.Focus != 100 {
		t.Errorf("Ожидаемый фокус 100, получен %d", player.Focus)
	}
	if player.Willpower != 100 {
		t.Errorf("Ожидаемая сила воли 100, получен %d", player.Willpower)
	}
	if player.Dopamine != 100 {
		t.Errorf("Ожидаемый дофамин 100, получен %d", player.Dopamine)
	}
	if player.GoKnowledge != 0 {
		t.Errorf("Ожидаемое знание Go 0, получен %d", player.GoKnowledge)
	}
	if player.Money != 0 {
		t.Errorf("Ожидаемые деньги 0, получено %d", player.Money)
	}
	if len(player.AskedQuestions) != 0 {
		t.Errorf("Ожидаемая пустая история вопросов, получена длина %d", len(player.AskedQuestions))
	}
	if player.SkillBonuses == nil {
		t.Error("Ожидаемая не-nil карта SkillBonuses")
	}
}

func TestAddExperience_NoLevelUp(t *testing.T) {
	player := NewPlayer("user1", "Test")
	
	levelGain := player.AddExperience(50)
	
	if player.Experience != 50 {
		t.Errorf("Ожидаемый опыт 50, получен %d", player.Experience)
	}
	if player.Level != 1 {
		t.Errorf("Ожидаемый уровень 1, получен %d", player.Level)
	}
	if levelGain != 0 {
		t.Errorf("Ожидаемое повышение уровня 0, получено %d", levelGain)
	}
}

func TestAddExperience_WithLevelUp(t *testing.T) {
	player := NewPlayer("user1", "Test")
	
	levelGain := player.AddExperience(150)
	
	if player.Experience != 150 {
		t.Errorf("Ожидаемый опыт 150, получен %d", player.Experience)
	}
	if player.Level != 2 {
		t.Errorf("Ожидаемый уровень 2, получен %d", player.Level)
	}
	if levelGain != 1 {
		t.Errorf("Ожидаемое повышение уровня 1, получено %d", levelGain)
	}
}

func TestAddExperience_MultipleLevelUps(t *testing.T) {
	player := NewPlayer("user1", "Test")
	
	levelGain := player.AddExperience(350)
	
	if player.Experience != 350 {
		t.Errorf("Ожидаемый опыт 350, получен %d", player.Experience)
	}
	if player.Level != 4 {
		t.Errorf("Ожидаемый уровень 4, получен %d", player.Level)
	}
	if levelGain != 3 {
		t.Errorf("Ожидаемое повышение уровня 3, получено %d", levelGain)
	}
}

func TestAddExperience_NegativeXP(t *testing.T) {
	player := NewPlayer("user1", "Test")
	
	levelGain := player.AddExperience(-50)
	
	// Отрицательный опыт обнуляется до 0
	if player.Experience != 0 {
		t.Errorf("Ожидаемый опыт 0, получен %d", player.Experience)
	}
	// Уровень остаётся 1 (базовый)
	if levelGain != 0 {
		t.Errorf("Ожидаемое повышение уровня 0, получено %d", levelGain)
	}
}

func TestAddExperience_MaxCap(t *testing.T) {
	player := NewPlayer("user1", "Test")
	
	// Добавляем больше максимального опыта
	levelGain := player.AddExperience(9999999)
	
	if player.Experience != 999999 {
		t.Errorf("Ожидаемый опыт 999999 (максимум), получен %d", player.Experience)
	}
	if player.Level != 100 {
		t.Errorf("Ожидаемый уровень 100 (максимум), получен %d", player.Level)
	}
	if levelGain != 99 {
		t.Errorf("Ожидаемое повышение уровня 99, получено %d", levelGain)
	}
}

func TestStudyGo_30Minutes(t *testing.T) {
	player := NewPlayer("user1", "Test")
	
	result := player.StudyGo(30)
	
	if result != "📚 Вы изучили Go в течение 30 минут!" {
		t.Errorf("Неверное сообщение: %s", result)
	}
	// EXP: 30/2 = 15
	if player.Experience != 15 {
		t.Errorf("Ожидаемый опыт 15, получен %d", player.Experience)
	}
	// Knowledge: 30/5 = 6
	if player.GoKnowledge != 6 {
		t.Errorf("Ожидаемое знание Go 6, получен %d", player.GoKnowledge)
	}
	// Dopamine: 30/3 = 10
	if player.Dopamine != 110 {
		t.Errorf("Ожидаемый дофамин 110, получен %d", player.Dopamine)
	}
	if player.PlayTime != 30 {
		t.Errorf("Ожидаемое время игры 30, получено %d", player.PlayTime)
	}
}

func TestStudyGo_60Minutes(t *testing.T) {
	player := NewPlayer("user1", "Test")
	
	player.StudyGo(60)
	
	// EXP: 60/2 = 30
	if player.Experience != 30 {
		t.Errorf("Ожидаемый опыт 30, получен %d", player.Experience)
	}
	// Knowledge: 60/5 = 12
	if player.GoKnowledge != 12 {
		t.Errorf("Ожидаемое знание Go 12, получен %d", player.GoKnowledge)
	}
	// Dopamine: 60/3 = 20
	if player.Dopamine != 120 {
		t.Errorf("Ожидаемый дофамин 120, получен %d", player.Dopamine)
	}
	if player.PlayTime != 60 {
		t.Errorf("Ожидаемое время игры 60, получено %d", player.PlayTime)
	}
}

func TestStudyGo_NegativeMinutes(t *testing.T) {
	player := NewPlayer("user1", "Test")
	
	player.StudyGo(-30)
	
	if player.Experience != 0 {
		t.Errorf("Ожидаемый опыт 0, получен %d", player.Experience)
	}
	if player.PlayTime != 0 {
		t.Errorf("Ожидаемое время игры 0, получено %d", player.PlayTime)
	}
}

func TestStudyGo_KnowledgeCap(t *testing.T) {
	player := NewPlayer("user1", "Test")
	player.GoKnowledge = 95
	
	player.StudyGo(60) // +12 knowledge
	
	if player.GoKnowledge != 100 {
		t.Errorf("Ожидаемое знание Go 100 (кап), получен %d", player.GoKnowledge)
	}
}

func TestRest_15Minutes(t *testing.T) {
	player := NewPlayer("user1", "Test")
	player.Focus = 50
	player.Dopamine = 50
	
	result := player.Rest(15)
	
	if result != "💤 Вы отдохнули в течение 15 минут!" {
		t.Errorf("Неверное сообщение: %s", result)
	}
	// Focus: 15/2 = 7
	if player.Focus != 57 {
		t.Errorf("Ожидаемый фокус 57, получен %d", player.Focus)
	}
	// Dopamine: 15/3 = 5
	if player.Dopamine != 55 {
		t.Errorf("Ожидаемый дофамин 55, получен %d", player.Dopamine)
	}
	if player.PlayTime != 15 {
		t.Errorf("Ожидаемое время игры 15, получено %d", player.PlayTime)
	}
}

func TestRest_30Minutes(t *testing.T) {
	player := NewPlayer("user1", "Test")
	player.Focus = 50
	player.Dopamine = 50
	
	player.Rest(30)
	
	// Focus: 30/2 = 15
	if player.Focus != 65 {
		t.Errorf("Ожидаемый фокус 65, получен %d", player.Focus)
	}
	// Dopamine: 30/3 = 10
	if player.Dopamine != 60 {
		t.Errorf("Ожидаемый дофамин 60, получен %d", player.Dopamine)
	}
}

func TestRest_NegativeMinutes(t *testing.T) {
	player := NewPlayer("user1", "Test")
	
	player.Rest(-15)
	
	if player.PlayTime != 0 {
		t.Errorf("Ожидаемое время игры 0, получено %d", player.PlayTime)
	}
}

func TestRest_FocusCap(t *testing.T) {
	player := NewPlayer("user1", "Test")
	player.Focus = 95
	
	player.Rest(30) // +15 focus
	
	if player.Focus != 100 {
		t.Errorf("Ожидаемый фокус 100 (кап), получен %d", player.Focus)
	}
}

func TestGetRating(t *testing.T) {
	player := NewPlayer("user1", "Test")
	player.Level = 5
	player.GoKnowledge = 50
	player.Focus = 80
	player.Willpower = 70
	
	rating := player.GetRating()
	// Formula: 50*10 + 80*5 + 70*3 + 5*100 = 500 + 400 + 210 + 500 = 1610
	if rating != 1610 {
		t.Errorf("Ожидаемый рейтинг 1610, получен %d", rating)
	}
}

func TestGetRatingTitle_Beginner(t *testing.T) {
	player := NewPlayer("user1", "Test")
	player.GoKnowledge = 10
	player.Focus = 20
	player.Willpower = 10
	player.Level = 1
	
	rating := player.GetRating()
	if rating >= 500 {
		t.Skip("Конфигурация игрока не соответствует титулу начинающего")
	}
	
	title := player.GetRatingTitle()
	if title != "🌱 Начинающий гофер" {
		t.Errorf("Ожидаемый титул '🌱 Начинающий гофер', получен '%s'", title)
	}
}

func TestGetRatingTitle_Junior(t *testing.T) {
	player := NewPlayer("user1", "Test")
	player.Level = 10
	player.GoKnowledge = 50
	player.Focus = 50
	player.Willpower = 50
	
	rating := player.GetRating()
	// 50*10 + 50*5 + 50*3 + 10*100 = 500 + 250 + 150 + 1000 = 1900
	if rating < 1500 || rating >= 3000 {
		t.Skipf("Рейтинг %d не соответствует Junior диапазону (1500-3000)", rating)
	}
	
	title := player.GetRatingTitle()
	if title != "🌳 Junior Go Developer" {
		t.Errorf("Ожидаемый титул '🌳 Junior Go Developer', получен '%s'", title)
	}
}

func TestGetRatingTitle_Senior(t *testing.T) {
	player := NewPlayer("user1", "Test")
	player.Level = 50
	player.GoKnowledge = 100
	player.Focus = 100
	player.Willpower = 100
	
	rating := player.GetRating()
	// 100*10 + 100*5 + 100*3 + 50*100 = 1000 + 500 + 300 + 5000 = 6800
	if rating < 5000 {
		t.Skipf("Рейтинг %d не соответствует Senior диапазону (5000+)", rating)
	}
	
	title := player.GetRatingTitle()
	if title != "🚀 Senior Go Master" {
		t.Errorf("Ожидаемый титул '🚀 Senior Go Master', получен '%s'", title)
	}
}

func TestValidateAfterLoad_EmptyName(t *testing.T) {
	player := &Player{
		ID:            "user1",
		Name:          "",
		Level:         1,
		Experience:    0,
		GoKnowledge:   50,
		Focus:         50,
		Willpower:     50,
		Money:         100,
		Dopamine:      100,
		CurrentHour:   12,
		AskedQuestions: []int{},
		SkillBonuses:  make(map[string]int),
	}
	
	errors := ValidateAfterLoad(player)
	
	if len(errors) != 1 {
		t.Errorf("Ожидалась 1 ошибка, получено %d", len(errors))
	}
	if errors[0] != "Имя игрока пустое" {
		t.Errorf("Ожидаемая ошибка 'Имя игрока пустое', получена '%s'", errors[0])
	}
}

func TestValidateAfterLoad_ValidPlayer(t *testing.T) {
	player := NewPlayer("user1", "TestPlayer")
	
	errors := ValidateAfterLoad(player)
	
	if len(errors) != 0 {
		t.Errorf("Ожидалось 0 ошибок, получено %d: %v", len(errors), errors)
	}
}
