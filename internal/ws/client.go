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
func NewClient(options ClientOptions) *Client {
	return &Client{
		hub:          options.Hub,
		conn:         options.Conn,
		sendCh:       make(chan models.Message, options.SendBuffer),
		username:     options.Username,
		room:         options.Room,
		readLimit:    options.ReadLimit,
		readTimeout:  options.ReadTimeout,
		writeTimeout: options.WriteTimeout,
		pingInterval: options.PingInterval,
	}
}

func (client *Client) Username() string { return client.username }
func (client *Client) Room() string     { return client.room }

// Deliver intenta encolar un mensaje. Devuelve false si el buffer está lleno
// (cliente lento) para que el hub pueda desconectarlo.
func (client *Client) Deliver(msg models.Message) bool {
	select {
	case client.sendCh <- msg:
		return true
	default:
		return false
	}
}

// Close cierra el canal de envío de forma idempotente.
func (client *Client) Close() {
	client.closeOnce.Do(func() { close(client.sendCh) })
}

// Start lanza las dos goroutines de lectura y escritura.
func (client *Client) Start() {
	go client.writeLoop()
	go client.readLoop()
}

func (client *Client) readLoop() {
	defer func() {
		client.hub.Unregister(client)
		client.conn.Close()
	}()

	client.conn.SetReadLimit(client.readLimit)
	_ = client.conn.SetReadDeadline(time.Now().Add(client.readTimeout))
	client.conn.SetPongHandler(func(string) error {
		return client.conn.SetReadDeadline(time.Now().Add(client.readTimeout))
	})

	for {
		_, raw, err := client.conn.ReadMessage()
		if err != nil {
			return
		}

		var msg models.Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Printf("ws: error al decodificar mensaje: %v", err)
			continue
		}

		// Sobrescribimos campos controlados por el servidor para evitar spoofing.
		msg.Username = client.username
		msg.Room = client.room
		msg.Timestamp = time.Now()
		msg.Type = models.TypeChat

		client.hub.Publish(msg)
	}
}

func (client *Client) writeLoop() {
	ticker := time.NewTicker(client.pingInterval)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-client.sendCh:
			_ = client.conn.SetWriteDeadline(time.Now().Add(client.writeTimeout))
			if !ok {
				_ = client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := client.conn.WriteJSON(msg); err != nil {
				return
			}

		case <-ticker.C:
			_ = client.conn.SetWriteDeadline(time.Now().Add(client.writeTimeout))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
