package mongo

import (
  "context"
	"testing"

  "go.mongodb.org/mongo-driver/v2/bson"
)

func TestFindOne_Default(t *testing.T) {
	client := GetMongoClient()

  data := map[string]any {
    "_id": "su_123459",
    "name": "MyNameFindOne",
    "age": int32(25),
  }

  o, err := client.InsertOne(context.Background(), "mydb", "mycollection", data)
	if err != nil {
		t.Fatalf("Failed to insertOne %#v", err)
	}
  filter := bson.M{"name": "MyNameFindOne"}
  found, err := client.FindOne(context.Background(), "mydb", "mycollection", filter)
  if err != nil {
    t.Fatalf("Failed to findOne %#v", err)
  }

  AssertSubset(t, o, found, "Should have been equal")
}
