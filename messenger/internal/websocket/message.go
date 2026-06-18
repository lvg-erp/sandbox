package websocket

import "encoding/json"

type ClientMessage struct {
	Type      string          `json:"type"`
	RequestID string          `json:"request_id"`
	Payload   json.RawMessage `json:"payload"`
}

type ServerMessage struct {
	Type      string      `json:"type"`
	RequestID string      `json:"request_id,omitempty"`
	Payload   interface{} `json:"payload"`
}

type ErrorPayload struct {
	Error string `json:"error"`
}
