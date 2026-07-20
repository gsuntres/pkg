package mongo

import (
	"os"
	"fmt"
	"context"
	"strings"
	"testing"
	"encoding/json"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestInsertOne(t *testing.T) {
	client := GetMongoClient()

	data := map[string]any {
		"_id": "su_123456",
		"name": "MyName",
		"age": int32(25),
	}

	o, err := client.InsertOne(context.Background(), "mydb", "mycollection", data)
	if err != nil {
		t.Fatalf("Failed to insertOne %#v", err)
	}

  // raw query
  var results bson.M
  filter := bson.M{ "name": "MyName" }

  c := client.Client.Database("mydb").Collection("mycollection")
  c.FindOne(context.Background(), filter).Decode(&results)

	AssertSubset(t, o, results, "Should have been equal")
}

func TestInsertOne_DataStruct(t *testing.T) {
	client := GetMongoClient()

	data := &Sample {
		Id: "su_123457",
		Name: "MyName200",
		Age: 25,
	}

	o, err := client.InsertOneFromStruct(context.Background(), "mydb", "mycollection", data)
	if err != nil {
		t.Fatalf("Failed to insertOne %#v", err)
	}

  // raw query
  var results bson.M
  filter := map[string]any { "name": "MyName200" }

  c := client.Client.Database("mydb").Collection("mycollection")
  c.FindOne(context.Background(), filter).Decode(&results)

  AssertSubset(t, o, results, "Should have been equal")
}

func TestInsertOne_WithIdPrefix(t *testing.T) {
	// Read JSON file
	data, err := os.ReadFile("./.test/user.json")
	if err != nil {
		t.Fatal(err)
	}

	var user bson.M
	if err := json.Unmarshal(data, &user); err != nil {
		t.Fatalf("Length: %d, First bytes: %x\n", len(data), data[:4])
	}

	client := GetMongoClient()
	client.AddDefinition(user)

	in := map[string]any {
		"name": "MyName112",
		"age": int32(25),
	}

	o, err := client.InsertOne(context.Background(), "mydb", "user", in)
	if err != nil {
		t.Fatalf("Failed to insertOne %#v", err)
	}

	// raw query
  var results bson.M
	filter := map[string]any { "name": "MyName112" }
	c := client.Client.Database("mydb").Collection("user")
  c.FindOne(context.Background(), filter).Decode(&results)

	if !strings.HasPrefix(results["_id"].(string), "usr_") {
		t.Fatal("_id should have been prefixed")
	}

  AssertSubset(t, o, results, "Should have been equal")
}

func TestPrepareForInsert_WithoutId(t *testing.T) {
	data := map[string]any {
		"name": "My Name Is",
	}

	prepareForInsert(data, "")

	if id, okid := data["_id"]; !okid || id == "" {
		t.Fatal("Failed to add Id")
	}
}

func TestPrepareForInsert_ExistingId(t *testing.T) {
	data := map[string]any {
		"_id": "myidxxxxxx",
		"name": "My Name Is",
	}

	prepareForInsert(data, "")

	if data["_id"] == "" {
		t.Fatal("id was updated")
	}
}

func TestInsertOne_Discriminate(t *testing.T) {
	// 1. Register schemas
	schemaStore, err := os.ReadFile("./.test/store.json")
	if err != nil { t.Fatal(err) }

	schemaOffer, err := os.ReadFile("./.test/offer.json")
	if err != nil { t.Fatal(err) }

	var store bson.M
	if err := json.Unmarshal(schemaStore, &store); err != nil {
		t.Fatalf("Length: %d, First bytes: %x\n", len(schemaStore), schemaStore[:4])
	}

	var offer bson.M
	if err := json.Unmarshal(schemaOffer, &offer); err != nil {
		t.Fatalf("Length: %d, First bytes: %x\n", len(schemaOffer), schemaOffer[:4])
	}

	client := GetMongoClient()
	client.AddDefinition(store)
	client.AddDefinition(offer)

	// Test
	ctx := context.Background()
	ctx = context.WithValue(ctx, "account", "xxxxxx")
	ctx = context.WithValue(ctx, "store", "str_1234")

	// 2. Insert data
  offer1 := map[string]any {
    "name": "Offer 1",
  }

  saved, err := client.InsertOne(ctx, "mydb", "offer", offer1)
  if err != nil { t.Fatalf("Failed to insertOne %#v", err) }

  v, ok := saved["store"]
  if !ok || v != "str_1234" {
  	t.Fatalf("Should have set store not %s", v)
  }
}

func TestInsertOne_Discriminate_NoStore(t *testing.T) {
	// 1. Register schemas
	schemaStore, err := os.ReadFile("./.test/store.json")
	if err != nil { t.Fatal(err) }

	schemaOffer, err := os.ReadFile("./.test/offer.json")
	if err != nil { t.Fatal(err) }

	var store bson.M
	if err := json.Unmarshal(schemaStore, &store); err != nil {
		t.Fatalf("Length: %d, First bytes: %x\n", len(schemaStore), schemaStore[:4])
	}

	var offer bson.M
	if err := json.Unmarshal(schemaOffer, &offer); err != nil {
		t.Fatalf("Length: %d, First bytes: %x\n", len(schemaOffer), schemaOffer[:4])
	}

	client := GetMongoClient()
	client.AddDefinition(store)
	client.AddDefinition(offer)

	// Test
	ctx := context.Background()

	// 2. Insert data
  offer1 := map[string]any {
    "name": "Offer 1",
  }

  _, err = client.InsertOne(ctx, "mydb", "offer", offer1)
  if err == nil {
  	t.Fatal("Should have required store")
  }
}

func TestInsertOne_Discriminate_EnsureStore(t *testing.T) {
	// 1. Register schemas
	schemaStore, err := os.ReadFile("./.test/store.json")
	if err != nil { t.Fatal(err) }

	schemaOffer, err := os.ReadFile("./.test/offer.json")
	if err != nil { t.Fatal(err) }

	var store bson.M
	if err := json.Unmarshal(schemaStore, &store); err != nil {
		t.Fatalf("Length: %d, First bytes: %x\n", len(schemaStore), schemaStore[:4])
	}

	var offer bson.M
	if err := json.Unmarshal(schemaOffer, &offer); err != nil {
		t.Fatalf("Length: %d, First bytes: %x\n", len(schemaOffer), schemaOffer[:4])
	}

	client := GetMongoClient()
	client.AddDefinition(store)
	client.AddDefinition(offer)

	// Test
	ctx := context.Background()
	ctx = context.WithValue(ctx, "store", "str_1234")

	// 2. Insert data
  offer1 := map[string]any {
    "name": "Offer 100",
    "store": "str_0000",
  }

  saved, err := client.InsertOne(ctx, "mydb", "offer", offer1)
  
  if saved["store"] != "str_1234" {
  	t.Fatal("Wrong store")
  }
}

func TestInsertOne_WithAudit(t *testing.T) {
	client := GetMongoClient()

	var (
		onAudit_calls int
		onAudit_after any
		onAudit_context any
	)

	cancel := client.Subscribe(func(audit *AuditResult) error {
		onAudit_calls++
		onAudit_after = audit.After
		onAudit_context = audit.Context

		return nil
	})
	
	data := map[string]any {
		"_id": "su_123458",
		"name": "MyNameTODelete",
		"age": int32(25),
	}
	
	o, err := client.InsertOne(context.Background(), "mydb", "mycollection", data)
	if err != nil { t.Fatalf("Failed to insertOne %#v", err) }

	cancel()

	ctx := context.Background()
	ctx = context.WithValue(ctx, "account", "xxxxxx")
	ctx = context.WithValue(ctx, "store", "str_4321")
	err = client.DeleteOne(ctx, "mydb", "mycollection", bson.M{"_id": o["_id"].(string)})
	if err != nil { t.Fatalf("Failed to deleteOne %#v", err) }

  // raw query
  var results bson.M
  filter := bson.M{ "name": "MyNameTODelete" }

  c := client.Client.Database("mydb").Collection("mycollection")
  c.FindOne(context.Background(), filter).Decode(&results)

  if results != nil {
  	t.Fatalf("Should have no results not %v", results)	
  }
  
  if onAudit_calls != 1 {
  	t.Fatalf("ondelete should have been called once, not %d", onAudit_calls)
  }

  if onAudit_after != nil {
  	tp := fmt.Sprintf("%T", onAudit_after)
  	
  	expectedType := fmt.Sprintf("%T", map[string]any{})
  	if tp != expectedType {
  		t.Fatalf("after has the wrong type %s", tp)
  	}

  	after := onAudit_after.(map[string]any)

  	AssertSubset(t, after, o, "Should have been equal")
  }

  if onAudit_context != nil {
  	ctx := onAudit_context.(map[string]any)
  	AssertSubset(t, ctx, map[string]any{"account": "xxxxxx", "store": "str_4321"}, "Should have been equal")
  }
}