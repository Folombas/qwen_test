package auth

import (
	"encoding/json"
	"net/http"

	"qwen_test/internal/models"
	"qwen_test/internal/response"
)

// AuthHandler HTTP обработчики аутентификации
type AuthHandler struct {
	authService *AuthService
}

// NewAuthHandler создаёт новый обработчик
func NewAuthHandler(authService *AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register обрабатывает регистрацию
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w, r.Method)
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Валидация
	if req.Email == "" || req.Password == "" || req.Name == "" {
		response.BadRequest(w, "Email, password and name are required")
		return
	}

	if len(req.Password) < 6 {
		response.BadRequest(w, "Password must be at least 6 characters")
		return
	}

	// Регистрируем пользователя
	user, tokens, err := h.authService.Register(req)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	// Отправляем ответ
	response.JSON(w, http.StatusCreated, map[string]interface{}{
		"user":    user,
		"tokens":  tokens,
		"message": "Registration successful! Please check your email for verification.",
	})
}

// Login обрабатывает вход
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w, r.Method)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Авторизуем пользователя
	user, tokens, err := h.authService.Login(req)
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}

	// Отправляем ответ
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"user":   user,
		"tokens": tokens,
	})
}

// Logout обрабатывает выход
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w, r.Method)
		return
	}

	userID := GetUserID(r)
	if userID == 0 {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	var req models.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := h.authService.Logout(userID, req.RefreshToken); err != nil {
		response.InternalServerError(w, err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Logout successful")
}

// Refresh обрабатывает обновление токена
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w, r.Method)
		return
	}

	var req models.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	userID := GetUserID(r)
	if userID == 0 {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	tokens, err := h.authService.RefreshToken(req.RefreshToken, userID)
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{"tokens": tokens})
}

// Me возвращает текущего пользователя
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.MethodNotAllowed(w, r.Method)
		return
	}

	userID := GetUserID(r)
	if userID == 0 {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	user, err := h.authService.GetCurrentUser(userID)
	if err != nil {
		response.NotFound(w, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, user)
}

// VerifyEmail обрабатывает подтверждение email
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w, r.Method)
		return
	}

	var req models.EmailVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := h.authService.VerifyEmail(req.Token); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Email verified successfully")
}

// ForgotPassword обрабатывает запрос на сброс пароля
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w, r.Method)
		return
	}

	var req models.PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	resetToken, err := h.authService.RequestPasswordReset(req.Email)
	if err != nil {
		response.InternalServerError(w, err.Error())
		return
	}

	// В продакшене здесь отправляем email со ссылкой на сброс
	// Пока просто возвращаем токен (для тестирования)
	response.JSON(w, http.StatusOK, map[string]string{
		"message":     "Password reset link sent to your email",
		"reset_token": resetToken, // Удалить в продакшене!
	})
}

// ResetPassword обрабатывает сброс пароля
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w, r.Method)
		return
	}

	var req models.PasswordResetConfirm
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := h.authService.ResetPassword(req.Token, req.NewPassword); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Password reset successfully")
}

// ChangePassword обрабатывает смену пароля
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w, r.Method)
		return
	}

	userID := GetUserID(r)
	if userID == 0 {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := h.authService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Success(w, http.StatusOK, "Password changed successfully")
}
