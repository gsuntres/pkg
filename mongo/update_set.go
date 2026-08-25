package mongo

import (
	"fmt"
	"context"
	"slices"
	
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"github.com/gsuntres/pkg/commons"
)

// UpdateSet search documents using filter and updates the first it finds using the $set operator.
// Will return true if the update succeeds.
func (c *MongoClient) UpdateSet(ctx context.Context, database, name string, filter, original bson.M) (bson.M, error) {
	collection := c.GetCollection(database, name)
	
	data := commons.BsonClone(original)
	
	prepareForUpdateSet(data, c.Snake)

	result, err := c.DiscriminatorCheckAndApplyToFilter(ctx, name, filter)
	if err != nil {
		return  nil, fmt.Errorf("Failed to discriminate in filter: %w", err)
	}

	f := Mongofy(&Query{ 
		Filter: filter,
		WithDiscriminator: result.WithDiscriminator,
		DiscriminatorField: result.DiscriminatorField,	
	})

	if err := c.DiscriminatorOmitInData(name, data); err != nil {
		return  nil, fmt.Errorf("Failed to discriminate in data: %w", err)
	}

	var found bson.M
	if c.WithAudit {
		if err := collection.FindOne(ctx, f).Decode(&found); err != nil {
			return nil, fmt.Errorf("Failed to find: %w", err)
		}
	}
	
	opts := options.UpdateOne().SetUpsert(false)
	update := bson.M{ "$set": data }
	updateResult, err := collection.UpdateOne(ctx, f, update, opts)
	if err != nil {
		return nil, fmt.Errorf("Failed to update: %w", err)
	}

	changed := updateResult.ModifiedCount != 0

	ignoreAudit := slices.Contains(c.IgnoreAudit, name)

	var after bson.M
	if err := collection.FindOne(ctx, f).Decode(&after); err != nil {
		return nil, fmt.Errorf("Failed to find: %w", err)
	}

	if changed && c.WithAudit && !ignoreAudit {
		contx := commons.ContextSerialize(ctx, c.ContextFields)
		audit := &AuditResult {
			Entity: name,
			Op: OpUpdate,
			Data: original,
			Before: found,
			After: after,
			Context: contx,
		}

		(*c.OnAudit)(audit)
	}

	return after, nil
}

func prepareForUpdateSet(data bson.M, snake bool) {
	ensureUpdatedAt(data, snake)
}
