package mongo

import (
	"time"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestCursor_Default(t *testing.T) {
	now := time.Now()

	oid := bson.NewObjectIDFromTimestamp(now)

	okey := EncodeCursor(oid)

	id, err := DecodeCursor(okey)

	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if id != oid {
		t.Fatalf("Failed to decode id")
	}
}