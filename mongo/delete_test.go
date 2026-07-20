package mongo

import (
	"os"
	"fmt"
	"context"
	"testing"
	"encoding/json"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/gsuntres/pkg/commons"
)

func TestDeleteOne(t *testing.T) {
	client := GetMongoClient()

	data := map[string]any {
		"_id": "su_123458",
		"name": "MyNameTODelete",
		"age": int32(25),
	}

	o, err := client.InsertOne(context.Background(), "mydb", "mycollection", data)
	if err != nil { t.Fatalf("Failed to insertOne %#v", err) }

	// now delete it
	err = client.DeleteOne(context.Background(), "mydb", "mycollection", bson.M{"_id": o["_id"].(string)})
	if err != nil { t.Fatalf("Failed to deleteOne %#v", err) }

  // raw query
  var results bson.M
  filter := bson.M{ "name": "MyName" }

  c := client.Client.Database("mydb").Collection("mycollection")
  c.FindOne(context.Background(), filter).Decode(&results)

  if results != nil {
  	t.Fatalf("Should have no results not %v", results)	
  }
}

func TestDelete_Discriminator(t *testing.T) {
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

	// Store str_1234 has BOOMRAW 1.
	ctx1 := context.Background()
	ctx1 = context.WithValue(ctx1, "store", "str_1234")
  offer1 := map[string]any { "name": "BOOMRAW 1" }
  _, err = client.InsertOne(ctx1, "mydb", "offer", offer1)
  if err != nil { t.Fatalf("Failed to insertOne %#v", err) }

  // Store str_4321 has BOOMRAW 2
  ctx2 := context.Background()
	ctx2 = context.WithValue(ctx2, "store", "str_4321")
	offer2 := map[string]any { "name": "BOOMRAW 2" }
  _, err = client.InsertOne(ctx2, "mydb", "offer", offer2)
  if err != nil { t.Fatalf("Failed to insertOne %#v", err) }

  // Attempt to delete BOOMRAW2 in str_1234
  err = client.DeleteOne(ctx1, "mydb", "offer", bson.M{"name": "BOOMRAW2"})
	
	// Within store str_4321 searching for BOOMRAW should return one result
	filter := bson.M{"name": bson.M{"$regex": "BOOMRAW*"}}
	findResult, err := client.Find(ctx2, "mydb", "offer", filter, &FindOptions{	Offset: int64(0) })
	if err != nil { t.Fatalf("Failed to find %#v", err) }

	dataAny, hasData := findResult["data"]
	if !hasData { t.Fatal("no data") }

	data := dataAny.(bson.A)

	if len(data) != 1 {
		t.Fatalf("Expected to return 1 document but got %d", len(data))
	}

	arr, _ := commons.BsonAToSlice(data)

	o := arr[0]

	name := o["name"]

	if name != "BOOMRAW 2" {
	 	t.Fatalf("Expected BOOMRAW 2 not %s", name)
	}
}

func TestDeleteOne_WithAudit(t *testing.T) {
	client := GetMongoClient()

	data := map[string]any {
		"_id": "su_123458",
		"name": "MyNameTODelete",
		"age": int32(25),
	}

	o, err := client.InsertOne(context.Background(), "mydb", "mycollection", data)
	if err != nil { t.Fatalf("Failed to insertOne %#v", err) }

	// now delete it

	var (
		onAudit_calls int
		onAudit_before any
		onAudit_context any
	)
	
	cancel := client.Subscribe(func(audit *AuditResult) error {
		onAudit_calls++
		onAudit_before = audit.Before
		onAudit_context = audit.Context
		
		return nil
	})
	
	ctx := context.Background()
	ctx = context.WithValue(ctx, "account", "xxxxxx")
	ctx = context.WithValue(ctx, "store", "str_4321")
	err = client.DeleteOne(ctx, "mydb", "mycollection", bson.M{"_id": o["_id"].(string)})
	if err != nil { t.Fatalf("Failed to deleteOne %#v", err) }
	
	cancel()
  
  // raw query
  var results bson.M
  filter := bson.M{ "name": "MyName" }

  c := client.Client.Database("mydb").Collection("mycollection")
  c.FindOne(context.Background(), filter).Decode(&results)

  if results != nil {
  	t.Fatalf("Should have no results not %v", results)	
  }
  
  if onAudit_calls != 1 {
  	t.Fatalf("ondelete should have been called once, not %d", onAudit_calls)
  }

  if onAudit_before != nil {
  	tp := fmt.Sprintf("%T", onAudit_before)
  	if tp != "bson.M" {
  		t.Fatalf("before has the wrong type %s", tp)
  	}

  	before := onAudit_before.(bson.M)

  	AssertSubset(t, before, o, "Should have been equal")
  }

  if onAudit_context != nil {
  	ctx := onAudit_context.(map[string]any)
  	AssertSubset(t, ctx, map[string]any{"account": "xxxxxx", "store": "str_4321"}, "Should have been equal")
  }
}

func TestDeleteOne_WithAuditButIgnored(t *testing.T) {
	client := GetMongoClient()
	client.IgnoreAudit = []string{"mycollection"}

	data := map[string]any {
		"_id": "su_123458",
		"name": "MyNameTODelete",
		"age": int32(25),
	}

	o, err := client.InsertOne(context.Background(), "mydb", "mycollection", data)
	if err != nil { t.Fatalf("Failed to insertOne %#v", err) }

	// now delete it

	var (
		onAudit_calls int
		onAudit_before any
	)

	cancel := client.Subscribe(func(audit *AuditResult) error {
		onAudit_calls++
		onAudit_before = audit.Before

		return nil
	})
	
	ctx := context.Background()
	ctx = context.WithValue(ctx, "account", "xxxxxx")
	ctx = context.WithValue(ctx, "store", "str_4321")
	err = client.DeleteOne(ctx, "mydb", "mycollection", bson.M{"_id": o["_id"].(string)})
	if err != nil { t.Fatalf("Failed to deleteOne %#v", err) }
	
	cancel()
	client.IgnoreAudit = []string{"event"}

  // raw query
  var results bson.M
  filter := bson.M{ "name": "MyName" }

  c := client.Client.Database("mydb").Collection("mycollection")
  c.FindOne(context.Background(), filter).Decode(&results)

  if results != nil {
  	t.Fatalf("Should have no results not %v", results)	
  }

  if onAudit_calls != 0 {
  	t.Fatalf("ondelete should have not been called. [%d]", onAudit_calls)
  }

  if onAudit_before != nil {
  	tp := fmt.Sprintf("%T", onAudit_before)
  	if tp != "bson.M" {
  		t.Fatalf("before has the wrong type %s", tp)
  	}

  	before := onAudit_before.(bson.M)

  	AssertSubset(t, before, o, "Should have been equal")
  }
}