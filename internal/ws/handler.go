package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"

	"github.com/jandreyh/chat-api/internal/config"
	"github.com/jandreyh/chat-api/internal/hub"
)

// Handler agrupa los endpoints WebSocket y REST relacionados con el chat.
type Handler struct {
	hub      *hub.Hub
	cfg      *config.Config
	upgrader websocket.Upgrader
}

// NewHandler construye un Handler con un upgrader configurado según la política
// de orígenes definida en la configuración.
func NewHandler(chatHub *hub.Hub, cfg *config.Config) *Handler {
	return &Handler{
		hub: chatHub,
		cfg: cfg,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     buildOriginChecker(cfg.AllowedOrigins),
		},
	}
}

// buildOriginChecker construye la función CheckOrigin a partir de una whitelist.
// "*" desactiva la validación (sólo recomendable en desarrollo).
func buildOriginChecker(allowed []string) func(*http.Request) bool {
	if len(allowed) == 1 && allowed[0] == "*" {
		return func(*http.Request) bool { return true }
	}
	whitelist := make(map[string]struct{}, len(allowed))
	for _, origin := range allowed {
		whitelist[origin] = struct{}{}
	}
	return func(request *http.Request) bool {
		_, allowed := whitelist[request.Header.Get("Origin")]
		return allowed
	}
}

// ServeWS atiende GET /ws?username=X&room=Y. Hace upgrade HTTP→WebSocket
// y registra al cliente en el hub.
func (handler *Handler) ServeWS(writer http.ResponseWriter, request *http.Request) {
	username := strings.TrimSpace(request.URL.Query().Get("username"))
	room := strings.TrimSpace(request.URL.Query().Get("room"))

	if username == "" {
		http.Error(writer, "username requerido", http.StatusBadRequest)
		return
	}
	if room == "" {
		room = "General"
	}

	if handler.hub.IsUsernameTaken(room, username) {
		http.Error(writer, "username ya en uso en esta sala", http.StatusConflict)
		return
	}

	conn, err := handler.upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Printf("ws: upgrade fallido: %v", err)
		return
	}

	client := NewClient(ClientOptions{
		Hub:          handler.hub,
		Conn:         conn,
		Username:     username,
		Room:         room,
		SendBuffer:   handler.cfg.SendBufferSize,
		ReadLimit:    handler.cfg.ReadLimit,
		ReadTimeout:  handler.cfg.ReadTimeout,
		WriteTimeout: handler.cfg.WriteTimeout,
		PingInterval: handler.cfg.PingInterval,
	})

	handler.hub.Register(client)
	client.Start()
}

// Rooms atiende GET /api/rooms y devuelve las salas activas.
func (handler *Handler) Rooms(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(writer).Encode(map[string]interface{}{
		"rooms": handler.hub.ActiveRooms(),
	})
}
