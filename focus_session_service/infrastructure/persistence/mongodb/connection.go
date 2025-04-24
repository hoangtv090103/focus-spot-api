package mongodb

import (
	"context"
	"focusspot/focussessionservice/config"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// NewMongoDBConnection configures connection to MongoDB and returns cliend and database
func NewMongoDBConnection(cfg *config.MongoDBConfig) (*mongo.Client, *mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Create MongoDB client
	clientOptions := options.Client().ApplyURI(cfg.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, nil, err
	}

	// Ping to check NewMongoDBConnection
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, nil, err
	}

	// Get database
	db := client.Database(cfg.Database)

	return client, db, nil
}

// CloseMongoDBConnection closes connection to MongoDB
func CloseMongoDBConnection(client *mongo.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.Disconnect(ctx)
}
