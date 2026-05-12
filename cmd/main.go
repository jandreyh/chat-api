package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"chat-app/internal/handlers"
	"chat-app/internal/hub"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ── Crear e iniciar el Hub ────────────────────────────────────────
	// El hub corre en su propia goroutine — gestiona todos los clientes
	h := hub.NewHub()
	go h.Run() // ← goroutine principal del chat

	// ── Handlers ─────────────────────────────────────────────────────
	chatHandler := handlers.NewChatHandler(h)

	// ── Rutas ────────────────────────────────────────────────────────
	mux := http.NewServeMux()

	// Servir el frontend (HTML/JS/CSS) desde la carpeta static/
	mux.Handle("/", http.FileServer(http.Dir("static")))

	// WebSocket endpoint — aquí ocurre la magia
	// ws://localhost:8080/ws?username=Edwin&room=Go
	mux.HandleFunc("/ws", chatHandler.ServeWS)

	// API REST auxiliar
	mux.HandleFunc("/api/rooms", chatHandler.RoomsHandler)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"status":"ok","app":"chat-app"}`)
	})

	// ── Iniciar servidor ──────────────────────────────────────────────
	fmt.Printf("🚀 Chat corriendo en http://localhost:%s\n", port)
	fmt.Println("📡 WebSocket en ws://localhost:" + port + "/ws")
	fmt.Println("─────────────────────────────────────────")
	fmt.Println("Conceptos de Go en este proyecto:")
	fmt.Println("  ✓ goroutines  — una por cliente (ReadPump + WritePump)")
	fmt.Println("  ✓ channels    — Register, Unregister, Broadcast, Send")
	fmt.Println("  ✓ select      — multiplexar canales en el Hub")
	fmt.Println("  ✓ maps        — salas y clientes")
	fmt.Println("  ✓ interfaces  — http.Handler")
	fmt.Println("─────────────────────────────────────────")

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
