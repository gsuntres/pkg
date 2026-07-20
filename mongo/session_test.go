package mongo

import (
	"os"
	"errors"
	"context"
	"encoding/json"
	"testing"
	
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/gsuntres/pkg/commons"
)

func registerDefs(t *testing.T, client *MongoClient) {
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

	client.AddDefinition(person)
	client.AddDefinition(activity)
}

func TestWithSession(t *testing.T) {
	client := GetMongoClient()
	client.SetRelaxed()
	defer client.SetStrict()

	ctx := context.Background()
	ctx = context.WithValue(ctx, "account", "xxxxxx")
	ctx = context.WithValue(ctx, "store", "str_4321")

	_, err := client.WithinSession(ctx, func(sessCtx context.Context, sess *mongo.Session) (any, error) {
		// 1. Register schemas
		registerDefs(t, client)

		// 2. Insert data
		p1 := map[string]any {
			"name": "MyName001",
			"age": int32(25),
		}

		o, err := client.InsertOne(sessCtx, "mydb", "person", p1)
		if err != nil { t.Fatalf("Failed to insertOne %#v", err) }

		a1 := map[string]any {
			"name": "Main activity 1",
			"person": o["_id"],
		}

		_, err = client.InsertOne(sessCtx, "mydb", "activity", a1)
		if err != nil { return nil, err }

		return nil, nil
	})

	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	// 3. Should have activityExpanded defined, let's query it.
	var results bson.M
	filter := map[string]any { "person.name": "MyName001" }
	c := client.Client.Database("mydb").Collection("activityExpanded")
  c.FindOne(ctx, filter).Decode(&results)
// t.Fatalf("Results %T", results)

	m := commons.BsonToMap(results)

  name := commons.GetString(m, "person.name")
  if name != "MyName001" {
  	t.Fatalf("should have saved %s", name)
  }
}

func TestWithSession_Rollback(t *testing.T) {
	client := GetMongoClient()
	client.SetRelaxed()
	defer client.SetStrict()

	var ErrFake error = errors.New("oups")

	ctx := context.Background()
	ctx = context.WithValue(ctx, "account", "xxxxxx")
	ctx = context.WithValue(ctx, "store", "str_4321")

	_, err := client.WithinSession(ctx, func(sessCtx context.Context, sess *mongo.Session) (any, error) {
		// 1. Register schemas
		registerDefs(t, client)

		// 2. Insert data
		p1 := map[string]any {
			"name": "MyName-d1edf78a-7393-11f1-9cdd-4763a3baaa87",
			"age": int32(25),
		}

		_, err := client.InsertOne(sessCtx, "mydb", "person", p1)
		if err != nil { t.Fatalf("Failed to insertOne %#v", err) }

		return nil, ErrFake
	})

	if !errors.Is(err, ErrFake) {
		t.Fatalf("Wrong error: %v", err)
	}

	// 3. Should have activityExpanded defined, let's query it.
	filter := map[string]any { "name": "MyName-d1edf78a-7393-11f1-9cdd-4763a3baaa87" }
	_, err = client.FindOne(ctx, "mydb", "person", filter)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("Should not have saved: %s", err)
	}
}