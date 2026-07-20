package mongo

import (
	"os"
	"fmt"
	"context"
	"testing"
	"encoding/json"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestReplace(t *testing.T) {
	data := map[string]any {
		"_id": "su_2b3c00",
		"name": "Ava Thompson",
		"age": int32(26),
	}

	ctx := context.Background()

	// Insert to mycollection
	o, err := client.InsertOne(ctx, "mydb", "mycollection", data)
	if err != nil {	t.Fatalf("Failed to insertOne %#v", err) }

	id := o["_id"].(string)

	// Retrieve from mycollection
	fetched, err := client.GetOne(ctx, "mydb", "mycollection", id)
	if err != nil {	t.Fatalf("Failed to fetch %#v", err) }
	
	fetched["name"] = "Noah Patel"


	// Replace in mycollection
	replaced, err := client.Replace(ctx, "mydb", "mycollection", id, fetched)
	if err != nil {	t.Fatalf("Failed to replace %#v", err) }

	if replaced["_id"] != fetched["_id"] {
		t.Fatalf("Not the same entity")
	}

	if replaced["name"] != "Noah Patel" {
		t.Fatal("unexpected name")
	}
}

// Make sure discrimination field is used when replacing.
func TestReplace_Discrimination(t *testing.T) {
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

	// Save data
	ctx1 := context.Background()
	ctx1 = context.WithValue(ctx1, "account", "xxxxxx")
	ctx1 = context.WithValue(ctx1, "store", "str_1234")

	// Save offer
  offer1 := map[string]any { "name": "OSRAM 10" }
  saved, err := client.InsertOne(ctx1, "mydb", "offer", offer1)
  if err != nil { t.Fatalf("Failed to insertOne %#v", err) }

  id := saved["_id"].(string)

  // // Now replace it
  saved["name"] = "** OSRAM 10 CHANGED **"
	replaced, err := client.Replace(ctx1, "mydb", "offer", id, saved)
	if err != nil { t.Fatalf("Failed to replace %#v", err) }

	if replaced["name"] != "** OSRAM 10 CHANGED **" {
		t.Fatal("Unexpected replaced object")
	}

	found, err := client.GetOne(ctx1, "mydb", "offer", id)
	if err != nil { t.Fatalf("Failed to find %#v", err) }
	
	if found["name"] != "** OSRAM 10 CHANGED **" {
		t.Fatal("Updates did not persist")
	}
}

func TestReplace_Discrimination_NoStore(t *testing.T) {
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

	// Save data
	ctx1 := context.Background()
	ctx1 = context.WithValue(ctx1, "store", "str_1234")
	
	// Save offer
  offer1 := map[string]any { "name": "OSRAM 11" }
  saved, err := client.InsertOne(ctx1, "mydb", "offer", offer1)
  if err != nil { t.Fatalf("Failed to insertOne %#v", err) }

  id := saved["_id"].(string)

  // // Now replace it
  saved["name"] = "** OSRAM 11 CHANGED **"
	_, err = client.Replace(context.Background(), "mydb", "offer", id, saved)
	if err == nil { 
		t.Fatal("should have required store") 
	}

}

// Should save the right store regardless of the input
func TestReplace_Discrimination_EnsureStore(t *testing.T) {
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

	// Save data
	ctx1 := context.Background()
	ctx1 = context.WithValue(ctx1, "store", "str_1234")
	
	// Save offer
  offer1 := map[string]any { "name": "OSRAM 12" }
  saved, err := client.InsertOne(ctx1, "mydb", "offer", offer1)
  if err != nil { t.Fatalf("Failed to insertOne %#v", err) }

  id := saved["_id"].(string)

  // // Now replace it
  saved["store"] = "str_0000"
	changed, err := client.Replace(ctx1, "mydb", "offer", id, saved)
	if err != nil { t.Fatalf("Failed to replace %v", err) }
	if changed["store"] != "str_1234" { 
		t.Fatal("Should have ensured store") 
	}
}

func TestReplace_WithAudit(t *testing.T) {
	client := GetMongoClient()

	// Insert sample
	data := map[string]any {
		"_id": "su_123458",
		"name": "MyNameTODelete",
		"age": int32(25),
	}
	
	before, err := client.InsertOne(context.Background(), "mydb", "mycollection", data)
	if err != nil { t.Fatalf("Failed to insertOne %#v", err) }

	var (
		onAudit_calls int
		onAudit_data any
		onAudit_before any
		onAudit_after any
		onAudit_context any
	)

	cancel := client.Subscribe(func(audit *AuditResult) error {
		onAudit_calls++
		onAudit_data = audit.Data
		onAudit_before = audit.Before
		onAudit_after = audit.After
		onAudit_context = audit.Context

		return nil
	})
	
	toupdate := map[string]any { "name": "** CHANGED NAME **"	}
	
	ctx := context.Background()
	ctx = context.WithValue(ctx, "account", "xxxxxx")
	ctx = context.WithValue(ctx, "store", "str_4321")
	o, err := client.Replace(ctx, "mydb", "mycollection", "su_123458", toupdate)
	if err != nil { t.Fatalf("Failed to replace %#v", err) }
	
	cancel()
  
  if onAudit_calls != 1 {
  	t.Fatalf("ondelete should have been called once, not %d", onAudit_calls)
  }

  if onAudit_data != nil {
  	dta := onAudit_data.(bson.M)
  	AssertSubset(t, dta, toupdate, "Should have been equal")
  } else {
  	t.Fatal("should have data")
  }

  if onAudit_before != nil {
		bf := onAudit_before.(bson.M)
  	AssertSubset(t, bf, before, "Should have been equal")
  } else {
  	t.Fatal("should have before")
  }

  if onAudit_after != nil {
  	tp := fmt.Sprintf("%T", onAudit_after)
  	
  	expectedType := fmt.Sprintf("%T", map[string]any{})
  	if tp != expectedType {
  		t.Fatalf("after has the wrong type %s", tp)
  	}

  	after := onAudit_after.(map[string]any)

  	AssertSubset(t, after, o, "Should have been equal")
  } else {
  	t.Fatal("should have after")
  }

  if onAudit_context != nil {
  	ctx := onAudit_context.(map[string]any)
  	AssertSubset(t, ctx, map[string]any{"account": "xxxxxx", "store": "str_4321"}, "Should have been equal")
  } else {
  	t.Fatal("should have context")
  }
}