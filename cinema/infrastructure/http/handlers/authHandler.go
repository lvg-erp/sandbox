package handlers

import (
	"cinema/domain/entities"
	"cinema/domain/ports"
	"cinema/internal/auth"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

func Register(repoUser ports.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u entities.User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if u.Email == "" || u.Pass == "" {
			http.Error(w, "email and password required", http.StatusBadRequest)
			return
		}

		// Проверка на существование
		_, err := repoUser.GetUser(u.Email)
		if err == nil {
			http.Error(w, "email already exists", http.StatusConflict)
			return
		}

		// Генерим пассворд
		hashed, err := bcrypt.GenerateFromPassword([]byte(u.Pass), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "failed to hash password", http.StatusInternalServerError)
			return
		}
		u.Pass = string(hashed)

		//
		if err := repoUser.CreateUser(u); err != nil {
			log.Printf("CreateUser error: %v", err)
			http.Error(w, "failed to create user", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func LoginHandler(repoUser ports.UserRepository, sessionRepo ports.SessionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Email string `json:"email"`
			Pass  string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		user, err := repoUser.GetUser(input.Email)
		if err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		if !auth.CheckPassword(input.Pass, user.Pass) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		if err := repoUser.DeleteUserSessions(user.ID); err != nil {
			log.Printf("Login: failed to delete old sessions: %v", err)
		}

		token, err := sessionRepo.CreateSession(user.ID, time.Now().Add(30*24*time.Hour))
		if err != nil {
			http.Error(w, "failed to create session", http.StatusInternalServerError)
			return
		}

		// ← ТОЛЬКО КУКА В ОТВЕТЕ
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    token,
			Expires:  time.Now().Add(30 * 24 * time.Hour),
			HttpOnly: true,
			Path:     "/",
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})

		json.NewEncoder(w).Encode(map[string]string{
			"session_id": token,
		})
	}
}

func Logout(repoSession ports.SessionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie("session_token")
		if cookie != nil {
			repoSession.DeleteSession(cookie.Value)
		}
		http.SetCookie(w, &http.Cookie{
			Name:   "session_token",
			Value:  "",
			MaxAge: -1,
			Path:   "/",
		})

	}
}

func Protected() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(*entities.User)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"email":  user.Email,
			"role":   user.Role,
		})
	}
}
