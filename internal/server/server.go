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
	cfg        *config.Config
	hub        *hub.Hub
	httpServer *http.Server
}

// New ensambla todas las dependencias y devuelve un Server listo para Run.
func New(cfg *config.Config) *Server {
	chatHub := hub.New(cfg.BroadcastBufferSize)
	handler := ws.NewHandler(chatHub, cfg)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(cfg.StaticDir)))
	mux.HandleFunc("/ws", handler.ServeWS)
	mux.HandleFunc("/api/rooms", handler.Rooms)
	mux.HandleFunc("/api/health", health)

	return &Server{
		cfg: cfg,
		hub: chatHub,
		httpServer: &http.Server{
			Addr:              ":" + cfg.Port,
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}
}

// Run arranca el hub y el servidor HTTP. Cancela limpiamente cuando ctx termina.
func (server *Server) Run(ctx context.Context) error {
	go server.hub.Run()

	serveErr := make(chan error, 1)
	go func() {
		log.Printf("chat-api escuchando en :%s", server.cfg.Port)
		if err := server.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
		close(serveErr)
	}()

	select {
	case <-ctx.Done():
		log.Printf("chat-api: cierre solicitado, drenando conexiones (%s)", server.cfg.ShutdownTimeout)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), server.cfg.ShutdownTimeout)
		defer cancel()
		return server.httpServer.Shutdown(shutdownCtx)
	case err := <-serveErr:
		return err
	}
}

func health(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(writer).Encode(map[string]string{
		"status": "ok",
		"app":    "chat-api",
	})
}
