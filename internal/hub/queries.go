package hub

// RoomInfo describe una sala activa en una respuesta de API.
type RoomInfo struct {
	Name  string `json:"name"`
	Users int    `json:"users"`
}

// IsUsernameTaken indica si un username ya está en uso dentro de una sala.
// Safe para llamar desde cualquier goroutine.
func (hub *Hub) IsUsernameTaken(room, username string) bool {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	members, exists := hub.rooms[room]
	if !exists {
		return false
	}
	for client := range members {
		if client.Username() == username {
			return true
		}
	}
	return false
}

// ActiveRooms retorna un snapshot de las salas activas con su conteo.
// Safe para llamar desde cualquier goroutine.
func (hub *Hub) ActiveRooms() []RoomInfo {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	rooms := make([]RoomInfo, 0, len(hub.rooms))
	for name, members := range hub.rooms {
		rooms = append(rooms, RoomInfo{Name: name, Users: len(members)})
	}
	return rooms
}
