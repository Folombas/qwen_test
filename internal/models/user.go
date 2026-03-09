// Package models предоставляет модели данных для приложения
package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User представляет пользователя в системе
type User struct {
	ID        int64      `json:"id"`
	Email     string     `json:"email"`
	Password  string     `json:"-"` // Никогда не возвращать в JSON
	Name      string     `json:"name"`
	Avatar    string     `json:"avatar"`
	Role      string     `json:"role"` // user, admin, moderator
	IsActive  bool       `json:"is_active"`
	IsBanned  bool       `json:"is_banned"`
	EmailVerified bool  `json:"email_verified"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	LastLogin *time.Time `json:"last_login,omitempty"`
}

// UserProfile — расширенный профиль пользователя
type UserProfile struct {
	UserID        int64     `json:"user_id"`
	Bio           string    `json:"bio"`
	Website       string    `json:"website"`
	Location      string    `json:"location"`
	BirthDate     *time.Time `json:"birth_date,omitempty"`
	Settings      UserSettings `json:"settings"`
	Stats         UserStats    `json:"stats"`
}

// UserSettings — настройки пользователя
type UserSettings struct {
	Theme           string `json:"theme"`            // dark, light
	Language        string `json:"language"`         // ru, en
	EmailNotifications bool  `json:"email_notifications"`
	PushNotifications  bool  `json:"push_notifications"`
	DailyReminder      bool  `json:"daily_reminder"`
	BattleReminder     bool  `json:"battle_reminder"`
}

// UserStats — статистика пользователя
type UserStats struct {
	TotalEXP        int `json:"total_exp"`
	Level           int `json:"level"`
	CorrectAnswers  int `json:"correct_answers"`
	WrongAnswers    int `json:"wrong_answers"`
	GoKnowledge     int `json:"go_knowledge"`
	Focus           int `json:"focus"`
	Willpower       int `json:"willpower"`
	TotalPlayTime   int `json:"total_play_time"` // в минутах
	DaysPlayed      int `json:"days_played"`
	CurrentStreak   int `json:"current_streak"`
	MaxStreak       int `json:"max_streak"`
	TournamentsWon  int `json:"tournaments_won"`
	TournamentsPlayed int `json:"tournaments_played"`
}

// Session представляет сессию пользователя
type Session struct {
	ID           string    `json:"id"`
	UserID       int64     `json:"user_id"`
	RefreshToken string    `json:"-"`
	UserAgent    string    `json:"user_agent"`
	IPAddress    string    `json:"ip_address"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	LastActive   time.Time `json:"last_active"`
}

// AuthTokens — пара токенов (access + refresh)
type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// LoginRequest — запрос на вход
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest — запрос на регистрацию
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// RefreshTokenRequest — запрос на обновление токена
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// PasswordResetRequest — запрос на сброс пароля
type PasswordResetRequest struct {
	Email string `json:"email"`
}

// PasswordResetConfirm — подтверждение сброса пароля
type PasswordResetConfirm struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// EmailVerificationRequest — подтверждение email
type EmailVerificationRequest struct {
	Token string `json:"token"`
}

// HashPassword хеширует пароль
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword проверяет пароль
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Sanitize очищает данные пользователя
func (u *User) Sanitize() {
	u.Password = "" // Никогда не возвращать хеш пароля
}

// GetDefaultSettings возвращает настройки по умолчанию
func GetDefaultSettings() UserSettings {
	return UserSettings{
		Theme:              "dark",
		Language:           "ru",
		EmailNotifications: true,
		PushNotifications:  true,
		DailyReminder:      true,
		BattleReminder:     true,
	}
}

// GetDefaultStats возвращает статистику по умолчанию
func GetDefaultStats() UserStats {
	return UserStats{
		TotalEXP:      0,
		Level:         1,
		CorrectAnswers: 0,
		WrongAnswers:   0,
		GoKnowledge:    0,
		Focus:          100,
		Willpower:      100,
		TotalPlayTime:  0,
		DaysPlayed:     1,
		CurrentStreak:  0,
		MaxStreak:      0,
	}
}
