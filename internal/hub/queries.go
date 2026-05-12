package hub

// RoomInfo describe una sala activa en una respuesta de API.
type RoomInfo struct {
	Name  string `json:"name"`
	Users int    `json:"users"`
}

// IsUsernameTaken indica si un username ya está en uso dentro de una sala.
// Safe para llamar desde cualquier goroutine.
func (h *Hub) IsUsernameTaken(room, username string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	members, ok := h.rooms[room]
	if !ok {
		return false
	}
	for c := range members {
		if c.Username() == username {
			return true
		}
	}
	return false
}

// ActiveRooms retorna un snapshot de las salas activas con su conteo.
// Safe para llamar desde cualquier goroutine.
func (h *Hub) ActiveRooms() []RoomInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()

	rooms := make([]RoomInfo, 0, len(h.rooms))
	for name, members := range h.rooms {
		rooms = append(rooms, RoomInfo{Name: name, Users: len(members)})
	}
	return rooms
}
