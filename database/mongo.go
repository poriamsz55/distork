package database

import (
	"context"
	config "drive/configs"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func Connect() (*mongo.Database, error) {

	if DB != nil {
		return DB, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, err
	}

	DB = client.Database(config.GetConfigDB().DatabaseName)

	return DB, nil
}

func Collection(name string) *mongo.Collection {

	return DB.Collection(name)
}
