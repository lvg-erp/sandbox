package handler

import (
	"encoding/json"
	"net/http"

	"log"
	"messanger/internal/auth"
	"messanger/internal/domain/service"

	"github.com/google/uuid"
)

type AuthHandler struct {
	jwtConfig   *auth.JWTConfig
	userService *service.UserService
}

func NewAuthHandler(jwtConfig *auth.JWTConfig, userService *service.UserService) *AuthHandler {
	return &AuthHandler{
		jwtConfig:   jwtConfig,
		userService: userService,
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
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
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

	log.Printf("📝 Register request for user: %s", req.Username)

	ctx := r.Context()

	// Проверяем, не существует ли уже пользователь
	existingUser, _ := h.userService.GetUserByUsername(ctx, req.Username)
	if existingUser != nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Создаем пользователя
	user, err := h.userService.CreateUser(ctx, req.Username)
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

	log.Printf("🔐 Login request for user: %s", req.Username)

	ctx := r.Context()

	// Находим пользователя
	user, err := h.userService.GetUserByUsername(ctx, req.Username)
	if err != nil || user == nil {
		log.Printf("❌ User not found: %s", req.Username)
		http.Error(w, "User not found. Please register first.", http.StatusUnauthorized)
		return
	}

	log.Printf("✅ User found: %s", user.Username)

	// Генерируем токены
	accessToken, err := h.jwtConfig.GenerateAccessToken(user.UUID, user.Username)
	if err != nil {
		log.Printf("❌ Failed to generate access token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	refreshToken := uuid.New().String()

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

// Refresh - обновление access токена
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

	// В упрощенной версии считаем, что refresh token - это UUID пользователя
	userUUID, err := uuid.Parse(req.RefreshToken)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	user, err := h.userService.GetUser(ctx, userUUID)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Генерируем новый access токен
	accessToken, err := h.jwtConfig.GenerateAccessToken(user.UUID, user.Username)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	resp := RefreshResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(h.jwtConfig.AccessExpiry.Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
