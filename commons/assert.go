package commons

import (
	"testing"
	"bytes"
	
	"go.mongodb.org/mongo-driver/v2/bson"
)

// AssertDeepEqual compares equality by comparing bytes.
func AssertDeepEqualStrict(t *testing.T, expected, actual any) {
	t.Helper()

	expJSON, _ := bson.MarshalExtJSON(expected, true, true)
	actJSON, _ := bson.MarshalExtJSON(actual, true, true)

	if !bytes.Equal(expJSON, actJSON) {
	    t.Fatalf("expected %s, got %s", expJSON, actJSON)
	}
}

func AssertDeepEqual(t *testing.T, expected, actual any) {
	t.Helper()

	expJSON, err := bson.MarshalExtJSON(expected, true, true)
	if err != nil {
		t.Fatalf("marshal expected: %v", err)
	}

	actJSON, err := bson.MarshalExtJSON(actual, true, true)
	if err != nil {
		t.Fatalf("marshal actual: %v", err)
	}

	var expValue, actValue any
	if err := json.Unmarshal(expJSON, &expValue); err != nil {
		t.Fatalf("decode expected: %v", err)
	}
	if err := json.Unmarshal(actJSON, &actValue); err != nil {
		t.Fatalf("decode actual: %v", err)
	}

	if !reflect.DeepEqual(expValue, actValue) {
		t.Fatalf("expected %s, got %s", expJSON, actJSON)
	}
}
