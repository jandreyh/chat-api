package hub

import "github.com/jandreyh/chat-api/internal/models"

// Register encola un nuevo cliente para que el hub lo añada a su sala.
func (h *Hub) Register(c Client) { h.registerCh <- c }

// Unregister encola la baja de un cliente.
func (h *Hub) Unregister(c Client) { h.unregisterCh <- c }

// Publish encola un mensaje para distribuir a todos los clientes de su sala.
func (h *Hub) Publish(msg models.Message) { h.broadcastCh <- msg }
