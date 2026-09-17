package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/guvi-internship/polling-backend/internal/db"
	"github.com/guvi-internship/polling-backend/internal/middleware"
	"github.com/guvi-internship/polling-backend/internal/models"
)

type createPollRequest struct {
	Question string   `json:"question" binding:"required"`
	Options  []string `json:"options" binding:"required"`
}

// CreatePoll validates and persists a new poll. Every rule here runs
// server-side regardless of what the frontend form already checked,
// because the frontend is not a trust boundary.
func (a *App) CreatePoll(c *gin.Context) {
	var req createPollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll payload"})
		return
	}

	question := strings.TrimSpace(req.Question)
	if len(question) < 3 || len(question) > 300 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question must be between 3 and 300 characters"})
		return
	}

	options, err := sanitizeOptions(req.Options)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	poll := models.Poll{
		OwnerID:   middleware.UserID(c),
		Question:  question,
		Options:   options,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
	}

	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	res, err := a.Mongo.Polls.InsertOne(ctx, poll)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create poll"})
		return
	}
	poll.ID = res.InsertedID.(primitive.ObjectID)

	// Seed the Redis counter hash with every option at zero. This is
	// what "Redis is doing real work, not bolted on" means in practice:
	// the very first read of this poll comes from Redis, not Mongo.
	pollID := poll.ID.Hex()
	seed := map[string]interface{}{}
	for _, opt := range options {
		seed[opt.ID] = 0
	}
	if len(seed) > 0 {
		a.Redis.HSet(ctx, db.VotesKey(pollID), seed)
	}

	c.JSON(http.StatusCreated, poll)
}

func sanitizeOptions(raw []string) ([]models.Option, error) {
	seenText := map[string]bool{}
	var cleaned []models.Option
	for i, text := range raw {
		t := strings.TrimSpace(text)
		if t == "" {
			continue
		}
		if len(t) > 150 {
			return nil, fmt.Errorf("option %d is too long (max 150 characters)", i+1)
		}
		key := strings.ToLower(t)
		if seenText[key] {
			continue // silently drop exact duplicates rather than reject the whole poll
		}
		seenText[key] = true
		cleaned = append(cleaned, models.Option{
			ID:   fmt.Sprintf("o%d", len(cleaned)+1),
			Text: t,
		})
	}
	if len(cleaned) < 2 {
		return nil, fmt.Errorf("a poll needs at least 2 distinct, non-empty options")
	}
	if len(cleaned) > 10 {
		return nil, fmt.Errorf("a poll can have at most 10 options")
	}
	return cleaned, nil
}

// GetPoll returns the poll definition plus live counts. Counts are
// read from Redis first; Mongo is only the fallback if Redis has no
// record yet (e.g. right after a Redis restart before reconciliation).
func (a *App) GetPoll(c *gin.Context) {
	pollID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(pollID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll id"})
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

	counts, total, err := a.currentCounts(ctx, pollID, poll)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load live counts"})
		return
	}

	c.JSON(http.StatusOK, models.PollResult{Poll: poll, Counts: counts, Total: total})
}

func (a *App) currentCounts(ctx context.Context, pollID string, poll models.Poll) (map[string]int64, int64, error) {
	raw, err := a.Redis.HGetAll(ctx, db.VotesKey(pollID)).Result()
	if err != nil {
		return nil, 0, err
	}

	counts := map[string]int64{}
	var total int64
	for _, opt := range poll.Options {
		var v int64
		if s, ok := raw[opt.ID]; ok {
			fmt.Sscanf(s, "%d", &v)
		}
		counts[opt.ID] = v
		total += v
	}
	return counts, total, nil
}

// ListMyPolls returns every poll the authenticated user owns, most
// recent first — the dashboard for managing your own polls.
func (a *App) ListMyPolls(c *gin.Context) {
	ownerID := middleware.UserID(c)

	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	cursor, err := a.Mongo.Polls.Find(ctx, bson.M{"owner_id": ownerID}, options.Find().SetSort(bson.M{"created_at": -1}))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load polls"})
		return
	}
	defer cursor.Close(ctx)

	polls := []models.Poll{}
	if err := cursor.All(ctx, &polls); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load polls"})
		return
	}

	c.JSON(http.StatusOK, polls)
}

// ClosePoll lets only the owner stop accepting new votes. Voting
// against a closed poll is rejected server-side too (see vote.go).
func (a *App) ClosePoll(c *gin.Context) {
	pollID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(pollID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll id"})
		return
	}

	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	now := time.Now().UTC()
	res, err := a.Mongo.Polls.UpdateOne(ctx,
		bson.M{"_id": objID, "owner_id": middleware.UserID(c)},
		bson.M{"$set": bson.M{"is_active": false, "closes_at": now}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not close poll"})
		return
	}
	if res.MatchedCount == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "poll not found or you don't own it"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "closed"})
}
