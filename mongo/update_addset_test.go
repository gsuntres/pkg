package mongo

import (
	"fmt"
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestUpdateAddSet_WithAudit(t *testing.T) {
	client := GetMongoClient()

	// Insert sample
	data := map[string]any {
		"_id": "su_122228",
		"names": bson.A{},
	}
	
	_, err := client.InsertOne(context.Background(), "mydb", "mycollection", data)
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
	
	toadd := bson.A{ "name 1", "name 2", "name 3", "name 1"}
	
	ctx := context.Background()
	ctx = context.WithValue(ctx, "account", "xxxxxx")
	ctx = context.WithValue(ctx, "store", "str_4321")
	ok, err := client.UpdateAddToSet(ctx, "mydb", "mycollection", bson.M{"_id": "su_122228"}, "names", toadd)
	if err != nil { t.Fatalf("Failed to update %#v", err) }
	
	if !ok { t.Fatal("Should have updated") }

	cancel()
  
	if onAudit_calls != 1 {
		t.Fatalf("ondelete should have been called once, not %d", onAudit_calls)
	}

	if onAudit_data != nil {
		dta := onAudit_data.([]any)
		if len(dta)  != len(toadd) {
			t.Fatal("wrong data")	
		}
	} else {
		t.Fatal("should have data")
	}

	if onAudit_before == nil {
		t.Fatal("should have before")
	}

	if onAudit_after != nil {
		tp := fmt.Sprintf("%T", onAudit_after)
		
		expectedType := fmt.Sprintf("%T", bson.M{})
		if tp != expectedType {
			t.Fatalf("after has the wrong type %s", tp)
		}

		doc := onAudit_after.(bson.M)
		delete(doc, "_id")
		delete(doc, "createdAt")
		delete(doc, "updatedAt")

		AssertSubset(t, doc, map[string]any{"names": bson.A{"name 1", "name 2", "name 3"}})
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