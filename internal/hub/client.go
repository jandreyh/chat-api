package hub

import (
	"encoding/json"
	"log"
	"time"

	"chat-app/internal/models"

	"github.com/gorilla/websocket"
)

// Client representa una conexión WebSocket activa
type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan models.Message // canal para enviar mensajes a este cliente
	Username string
	Room     string
}

// ReadPump escucha mensajes entrantes del navegador
// Corre en su propia goroutine por cada cliente
func (c *Client) ReadPump() {
	defer func() {
		// Al desconectarse, avisar al hub
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	// Límites de seguridad
	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, rawMsg, err := c.Conn.ReadMessage()
		if err != nil {
			// El cliente cerró la conexión — salir del loop
			break
		}

		// Deserializar el JSON que mandó el navegador
		var msg models.Message
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			log.Printf("Error al leer mensaje: %v", err)
			continue
		}

		// Completar los campos del mensaje
		msg.Username = c.Username
		msg.Room = c.Room
		msg.Timestamp = time.Now()
		msg.Type = models.TypeChat

		// Enviar al hub para que lo distribuya a todos en la sala
		c.Hub.Broadcast <- msg
	}
}

// WritePump envía mensajes desde el canal Send al navegador
// Corre en su propia goroutine por cada cliente
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// El hub cerró el canal
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			// Serializar y enviar como JSON
			if err := c.Conn.WriteJSON(msg); err != nil {
				return
			}

		case <-ticker.C:
			// Ping periódico para mantener la conexión viva
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
