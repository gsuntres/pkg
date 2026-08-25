package mongo

import (
	"context"
	"slices"
	
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/gsuntres/pkg/commons"
)

func (c *MongoClient) Replace(ctx context.Context, database, name string, id string, original bson.M) (bson.M, error) {
	collection := c.GetCollection(database, name)

	data := commons.BsonClone(original)

	filter := map[string]any { "_id": id }

	var found bson.M
	err := collection.FindOne(ctx, filter).Decode(&found)
	if err != nil {
		return nil, err
	}

	prepareForReplace(data, found, c.Snake)

	if err := c.DiscriminatorCheckAndApplyToData(ctx, name, data); err != nil {
		return  nil, err
	}
		
	updateResult, err := collection.ReplaceOne(ctx, filter, data)
	if err != nil {
		return nil, err
	}

	PostReplace(updateResult, data, id)

	ignoreAudit := slices.Contains(c.IgnoreAudit, name)

	if c.WithAudit && !ignoreAudit {
		contx := commons.ContextSerialize(ctx, c.ContextFields)
		audit := &AuditResult {
			Entity: name,
			Op: OpUpdate,
			Data: original,
			Before: found,
			After: data,
			Context: contx,
		}

		(*c.OnAudit)(audit)
	}

	return data, nil
}

func PostReplace(updateResult *mongo.UpdateResult, data bson.M, id string) {
	if updateResult.ModifiedCount == 1 {
		data["_id"] = id
	}
}

func prepareForReplace(data bson.M, old bson.M, snake bool) {
	ensureNoId(data)
	ensureUpdatedAt(data, snake)
	ensureExistingCreatedAt(data, old)
}
