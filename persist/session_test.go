package persist

import (
	"fmt"
	"errors"
	"context"
	"testing"

	v2mongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestWithiSession(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "account", "xxxxxx")
	ctx = context.WithValue(ctx, "store", "str_4321")

	_, err := WithinSession(ctx, func(sessCtx context.Context, p any) (any, error) {
		item := bson.M{
			"name": "Item-98a541e2-73a0-11f1-942b-2748d6589f29",
		}
		results, err := Call(p, "InsertItem", sessCtx, item)
		if err != nil { return nil, fmt.Errorf("call failed: %w", err) }

		if !results[1].IsNil() {
	    	err := results[1].Interface().(error)
			return nil, fmt.Errorf("insert failed %w", err)
		}

		return nil, nil
	})

	if err != nil { t.Fatalf("failed: %v", err)	}

	p := GetCurrent()
	filter := bson.M{
		"name": "Item-98a541e2-73a0-11f1-942b-2748d6589f29",
	}
	results, err := Call(p, "FindItem", ctx, filter)
	if err != nil { t.Fatalf("failed to call %v", err) }

	if !results[1].IsNil() {
		errResults := results[1].Interface().(error)
		t.Fatalf("failed to find: %v", errResults)
	}

	found := results[0].Interface().(bson.M)
	name, _ := found["name"]
	if name != "Item-98a541e2-73a0-11f1-942b-2748d6589f29" {
		t.Fatalf("should have saved %s", name)
	}
}

func TestWithiSession_Abort(t *testing.T) {
	// 1. Save within transaction
	ctx := context.Background()
	ctx = context.WithValue(ctx, "account", "xxxxxx")
	ctx = context.WithValue(ctx, "store", "str_4321")

	var ErrFake error = errors.New("oups")

	_, err := WithinSession(ctx, func(sessCtx context.Context, p any) (any, error) {
		item := bson.M{
			"name": "Item-3f52d6ea-73b4-11f1-9fb7-7f96dc3113bb",
		}
		results, err := Call(p, "InsertItem", sessCtx, item)
		if err != nil { return nil, fmt.Errorf("call failed: %w", err) }

		if !results[1].IsNil() {
	    err := results[1].Interface().(error)

			return nil, fmt.Errorf("insert failed %w", err)
		}

		return nil, ErrFake
	})

	if !errors.Is(err, ErrFake) {
		t.Fatalf("Wrong error: %v", err)
	}

	// 2. Check if saved
	p := GetCurrent()
	filter := bson.M{
		"name": "Item-3f52d6ea-73b4-11f1-9fb7-7f96dc3113bb",
	}
	results, err := Call(p, "FindItem", ctx, filter)
	if err != nil { t.Fatalf("failed to call %v", err) }

	if !results[1].IsNil() {
		errResults := results[1].Interface().(error)
		if errResults != nil && !errors.Is(errResults, v2mongo.ErrNoDocuments) {
			t.Fatalf("Should not have saved: %s", errResults)
		}
	}
}