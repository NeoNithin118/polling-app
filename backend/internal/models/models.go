package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"password_hash" json:"-"`
	CreatedAt    time.Time          `bson:"created_at" json:"createdAt"`
}

type Option struct {
	ID   string `bson:"id" json:"id"`
	Text string `bson:"text" json:"text"`
}

type Poll struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OwnerID   string             `bson:"owner_id" json:"ownerId"`
	Question  string             `bson:"question" json:"question"`
	Options   []Option           `bson:"options" json:"options"`
	IsActive  bool               `bson:"is_active" json:"isActive"`
	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`
	ClosesAt  *time.Time         `bson:"closes_at,omitempty" json:"closesAt,omitempty"`
}

// Vote is the permanent audit-log entry. It is NOT what powers the
// live counter (Redis owns that) — this is what lets us reconcile,
// export, or rebuild Redis state if it's ever lost.
type Vote struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PollID           string             `bson:"poll_id" json:"pollId"`
	OptionID         string             `bson:"option_id" json:"optionId"`
	VoterFingerprint string             `bson:"voter_fingerprint" json:"-"`
	CreatedAt        time.Time          `bson:"created_at" json:"createdAt"`
}

// PollResult is what the API returns for "current standings":
// the poll shape plus live counts merged in from Redis.
type PollResult struct {
	Poll   Poll             `json:"poll"`
	Counts map[string]int64 `json:"counts"`
	Total  int64            `json:"total"`
}
