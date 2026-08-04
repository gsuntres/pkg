package mongo

import (
	"log"
	"context"
	"slices"
	
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/gsuntres/pkg/commons"
)

// InsertOneWithStruct can be used to insert defined structs.
func (c *MongoClient) InsertOneFromStruct(ctx context.Context, database, name string, data any) (bson.M, error) {
	o, err := ToMap(data)
	if err != nil {
		return nil, err
	}

	return c.InsertOne(ctx, database, name, o)
}

// InsertOne will add missing ids and the created date before saving to the database.
func (c *MongoClient) InsertOne(ctx context.Context, database, name string, original bson.M) (bson.M, error) {
	collection := c.GetCollection(database, name)

	data := commons.BsonClone(original)

	prepareForInsert(data, c.GetIdPrefix(name))

	if err := c.DiscriminatorCheckAndApplyToData(ctx, name, data); err != nil {
		return nil, err
	}

	if _, err := collection.InsertOne(ctx, data); err != nil {
		return nil, err
	}
		
	ignoreAudit := slices.Contains(c.IgnoreAudit, name)
	log.Printf("WITH AUDIT? %v", c.WithAudit)
	if c.WithAudit && !ignoreAudit {
		log.Println("trigger audit")
		contx := commons.ContextSerialize(ctx, c.ContextFields)
		audit := &AuditResult {
			Entity: name,
			Op: OpInsert,
			Data: original,
			After: data,
			Context: contx,
		}

		if c.OnAudit == nil {
			log.Println("Audit failed because OnAudit is not defined")
		} else {
			(*c.OnAudit)(audit)
		}
	}

	return data, nil
}

// prepareForInsert takes a map[string]any and:
//	* adds a new _id if property does not exist or is an empty string
//	* adds/updates property created_at using the current timestamp.
func prepareForInsert(data bson.M, idPrefix string) {
	ensureId(data, idPrefix)
	ensureCreatedAt(data)
	ensureUpdatedAt(data)
}
