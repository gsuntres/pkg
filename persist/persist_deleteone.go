package persist

import (
	"fmt"
	"context"
	"reflect"

	"github.com/gsuntres/pkg/commons"
	"github.com/gsuntres/pkg/mongo"
)

func BuildDeleteOne(col map[string]any, report *InitReport) {
	name := col["_name"].(string)
	singular := col["singular"].(string)

	// prepare input arguments and return results
	in := []reflect.Type{
		reflect.TypeOf((*context.Context)(nil)).Elem(),
		reflect.TypeOf((*map[string]any)(nil)).Elem(),
	}
	out := []reflect.Type{
		reflect.TypeOf((*error)(nil)).Elem(),
	}

	// create function signature
	variadic := false
	funcType := reflect.FuncOf(in, out, variadic)
  deleteOneName := fmt.Sprintf("%s%s", "Delete", commons.StringTitle(singular))
	fields = append(fields, reflect.StructField{
    Name: deleteOneName,
		Type: funcType,
	})

	isSystem := false
	if v, ok := col["system"]; ok {
		isSystem = v.(bool)
	}

	report.AddField(deleteOneName)

	mc := mongo.GetMongoClient()

	// we defer function's implementation until we create the actual struct
  deferedFuncs[deleteOneName] = func(ctx context.Context, filter map[string]any) error {
		db := "__undefined__"
		if isSystem {
			db = sysDb
		} else {
			account := ctx.Value("account").(string)
			if account != "" {
				db = mc.GetName(account)
			}
		}

    return mc.DeleteOne(ctx, db, name, filter)
	}
}
