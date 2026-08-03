package handler

import (
	"context"
	"messanger/internal/auth"
	"net/http"

	"log"
	"messanger/internal/domain/service"
	ws "messanger/internal/websocket"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebSocketHandler struct {
	hub         *ws.Hub
	userService *service.UserService
	jwtConfig   *auth.JWTConfig
}

func NewWebSocketHandler(
	userService *service.UserService,
	chatService *service.ChatService,
	messageService *service.MessageService,
	jwtConfig *auth.JWTConfig,
) *WebSocketHandler {
	hub := ws.NewHub(userService, chatService, messageService)
	go hub.Run()

	return &WebSocketHandler{
		hub:         hub,
		userService: userService,
		jwtConfig:   jwtConfig,
	}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		log.Printf("❌ Missing token")
		http.Error(w, "token required", http.StatusUnauthorized)
		return
	}

	claims, err := h.jwtConfig.ValidateToken(token)
	if err != nil {
		log.Printf("❌ Invalid token: %v", err)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	log.Printf("🔥 WebSocket connection for user: %s (UUID: %s)", claims.Username, claims.UserUUID)

	ctx := context.Background()
	userUUID, err := uuid.Parse(claims.UserUUID)
	if err != nil {
		log.Printf("❌ Invalid user UUID: %v", err)
		http.Error(w, "invalid user", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.GetUser(ctx, userUUID)
	if err != nil || user == nil {
		log.Printf("❌ User not found: %s", claims.Username)
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ Upgrade failed: %v", err)
		return
	}

	log.Printf("✅ WebSocket connected for user: %s (UUID: %s)", user.Username, user.UUID)

	client := &ws.Client{
		Hub:      h.hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Username: user.Username, // ПЕРЕДАЕМ USERNAME!
		UserUUID: user.UUID,
	}

	h.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
