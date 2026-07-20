package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// FindOne will return at most one document based on the filter provided.
func (c *MongoClient) FindOne(ctx context.Context, database, name string, filter bson.M, opts ...options.Lister[options.FindOneOptions]) (bson.M, error) {

	// 1. Prepare to query.
	collection := c.GetCollection(database, name)

	// 2. Query
	var out bson.M
	err := collection.FindOne(ctx, filter, opts...).Decode(&out)
	if err != nil {
		return nil, err
	}

	return out, nil
}
