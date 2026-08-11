package mongo

import (
	"context"
	"slices"
	
	"go.mongodb.org/mongo-driver/v2/bson"
	"github.com/gsuntres/pkg/commons"
)

// Example: items []any{
//   bson.M{"foo": "one"},
//   bson.M{"foo": "two"},
// }
func (c *MongoClient) UpdateAddToSet(ctx context.Context, database, name string, filter map[string]any, targetField string, items []any) (bool, error) {
	collection := c.GetCollection(database, name)

	result, err := c.DiscriminatorCheckAndApplyToFilter(ctx, name, filter)
	if err != nil {
		return  false, err
	}

	f := Mongofy(&Query{ 
		Filter: filter,
		WithDiscriminator: result.WithDiscriminator,
		DiscriminatorField: result.DiscriminatorField,	
	})

	var found bson.M
	if err := collection.FindOne(ctx, f).Decode(&found); err != nil {
		return false, err
	}

	data := commons.MapPick(found, "updatedAt")
	
	prepareForUpdateAddToSet(data)
	
	update := bson.M{ 
		"$addToSet": bson.M{
			targetField: bson.M{
				"$each": items,
			},
		},
		"$set": data,
	}
	updateResult, err := collection.UpdateOne(ctx, f, update)
	if err != nil {
		return false, err
	}

	changed := updateResult.ModifiedCount != 0

	ignoreAudit := slices.Contains(c.IgnoreAudit, name)

	if changed && c.WithAudit && !ignoreAudit {
		var after bson.M
		if err := collection.FindOne(ctx, f).Decode(&after); err != nil {
			return false, err
		}

		contx := commons.ContextSerialize(ctx, c.ContextFields)
		audit := &AuditResult {
			Entity: name,
			Op: OpUpdate,
			Data: items,
			Before: found,
			After: after,
			Context: contx,
		}

		(*c.OnAudit)(audit)
	}

	return changed, nil
}

func prepareForUpdateAddToSet(data bson.M) {
	ensureUpdatedAt(data)
}
