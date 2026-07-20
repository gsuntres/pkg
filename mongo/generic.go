package mongo

import (
	"errors"
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (c *MongoClient) GenericFind(ctx context.Context, payload *FindRequest) (*DataResults, error) {
	account := ctx.Value("account").(string)
	if account == "" {
		return nil, errors.New("account required")
	}

	if payload.Entity == "" {
		return nil, errors.New("entity required")
	}

	name := payload.Entity

	database := c.GetName(account)

	collection := c.GetCollection(database, name)

	var filter bson.D
	var err error

	filter, err = MapToBsonD(payload.Filter)
	if err != nil {
		return nil, err
	}

	opts := options.Find()

	var cursor *mongo.Cursor
	cursor, err = collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}

	results := make([]map[string]any, 0)
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	dataResults := &DataResults{
		Data: results,
	}

	return dataResults, nil
}

func GenericInsertOne(ctx context.Context, entityType string, data any) (any, error) {
	// coll := getCollection("sampledb", "dummy")

	return nil, nil
}
