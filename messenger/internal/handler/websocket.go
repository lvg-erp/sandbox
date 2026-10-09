package handler

import (
	"log"
	"messanger/internal/auth"
	"messanger/internal/domain/service"
	ws "messanger/internal/websocket"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// sameOriginOrLocal разрешает соединения с того же хоста и localhost
// (удобно для разработки), остальное отклоняет.
func sameOriginOrLocal(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true // не браузерный клиент
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if u.Host == r.Host {
		return true
	}
	h := u.Hostname()
	return h == "localhost" || h == "127.0.0.1" || h == "::1"
}

var upgrader = websocket.Upgrader{
	CheckOrigin:     sameOriginOrLocal,
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

	ctx := r.Context()
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
		Done:     make(chan struct{}),
		Username: user.Username, // ПЕРЕДАЕМ USERNAME!
		UserUUID: user.UUID,
	}

	h.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
