// Package auth предоставляет функционал аутентификации и авторизации
package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig конфигурация JWT
type JWTConfig struct {
	SecretKey      string
	AccessDuration time.Duration
	RefreshDuration time.Duration
	Issuer         string
}

// Claims представляет claims для JWT токена
type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWTService сервис для работы с JWT
type JWTService struct {
	config JWTConfig
}

// TokenPair пара токенов (access + refresh)
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// NewJWTService создаёт новый JWT сервис
func NewJWTService(secretKey string, accessDuration, refreshDuration time.Duration) *JWTService {
	return &JWTService{
		config: JWTConfig{
			SecretKey:       secretKey,
			AccessDuration:  accessDuration,
			RefreshDuration: refreshDuration,
			Issuer:          "go-quiz",
		},
	}
}

// GenerateTokens генерирует пару токенов
func (s *JWTService) GenerateTokens(userID int64, email, role string) (*TokenPair, error) {
	accessToken, err := s.generateAccessToken(userID, email, role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.config.AccessDuration.Seconds()),
	}, nil
}

// generateAccessToken генерирует access токен
func (s *JWTService) generateAccessToken(userID int64, email, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.AccessDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.Issuer,
			Subject:   strconv.FormatInt(userID, 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.SecretKey))
}

// generateRefreshToken генерирует refresh токен
func (s *JWTService) generateRefreshToken(userID int64) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(s.config.RefreshDuration)),
		IssuedAt:  jwt.NewNumericDate(now),
		Issuer:    s.config.Issuer,
		Subject:   strconv.FormatInt(userID, 10),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.SecretKey))
}

// ValidateAccessToken валидирует access токен
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.config.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateRefreshToken валидирует refresh токен
func (s *JWTService) ValidateRefreshToken(tokenString string) (jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.config.SecretKey), nil
	})

	if err != nil {
		return jwt.RegisteredClaims{}, err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return *claims, nil
	}

	return jwt.RegisteredClaims{}, errors.New("invalid refresh token")
}

// RefreshTokens обновляет пару токенов
func (s *JWTService) RefreshTokens(refreshToken string, userID int64, email, role string) (*TokenPair, error) {
	// Валидируем refresh токен
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Проверяем что токен принадлежит этому пользователю
	if claims.Subject != strconv.FormatInt(userID, 10) {
		return nil, errors.New("token mismatch")
	}

	// Генерируем новые токены
	return s.GenerateTokens(userID, email, role)
}

// GetTokenExpiry возвращает время истечения токена
func (s *JWTService) GetTokenExpiry() (accessExpiry, refreshExpiry time.Duration) {
	return s.config.AccessDuration, s.config.RefreshDuration
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() JWTConfig {
	return JWTConfig{
		SecretKey:       "your-secret-key-change-in-production",
		AccessDuration:  15 * time.Minute,
		RefreshDuration: 7 * 24 * time.Hour, // 7 дней
		Issuer:          "go-quiz",
	}
}
