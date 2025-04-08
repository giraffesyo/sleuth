package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Article represents a news article model.
type Article struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"` // MongoDB document ID
	Title       string             `bson:"title" json:"title"`
	Url         string             `bson:"url" json:"url"`
	Date        string             `bson:"date" json:"date"`
	Description string             `bson:"description" json:"description"`
	Provider    string             `bson:"provider" json:"provider"`
}

// CreateArticle inserts a new article into the provided MongoDB collection.
// If the insertion is successful, it assigns the generated ID to article.Id.
func (c *Mongo) CreateArticle(ctx context.Context, article *Article) error {
	// Set a timeout to avoid hanging operations.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Insert the article document into the collection.
	result, err := c.articles().InsertOne(ctx, article)
	if err != nil {
		return err
	}

	// Set the ID field on the article.
	article.Id = result.InsertedID.(primitive.ObjectID)
	return nil
}

// // FindArticleByID searches for an article by its MongoDB ObjectID.
// // It returns the article if found, or an error if no document matches the given ID.
// func (c *Mongo) FindArticleByID(ctx context.Context, id primitive.ObjectID) (*Article, error) {
// 	// Set a timeout for the operation.
// 	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
// 	defer cancel()

// 	var article Article
// 	filter := bson.M{"_id": id}

// 	// Find the document with the matching _id.
// 	err := c.ArticlesCollection().FindOne(ctx, filter).Decode(&article)
// 	if err != nil {
// 		if err == mongo.ErrNoDocuments {
// 			return nil, errors.New("article not found")
// 		}
// 		return nil, err
// 	}

// 	return &article, nil
// }

// // DeleteArticle removes an article from the collection by its MongoDB ObjectID.
// // It returns an error if no article is found or if the operation fails.
// func (c *Mongo) DeleteArticle(ctx context.Context, id primitive.ObjectID) error {
// 	// Set a timeout for the operation.
// 	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
// 	defer cancel()

// 	filter := bson.M{"_id": id}

// 	// Delete the document matching the given _id.
// 	result, err := c.ArticlesCollection().DeleteOne(ctx, filter)
// 	if err != nil {
// 		return err
// 	}
// 	if result.DeletedCount == 0 {
// 		return errors.New("no article found to delete")
// 	}
// 	return nil
// }
