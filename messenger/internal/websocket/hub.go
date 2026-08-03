package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"messanger/internal/domain/entity"
	"messanger/internal/domain/service"

	"github.com/google/uuid"
)

type Hub struct {
	// Экспортируемые поля (с большой буквы)
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *BroadcastMessage

	// Приватные поля (с маленькой буквы)
	clients        map[uuid.UUID]*Client
	mu             sync.RWMutex
	userService    *service.UserService
	chatService    *service.ChatService
	messageService *service.MessageService
}

type BroadcastMessage struct {
	ChatUUID   uuid.UUID
	Message    []byte
	ExcludeUID uuid.UUID
}

func NewHub(
	userService *service.UserService,
	chatService *service.ChatService,
	messageService *service.MessageService,
) *Hub {
	return &Hub{
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		Broadcast:      make(chan *BroadcastMessage, 256),
		clients:        make(map[uuid.UUID]*Client),
		userService:    userService,
		chatService:    chatService,
		messageService: messageService,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client.UserUUID] = client
			h.mu.Unlock()
			log.Printf("Client registered: %s", client.UserUUID)

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserUUID]; ok {
				delete(h.clients, client.UserUUID)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("Client unregistered: %s", client.UserUUID)

		case message := <-h.Broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				if client.UserUUID != message.ExcludeUID {
					select {
					case client.Send <- message.Message:
					default:
						close(client.Send)
						delete(h.clients, client.UserUUID)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// HandleMessage - экспортируемый метод
func (h *Hub) HandleMessage(client *Client, msg *ClientMessage) {
	ctx := context.Background()

	switch msg.Type {
	case "chat.create.personal":
		var payload struct {
			ReceiverUUID string `json:"receiver_uuid"`
			ReceiverName string `json:"receiver_name"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			h.sendError(client, msg.RequestID, "Invalid payload")
			return
		}

		ctx := context.Background()
		var receiverUUID uuid.UUID
		var err error

		if payload.ReceiverName != "" {
			// Ищем по username
			user, err := h.userService.GetUserByUsername(ctx, payload.ReceiverName)
			if err != nil || user == nil {
				h.sendError(client, msg.RequestID, "User not found: "+payload.ReceiverName)
				return
			}
			receiverUUID = user.UUID
		} else if payload.ReceiverUUID != "" {
			// Ищем по UUID
			receiverUUID, err = uuid.Parse(payload.ReceiverUUID)
			if err != nil {
				h.sendError(client, msg.RequestID, "Invalid receiver UUID")
				return
			}
		} else {
			h.sendError(client, msg.RequestID, "receiver_name or receiver_uuid required")
			return
		}

		chat, err := h.chatService.CreatePersonalChat(ctx, client.UserUUID, receiverUUID)
		if err != nil {
			h.sendError(client, msg.RequestID, err.Error())
			return
		}

		h.sendResponse(client, msg.RequestID, map[string]interface{}{
			"chat_uuid": chat.UUID.String(),
			"chat_type": chat.Type,
		})

	case "message.send":
		var payload struct {
			ChatUUID string `json:"chat_uuid"`
			Body     string `json:"body"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			h.sendError(client, msg.RequestID, "Invalid payload")
			return
		}

		chatUUID, err := uuid.Parse(payload.ChatUUID)
		if err != nil {
			h.sendError(client, msg.RequestID, "Invalid chat UUID")
			return
		}

		// Отправляем сообщение
		message, err := h.messageService.SendMessage(ctx, chatUUID, client.UserUUID, payload.Body)
		if err != nil {
			h.sendError(client, msg.RequestID, err.Error())
			return
		}

		log.Printf("✅ Message sent: %s from %s", message.UUID, client.Username)

		// Ответ отправителю
		h.sendResponse(client, msg.RequestID, map[string]interface{}{
			"message_uuid": message.UUID.String(),
			"created_at":   message.CreatedAt.Format(time.RFC3339),
		})

		// Рассылка всем участникам
		h.broadcastToChat(ctx, chatUUID, message, client.UserUUID)

	case "chat.list":
		chats, err := h.chatService.GetUserChats(ctx, client.UserUUID)
		if err != nil {
			h.sendError(client, msg.RequestID, err.Error())
			return
		}
		h.sendResponse(client, msg.RequestID, chats)

	case "chat.get":
		var payload struct {
			ChatUUID string `json:"chat_uuid"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			h.sendError(client, msg.RequestID, "Invalid payload")
			return
		}

		chatUUID, err := uuid.Parse(payload.ChatUUID)
		if err != nil {
			h.sendError(client, msg.RequestID, "Invalid chat UUID")
			return
		}

		chatData, err := h.chatService.GetChatWithLastMessage(ctx, chatUUID, client.UserUUID)
		if err != nil {
			h.sendError(client, msg.RequestID, err.Error())
			return
		}
		h.sendResponse(client, msg.RequestID, chatData)

	case "message.history":
		var payload struct {
			ChatUUID string `json:"chat_uuid"`
			Limit    int    `json:"limit"`
			Offset   int    `json:"offset"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			h.sendError(client, msg.RequestID, "Invalid payload")
			return
		}

		chatUUID, err := uuid.Parse(payload.ChatUUID)
		if err != nil {
			h.sendError(client, msg.RequestID, "Invalid chat UUID")
			return
		}

		if payload.Limit == 0 {
			payload.Limit = 50
		}

		messages, err := h.messageService.GetChatMessages(ctx, chatUUID, payload.Limit, payload.Offset)
		if err != nil {
			h.sendError(client, msg.RequestID, err.Error())
			return
		}
		h.sendResponse(client, msg.RequestID, messages)

	case "user.update_last_seen":
		_ = h.userService.UpdateLastSeen(ctx, client.UserUUID)
	}
}

func (h *Hub) broadcastToChat(ctx context.Context, chatUUID uuid.UUID, message *entity.Message, senderUUID uuid.UUID) {
	// Получаем имя отправителя
	sender, err := h.userService.GetUser(ctx, senderUUID)
	senderUsername := ""
	if err == nil && sender != nil {
		senderUsername = sender.Username
	}

	log.Printf("📤 Broadcasting to chat %s: sender=%s (UUID=%s), body=%s",
		chatUUID, senderUsername, senderUUID, message.Body)

	data := map[string]interface{}{
		"type": "message.new",
		"payload": map[string]interface{}{
			"message_uuid":    message.UUID.String(),
			"chat_uuid":       message.ChatUUID.String(),
			"sender_uuid":     message.SenderUUID.String(),
			"sender_username": senderUsername, // ЭТО КЛЮЧЕВОЕ ПОЛЕ!
			"body":            message.Body,
			"created_at":      message.CreatedAt.Format(time.RFC3339),
		},
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("❌ Failed to marshal message: %v", err)
		return
	}

	h.Broadcast <- &BroadcastMessage{
		ChatUUID:   chatUUID,
		Message:    bytes,
		ExcludeUID: senderUUID,
	}
}

func (h *Hub) sendResponse(client *Client, requestID string, payload interface{}) {
	msg := ServerMessage{
		Type:      "response",
		RequestID: requestID,
		Payload:   payload,
	}
	bytes, _ := json.Marshal(msg)
	select {
	case client.Send <- bytes:
	default:
		log.Printf("Client %s send buffer full", client.UserUUID)
	}
}

func (h *Hub) sendError(client *Client, requestID string, errorMsg string) {
	msg := ServerMessage{
		Type:      "error",
		RequestID: requestID,
		Payload:   ErrorPayload{Error: errorMsg},
	}
	bytes, _ := json.Marshal(msg)
	select {
	case client.Send <- bytes:
	default:
		log.Printf("Client %s send buffer full", client.UserUUID)
	}
}
