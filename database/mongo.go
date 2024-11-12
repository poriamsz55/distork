package database

import (
	"context"
	"log"
	"time"

	config "github.com/poriamsz55/distork/configs"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	DB     *mongo.Database
	ctx    context.Context
	cancel context.CancelFunc
)

func Connect() (*mongo.Database, error) {
	if DB != nil {
		return DB, nil
	}

	// Persistent context
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, err
	}

	// Ensure the connection is established
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Initialize the database object
	DB = client.Database(config.GetConfigDB().DatabaseName)

	return DB, nil
}

func Disconnect() {
	if DB != nil {
		err := DB.Client().Disconnect(context.Background())
		if err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}
}

func Collection(name string) *mongo.Collection {
	if DB == nil {
		panic("database not initialized. Call Connect first.")
	}
	return DB.Collection(name)
}
