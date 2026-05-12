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
func (hub *Hub) Run() {
	for {
		select {
		case client := <-hub.registerCh:
			hub.onRegister(client)
		case client := <-hub.unregisterCh:
			hub.onUnregister(client)
		case msg := <-hub.broadcastCh:
			hub.fanout(msg.Room, msg)
		}
	}
}

func (hub *Hub) onRegister(client Client) {
	room := client.Room()

	hub.mu.Lock()
	if _, exists := hub.rooms[room]; !exists {
		hub.rooms[room] = make(map[Client]struct{})
	}
	hub.rooms[room][client] = struct{}{}
	occupants := len(hub.rooms[room])
	users := hub.usernamesLocked(room)
	hub.mu.Unlock()

	log.Printf("[%s] %s se conectó (%d en sala)", room, client.Username(), occupants)
	hub.fanout(room, models.Message{
		Type:     models.TypeJoin,
		Username: client.Username(),
		Content:  client.Username() + " se unió al chat 👋",
		Room:     room,
		Users:    users,
	})
}

func (hub *Hub) onUnregister(client Client) {
	room := client.Room()

	hub.mu.Lock()
	members, roomExists := hub.rooms[room]
	if !roomExists {
		hub.mu.Unlock()
		return
	}
	if _, isMember := members[client]; !isMember {
		hub.mu.Unlock()
		return
	}
	delete(members, client)
	remaining := len(members)
	var users []string
	if remaining == 0 {
		delete(hub.rooms, room)
	} else {
		users = hub.usernamesLocked(room)
	}
	hub.mu.Unlock()

	client.Close()
	log.Printf("[%s] %s se desconectó (%d en sala)", room, client.Username(), remaining)

	if remaining > 0 {
		hub.fanout(room, models.Message{
			Type:     models.TypeLeave,
			Username: client.Username(),
			Content:  client.Username() + " salió del chat",
			Room:     room,
			Users:    users,
		})
	}
}

// fanout entrega un mensaje a todos los clientes de una sala.
// Toma un snapshot bajo lock corto, entrega fuera del lock para evitar
// que un cliente lento bloquee a los demás.
func (hub *Hub) fanout(room string, msg models.Message) {
	hub.mu.RLock()
	members, exists := hub.rooms[room]
	if !exists {
		hub.mu.RUnlock()
		return
	}
	snapshot := make([]Client, 0, len(members))
	for client := range members {
		snapshot = append(snapshot, client)
	}
	hub.mu.RUnlock()

	var stale []Client
	for _, client := range snapshot {
		if !client.Deliver(msg) {
			stale = append(stale, client)
		}
	}
	if len(stale) == 0 {
		return
	}

	hub.mu.Lock()
	if members, exists := hub.rooms[room]; exists {
		for _, client := range stale {
			delete(members, client)
		}
		if len(members) == 0 {
			delete(hub.rooms, room)
		}
	}
	hub.mu.Unlock()

	for _, client := range stale {
		client.Close()
	}
}

func (hub *Hub) usernamesLocked(room string) []string {
	members := hub.rooms[room]
	users := make([]string, 0, len(members))
	for client := range members {
		users = append(users, client.Username())
	}
	return users
}
