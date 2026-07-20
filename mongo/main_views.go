package mongo

import (
	"context"
  "log"

  "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// CreateViews will create views for the given collection and definition.
func (c *MongoClient) CreateViews(db *mongo.Database, cdef *CollectionDefinition) bool {
  if cdef == nil || cdef.Views == nil {
    log.Printf("No definition for views found.")

    return false
  }

  viewCreated := false

  for name, defVal := range cdef.Views {

  	// 1. Decode definition
  	v, err := bson.Marshal(defVal)
  	if err != nil {
      log.Printf("failed to marshal %v", err)

      continue
    }

    // 2. Take the raw representation
    vRaw := bson.Raw(v)
    if err := vRaw.Validate(); err != nil {
      log.Printf("failed to validate bson raw: %v", err)

      continue
    }

    pipelineVal := vRaw.Lookup("pipeline")
    pipelineArr, ok := pipelineVal.ArrayOK()
    if !ok {
      log.Printf("Unable to extract pipeline")

      continue
    }
    
    pipeline := mongo.Pipeline{}

    successPipeline := true
    pipelineValues, _ := pipelineArr.Values()
    for _, o := range pipelineValues {
      var stage bson.D
      if err := bson.Unmarshal(o.Value, &stage); err != nil {
        log.Printf("failed to unmarshal stage %v, %v", o, err)

        successPipeline = false
        break
      }

      pipeline = append(pipeline, stage)
    }

    if successPipeline == false {
      continue
    }
    
		// Specify the Collation option to set a default collation for the view.
		opts := options.CreateView().SetCollation(&options.Collation{
			Locale: "en_US",
		})

		viewonVal := vRaw.Lookup("viewOn")
    viewOn, ok := viewonVal.StringValueOK(); 
    if !ok {
    	log.Printf("failed to find viewOn")

      continue
    }

		err = db.CreateView(context.TODO(), name, viewOn, pipeline, opts)
		if err != nil {
			log.Printf("failed to create view %v", err)

      continue
		}

    viewCreated = true
  }

  return viewCreated
}