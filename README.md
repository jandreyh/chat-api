# 💬 GoChat — Chat en Tiempo Real con WebSocket

Chat en tiempo real construido con Go puro, WebSocket y una interfaz web moderna.
Sin frameworks externos — solo `net/http`, `goroutines`, `channels` y `gorilla/websocket`.

---

## 🚀 Cómo ejecutar

```bash
# 1. Instalar dependencias
go mod tidy

# 2. Ejecutar
go run ./cmd/chat-api

# 3. Abrir en el navegador
open http://localhost:8080
```

Abre **varias pestañas** del navegador para simular múltiples usuarios chateando.

### Compilar binario

```bash
go build -o bin/chat-api ./cmd/chat-api
./bin/chat-api
```

### Exponer públicamente con Cloudflare Tunnel

```bash
cloudflared tunnel --url http://localhost:8080
```

---

## ⚙️ Configuración (variables de entorno)

Todos los parámetros tienen un valor por defecto razonable. Sobreescríbelos vía env vars:

| Variable | Default | Descripción |
|----------|---------|-------------|
| `PORT` | `8080` | Puerto HTTP |
| `STATIC_DIR` | `static` | Directorio del frontend |
| `ALLOWED_ORIGINS` | `*` | Orígenes permitidos para el WebSocket (CSV o `*`) |
| `WS_READ_LIMIT_BYTES` | `512` | Tamaño máximo de mensaje entrante |
| `WS_READ_TIMEOUT` | `60s` | Timeout de lectura (sin pong = desconexión) |
| `WS_WRITE_TIMEOUT` | `10s` | Timeout de escritura |
| `WS_PING_INTERVAL` | `54s` | Intervalo entre pings (keepalive) |
| `WS_SEND_BUFFER` | `256` | Buffer del canal Send por cliente |
| `HUB_BROADCAST_BUFFER` | `256` | Buffer del canal Broadcast del hub |
| `SHUTDOWN_TIMEOUT` | `5s` | Tiempo para drenar conexiones al cerrar |

Ejemplo en producción:

```bash
PORT=3000 \
ALLOWED_ORIGINS="https://midominio.com,https://www.midominio.com" \
./bin/chat-api
```

---

## 🗂️ Estructura del proyecto

```
chat-api/
├── cmd/chat-api/
│   └── main.go                    ← Punto de entrada (config + señales)
├── internal/
│   ├── config/
│   │   └── config.go              ← Carga env vars con defaults
│   ├── server/
│   │   └── server.go              ← http.Server + shutdown graceful
│   ├── hub/
│   │   ├── hub.go                 ← Loop principal + Run() + interfaz Client
│   │   ├── commands.go            ← Register / Unregister / Publish
│   │   └── queries.go             ← IsUsernameTaken / ActiveRooms (RLock)
│   ├── ws/
│   │   ├── client.go              ← Implementación WS de hub.Client (Read/Write loops)
│   │   └── handler.go             ← Upgrade HTTP→WebSocket + endpoint /api/rooms
│   └── models/
│       └── message.go             ← Struct Message + MessageType
└── static/
    ├── index.html                 ← Markup
    ├── css/styles.css             ← Estilos
    └── js/app.js                  ← Lógica del cliente
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
Encapsulados dentro del hub y expuestos vía métodos (API limpia):
```go
hub.registerCh   chan Client          // cliente nuevo
hub.unregisterCh chan Client          // cliente se va
hub.broadcastCh  chan models.Message  // mensaje para distribuir
client.sendCh    chan models.Message  // mensajes para este cliente

// API pública en commands.go:
hub.Register(c)     // → registerCh <- c
hub.Unregister(c)   // → unregisterCh <- c
hub.Publish(msg)    // → broadcastCh <- msg
```

### Select
El Hub multiplexa eventos con `select`:
```go
select {
case c   := <-h.registerCh:   // nuevo cliente
case c   := <-h.unregisterCh: // cliente se va
case msg := <-h.broadcastCh:  // distribuir mensaje
}
```

### Maps
```go
rooms map[string]map[Client]struct{}
// "General" → {cliente1: {}, cliente2: {}}
// "Go"      → {cliente3: {}}
```

### Interfaces (Dependency Inversion)
El hub no conoce WebSocket — depende de la interfaz `Client`:
```go
type Client interface {
    Username() string
    Room() string
    Deliver(msg models.Message) bool
    Close()
}
```
Cualquier transporte futuro (gRPC, SSE) puede plugearse sin tocar el hub.

### Concurrencia: channels + sync.RWMutex
- Las **mutaciones** se serializan vía channels en `Run()` (patrón actor).
- Las **consultas externas** (`IsUsernameTaken`, `ActiveRooms`) usan `RLock`/`RUnlock` para evitar data races.
- `Client.Close()` usa `sync.Once` para que el cierre del canal sea idempotente.

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
