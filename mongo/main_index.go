package mongo

import (
  "context"
  "log"

  "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// CreateIndexes will create indexes for the given collection and definition.
func (c *MongoClient) CreateIndexes(collection *mongo.Collection, cdef *CollectionDefinition) {
  if cdef == nil {
    log.Printf("No definitions will not create indexes")

    return
  }

  // handle indexes
  indexModels := make([]mongo.IndexModel, 0)

  for _, keyDef := range cdef.IndexSpecs {
    log.Printf("Key Definition [%s]%s", cdef.Name, keyDef["name"])
    
    kdb, err := bson.Marshal(keyDef)
    if err != nil {
      log.Printf("failed to marshal %v", err)

      continue
    }

    kdRaw := bson.Raw(kdb)
    if err := kdRaw.Validate(); err != nil {
      log.Printf("failed to validate bson raw: %v", err)

      continue
    }
    //
    idxModel := mongo.IndexModel{}

    opts := options.Index()

    keysVal := kdRaw.Lookup("keys")

    var keysBson bson.D
    if err := bson.Unmarshal(keysVal.Value, &keysBson); err != nil {
      log.Printf("failed to unmarshal keys %v, %v", keysVal.Value, err)

      continue
    }

    idxModel.Keys = keysBson

    nameVal := kdRaw.Lookup("name")
    if name, ok := nameVal.StringValueOK(); ok {
      opts = opts.SetName(name)
    }

    uniqueVal := kdRaw.Lookup("unique")
    if unique, ok := uniqueVal.BooleanOK(); ok {
      opts = opts.SetUnique(unique)
    }

    partialVal := kdRaw.Lookup("partialFilterExpression")
    if partialVal.Type == bson.TypeEmbeddedDocument {
      var partialFilterExpression bson.M
      if err := bson.Unmarshal(partialVal.Value, &partialFilterExpression); err != nil {
        log.Printf("failed to unmarshal partial filter: %v", err)
      }

      opts = opts.SetPartialFilterExpression(partialFilterExpression)
    }

    idxModel.Options = opts

    indexModels = append(indexModels, idxModel)
  }

  collection.Indexes().CreateMany(context.Background(), indexModels)
}
