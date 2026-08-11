package mongo

import (
	"fmt"
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// GetByIds return an array of documents matching provided ids.
func (c *MongoClient) GetByIds(ctx context.Context, database, name string, ids []string) (bson.A, error) {

	// 1. Prepare to query.
	collection := c.GetCollection(database, name)

  	// 2. Query
  	filter := bson.M{ "_id": bson.M{ "$in": ids } }

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("Failed to find: %w", err)
	}

	var results bson.A
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, fmt.Errorf("Failed to decode cursor: %w", err)
	}

	return results, nil
}
