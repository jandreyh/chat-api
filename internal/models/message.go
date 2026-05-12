package models

import "time"

// MessageType discrimina el tipo de mensaje en el protocolo WebSocket.
type MessageType string

const (
	TypeChat  MessageType = "chat"
	TypeJoin  MessageType = "join"
	TypeLeave MessageType = "leave"
	TypeUsers MessageType = "users"
	TypeError MessageType = "error"
)

// Message es la unidad de datos que viaja por WebSocket en formato JSON.
type Message struct {
	Type      MessageType `json:"type"`
	Username  string      `json:"username"`
	Content   string      `json:"content"`
	Room      string      `json:"room"`
	Timestamp time.Time   `json:"timestamp"`
	Users     []string    `json:"users,omitempty"`
}
