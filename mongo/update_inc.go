package mongo

import (
	"context"
	"slices"
	"errors"
	
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"go.mongodb.org/mongo-driver/v2/bson"
	"github.com/gsuntres/pkg/commons"
)

// UpdateInc find one document and update with the $inc operator.
// Will return true if the update succeeds.
func (c *MongoClient) UpdateInc(ctx context.Context, database, name string, filter, original bson.M) (bson.M, error) {
	collection := c.GetCollection(database, name)
	
	data := bson.M{}
	
	for k, v := range original {
		var normalized int64
		switch v.(type) {
		case int32:
			normalized = int64(v.(int32))
		case int64:
			normalized = v.(int64)
		default:
			continue
		}

		data[k] = normalized
	}
	
	result, err := c.DiscriminatorCheckAndApplyToFilter(ctx, name, filter)
	if err != nil {
		return  nil, err
	}

	f := Mongofy(&Query{ 
		Filter: filter,
		WithDiscriminator: result.WithDiscriminator,
		DiscriminatorField: result.DiscriminatorField,	
	})

	if err := c.DiscriminatorOmitInData(name, data); err != nil {
		return  nil, err
	}

	var found bson.M
	if c.WithAudit {
		if err := collection.FindOne(ctx, f).Decode(&found); err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				return nil, err	
			}			
		}
	}

	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	update := bson.M{ "$inc": data }
	var after bson.M
	errUp := collection.FindOneAndUpdate(ctx, f, update, opts).Decode(&after)
	if errUp != nil {
		return nil, errUp
	}

	ignoreAudit := slices.Contains(c.IgnoreAudit, name)

	if c.WithAudit && !ignoreAudit {
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
