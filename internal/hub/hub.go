package hub

import (
	"log"

	"chat-app/internal/models"
)

// Hub es el corazón del chat: gestiona clientes y distribuye mensajes
// Todo pasa por canales — esto es Go concurrente en acción
type Hub struct {
	// Clientes activos agrupados por sala: map[sala]map[cliente]bool
	Rooms map[string]map[*Client]bool

	// Canales de comunicación (la magia de Go)
	Register   chan *Client        // cliente nuevo se conecta
	Unregister chan *Client        // cliente se desconecta
	Broadcast  chan models.Message // mensaje para distribuir
}

// NewHub crea e inicializa el hub
func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[string]map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan models.Message, 256),
	}
}

// Run es el loop principal del hub — corre en una goroutine dedicada
// Procesa eventos de forma secuencial y segura (sin mutex necesario)
func (h *Hub) Run() {
	for {
		select {

		// ── Nuevo cliente ────────────────────────────────────────────
		case client := <-h.Register:
			// Crear la sala si no existe
			if _, ok := h.Rooms[client.Room]; !ok {
				h.Rooms[client.Room] = make(map[*Client]bool)
			}
			h.Rooms[client.Room][client] = true

			log.Printf("✅ [%s] %s se conectó (%d en sala)",
				client.Room, client.Username, len(h.Rooms[client.Room]))

			// Notificar a todos que alguien llegó
			h.broadcastToRoom(client.Room, models.Message{
				Type:     models.TypeJoin,
				Username: client.Username,
				Content:  client.Username + " se unió al chat 👋",
				Room:     client.Room,
				Users:    h.usersInRoom(client.Room),
			})

		// ── Cliente se desconecta ─────────────────────────────────────
		case client := <-h.Unregister:
			if room, ok := h.Rooms[client.Room]; ok {
				if _, ok := room[client]; ok {
					delete(room, client)
					close(client.Send)

					log.Printf("❌ [%s] %s se desconectó (%d en sala)",
						client.Room, client.Username, len(h.Rooms[client.Room]))

					// Notificar a todos que alguien salió
					if len(room) > 0 {
						h.broadcastToRoom(client.Room, models.Message{
							Type:     models.TypeLeave,
							Username: client.Username,
							Content:  client.Username + " salió del chat",
							Room:     client.Room,
							Users:    h.usersInRoom(client.Room),
						})
					}
					// Limpiar sala vacía
					if len(room) == 0 {
						delete(h.Rooms, client.Room)
					}
				}
			}

		// ── Mensaje para distribuir ───────────────────────────────────
		case msg := <-h.Broadcast:
			h.broadcastToRoom(msg.Room, msg)
		}
	}
}

// broadcastToRoom envía un mensaje a todos los clientes de una sala
func (h *Hub) broadcastToRoom(room string, msg models.Message) {
	clients, ok := h.Rooms[room]
	if !ok {
		return
	}
	for client := range clients {
		select {
		case client.Send <- msg:
		default:
			// Si el canal está lleno, desconectar al cliente
			close(client.Send)
			delete(clients, client)
		}
	}
}

// usersInRoom retorna los nombres de todos en una sala
func (h *Hub) usersInRoom(room string) []string {
	var users []string
	for client := range h.Rooms[room] {
		users = append(users, client.Username)
	}
	return users
}
