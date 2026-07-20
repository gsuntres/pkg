package persist

import (
	"fmt"
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/gsuntres/pkg/commons"
	"github.com/gsuntres/pkg/mongo"
)

func BuildInsertOne(col map[string]any, report *InitReport) {
	name := col["_name"].(string)
	singular := col["singular"].(string)

	// prepare input arguments and return results
	in := []reflect.Type{
		reflect.TypeOf((*context.Context)(nil)).Elem(),
		reflect.TypeOf((*map[string]any)(nil)).Elem(),
	}
	out := []reflect.Type{
		reflect.TypeOf(bson.M{}),
		reflect.TypeOf((*error)(nil)).Elem(),
	}

	// create function signature
	variadic := false
	funcType := reflect.FuncOf(in, out, variadic)
	insertOneName := fmt.Sprintf("%s%s", "Insert", commons.StringTitle(singular))
	fields = append(fields, reflect.StructField{
		Name: insertOneName,
		Type: funcType,
	})

	isSystem := false
	if v, ok := col["system"]; ok {
		isSystem = v.(bool)
	}

	report.AddField(insertOneName)

	mc := mongo.GetMongoClient()

	// we defer function's implementation until we create the actual struct
	deferedFuncs[insertOneName] = func(ctx context.Context, data map[string]any) (bson.M, error) {
		db := "__undefined__"
		if isSystem {
			db = sysDb
		} else {
			accountAny := ctx.Value("account")
			if accountAny == nil {
				return nil, fmt.Errorf("account required for %s", name)
			}
			account := accountAny.(string)
			if account != "" {
				db = mc.GetName(account)
			}
		}

		return mc.InsertOne(ctx, db, name, data)
	}
}
