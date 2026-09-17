package handlers

import (
	"github.com/redis/go-redis/v9"

	"github.com/guvi-internship/polling-backend/internal/db"
	"github.com/guvi-internship/polling-backend/internal/ws"
)

// App carries every dependency a handler might need. Handlers are
// methods on *App instead of free functions with globals, which
// keeps the package testable and makes the wiring in main.go explicit.
type App struct {
	Mongo     *db.Mongo
	Redis     *redis.Client
	Hub       *ws.Hub
	JWTSecret string
	TokenTTLH int
}
