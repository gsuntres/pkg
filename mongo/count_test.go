package mongo

import (
  "context"
	"testing"

  "go.mongodb.org/mongo-driver/v2/bson"
)

func TestCount_Default(t *testing.T) {
	client := GetMongoClient()
  
  count, err := client.Count(context.Background(), "mydb", "mycollection", bson.M{})
  if err != nil {
    t.Fatalf("Failed to findOne %#v", err)
  }

  // TODO: this will change as tests change
  if count != 3 {
    t.Fatalf("Wrong count %d", count)
  }
}
