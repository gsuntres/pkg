package mongo

import (
	"fmt"
	"testing"
	"time"
	"context"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

const SAMPLE_DB = "mydb"
const SAMPLE_COLLECTION = "mycollection"

// AssertSubset compares a subset map against a superset (from Mongo).
func AssertSubset(t *testing.T, subset, superset map[string]any, msgAndArgs ...any) bool {
	t.Helper()

	// 1. Filter: Ignore keys in the superset that aren't in our expected subset
	ignoreExtra := cmpopts.IgnoreMapEntries(func(k string, v any) bool {
		_, ok := subset[k]
		return !ok
	})

	// 2. Transformer: Convert Mongo v2 int64 back to time.Time for comparison
	// 1. Define the logic
	timeLogic := func(v any) any {
		switch val := v.(type) {
    case time.Time:
        return val.UTC()
    case bson.DateTime:
        return val.Time().UTC()
    case int64:
        return time.UnixMilli(val).UTC()
    default:
        return time.Time{}
    }
	}

	// 2. Wrap it in a Filter so it's not an "unfiltered" option
	timeTransform := cmp.FilterValues(func(x, y any) bool {
	  _, xisbsontime := x.(bson.DateTime)
    _, yistime := y.(time.Time)
    _, xistime := x.(time.Time)
    _, yisbsontime := y.(bson.DateTime)

    return (xisbsontime && yistime) || (yisbsontime && xistime)
	}, cmp.Transformer("BsonTime", timeLogic))

	// 3. Execute Diff
	diff := cmp.Diff(subset, superset,
		ignoreExtra,
		timeTransform,
		cmpopts.EquateEmpty(), // Treats nil slice == empty slice
	)

	if diff != "" {
		msg := "Maps do not match (subset != superset)"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		t.Errorf("%s:\n%s", msg, diff)

		return false
	}

	return true
}

func LoadTestSample() error {
	// Read JSON file
	jsonData, _ := os.ReadFile("./.test/sample.json")

	var docs []bson.M
	if err := bson.UnmarshalExtJSON(jsonData, true, &docs); err != nil {
	  return err
	}

	// Insert into MongoDB
	client := GetMongoClient()
	collection := client.Client.Database(SAMPLE_DB).Collection(SAMPLE_COLLECTION)
	if _, err := collection.InsertMany(context.TODO(), docs); err != nil {
		return err
	}

	return nil
}
