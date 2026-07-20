package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// FindOne will return at most one document based on the filter provided.
func (c *MongoClient) Count(ctx context.Context, database, name string, filter bson.M) (int64, error) {
	var zero int64

	// 1. Prepare to query.
	collection := c.GetCollection(database, name)

	// 2. Check discriminator and mongofy
	result, err := c.DiscriminatorCheckAndApplyToFilter(ctx, name, filter)
	if err != nil {
		return zero, err
	}

	f := Mongofy(&Query{ 
		Filter: filter,
		WithDiscriminator: result.WithDiscriminator,
		DiscriminatorField: result.DiscriminatorField,	
	})

	// 2. Query
	count, err := collection.CountDocuments(ctx, f)
	if err != nil {
		return zero, err
	}

	return count, nil
}
