package db

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	client *mongo.Client
}

var Models *Mongo

func GetMongoURI() string {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal().Msg("MONGODB_URI environment variable is not set")
	}
	return uri
}

func init() {
	Models = &Mongo{}
}

func (c *Mongo) articles() *mongo.Collection {
	return c.client.Database("sleuth").Collection("articles")
}

func (c *Mongo) ConnectDatabase(uri string) error {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	c.client = client

	log.Info().Msg("Connected to MongoDB!")
	return nil
}
