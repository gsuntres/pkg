package mongo

import (
  "fmt"
  "os"
  "context"
  "testing"
  "time"
  "encoding/json"

  "go.mongodb.org/mongo-driver/v2/bson"
)

func TestCreateIndexes(t *testing.T) {
  cd := &CollectionDefinition{
    IndexSpecs: []map[string]any{
      {
        "keys": map[string]any{
          "code": 1,
        },
        "name": "idx_1",
      },
    },
  }

  c := GetMongoClient()

  c.SetRelaxed()

  collection := c.GetCollection("mydb", "mycol")

  c.SetStrict()

  c.CreateIndexes(collection, cd)

  indexView := collection.Indexes()

  // Specify a timeout to limit the amount of time the operation can run on
	// the server.
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	cursor, err := indexView.List(ctx, nil)
	if err != nil {
    t.Fatal(err)
	}

	// Get a slice of all indexes returned and print them out.
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		t.Fatal(err)
	}

  idx1 := results[1]

  if len(results) != 2 {
    t.Fatal("Unexpected number of indexes")
  }

  if idx1["name"] != "idx_1" {
      t.Fatal("Should have register index idx_1")
  }
}

func TestCreateTimeSeries(t *testing.T) {
  // 1. Register schemas
  schemaEvent, err := os.ReadFile("./.test/event.json")
  if err != nil { t.Fatal(err) }

  var event bson.M
  if err := json.Unmarshal(schemaEvent, &event); err != nil {
    t.Fatalf("Length: %d, First bytes: %x\n", len(schemaEvent), schemaEvent[:4])
  }

  client := GetMongoClient()
  client.AddDefinition(event)
  client.GetCollectionWithCheck("mydb", "event")

  // 2. Insert data
  event1 := map[string]any {
    "name": "event1",
    "createdAt": time.Now(),
  }

  _, err = client.InsertOne(context.Background(), "mydb", "event", event1)
  if err != nil { t.Fatalf("Failed to insertOne %#v", err) }

  c := client.GetCollection("mydb", "event")
  db := c.Database()

  cmd := bson.D{
    {Key: "listCollections", Value: 1},
    {Key: "filter", Value: bson.D{
      {Key: "name", Value: "event"},
    }},
  }

  var result struct {
    Cursor struct {
      FirstBatch []bson.M `bson:"firstBatch"`
    } `bson:"cursor"`
  }
  err = db.RunCommand(context.Background(), cmd).Decode(&result)
  if err != nil {
      t.Fatal(err)
  }

  coll := result.Cursor.FirstBatch[0]

  if coll["type"] != "timeseries" {
    fmt.Println("❌ time series collection")
  }
  
  if opts, ok := coll["options"].(bson.M); ok {
    if _, ok := opts["timeseries"]; !ok {
      t.Fatal("❌ time series collection")
    }
  }

}