package commons

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// AssertDeepEqual compares equality by comparing bytes.
func AssertDeepEqualStrict(t *testing.T, expected, actual any, ignore ...string) {
	t.Helper()

	expJSON, err := bson.MarshalExtJSON(expected, true, false)
	if err != nil {
		t.Fatalf("marshal expected: %v", err)
	}

	actJSON, err := bson.MarshalExtJSON(actual, true, false)
	if err != nil {
		t.Fatalf("marshal actual: %v", err)
	}

	ignored := make(map[string]bool, len(ignore))
	for _, key := range ignore {
		ignored[key] = true
	}

	if diff := cmp.Diff(
		prettyJSON(t, expJSON, ignore...),
		prettyJSON(t, actJSON, ignore...),
	); diff != "" {
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

func prettyJSON(t *testing.T, data []byte, ignore ...string) string {
	t.Helper()

	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}

	ignored := make(map[string]bool, len(ignore))
	for _, key := range ignore {
		ignored[key] = true
	}
	removeFields(v, ignored)

	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("pretty-print JSON: %v", err)
	}
	return string(out)
}

func removeFields(v any, ignored map[string]bool) {
	switch x := v.(type) {
	case map[string]any:
		for key, value := range x {
			if ignored[key] {
				delete(x, key)
			} else {
				removeFields(value, ignored)
			}
		}
	case []any:
		for _, value := range x {
			removeFields(value, ignored)
		}
	}
}
