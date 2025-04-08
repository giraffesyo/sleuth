package db

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	Client *mongo.Client
}

func NewMongo() *Mongo {
	return &Mongo{}
}

func (c *Mongo) ArticlesCollection() *mongo.Collection {
	return c.Client.Database("sleuth").Collection("articles")
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

	c.Client = client

	log.Info().Msg("Connected to MongoDB!")
	return nil
}
