# 💬 GoChat — Chat en Tiempo Real con WebSocket

Chat en tiempo real construido con Go puro, WebSocket y una interfaz web moderna.
Sin frameworks externos — solo `net/http`, `goroutines`, `channels` y `gorilla/websocket`.

---

## 🚀 Cómo ejecutar

```bash
# 1. Instalar dependencias
go mod tidy

# 2. Ejecutar
go run cmd/main.go

# 3. Abrir en el navegador
open http://localhost:8080
```

Abre **varias pestañas** del navegador para simular múltiples usuarios chateando.

---

## 🗂️ Estructura del proyecto

```
chat-app/
├── cmd/
│   └── main.go                    ← Servidor, rutas, punto de entrada
├── internal/
│   ├── hub/
│   │   ├── hub.go                 ← Hub central (goroutine + channels)
│   │   ├── client.go              ← Cliente WebSocket (ReadPump + WritePump)
│   │   └── helpers.go             ← Métodos auxiliares del Hub
│   ├── handlers/
│   │   └── ws.go                  ← Upgrade HTTP→WebSocket
│   └── models/
│       └── models.go              ← Structs: Message, MessageType
└── static/
    └── index.html                 ← Frontend completo (HTML + CSS + JS)
```

---

## 📡 Endpoints

| Endpoint | Descripción |
|----------|-------------|
| `GET /` | Interfaz web del chat |
| `GET /ws?username=X&room=Y` | Conexión WebSocket |
| `GET /api/rooms` | Salas activas (JSON) |
| `GET /api/health` | Estado del servidor |

---

## 🧠 Conceptos de Go aplicados

### Goroutines
Cada cliente tiene **2 goroutines** corriendo concurrentemente:
```
Cliente conectado
├── go ReadPump()   ← escucha mensajes del navegador
└── go WritePump()  ← envía mensajes al navegador
```

### Channels
```go
Hub.Register   chan *Client        // cliente nuevo
Hub.Unregister chan *Client        // cliente se va
Hub.Broadcast  chan models.Message // mensaje para todos
Client.Send    chan models.Message // mensajes para este cliente
```

### Select
El Hub usa `select` para procesar múltiples canales:
```go
select {
case client := <-h.Register:   // nuevo cliente
case client := <-h.Unregister: // cliente se va
case msg    := <-h.Broadcast:  // distribuir mensaje
}
```

### Maps
```go
Rooms map[string]map[*Client]bool
// "General" → {cliente1: true, cliente2: true}
// "Go"      → {cliente3: true}
```

---

## 📅 Plan Semana a Semana (15 horas)

### Semana 1 — WebSocket y primer mensaje (3h)
- Qué es WebSocket vs HTTP (handshake, full-duplex)
- Instalar gorilla/websocket, hacer el upgrade
- Enviar y recibir primer mensaje JSON
- Entregable: servidor que hace eco de cada mensaje

### Semana 2 — Hub y múltiples clientes (3h)
- Struct Hub con channels Register/Unregister/Broadcast
- Lanzar goroutine `go h.Run()`
- Broadcast a todos los clientes conectados
- Entregable: varios usuarios chateando en tiempo real

### Semana 3 — Salas de chat (3h)
- Map de salas: `map[string]map[*Client]bool`
- Filtrar broadcast por sala
- Parámetros `?username=X&room=Y` en la URL WebSocket
- Entregable: salas independientes funcionando

### Semana 4 — Frontend y notificaciones (3h)
- Interfaz HTML+CSS+JS que consume el WebSocket
- Mensajes de sistema: join/leave
- Lista de usuarios en tiempo real
- Entregable: chat con interfaz visual completa

### Semana 5 — Pulido y demo final (3h)
- Manejo de errores y desconexiones inesperadas
- Ping/Pong para mantener conexiones vivas
- Variables de entorno, configuración
- Demo en vivo con múltiples usuarios

---

## 🔮 Ideas para extender (retos opcionales)

- [ ] Historial de mensajes con SQLite (aplicar lo de todo-api)
- [ ] Mensajes privados entre usuarios
- [ ] Indicador "escribiendo..."
- [ ] Emojis y reacciones
- [ ] Autenticación JWT (conectar con todo-api)
- [ ] Deploy en Railway o Render
- [ ] Notificaciones de sonido

---

## 🛠 Tecnologías

| Tecnología | Uso |
|------------|-----|
| Go 1.21+ | Lenguaje principal |
| `net/http` | Servidor HTTP |
| `gorilla/websocket` | Protocolo WebSocket |
| Goroutines + Channels | Concurrencia |
| HTML + CSS + JS vanilla | Frontend |
