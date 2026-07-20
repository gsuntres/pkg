package mongo

import (
	"os"
	"context"
	"strings"
	"encoding/json"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestCreateViews(t *testing.T) {
	// 1. Register schemas
	schemaPerson, err := os.ReadFile("./.test/person.json")
	if err != nil { t.Fatal(err) }

	schemaActivity, err := os.ReadFile("./.test/activity.json")
	if err != nil { t.Fatal(err) }

	var person bson.M
	if err := json.Unmarshal(schemaPerson, &person); err != nil {
		t.Fatalf("Length: %d, First bytes: %x\n", len(schemaPerson), schemaPerson[:4])
	}

	var activity bson.M
	if err := json.Unmarshal(schemaActivity, &activity); err != nil {
		t.Fatalf("Length: %d, First bytes: %x\n", len(schemaActivity), schemaActivity[:4])
	}

	client := GetMongoClient()
	client.AddDefinition(person)
	client.AddDefinition(activity)
	client.GetCollectionWithCheck("mydb", "person")
	client.GetCollectionWithCheck("mydb", "activity")

	// 2. Insert data
	p1 := map[string]any {
		"name": "MyName112",
		"age": int32(25),
	}

	o, err := client.InsertOne(context.Background(), "mydb", "person", p1)
	if err != nil { t.Fatalf("Failed to insertOne %#v", err) }

	a1 := map[string]any {
		"name": "Main activity",
		"person": o["_id"],
	}

	o, err = client.InsertOne(context.Background(), "mydb", "activity", a1)
	if err != nil { t.Fatalf("Failed to insertOne %#v", err) }

	// 3. Should have activityExpanded defined, let's query it.
	var results bson.M
	filter := map[string]any { "person.name": "MyName112" }
	c := client.Client.Database("mydb").Collection("activityExpanded")
  c.FindOne(context.Background(), filter).Decode(&results)

  if !strings.HasPrefix(results["_id"].(string), "act_") {
		t.Fatal("_id should have been prefixed")
	}
}