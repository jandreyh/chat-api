package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/jandreyh/chat-api/internal/config"
	"github.com/jandreyh/chat-api/internal/hub"
	"github.com/jandreyh/chat-api/internal/ws"
)

// Server compone la aplicación: hub + handlers + http.Server.
type Server struct {
	cfg  *config.Config
	hub  *hub.Hub
	http *http.Server
}

// New ensambla todas las dependencias y devuelve un Server listo para Run.
func New(cfg *config.Config) *Server {
	h := hub.New(cfg.BroadcastBufferSize)
	handler := ws.NewHandler(h, cfg)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(cfg.StaticDir)))
	mux.HandleFunc("/ws", handler.ServeWS)
	mux.HandleFunc("/api/rooms", handler.Rooms)
	mux.HandleFunc("/api/health", health)

	return &Server{
		cfg: cfg,
		hub: h,
		http: &http.Server{
			Addr:              ":" + cfg.Port,
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}
}

// Run arranca el hub y el servidor HTTP. Cancela limpiamente cuando ctx termina.
func (s *Server) Run(ctx context.Context) error {
	go s.hub.Run()

	serveErr := make(chan error, 1)
	go func() {
		log.Printf("chat-api escuchando en :%s", s.cfg.Port)
		if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
		close(serveErr)
	}()

	select {
	case <-ctx.Done():
		log.Printf("chat-api: cierre solicitado, drenando conexiones (%s)", s.cfg.ShutdownTimeout)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()
		return s.http.Shutdown(shutdownCtx)
	case err := <-serveErr:
		return err
	}
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"app":    "chat-api",
	})
}
