package websocket

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"messanger/internal/domain/entity"
	"messanger/internal/domain/service"

	"github.com/google/uuid"
)

// Hub управляет всеми WS-соединениями. Карта clients доступна только из
// горутины Run (владелец), поэтому мьютекс не нужен.
type Hub struct {
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *BroadcastMessage

	clients        map[uuid.UUID]map[*Client]struct{}
	userService    *service.UserService
	chatService    *service.ChatService
	messageService *service.MessageService
}

type BroadcastMessage struct {
	ChatUUID   uuid.UUID
	Recipients map[uuid.UUID]struct{}
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
		clients:        make(map[uuid.UUID]map[*Client]struct{}),
		userService:    userService,
		chatService:    chatService,
		messageService: messageService,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.register(client)

		case client := <-h.Unregister:
			h.unregister(client)

		case message := <-h.Broadcast:
			h.deliver(message)
		}
	}
}

func (h *Hub) register(client *Client) {
	set, ok := h.clients[client.UserUUID]
	if !ok {
		set = make(map[*Client]struct{})
		h.clients[client.UserUUID] = set
	}
	set[client] = struct{}{}
	log.Printf("Client registered: %s (%d connections)", client.UserUUID, len(set))
}

// unregister убирает конкретное соединение; другие соединения того же
// пользователя (вторая вкладка, телефон) продолжают работать.
func (h *Hub) unregister(client *Client) {
	set, ok := h.clients[client.UserUUID]
	if !ok {
		return
	}
	if _, ok := set[client]; !ok {
		return
	}
	delete(set, client)
	if len(set) == 0 {
		delete(h.clients, client.UserUUID)
	}
	close(client.Done)
	log.Printf("Client unregistered: %s", client.UserUUID)
}

func (h *Hub) deliver(message *BroadcastMessage) {
	for userUUID := range message.Recipients {
		for client := range h.clients[userUUID] {
			if client.UserUUID == message.ExcludeUID {
				continue
			}
			select {
			case client.Send <- message.Message:
			default:
				// Клиент не успевает забирать сообщения — отключаем его
				h.unregister(client)
			}
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

		// Рассылка только участникам чата
		h.broadcastToChat(ctx, chatUUID, message, client.UserUUID)

	case "chat.list":
		chats, err := h.chatService.GetUserChats(ctx, client.UserUUID)
		if err != nil {
			h.sendError(client, msg.RequestID, err.Error())
			return
		}
		if chats == nil {
			chats = []*entity.Chat{}
		}
		// Для личных чатов фронтенд показывает собеседника, поэтому прикладываем участников
		result := make([]map[string]interface{}, 0, len(chats))
		for _, c := range chats {
			participants, err := h.chatService.GetParticipants(ctx, c.UUID)
			if err != nil {
				participants = nil
			}
			result = append(result, map[string]interface{}{
				"uuid":         c.UUID.String(),
				"name":         c.Name,
				"type":         c.Type,
				"created_at":   c.CreatedAt.Format(time.RFC3339),
				"updated_at":   c.UpdatedAt.Format(time.RFC3339),
				"participants": participants,
			})
		}
		h.sendResponse(client, msg.RequestID, result)

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

		messages, err := h.messageService.GetChatMessages(ctx, chatUUID, client.UserUUID, payload.Limit, payload.Offset)
		if err != nil {
			h.sendError(client, msg.RequestID, err.Error())
			return
		}
		if messages == nil {
			messages = []*entity.Message{}
		}
		h.sendResponse(client, msg.RequestID, messages)

	case "message.read":
		// Прочитанность сообщений пока не реализована

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

	// Рассылаем только участникам чата
	participants, err := h.chatService.GetParticipants(ctx, chatUUID)
	if err != nil {
		log.Printf("❌ Failed to get participants of chat %s: %v", chatUUID, err)
		return
	}
	recipients := make(map[uuid.UUID]struct{}, len(participants))
	for _, p := range participants {
		recipients[p.UUID] = struct{}{}
	}

	data := map[string]interface{}{
		"type": "message.new",
		"payload": map[string]interface{}{
			"message_uuid":    message.UUID.String(),
			"chat_uuid":       message.ChatUUID.String(),
			"sender_uuid":     message.SenderUUID.String(),
			"sender_username": senderUsername,
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
		Recipients: recipients,
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
