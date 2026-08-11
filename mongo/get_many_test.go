package mongo

import (
  	"context"
	"testing"

	"github.com/gsuntres/pkg/commons"
)

func TestGetByIds(t *testing.T) {
	client := GetMongoClient()

	data := []map[string]any {
		map[string]any{
			"_id": "su_xxxxx1",
			"name": "MyNameGetOne",
			"age": int32(25),
		}, map[string]any{
			"_id": "su_xxxxx2",
			"name": "MyNameGetTwo",
			"age": int32(26),
		}, map[string]any{
			"_id": "su_xxxxx3",
			"name": "MyNameGetThree",
			"age": int32(6),
		},
	}

	for _, o := range data {
		_, err := client.InsertOne(context.Background(), "mydb", "mycollection", o)
		if err != nil {
			t.Fatalf("Failed to insertOne %#v", err)
		}
	}

	out, err := client.GetByIds(context.Background(), "mydb", "mycollection", []string{"su_xxxxx1", "su_xxxxx3"})
	if err != nil {
		t.Fatalf("Failed to get by ids %#v", err)
	}

	commons.AssertDeepEqual(t, out, data)
}

