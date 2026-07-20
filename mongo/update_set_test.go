package mongo

import (
	"fmt"
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestUpdateSet_WithAudit(t *testing.T) {
	client := GetMongoClient()

	// Insert sample
	data := map[string]any {
		"_id": "su_122258",
		"name": "MyNameToUpdateSet",
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
	
	toupdate := map[string]any { "name": "** CHANGED USING UPDATE SET **"	}
	
	ctx := context.Background()
	ctx = context.WithValue(ctx, "account", "xxxxxx")
	ctx = context.WithValue(ctx, "store", "str_4321")
	ok, err := client.UpdateSet(ctx, "mydb", "mycollection", bson.M{"_id": "su_122258"}, toupdate)
	if err != nil { t.Fatalf("Failed to update %#v", err) }
	
	if !ok { t.Fatal("Should have updated") }

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
  	
  	expectedType := fmt.Sprintf("%T", bson.M{})
  	if tp != expectedType {
  		t.Fatalf("after has the wrong type %s", tp)
  	}

  	after := onAudit_after.(bson.M)

  	before["name"] = "** CHANGED USING UPDATE SET **"
  	delete(before, "updatedAt")
  	delete(after, "updatedAt")

  	AssertSubset(t, after, before, "Should have been equal")
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