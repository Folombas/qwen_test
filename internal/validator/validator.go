// Package validator предоставляет функции для валидации и санитизации игровых данных.
package validator

import (
	"fmt"
	"log"
)

// Константы валидации
const (
	// Характеристики (0-100)
	MinStatValue     = 0
	MaxStatValue     = 100

	// Деньги и дофамин
	MinMoneyValue    = 0
	MaxMoneyValue    = 999999
	MinDopamineValue = 0
	MaxDopamineValue = 999

	// Опыт и уровень
	MinLevel       = 1
	MaxLevel       = 100
	MinExperience  = 0
	MaxExperience  = 999999

	// Время
	MinPlayTime   = 0
	MaxPlayTime   = 999999
	MinDaysPlayed = 1
	MaxDaysPlayed = 9999
	MinHour       = 0
	MaxHour       = 23

	// Длины строк
	MaxNameLength        = 50
	MaxAchievementLength = 200
	MaxTemptationLength  = 100
)

// ClampInt ограничивает целое число диапазоном [min, max]
func ClampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// ClampStat ограничивает характеристику диапазоном [0, 100]
func ClampStat(value int) int {
	return ClampInt(value, MinStatValue, MaxStatValue)
}

// ClampMoney ограничивает деньги диапазоном [0, 999999]
func ClampMoney(value int) int {
	return ClampInt(value, MinMoneyValue, MaxMoneyValue)
}

// ClampDopamine ограничивает дофамин диапазоном [0, 999]
func ClampDopamine(value int) int {
	return ClampInt(value, MinDopamineValue, MaxDopamineValue)
}

// ClampExperience ограничивает опыт диапазоном [0, 999999]
func ClampExperience(value int) int {
	return ClampInt(value, MinExperience, MaxExperience)
}

// ClampLevel ограничивает уровень диапазоном [1, 100]
func ClampLevel(value int) int {
	return ClampInt(value, MinLevel, MaxLevel)
}

// ClampHour ограничивает час диапазоном [0, 23]
func ClampHour(value int) int {
	return ClampInt(value, MinHour, MaxHour)
}

// ClampStringLength ограничивает длину строки
func ClampStringLength(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}

// IsValidName проверяет корректность имени
func IsValidName(name string) bool {
	return len(name) > 0 && len(name) <= MaxNameLength
}

// FormatValidationErrors форматирует ошибки для вывода
func FormatValidationErrors(errors []string) string {
	if len(errors) == 0 {
		return ""
	}
	result := "⚠️  WARNING: Валидация:\n"
	for _, err := range errors {
		result += fmt.Sprintf("  - %s\n", err)
	}
	return result
}

// LogValidationErrors логирует ошибки валидации
func LogValidationErrors(context string, errors []string) {
	if len(errors) == 0 {
		return
	}
	log.Printf("⚠️  WARNING: Валидация %s:\n%s", context, FormatValidationErrors(errors))
}
