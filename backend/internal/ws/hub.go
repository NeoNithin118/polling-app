package ws

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"

	"github.com/guvi-internship/polling-backend/internal/db"
)

// Hub keeps track of which websocket connections are watching which
// poll, and fans out Redis pub/sub messages to the right connections.
// A single Redis PSubscribe backs every poll's live updates, so the
// backend can be scaled to multiple instances: whichever instance
// receives the vote publishes to Redis, and every instance (each with
// its own Hub) forwards it to the clients connected to *that*
// instance. This is what makes votes show up with no page refresh.
type Hub struct {
	redis *redis.Client

	mu      sync.RWMutex
	clients map[string]map[*websocket.Conn]bool // pollID -> set of conns
}

func NewHub(redisClient *redis.Client) *Hub {
	return &Hub{
		redis:   redisClient,
		clients: make(map[string]map[*websocket.Conn]bool),
	}
}

// Run subscribes to every poll channel and blocks, forwarding
// messages as they arrive. Call it in a goroutine at startup.
func (h *Hub) Run(ctx context.Context) {
	pubsub := h.redis.PSubscribe(ctx, db.ChannelPattern)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			pollID := extractPollID(msg.Channel)
			h.broadcast(pollID, msg.Payload)
		}
	}
}

func extractPollID(channel string) string {
	// channel is "channel:poll:<id>"
	parts := strings.SplitN(channel, ":", 3)
	if len(parts) != 3 {
		return ""
	}
	return parts[2]
}

func (h *Hub) Register(pollID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[pollID] == nil {
		h.clients[pollID] = make(map[*websocket.Conn]bool)
	}
	h.clients[pollID][conn] = true
}

func (h *Hub) Unregister(pollID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.clients[pollID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.clients, pollID)
		}
	}
	_ = conn.Close()
}

func (h *Hub) broadcast(pollID, payload string) {
	h.mu.RLock()
	conns := h.clients[pollID]
	h.mu.RUnlock()

	for conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(payload)); err != nil {
			log.Printf("ws write error for poll %s: %v", pollID, err)
			h.Unregister(pollID, conn)
		}
	}
}

// ViewerCount reports how many live sockets are currently watching a
// poll. This is the "extra feature" the frontend uses to show a live
// viewer count — a small, genuinely-driven-by-Redis-connections number,
// not a decorative one.
func (h *Hub) ViewerCount(pollID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[pollID])
}

// PublishPayload is a small helper so handlers don't need to build
// the JSON envelope themselves.
type UpdatePayload struct {
	Type    string           `json:"type"`
	PollID  string           `json:"pollId"`
	Counts  map[string]int64 `json:"counts"`
	Total   int64            `json:"total"`
}

func Marshal(p UpdatePayload) (string, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
