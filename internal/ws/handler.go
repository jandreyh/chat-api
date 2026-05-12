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
func NewHandler(h *hub.Hub, cfg *config.Config) *Handler {
	return &Handler{
		hub: h,
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
	set := make(map[string]struct{}, len(allowed))
	for _, o := range allowed {
		set[o] = struct{}{}
	}
	return func(r *http.Request) bool {
		_, ok := set[r.Header.Get("Origin")]
		return ok
	}
}

// ServeWS atiende GET /ws?username=X&room=Y. Hace upgrade HTTP→WebSocket
// y registra al cliente en el hub.
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.URL.Query().Get("username"))
	room := strings.TrimSpace(r.URL.Query().Get("room"))

	if username == "" {
		http.Error(w, "username requerido", http.StatusBadRequest)
		return
	}
	if room == "" {
		room = "General"
	}

	if h.hub.IsUsernameTaken(room, username) {
		http.Error(w, "username ya en uso en esta sala", http.StatusConflict)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws: upgrade fallido: %v", err)
		return
	}

	client := NewClient(ClientOptions{
		Hub:          h.hub,
		Conn:         conn,
		Username:     username,
		Room:         room,
		SendBuffer:   h.cfg.SendBufferSize,
		ReadLimit:    h.cfg.ReadLimit,
		ReadTimeout:  h.cfg.ReadTimeout,
		WriteTimeout: h.cfg.WriteTimeout,
		PingInterval: h.cfg.PingInterval,
	})

	h.hub.Register(client)
	client.Start()
}

// Rooms atiende GET /api/rooms y devuelve las salas activas.
func (h *Handler) Rooms(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"rooms": h.hub.ActiveRooms(),
	})
}
