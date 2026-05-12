package hub

import (
	"log"
	"sync"

	"github.com/jandreyh/chat-api/internal/models"
)

// Client es la interfaz que cualquier transporte debe cumplir para
// formar parte del hub. Desacopla al hub del WebSocket (Dependency Inversion).
type Client interface {
	Username() string
	Room() string
	Deliver(msg models.Message) bool
	Close()
}

// Hub coordina las salas y la distribución de mensajes.
// Aplica el patrón actor: las mutaciones se serializan en Run().
// Las consultas externas usan un RWMutex para evitar data races.
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[Client]struct{}

	registerCh   chan Client
	unregisterCh chan Client
	broadcastCh  chan models.Message
}

// New construye un hub con buffers configurables.
func New(broadcastBuffer int) *Hub {
	return &Hub{
		rooms:        make(map[string]map[Client]struct{}),
		registerCh:   make(chan Client),
		unregisterCh: make(chan Client),
		broadcastCh:  make(chan models.Message, broadcastBuffer),
	}
}

// Run ejecuta el loop principal del hub. Debe correr en su propia goroutine.
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.registerCh:
			h.onRegister(c)
		case c := <-h.unregisterCh:
			h.onUnregister(c)
		case msg := <-h.broadcastCh:
			h.fanout(msg.Room, msg)
		}
	}
}

func (h *Hub) onRegister(c Client) {
	room := c.Room()

	h.mu.Lock()
	if _, ok := h.rooms[room]; !ok {
		h.rooms[room] = make(map[Client]struct{})
	}
	h.rooms[room][c] = struct{}{}
	occupants := len(h.rooms[room])
	users := h.usernamesLocked(room)
	h.mu.Unlock()

	log.Printf("[%s] %s se conectó (%d en sala)", room, c.Username(), occupants)
	h.fanout(room, models.Message{
		Type:     models.TypeJoin,
		Username: c.Username(),
		Content:  c.Username() + " se unió al chat 👋",
		Room:     room,
		Users:    users,
	})
}

func (h *Hub) onUnregister(c Client) {
	room := c.Room()

	h.mu.Lock()
	clients, ok := h.rooms[room]
	if !ok {
		h.mu.Unlock()
		return
	}
	if _, ok := clients[c]; !ok {
		h.mu.Unlock()
		return
	}
	delete(clients, c)
	remaining := len(clients)
	var users []string
	if remaining == 0 {
		delete(h.rooms, room)
	} else {
		users = h.usernamesLocked(room)
	}
	h.mu.Unlock()

	c.Close()
	log.Printf("[%s] %s se desconectó (%d en sala)", room, c.Username(), remaining)

	if remaining > 0 {
		h.fanout(room, models.Message{
			Type:     models.TypeLeave,
			Username: c.Username(),
			Content:  c.Username() + " salió del chat",
			Room:     room,
			Users:    users,
		})
	}
}

// fanout entrega un mensaje a todos los clientes de una sala.
// Toma un snapshot bajo lock corto, entrega fuera del lock para evitar
// que un cliente lento bloquee a los demás.
func (h *Hub) fanout(room string, msg models.Message) {
	h.mu.RLock()
	members, ok := h.rooms[room]
	if !ok {
		h.mu.RUnlock()
		return
	}
	snapshot := make([]Client, 0, len(members))
	for c := range members {
		snapshot = append(snapshot, c)
	}
	h.mu.RUnlock()

	var stale []Client
	for _, c := range snapshot {
		if !c.Deliver(msg) {
			stale = append(stale, c)
		}
	}
	if len(stale) == 0 {
		return
	}

	h.mu.Lock()
	if members, ok := h.rooms[room]; ok {
		for _, c := range stale {
			delete(members, c)
		}
		if len(members) == 0 {
			delete(h.rooms, room)
		}
	}
	h.mu.Unlock()

	for _, c := range stale {
		c.Close()
	}
}

func (h *Hub) usernamesLocked(room string) []string {
	members := h.rooms[room]
	users := make([]string, 0, len(members))
	for c := range members {
		users = append(users, c.Username())
	}
	return users
}
