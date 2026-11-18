package commons

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// AssertDeepEqual compares equality by comparing bytes.
func AssertDeepEqualStrict(t *testing.T, expected, actual any) {
	t.Helper()

	expJSON, err := bson.MarshalExtJSON(expected, true, false)
	if err != nil {
		t.Fatalf("marshal expected: %v", err)
	}

	actJSON, err := bson.MarshalExtJSON(actual, true, false)
	if err != nil {
		t.Fatalf("marshal actual: %v", err)
	}

	if diff := cmp.Diff(prettyJSON(t, expJSON), prettyJSON(t, actJSON)); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
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

func prettyJSON(t *testing.T, data []byte) string {
	t.Helper()

	var out bytes.Buffer
	if err := json.Indent(&out, data, "", "  "); err != nil {
		t.Fatalf("indent JSON: %v", err)
	}
	return out.String()
}
