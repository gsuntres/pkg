package persist

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func addItems(t *testing.T) {
	names := []string{
		"item1",
		"item2",
		"item3",
		"item4",
	}

	p := GetCurrent()
	ctx := context.Background()
	ctx = context.WithValue(ctx, "account", "xxxxxx")

	for _, item := range names {
		data := bson.M{
			"name": item,
		}
		results, err := Call(p, "InsertItem", ctx, data)
		if err != nil {
			t.Fatalf("Failed to call %v", err)
		}
		if !results[1].IsNil() {
	    err := results[1].Interface().(error)

			t.Fatalf("Call failed %v", err)
		}
	}
}

func TestCount(t *testing.T) {
	addItems(t)

	p := GetCurrent()
	ctx := context.Background()
	ctx = context.WithValue(ctx, "account", "xxxxxx")
	results, err := Call(p, "CountItems", ctx, bson.M{})
	if err != nil {
		t.Fatalf("Failed to call %v", err)
	}
	if !results[1].IsNil() {
    err := results[1].Interface().(error)

		t.Fatalf("Call failed %v", err)
	}

	count := results[0].Interface().(int64)
	if count != 4 {
		t.Fatalf("Wrong count %d", count)
	}
}