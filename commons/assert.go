package commons

import (
	"testing"
	"bytes"
	
	"go.mongodb.org/mongo-driver/v2/bson"
)

// AssertDeepEqual compares equality by comparing bytes.
func AssertDeepEqual(t *testing.T, expected, actual any) {
	t.Helper()

	expJSON, _ := bson.MarshalExtJSON(expected, true, true)
	actJSON, _ := bson.MarshalExtJSON(actual, true, true)

	if !bytes.Equal(expJSON, actJSON) {
	    t.Fatalf("expected %s, got %s", expJSON, actJSON)
	}
}
