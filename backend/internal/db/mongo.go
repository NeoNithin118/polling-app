package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Mongo bundles the handles the rest of the app actually needs.
// Everything is resolved once at startup so handlers never touch
// connection strings.
type Mongo struct {
	Client *mongo.Client
	Users  *mongo.Collection
	Polls  *mongo.Collection
	Votes  *mongo.Collection
}

func ConnectMongo(uri, dbName string) (*Mongo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	database := client.Database(dbName)
	m := &Mongo{
		Client: client,
		Users:  database.Collection("users"),
		Polls:  database.Collection("polls"),
		Votes:  database.Collection("votes"),
	}

	if err := m.ensureIndexes(ctx); err != nil {
		return nil, err
	}
	return m, nil
}

// ensureIndexes sets up the constraints that matter for correctness:
// unique emails, and fast lookups of polls by owner / votes by poll.
func (m *Mongo) ensureIndexes(ctx context.Context) error {
	_, err := m.Users.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"email": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	_, err = m.Polls.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{"owner_id": 1},
	})
	if err != nil {
		return err
	}

	_, err = m.Votes.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{"poll_id": 1},
	})
	return err
}
