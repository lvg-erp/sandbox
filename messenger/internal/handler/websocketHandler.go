package handler

import (
	"context"
	"net/http"

	_ "github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"messanger/internal/domain/service"
	ws "messanger/internal/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		log.Printf("WebSocket origin: %s", r.Header.Get("Origin"))
		return true // Временно разрешаем все
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		log.Printf("WebSocket error: status=%d, reason=%v", status, reason)
	},
}

type WebSocketHandler struct {
	hub *ws.Hub
}

func NewWebSocketHandler(
	userService *service.UserService,
	chatService *service.ChatService,
	messageService *service.MessageService,
) *WebSocketHandler {
	hub := ws.NewHub(userService, chatService, messageService)
	go hub.Run()

	return &WebSocketHandler{hub: hub}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔥 WebSocket request received! URL: %s, username: %s",
		r.URL.Path,
		r.URL.Query().Get("username")) // Изменили с AccountUUID на username

	username := r.URL.Query().Get("username") // Изменили с AccountUUID на username
	if username == "" {
		log.Printf("❌ Missing username")
		http.Error(w, "username required", http.StatusBadRequest)
		return
	}

	log.Printf("👤 User: %s", username)

	// Получаем или создаем пользователя по username
	ctx := context.Background()
	user, err := h.userService.GetUserByUsername(ctx, username)
	if err != nil || user == nil {
		// Создаем пользователя если не существует
		user, err = h.userService.CreateUser(ctx, username)
		if err != nil {
			log.Printf("❌ Failed to create user: %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
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
		UserUUID: user.UUID, // Используем UUID из базы
	}

	h.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
