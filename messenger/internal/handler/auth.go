package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"log"
	"messanger/internal/auth"
	"messanger/internal/domain/service"
)

type AuthHandler struct {
	jwtConfig   *auth.JWTConfig
	userService *service.UserService
	authService *service.AuthService
}

func NewAuthHandler(jwtConfig *auth.JWTConfig, userService *service.UserService, authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		jwtConfig:   jwtConfig,
		userService: userService,
		authService: authService,
	}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	UUID     string `json:"uuid"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	Username     string `json:"username"`
	UUID         string `json:"uuid"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Register - регистрация нового пользователя
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	if len(req.Password) < service.MinPasswordLength {
		http.Error(w, fmt.Sprintf("Password must be at least %d characters", service.MinPasswordLength), http.StatusBadRequest)
		return
	}

	log.Printf("📝 Register request for user: %s", req.Username)

	ctx := r.Context()

	// Проверяем, не существует ли уже пользователь
	existingUser, _ := h.userService.GetUserByUsername(ctx, req.Username)
	if existingUser != nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Создаем пользователя
	user, err := h.userService.CreateUser(ctx, req.Username, req.Password)
	if err != nil {
		log.Printf("❌ Failed to create user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ User registered: %s (UUID: %s)", user.Username, user.UUID)

	resp := RegisterResponse{
		UUID:     user.UUID.String(),
		Username: user.Username,
		Message:  "User registered successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Login - вход пользователя
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	log.Printf("🔐 Login request for user: %s", req.Username)

	ctx := r.Context()

	// Проверяем имя пользователя и пароль
	user, err := h.userService.Authenticate(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			log.Printf("❌ Login failed for user: %s", req.Username)
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}
		log.Printf("❌ Login error for user %s: %v", req.Username, err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Генерируем токены
	accessToken, err := h.jwtConfig.GenerateAccessToken(user.UUID, user.Username)
	if err != nil {
		log.Printf("❌ Failed to generate access token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.authService.IssueRefreshToken(ctx, user.UUID)
	if err != nil {
		log.Printf("❌ Failed to issue refresh token: %v", err)
		http.Error(w, "Failed to issue refresh token", http.StatusInternalServerError)
		return
	}

	resp := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(h.jwtConfig.AccessExpiry.Seconds()),
		Username:     user.Username,
		UUID:         user.UUID.String(),
	}

	log.Printf("✅ Login successful for user: %s", user.Username)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Refresh - обновление access токена по refresh токену (с ротацией)
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, "Refresh token is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	user, newRefreshToken, err := h.authService.RefreshTokens(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRefreshToken) {
			http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
			return
		}
		log.Printf("❌ Refresh error: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	accessToken, err := h.jwtConfig.GenerateAccessToken(user.UUID, user.Username)
	if err != nil {
		log.Printf("❌ Failed to generate access token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	resp := RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(h.jwtConfig.AccessExpiry.Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Logout отзывает refresh токен. Ответ всегда успешный, чтобы не раскрывать
// валидность переданного токена.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, "Refresh token is required", http.StatusBadRequest)
		return
	}

	if err := h.authService.RevokeToken(r.Context(), req.RefreshToken); err != nil {
		log.Printf("❌ Logout error: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "logged out"})
}
