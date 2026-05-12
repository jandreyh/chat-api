package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/jandreyh/chat-api/internal/hub"
	"github.com/jandreyh/chat-api/internal/models"
)

// Client representa una conexión WebSocket activa. Implementa hub.Client.
type Client struct {
	hub          *hub.Hub
	conn         *websocket.Conn
	sendCh       chan models.Message
	username     string
	room         string
	readLimit    int64
	readTimeout  time.Duration
	writeTimeout time.Duration
	pingInterval time.Duration
	closeOnce    sync.Once
}

// ClientOptions agrupa los parámetros para construir un Client.
type ClientOptions struct {
	Hub          *hub.Hub
	Conn         *websocket.Conn
	Username     string
	Room         string
	SendBuffer   int
	ReadLimit    int64
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PingInterval time.Duration
}

// NewClient construye un Client listo para registrar en el hub.
func NewClient(opts ClientOptions) *Client {
	return &Client{
		hub:          opts.Hub,
		conn:         opts.Conn,
		sendCh:       make(chan models.Message, opts.SendBuffer),
		username:     opts.Username,
		room:         opts.Room,
		readLimit:    opts.ReadLimit,
		readTimeout:  opts.ReadTimeout,
		writeTimeout: opts.WriteTimeout,
		pingInterval: opts.PingInterval,
	}
}

func (c *Client) Username() string { return c.username }
func (c *Client) Room() string     { return c.room }

// Deliver intenta encolar un mensaje. Devuelve false si el buffer está lleno
// (cliente lento) para que el hub pueda desconectarlo.
func (c *Client) Deliver(msg models.Message) bool {
	select {
	case c.sendCh <- msg:
		return true
	default:
		return false
	}
}

// Close cierra el canal de envío de forma idempotente.
func (c *Client) Close() {
	c.closeOnce.Do(func() { close(c.sendCh) })
}

// Start lanza las dos goroutines de lectura y escritura.
func (c *Client) Start() {
	go c.writeLoop()
	go c.readLoop()
}

func (c *Client) readLoop() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(c.readLimit)
	_ = c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		var msg models.Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Printf("ws: error al decodificar mensaje: %v", err)
			continue
		}

		// Sobrescribimos campos controlados por el servidor para evitar spoofing.
		msg.Username = c.username
		msg.Room = c.room
		msg.Timestamp = time.Now()
		msg.Type = models.TypeChat

		c.hub.Publish(msg)
	}
}

func (c *Client) writeLoop() {
	ticker := time.NewTicker(c.pingInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.sendCh:
			_ = c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(msg); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
