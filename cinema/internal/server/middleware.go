package server

import (
	"cinema/internal/db"
	"context"
	"log"
	"net/http"
)

func AuthMiddleware(repo db.Repository, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			log.Printf("Auth: NO COOKIE")
			http.Error(w, "no session", http.StatusUnauthorized)
			return
		}
		log.Printf("Auth: COOKIE FOUND: %s", cookie.Value)

		session, err := repo.GetSession(cookie.Value)
		if err != nil {
			log.Printf("Auth: SESSION NOT FOUND OR EXPIRED: %v", err)
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}
		log.Printf("Auth: SESSION OK: user_id=%d, expires=%s", session.UserID, session.ExpiresAt)

		user, err := repo.GetUserByID(session.UserID)
		if err != nil {
			log.Printf("Auth: USER NOT FOUND: %v", err)
			http.Error(w, "user not found", http.StatusUnauthorized)
			return
		}
		log.Printf("Auth: USER LOADED: ID=%d, Email=%s, Role=%s", user.ID, user.Email, user.Role)

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func RoleMiddleware(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value("user").(db.User)
		//log.Printf("Role check: user=%v, role=%s, required=%s", user.ID, user.Role, requiredRole)
		if !ok || user.Role != requiredRole {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
}
