package hub

import "github.com/jandreyh/chat-api/internal/models"

// Register encola un nuevo cliente para que el hub lo añada a su sala.
func (hub *Hub) Register(client Client) { hub.registerCh <- client }

// Unregister encola la baja de un cliente.
func (hub *Hub) Unregister(client Client) { hub.unregisterCh <- client }

// Publish encola un mensaje para distribuir a todos los clientes de su sala.
func (hub *Hub) Publish(msg models.Message) { hub.broadcastCh <- msg }
