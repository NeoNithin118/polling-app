package db

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(addr, password string, dbIndex int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       dbIndex,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return client, nil
}

// Key helpers centralised so the naming scheme lives in exactly one place.

func VotesKey(pollID string) string {
	return "poll:" + pollID + ":votes"
}

func VotersKey(pollID string) string {
	return "poll:" + pollID + ":voters"
}

func Channel(pollID string) string {
	return "channel:poll:" + pollID
}

// ChannelPattern is used by the WebSocket hub to subscribe to every
// poll's channel with a single PSubscribe call instead of one
// subscription per poll.
const ChannelPattern = "channel:poll:*"
