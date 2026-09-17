package db

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(addr, password string, dbIndex int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       dbIndex,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
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

const ChannelPattern = "channel:poll:*"