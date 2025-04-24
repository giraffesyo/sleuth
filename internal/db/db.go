package db

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
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
		uri = "mongodb://localhost:9000/?directConnection=true"
	}
	return uri
}

func init() {
	Models = &Mongo{}
}

func (c *Mongo) articles() *mongo.Collection {
	return c.client.Database("sleuth").Collection("articles")
}

func (c *Mongo) queries() *mongo.Collection {
	return c.client.Database("sleuth").Collection("queries")
}

func ensureUrlUniqueIndex(collection *mongo.Collection) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.M{"url": 1},
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	log.Info().Msg("Index created successfully")
	return nil
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

	if err := ensureUrlUniqueIndex(c.articles()); err != nil {
		return err
	}
	log.Info().Msg("Connected to MongoDB!")
	return nil
}
