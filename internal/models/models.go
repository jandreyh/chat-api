package models

import "time"

// MessageType define el tipo de mensaje
type MessageType string

const (
	TypeChat   MessageType = "chat"    // mensaje normal
	TypeJoin   MessageType = "join"    // usuario entró
	TypeLeave  MessageType = "leave"   // usuario salió
	TypeUsers  MessageType = "users"   // lista de usuarios
	TypeError  MessageType = "error"   // error
)

// Message es la estructura que viaja por WebSocket (JSON)
type Message struct {
	Type      MessageType `json:"type"`
	Username  string      `json:"username"`
	Content   string      `json:"content"`
	Room      string      `json:"room"`
	Timestamp time.Time   `json:"timestamp"`
	Users     []string    `json:"users,omitempty"`
}

// JoinRequest es lo que manda el cliente al conectarse
type JoinRequest struct {
	Username string `json:"username"`
	Room     string `json:"room"`
}
