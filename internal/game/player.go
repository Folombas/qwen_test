// Package game предоставляет игровую логику для веб-приложения.
package game

import (
	"math"
	"time"

	"qwen_test/internal/validator"
)

// Player представляет игрока
type Player struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Level         int       `json:"level"`
	Experience    int       `json:"experience"`
	GoKnowledge   int       `json:"go_knowledge"`
	Focus         int       `json:"focus"`
	Willpower     int       `json:"willpower"`
	Money         int       `json:"money"`
	Dopamine      int       `json:"dopamine"`
	PlayTime      int       `json:"play_time"`
	DaysPlayed    int       `json:"days_played"`
	CurrentDay    int       `json:"current_day"`
	CurrentHour   int       `json:"current_hour"`
	CorrectAnswers int      `json:"correct_answers"`
	WrongAnswers  int       `json:"wrong_answers"`
	AskedQuestions []int    `json:"asked_questions"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	SkillBonuses  map[string]int `json:"skill_bonuses"`
}

// NewPlayer создаёт нового игрока
func NewPlayer(id, name string) *Player {
	now := time.Now()
	player := &Player{
		ID:             id,
		Name:           name,
		Level:          1,
		Experience:     0,
		GoKnowledge:    0,
		Focus:          100,
		Willpower:      100,
		Money:          0,
		Dopamine:       100,
		PlayTime:       0,
		DaysPlayed:     1,
		CurrentDay:     1,
		CurrentHour:    9,
		CorrectAnswers: 0,
		WrongAnswers:   0,
		AskedQuestions: []int{},
		CreatedAt:      now,
		UpdatedAt:      now,
		SkillBonuses:   make(map[string]int),
	}

	// Валидация после создания
	if err := ValidateAfterLoad(player); err != nil {
		validator.LogValidationErrors("нового игрока", err)
	}

	return player
}

// AddExperience добавляет опыт игроку
func (p *Player) AddExperience(xp int) int {
	if xp < 0 {
		xp = 0
	}

	oldLevel := p.Level
	p.Experience = validator.ClampExperience(p.Experience + xp)

	// Проверяем повышение уровня
	newLevel := int(math.Floor(float64(p.Experience)/100)) + 1
	if newLevel > p.Level {
		p.Level = validator.ClampLevel(newLevel)
	}

	return p.Level - oldLevel // Возвращаем количество повышений уровня
}

// StudyGo изучает Go
func (p *Player) StudyGo(minutes int) string {
	if minutes < 0 {
		minutes = 0
	}

	expGain := minutes / 2
	knowledgeGain := minutes / 5
	dopamineGain := minutes / 3

	p.Experience = validator.ClampExperience(p.Experience + expGain)
	p.GoKnowledge = validator.ClampStat(p.GoKnowledge + knowledgeGain)
	p.Dopamine = validator.ClampDopamine(p.Dopamine + dopamineGain)
	p.PlayTime = validator.ClampInt(p.PlayTime+minutes, 0, validator.MaxPlayTime)

	// Обновляем время
	p.CurrentHour = validator.ClampHour(p.CurrentHour + minutes/60)

	return "📚 Вы изучили Go в течение " + string(rune(minutes)) + " минут!"
}

// Rest отдыхает
func (p *Player) Rest(minutes int) string {
	if minutes < 0 {
		minutes = 0
	}

	focusGain := minutes / 2
	dopamineGain := minutes / 3

	p.Focus = validator.ClampStat(p.Focus + focusGain)
	p.Dopamine = validator.ClampDopamine(p.Dopamine + dopamineGain)
	p.PlayTime = validator.ClampInt(p.PlayTime+minutes, 0, validator.MaxPlayTime)

	return "💤 Вы отдохнули в течение " + string(rune(minutes)) + " минут!"
}

// GetRating возвращает рейтинг игрока
func (p *Player) GetRating() int {
	// Формула рейтинга как в focusgo
	rating := p.GoKnowledge*10 + p.Focus*5 + p.Willpower*3 + p.Level*100
	return rating
}

// GetRatingTitle возвращает название рейтинга
func (p *Player) GetRatingTitle() string {
	rating := p.GetRating()
	if rating < 500 {
		return "🌱 Начинающий гофер"
	} else if rating < 1500 {
		return "🌿 Ученик разработчика"
	} else if rating < 3000 {
		return "🌳 Junior Go Developer"
	} else if rating < 5000 {
		return "🏢 Middle Go Developer"
	}
	return "🚀 Senior Go Master"
}

// ApplySkillBonuses применяет бонусы от навыков
func (p *Player) ApplySkillBonuses(tree *SkillTree) {
	if tree == nil {
		return
	}

	bonuses := tree.GetTotalBonuses()
	p.SkillBonuses = bonuses

	p.Focus = validator.ClampStat(p.Focus + bonuses["focus"])
	p.Willpower = validator.ClampStat(p.Willpower + bonuses["willpower"])
	p.GoKnowledge = validator.ClampStat(p.GoKnowledge + bonuses["knowledge"])
	p.Money = validator.ClampMoney(p.Money + bonuses["money"])
	p.Dopamine = validator.ClampDopamine(p.Dopamine + bonuses["dopamine"])
}

// ValidateAfterLoad проверяет валидность данных игрока после загрузки
func ValidateAfterLoad(p *Player) []string {
	var errors []string

	if p.Name == "" {
		errors = append(errors, "Имя игрока пустое")
	} else if len(p.Name) > validator.MaxNameLength {
		errors = append(errors, "Имя игрока слишком длинное")
		p.Name = validator.ClampStringLength(p.Name, validator.MaxNameLength)
	}

	if p.Level < validator.MinLevel || p.Level > validator.MaxLevel {
		errors = append(errors, "Уровень вне диапазона")
		p.Level = validator.ClampLevel(p.Level)
	}

	if p.Experience < validator.MinExperience || p.Experience > validator.MaxExperience {
		errors = append(errors, "Опыт вне диапазона")
		p.Experience = validator.ClampExperience(p.Experience)
	}

	if p.GoKnowledge < validator.MinStatValue || p.GoKnowledge > validator.MaxStatValue {
		errors = append(errors, "Знание Go вне диапазона")
		p.GoKnowledge = validator.ClampStat(p.GoKnowledge)
	}

	if p.Focus < validator.MinStatValue || p.Focus > validator.MaxStatValue {
		errors = append(errors, "Фокус вне диапазона")
		p.Focus = validator.ClampStat(p.Focus)
	}

	if p.Willpower < validator.MinStatValue || p.Willpower > validator.MaxStatValue {
		errors = append(errors, "Сила воли вне диапазона")
		p.Willpower = validator.ClampStat(p.Willpower)
	}

	if p.Money < validator.MinMoneyValue || p.Money > validator.MaxMoneyValue {
		errors = append(errors, "Деньги вне диапазона")
		p.Money = validator.ClampMoney(p.Money)
	}

	if p.Dopamine < validator.MinDopamineValue || p.Dopamine > validator.MaxDopamineValue {
		errors = append(errors, "Дофамин вне диапазона")
		p.Dopamine = validator.ClampDopamine(p.Dopamine)
	}

	if p.CurrentHour < validator.MinHour || p.CurrentHour > validator.MaxHour {
		errors = append(errors, "Час вне диапазона")
		p.CurrentHour = validator.ClampHour(p.CurrentHour)
	}

	return errors
}
