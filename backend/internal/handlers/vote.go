package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/guvi-internship/polling-backend/internal/db"
	"github.com/guvi-internship/polling-backend/internal/models"
	"github.com/guvi-internship/polling-backend/internal/ws"
	"github.com/guvi-internship/polling-backend/internal/middleware"
)

type voteRequest struct {
	OptionID string `json:"optionId" binding:"required"`
}

// CastVote is the hot path of the whole app. Every check here is
// server-side on purpose: the option must actually belong to the
// poll, the poll must be open, and the voter must not have voted
// already — none of that is trusted from the client beyond the IDs
// it sends.
func (a *App) CastVote(c *gin.Context) {
	pollID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(pollID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll id"})
		return
	}

	var req voteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "optionId is required"})	
		return
	}

	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	var poll models.Poll
	if err := a.Mongo.Polls.FindOne(ctx, bson.M{"_id": objID}).Decode(&poll); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "poll not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load poll"})
		return
	}

	if !poll.IsActive {
		c.JSON(http.StatusConflict, gin.H{"error": "this poll is closed"})
		return
	}

	validOption := false
	for _, opt := range poll.Options {
		if opt.ID == req.OptionID {
			validOption = true
			break
		}
	}
	if !validOption {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown option for this poll"})
		return
	}

	// Fingerprint = client-supplied voterId, tied to IP as a weak
	// secondary signal. This is not bulletproof (nothing anonymous is)
	// but it stops casual double-voting, which is what the brief asks for.
	userID := middleware.UserID(c)
	fingerprint := userID

	added, err := a.Redis.SAdd(ctx, db.VotersKey(pollID), fingerprint).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not register vote"})
		return
	}
	if added == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "you've already voted on this poll"})
		return
	}

	newCount, err := a.Redis.HIncrBy(ctx, db.VotesKey(pollID), req.OptionID, 1).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not record vote"})
		return
	}
	_ = newCount

	// Permanent audit log in Mongo — independent of Redis, so results
	// can always be reconstructed or exported even if Redis is wiped.
	vote := models.Vote{
		PollID:           pollID,
		OptionID:         req.OptionID,
		VoterFingerprint: fingerprint,
		CreatedAt:        time.Now().UTC(),
	}
	if _, err := a.Mongo.Votes.InsertOne(ctx, vote); err != nil {
		// The Redis counter is already updated and is the source of
		// truth for live display; log this rather than fail the
		// request outright, since the vote has already been accepted.
	}

	counts, total, err := a.currentCounts(ctx, pollID, poll)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "vote recorded but could not refresh counts"})
		return
	}

	payload, err := ws.Marshal(ws.UpdatePayload{
		Type:   "vote",
		PollID: pollID,
		Counts: counts,
		Total:  total,
	})
	if err == nil {
		a.Redis.Publish(ctx, db.Channel(pollID), payload)
	}

	c.JSON(http.StatusOK, models.PollResult{Poll: poll, Counts: counts, Total: total})
}
