package mongo

import (
	"testing"
	"time"
	"testing/synctest"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestEnsureId(t *testing.T) {
	data := map[string]any {
		"Name": "My Name Is",
	}

	ensureId(data, "")

	if id, okid := data["_id"]; !okid || id == "" {
		t.Fatal("Failed to add Id")
	}
}

func TestEnsureId_EmptyId(t *testing.T) {
	data := map[string]any {
		"_id": "myidxxxxxx",
		"name": "My Name Is",
	}

	ensureId(data, "")

	if data["_id"] == "" {
		t.Fatal("Id was updated")
	}
}

func TestEnsureId_ExistingId(t *testing.T) {
	data := map[string]any {
		"_id": "",
		"name": "My Name Is",
	}

	ensureId(data, "")

	if id, okid := data["_id"]; !okid || id == "" {
		t.Fatal("Failed to add Id")
	}
}

func TestEnsureCreatedAt(t *testing.T) {
	data := map[string]any {
		"name": "My Name Is",
	}

	synctest.Test(t, func(t *testing.T) {
		now := ensureCreatedAt(data)

		if createdAt, _ := data["createdAt"]; createdAt != now {
			t.Fatal("Failed to add CreatedAt")
		}
	})
}

func TestEnsureExistingCreatedAt(t *testing.T) {
	tm, err := time.Parse(time.RFC3339, "2025-12-24T13:00:11Z")
	if err != nil { t.Fatal(err) }

	// t.Fatalf("==> %T", tm)

	old := bson.M{"createdAt": tm}

	data := bson.M{}

	ensureExistingCreatedAt(data, old)

	if data["createdAt"] != tm {
		t.Fatal("wrong date")
	}
}

func TestEnsureExistingCreatedAt_NoField(t *testing.T) {
	old := bson.M{"name": "I have no created at"}

	data := bson.M{}

	ensureExistingCreatedAt(data, old)

	if _, ok := data["createdAt"]; ok {
		t.Fatal("Should had no createdAt")
	}
}