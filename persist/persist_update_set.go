package persist

import (
	"fmt"
	"context"
	"reflect"

	"github.com/gsuntres/pkg/commons"
	"github.com/gsuntres/pkg/mongo"
)

func BuildUpdateSet(col map[string]any, report *InitReport) {
	name := col["_name"].(string)
	singular := col["singular"].(string)

	// prepare input arguments and return results
	in := []reflect.Type{
		reflect.TypeOf((*context.Context)(nil)).Elem(),
		reflect.TypeOf((*map[string]any)(nil)).Elem(),
		reflect.TypeOf((*map[string]any)(nil)).Elem(),
	}
	out := []reflect.Type{
		reflect.TypeOf((*map[string]any)(nil)).Elem(),
		reflect.TypeOf((*error)(nil)).Elem(),
	}

	// create function signature
	variadic := false
	funcType := reflect.FuncOf(in, out, variadic)
	fname := fmt.Sprintf("%s%s%s", "Update", commons.StringTitle(singular), "Set")
	fields = append(fields, reflect.StructField{
		Name: fname,
		Type: funcType,
	})

	isSystem := false
	if v, ok := col["system"]; ok {
		isSystem = v.(bool)
	}

	report.AddField(fname)

	mc := mongo.GetMongoClient()

	// we defer function's implementation until we create the actual struct
	deferedFuncs[fname] = func(ctx context.Context, filter map[string]any, data map[string]any) (map[string]any, error) {
		db := "__undefined__"
		if isSystem {
			db = sysDb
		} else {
			account := ctx.Value("account").(string)
			if account != "" {
				db = mc.GetName(account)
			}
		}

		return mc.UpdateSet(ctx, db, name, filter, data)
	}
}
