package mongo

import (
	"fmt"
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/gsuntres/pkg/commons"
)

func TestUpdateInc_Default(t *testing.T) {
	client := GetMongoClient()
	client.SetRelaxed()
	defer client.SetStrict()

	toupdate := map[string]any { 
		"counter": int64(1),
		"counter2": int64(2),
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "account", "xxxxxx")
	ctx = context.WithValue(ctx, "store", "str_4321")
	filter := bson.M{
		"type": "counter",
		"name": "mycounter2",
	}
	after, err := client.UpdateInc(ctx, "mydb", "mycounter", filter, toupdate)
	if err != nil { t.Fatalf("Failed to update %#v", err) }

	counter, _ := commons.GetGeneric[int64](after, "counter")
	counter2, _ := commons.GetGeneric[int64](after, "counter2")

	if counter != 1 || counter2 != 2 {
		t.Fatalf("wrong counters %d, %d", counter, counter2)
	}
}

func TestUpdateInc_WithAudit(t *testing.T) {
	client := GetMongoClient()
	
	client.SetRelaxed()
	defer client.SetStrict()

	// Insert sample
	data := map[string]any {
		"_id": "xxxxxx",
		"type": "counter",
		"name": "mycounter",
		"counter": int64(0),
		"counter2": int64(25),
	}
	
	before, err := client.InsertOne(context.Background(), 
		"mydb", 
		"mycounter", 
		data,
	)
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

	toupdate := map[string]any { 
		"counter": int64(1),
		"counter2": int64(1),
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "account", "xxxxxx")
	ctx = context.WithValue(ctx, "store", "str_4321")
	filter := bson.M{
		"type": "counter",
		"name": "mycounter",
	}
	after, err := client.UpdateInc(ctx, "mydb", "mycounter", filter, toupdate)
	if err != nil { t.Fatalf("Failed to update %#v", err) }

	counter, _ := commons.GetGeneric[int64](after, "counter")
	counter2, _ := commons.GetGeneric[int64](after, "counter2")

	if counter != 1 || counter2 != 26 {
		t.Fatalf("Wrong counters counter: %d, counter2: %d", counter, counter2)
	}

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

  	delete(before, "updatedAt")
  	delete(before, "createdAt")

  	before_counter, _ := commons.GetGeneric[int64](before, "counter")
  	before_counter2, _ := commons.GetGeneric[int64](before, "counter2")
  	if before_counter != 0 && before_counter2 != 0 {
  		t.Fatalf("wrong counters %d, %d", before_counter, before_counter2)
  	}

  	delete(after, "createdAt")
  	delete(after, "updatedAt")
  	
  	after_counter, _ := commons.GetGeneric[int64](after, "counter")
  	after_counter2, _ := commons.GetGeneric[int64](after, "counter2")
  	if after_counter != 1 && after_counter2 != 26 {
  		t.Fatalf("wrong counters %d, %d", after_counter, after_counter2)
  	}
  } else {
  	t.Fatal("should have after")
  }

  if onAudit_context != nil {
  	ctx := onAudit_context.(map[string]any)
  	AssertSubset(t, 
  		ctx, 
  		map[string]any{
	  		"account": "xxxxxx", 
	  		"store": "str_4321",
	  	},
  		"Should have been equal",
  	)
  } else {
  	t.Fatal("should have context")
  }
}