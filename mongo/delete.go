package mongo

import (
	"context"
	"slices"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/gsuntres/pkg/commons"
)

// DeleteOne will delete the first document that matches the filter.
func (c *MongoClient) DeleteOne(ctx context.Context, database, name string, filter bson.M) error {
	var err error

	// 1. Prepare query.
	collection := c.GetCollection(database, name)

	// 2. Check discriminator and mongofy
	result, err := c.DiscriminatorCheckAndApplyToFilter(ctx, name, filter)
	if err != nil {
		return err
	}

	f := Mongofy(&Query{ 
		Filter: filter,
		WithDiscriminator: result.WithDiscriminator,
		DiscriminatorField: result.DiscriminatorField,	
	})

	var found bson.M
	if c.WithAudit {
		err = collection.FindOne(ctx, f).Decode(&found)
		if err != nil {
			return err
		}
	}
	
	// 3. Delete
  _, err = collection.DeleteOne(ctx, f)
	if err != nil {
		return err
	}

	if c.WithAudit && !slices.Contains(c.IgnoreAudit, name) {
		contx := commons.ContextSerialize(ctx, c.ContextFields)
		audit := &AuditResult {
			Entity: name,
			Op: OpDelete,
			Before: found,
			Context: contx,
		}

		(*c.OnAudit)(audit)
	}

	return nil
}
