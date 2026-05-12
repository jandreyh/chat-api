package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config agrupa todos los parámetros configurables del servidor.
// Cargados desde variables de entorno con valores por defecto seguros.
type Config struct {
	Port                string
	StaticDir           string
	AllowedOrigins      []string
	ReadLimit           int64
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	PingInterval        time.Duration
	SendBufferSize      int
	BroadcastBufferSize int
	ShutdownTimeout     time.Duration
}

// Load lee la configuración desde el entorno aplicando valores por defecto.
func Load() *Config {
	return &Config{
		Port:                getString("PORT", "8080"),
		StaticDir:           getString("STATIC_DIR", "static"),
		AllowedOrigins:      parseCSV(getString("ALLOWED_ORIGINS", "*")),
		ReadLimit:           int64(getInt("WS_READ_LIMIT_BYTES", 512)),
		ReadTimeout:         getDuration("WS_READ_TIMEOUT", 60*time.Second),
		WriteTimeout:        getDuration("WS_WRITE_TIMEOUT", 10*time.Second),
		PingInterval:        getDuration("WS_PING_INTERVAL", 54*time.Second),
		SendBufferSize:      getInt("WS_SEND_BUFFER", 256),
		BroadcastBufferSize: getInt("HUB_BROADCAST_BUFFER", 256),
		ShutdownTimeout:     getDuration("SHUTDOWN_TIMEOUT", 5*time.Second),
	}
}

func getString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func parseCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
