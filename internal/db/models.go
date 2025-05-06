package db

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// RelevantTimestamp represents one extracted segment where
// a body discovery (or related forensic event) was mentioned.
type RelevantTimestamp struct {
	Start       string `bson:"start"       json:"start"`        // e.g. "00:25"
	End         string `bson:"end"         json:"end"`          // e.g. "00:28"
	TextSnippet string `bson:"textSnippet" json:"text_snippet"` // the transcript excerpt
	Location    string `bson:"location"    json:"location"`     // if extracted
	TimeDetail  string `bson:"timeDetail"  json:"time_detail"`  // e.g. "yesterday evening"
}

// Article represents a news article model.
type Article struct {
	Id                                primitive.ObjectID  `bson:"_id,omitempty" json:"id,omitempty"` // MongoDB document ID
	Title                             string              `bson:"title" json:"title"`
	Url                               string              `bson:"url" json:"url"`
	Date                              string              `bson:"date" json:"date"`
	Description                       string              `bson:"description" json:"description"`
	Provider                          string              `bson:"provider" json:"provider"`
	AiHasCheckedIfShouldDownloadVideo bool                `bson:"aiHasCheckedIfShouldDownloadVideo" json:"AiHasCheckedIfShouldDownloadVideo"`
	AiSuggestsDownloadingVideo        bool                `bson:"aiSuggestsDownloadingVideo" json:"AiSuggestsDownloadingVideo"`
	VideoPath                         string              `bson:"videoPath" json:"videoPath"` // Path to the downloaded video file
	VideoUrl                          string              `bson:"videoUrl" json:"videoUrl"`   // Direct URL to the video file
	RelevantTimestamps                []RelevantTimestamp `bson:"relevantTimestamps" json:"relevantTimestamps"`
	VictimNames                       []string            `bson:"victimNames" json:"victimNames"`
	Location                          string              `bson:"location" json:"location"`
	CaseId                            int32               `bson:"caseId" json:"caseId"` // For case grouping
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

// FindArticleByID searches for an article by its MongoDB ObjectID.
// It returns the article if found, or an error if no document matches the given ID.
func (c *Mongo) FindArticleByUrl(ctx context.Context, url string) (*Article, error) {
	// Set a timeout for the operation.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var article Article
	filter := bson.M{"url": url}

	// Find the document with the matching url.
	err := c.articles().FindOne(ctx, filter).Decode(&article)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("article not found")
		}
		return nil, err
	}

	return &article, nil
}

func (c *Mongo) FindAllArticles(ctx context.Context) ([]*Article, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var articles []*Article
	filter := bson.M{} // Empty filter to get all documents

	cursor, err := c.articles().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode each document into an Article struct
	for cursor.Next(ctx) {
		var article Article
		if err := cursor.Decode(&article); err != nil {
			return nil, err
		}
		articles = append(articles, &article)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

// FindAllArticlesNotChecked searches for all articles that have not been checked by AI.
// It returns a slice of articles and an error if the operation fails.
func (c *Mongo) FindAllArticlesNotChecked(ctx context.Context) ([]*Article, error) {
	// Set a timeout for the operation.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var articles []*Article
	filter := bson.M{"aiHasCheckedIfShouldDownloadVideo": false}

	// Find all documents that match the filter.
	cursor, err := c.articles().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode each document into an Article struct.
	for cursor.Next(ctx) {
		var article Article
		if err := cursor.Decode(&article); err != nil {
			return nil, err
		}
		articles = append(articles, &article)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

// UpdateArticle updates an article in the collection by its MongoDB ObjectID.
// It returns an error if no article is found or if the operation fails.
func (c *Mongo) UpdateArticle(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	// Set a timeout for the operation.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}

	// Update the document matching the given _id.
	result, err := c.articles().UpdateOne(ctx, filter, bson.M{"$set": update})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("no article found to update")
	}
	return nil
}

// FindArticlesByFilter searches for articles based on a filter.
// It returns a slice of articles that match the filter criteria.
func (c *Mongo) FindArticlesByFilter(ctx context.Context, filter bson.M) ([]*Article, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var articles []*Article

	cursor, err := c.articles().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode each document into an Article struct
	for cursor.Next(ctx) {
		var article Article
		if err := cursor.Decode(&article); err != nil {
			return nil, err
		}
		articles = append(articles, &article)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

// Query represents a search query generated by AI
type Query struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"` // MongoDB document ID
	Query       string             `bson:"query" json:"query"`                // The search query
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`        // When the query was generated
	Used        bool               `bson:"used" json:"used"`                  // Whether this query has been used in a search
	Description string             `bson:"description" json:"description"`    // Optional description or context for the query
}

// CreateQuery inserts a new query into the queries collection
func (c *Mongo) CreateQuery(ctx context.Context, query *Query) error {
	// Set a timeout to avoid hanging operations
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Set creation time if not already set
	if query.CreatedAt.IsZero() {
		query.CreatedAt = time.Now()
	}

	// Insert the query document into the collection
	result, err := c.queries().InsertOne(ctx, query)
	if err != nil {
		return err
	}

	// Set the ID field on the query
	query.Id = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindQueryByValue searches for a query by its string value
func (c *Mongo) FindQueryByValue(ctx context.Context, queryString string) (*Query, error) {
	// Set a timeout for the operation
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var query Query
	filter := bson.M{"query": queryString}

	// Find the document with the matching query string
	err := c.queries().FindOne(ctx, filter).Decode(&query)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("query not found")
		}
		return nil, err
	}

	return &query, nil
}

// FindAllQueries returns all search queries in the collection
func (c *Mongo) FindAllQueries(ctx context.Context) ([]*Query, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var queries []*Query
	filter := bson.M{} // Empty filter to get all documents

	cursor, err := c.queries().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode each document into a Query struct
	for cursor.Next(ctx) {
		var query Query
		if err := cursor.Decode(&query); err != nil {
			return nil, err
		}
		queries = append(queries, &query)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return queries, nil
}

// FindUnusedQueries returns all queries that haven't been used yet
func (c *Mongo) FindUnusedQueries(ctx context.Context) ([]*Query, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var queries []*Query
	filter := bson.M{"used": false} // Only get unused queries

	cursor, err := c.queries().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode each document into a Query struct
	for cursor.Next(ctx) {
		var query Query
		if err := cursor.Decode(&query); err != nil {
			return nil, err
		}
		queries = append(queries, &query)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return queries, nil
}

// UpdateQuery updates a query in the collection by its MongoDB ObjectID
func (c *Mongo) UpdateQuery(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	// Set a timeout for the operation
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}

	// Update the document matching the given _id
	result, err := c.queries().UpdateOne(ctx, filter, bson.M{"$set": update})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("no query found to update")
	}
	return nil
}

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
