package persist

import (
	"fmt"
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/gsuntres/pkg/commons"
	"github.com/gsuntres/pkg/mongo"
)

func BuildFind(col map[string]any, report *InitReport) {
	name := col["_name"].(string)
  plural := col["plural"].(string)

	// prepare input arguments and return results
	in := []reflect.Type{
		reflect.TypeOf((*context.Context)(nil)).Elem(),
		reflect.TypeOf((*map[string]any)(nil)).Elem(),
		reflect.TypeOf(&mongo.FindOptions{}),
	}
	out := []reflect.Type{
		reflect.TypeOf(bson.M{}),
		reflect.TypeOf((*error)(nil)).Elem(),
	}

	// create function signature
	variadic := false
	funcType := reflect.FuncOf(in, out, variadic)
  funcName := fmt.Sprintf("%s%s", "Find", commons.StringTitle(plural))
	fields = append(fields, reflect.StructField{
		Name: funcName,
		Type: funcType,
	})

	isSystem := false
	if v, ok := col["system"]; ok {
		isSystem = v.(bool)
	}

	report.AddField(funcName)

	mc := mongo.GetMongoClient()

	// we defer function's implementation until we create the actual struct
	deferedFuncs[funcName] = func(ctx context.Context, filter map[string]any, opts *mongo.FindOptions) (bson.M, error) {
		db := "__undefined__"
		if isSystem {
			db = sysDb
		} else {
			account := ctx.Value("account").(string)
			if account != "" {
				db = mc.GetName(account)
			}
		}

		return mc.Find(ctx, db, name, filter, opts)
	}
}

type FindResults struct {
	Data				[]map[string]any `json:"data"`
	Count 			int `json:"count"`
	HasMore 		bool `json:"has_more"`
	HasPrevious bool `json:"has_previous"`
	Offset 			int64 `json:"offset"`
	Limit 			int64 `json:"limit"`
}

func GetFindResults(m bson.M) *FindResults {
	o := &FindResults{}
	commons.BsonToStruct(m, o)

	return o
}