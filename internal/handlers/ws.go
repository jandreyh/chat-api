package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"chat-app/internal/hub"
	"chat-app/internal/models"

	"github.com/gorilla/websocket"
)

// upgrader convierte una conexión HTTP en WebSocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// CheckOrigin permite conexiones desde cualquier origen (desarrollo)
	// En producción debes validar el origen
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ChatHandler gestiona las conexiones WebSocket
type ChatHandler struct {
	Hub *hub.Hub
}

func NewChatHandler(h *hub.Hub) *ChatHandler {
	return &ChatHandler{Hub: h}
}

// ServeWS maneja el upgrade HTTP → WebSocket
// GET /ws?username=Edwin&room=Go
func (h *ChatHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Leer parámetros de la URL
	username := strings.TrimSpace(r.URL.Query().Get("username"))
	room := strings.TrimSpace(r.URL.Query().Get("room"))

	// Validaciones
	if username == "" {
		http.Error(w, "username requerido", http.StatusBadRequest)
		return
	}
	if room == "" {
		room = "General" // sala por defecto
	}

	// Verificar que el username no esté en uso en esa sala
	if h.Hub.IsUsernameTaken(room, username) {
		http.Error(w, "username ya en uso en esta sala", http.StatusConflict)
		return
	}

	// Hacer el upgrade de HTTP a WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error al hacer upgrade WebSocket: %v", err)
		return
	}

	// Crear el cliente
	client := &hub.Client{
		Hub:      h.Hub,
		Conn:     conn,
		Send:     make(chan models.Message, 256),
		Username: username,
		Room:     room,
	}

	// Registrar en el hub
	h.Hub.Register <- client

	// Lanzar las goroutines de lectura y escritura
	// Cada cliente tiene 2 goroutines corriendo concurrentemente
	go client.WritePump()
	go client.ReadPump()
}

// RoomsHandler retorna las salas activas
func (h *ChatHandler) RoomsHandler(w http.ResponseWriter, r *http.Request) {
	rooms := h.Hub.ActiveRooms()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rooms": rooms,
	})
}
