package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// The frontend and backend are on different origins once deployed
	// (e.g. vercel.app talking to onrender.com), so the default
	// same-origin check would reject every real connection. We already
	// gate writes with JWT/ownership checks where it matters (poll
	// management), and this endpoint is read-only broadcast, so a
	// permissive origin check here is an intentional, scoped trade-off.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// PollSocket upgrades the connection and registers it with the Hub so
// it receives every subsequent vote broadcast for this poll, with no
// polling and no page refresh needed on the client.
func (a *App) PollSocket(c *gin.Context) {
	pollID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(pollID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll id"})
		return
	}

	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()
	count, err := a.Mongo.Polls.CountDocuments(ctx, bson.M{"_id": objID})
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "poll not found"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}

	a.Hub.Register(pollID, conn)
	defer a.Hub.Unregister(pollID, conn)

	// We don't expect meaningful messages from the client on this
	// socket, but we still need to read in a loop: it's how gorilla
	// detects the connection closing (browser tab closed, network
	// drop, etc.) so we can clean up the registration promptly.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
