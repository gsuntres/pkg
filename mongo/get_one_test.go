package mongo

import (
  "context"
	"testing"
)

func TestGetOne_ByIdString(t *testing.T) {
	client := GetMongoClient()

  data := map[string]any {
    "_id": "su_156345",
    "name": "MyNameGetOne",
    "age": int32(25),
  }

  o, err := client.InsertOne(context.Background(), "mydb", "mycollection", data)
	if err != nil {
		t.Fatalf("Failed to insertOne %#v", err)
	}

  found, err := client.GetOne(context.Background(), "mydb", "mycollection", "su_156345")
  if err != nil {
    t.Fatalf("Failed to findOne %#v", err)
  }

  AssertSubset(t, o, found, "Should have been equal")
}

func TestGetOne_ByObjectId(t *testing.T) {
	client := GetMongoClient()

  data := map[string]any {
    "name": "MyNameGetOne",
    "age": int32(25),
  }

  o, err := client.InsertOne(context.Background(), "mydb", "mycollection", data)
	if err != nil {
		t.Fatalf("Failed to insertOne %#v", err)
	}

  found, err := client.GetOne(context.Background(), "mydb", "mycollection", o["_id"])
  if err != nil {
    t.Fatalf("Failed to findOne %#v", err)
  }

  AssertSubset(t, o, found, "Should have been equal")
}
