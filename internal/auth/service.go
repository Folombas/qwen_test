package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"qwen_test/internal/models"
)

// AuthService сервис аутентификации
type AuthService struct {
	db         *sql.DB
	jwtService *JWTService
}

// NewAuthService создаёт новый сервис аутентификации
func NewAuthService(db *sql.DB, jwtService *JWTService) *AuthService {
	return &AuthService{
		db:         db,
		jwtService: jwtService,
	}
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(req models.RegisterRequest) (*models.User, *models.AuthTokens, error) {
	// Проверяем существует ли пользователь с таким email
	var exists int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", req.Email).Scan(&exists)
	if err != nil {
		return nil, nil, err
	}
	if exists > 0 {
		return nil, nil, errors.New("user with this email already exists")
	}

	// Хешируем пароль
	hashedPassword, err := models.HashPassword(req.Password)
	if err != nil {
		return nil, nil, err
	}

	// Создаём пользователя
	var userID int64
	err = s.db.QueryRow(`
		INSERT INTO users (email, password, name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, req.Email, hashedPassword, req.Name, "user", true, time.Now(), time.Now()).Scan(&userID)
	if err != nil {
		return nil, nil, err
	}

	// Создаём профиль
	defaultSettings := models.GetDefaultSettings()
	defaultStats := models.GetDefaultStats()
	
	_, err = s.db.Exec(`
		INSERT INTO user_profiles (user_id, settings, stats, created_at)
		VALUES ($1, $2, $3, $4)
	`, userID, defaultSettings, defaultStats, time.Now())
	if err != nil {
		log.Printf("Warning: failed to create profile for user %d: %v", userID, err)
	}

	// Создаём токен верификации email
	verificationToken, err := s.createVerificationToken(userID)
	if err != nil {
		return nil, nil, err
	}

	// Сохраняем токен
	_, err = s.db.Exec(`
		INSERT INTO email_verification_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`, userID, verificationToken, time.Now().Add(24*time.Hour))
	if err != nil {
		log.Printf("Warning: failed to create verification token: %v", err)
	}

	// Генерируем JWT токены
	tokens, err := s.jwtService.GenerateTokens(userID, req.Email, "user")
	if err != nil {
		return nil, nil, err
	}

	user := &models.User{
		ID:            userID,
		Email:         req.Email,
		Name:          req.Name,
		Role:          "user",
		IsActive:      true,
		EmailVerified: false,
		CreatedAt:     time.Now(),
	}

	// Возвращаем пользователя без пароля и с токеном верификации
	user.Sanitize()
	return user, &models.AuthTokens{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    "Bearer",
	}, nil
}

// Login авторизует пользователя
func (s *AuthService) Login(req models.LoginRequest) (*models.User, *models.AuthTokens, error) {
	// Находим пользователя
	var user models.User
	var hashedPassword string
	err := s.db.QueryRow(`
		SELECT id, email, password, name, role, is_active, email_verified, created_at, updated_at
		FROM users WHERE email = $1
	`, req.Email).Scan(
		&user.ID, &user.Email, &hashedPassword, &user.Name, &user.Role,
		&user.IsActive, &user.EmailVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, errors.New("invalid email or password")
		}
		return nil, nil, err
	}

	// Проверяем пароль
	if !models.CheckPassword(req.Password, hashedPassword) {
		return nil, nil, errors.New("invalid email or password")
	}

	// Проверяем активен ли пользователь
	if !user.IsActive {
		return nil, nil, errors.New("account is deactivated")
	}

	// Обновляем last_login
	_, err = s.db.Exec("UPDATE users SET last_login = $1 WHERE id = $2", time.Now(), user.ID)
	if err != nil {
		log.Printf("Warning: failed to update last_login: %v", err)
	}

	// Генерируем JWT токены
	tokens, err := s.jwtService.GenerateTokens(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, nil, err
	}

	user.Sanitize()
	return &user, &models.AuthTokens{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    "Bearer",
	}, nil
}

// RefreshToken обновляет токены
func (s *AuthService) RefreshToken(refreshToken string, userID int64) (*models.AuthTokens, error) {
	// Находим пользователя
	var email, role string
	err := s.db.QueryRow("SELECT email, role FROM users WHERE id = $1", userID).Scan(&email, &role)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Обновляем токены
	tokens, err := s.jwtService.RefreshTokens(refreshToken, userID, email, role)
	if err != nil {
		return nil, err
	}

	return &models.AuthTokens{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		TokenType:    "Bearer",
	}, nil
}

// Logout завершает сессию
func (s *AuthService) Logout(userID int64, refreshToken string) error {
	// В продакшене здесь нужно инвалидировать refresh токен в Redis/БД
	_, err := s.db.Exec(`
		INSERT INTO revoked_tokens (user_id, token, revoked_at)
		VALUES ($1, $2, $3)
	`, userID, refreshToken, time.Now())
	return err
}

// GetCurrentUser получает текущего пользователя
func (s *AuthService) GetCurrentUser(userID int64) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, email, name, avatar, role, is_active, is_banned, email_verified, created_at, updated_at, last_login
		FROM users WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.Email, &user.Name, &user.Avatar, &user.Role,
		&user.IsActive, &user.IsBanned, &user.EmailVerified,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	user.Sanitize()
	return &user, nil
}

// createVerificationToken создаёт токен верификации
func (s *AuthService) createVerificationToken(userID int64) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// VerifyEmail подтверждает email
func (s *AuthService) VerifyEmail(token string) error {
	// Находим токен в БД
	var userID int64
	var expiresAt time.Time
	err := s.db.QueryRow(`
		SELECT user_id, expires_at FROM email_verification_tokens WHERE token = $1
	`, token).Scan(&userID, &expiresAt)
	if err != nil {
		return errors.New("invalid verification token")
	}

	// Проверяем не истёк ли токен
	if time.Now().After(expiresAt) {
		return errors.New("verification token expired")
	}

	// Подтверждаем email
	_, err = s.db.Exec("UPDATE users SET email_verified = true WHERE id = $1", userID)
	if err != nil {
		return err
	}

	// Удаляем использованный токен
	_, err = s.db.Exec("DELETE FROM email_verification_tokens WHERE token = $1", token)
	return err
}

// RequestPasswordReset запрашивает сброс пароля
func (s *AuthService) RequestPasswordReset(email string) (string, error) {
	var userID int64
	err := s.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		// Не показываем что email не найден (security best practice)
		return "", nil
	}

	// Создаём токен сброса
	resetToken, err := s.createVerificationToken(userID)
	if err != nil {
		return "", err
	}

	// Сохраняем токен
	_, err = s.db.Exec(`
		INSERT INTO password_reset_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`, userID, resetToken, time.Now().Add(time.Hour))
	if err != nil {
		return "", err
	}

	return resetToken, nil
}

// ResetPassword сбрасывает пароль
func (s *AuthService) ResetPassword(token, newPassword string) error {
	// Находим токен
	var userID int64
	var expiresAt time.Time
	err := s.db.QueryRow(`
		SELECT user_id, expires_at FROM password_reset_tokens WHERE token = $1
	`, token).Scan(&userID, &expiresAt)
	if err != nil {
		return errors.New("invalid reset token")
	}

	// Проверяем не истёк ли токен
	if time.Now().After(expiresAt) {
		return errors.New("reset token expired")
	}

	// Хешируем новый пароль
	hashedPassword, err := models.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Обновляем пароль
	_, err = s.db.Exec("UPDATE users SET password = $1 WHERE id = $2", hashedPassword, userID)
	if err != nil {
		return err
	}

	// Удаляем использованный токен
	_, err = s.db.Exec("DELETE FROM password_reset_tokens WHERE token = $1", token)
	return err
}

// ChangePassword меняет пароль
func (s *AuthService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	// Получаем текущий хеш пароля
	var hashedPassword string
	err := s.db.QueryRow("SELECT password FROM users WHERE id = $1", userID).Scan(&hashedPassword)
	if err != nil {
		return errors.New("user not found")
	}

	// Проверяем старый пароль
	if !models.CheckPassword(oldPassword, hashedPassword) {
		return errors.New("invalid current password")
	}

	// Хешируем новый пароль
	newHashedPassword, err := models.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Обновляем пароль
	_, err = s.db.Exec("UPDATE users SET password = $1 WHERE id = $2", newHashedPassword, userID)
	return err
}
