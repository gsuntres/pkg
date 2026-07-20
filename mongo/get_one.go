package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// FindOne will return at most one document based on the filter provided.
func (c *MongoClient) GetOne(ctx context.Context, database, name string, id any) (bson.M, error) {

	// 1. Prepare to query.
	collection := c.GetCollection(database, name)

  // 2. Query
  filter := map[string]any { "_id": id }

	var out bson.M
	err := collection.FindOne(ctx, filter).Decode(&out)
	if err != nil {
		return nil, err
	}

	return out, nil
}
