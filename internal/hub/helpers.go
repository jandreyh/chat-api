package hub

// IsUsernameTaken verifica si un username ya está en uso en una sala
func (h *Hub) IsUsernameTaken(room, username string) bool {
	clients, ok := h.Rooms[room]
	if !ok {
		return false
	}
	for client := range clients {
		if client.Username == username {
			return true
		}
	}
	return false
}

// ActiveRooms retorna info de todas las salas activas
func (h *Hub) ActiveRooms() []map[string]interface{} {
	var rooms []map[string]interface{}
	for name, clients := range h.Rooms {
		rooms = append(rooms, map[string]interface{}{
			"name":  name,
			"users": len(clients),
		})
	}
	return rooms
}
