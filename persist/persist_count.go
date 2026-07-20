package persist

import (
	"fmt"
	"context"
	"reflect"

	"github.com/gsuntres/pkg/commons"
	"github.com/gsuntres/pkg/mongo"
)

func BuildCount(col map[string]any, report *InitReport) {
	name := col["_name"].(string)
  plural := col["plural"].(string)

	// prepare input arguments and return results
	in := []reflect.Type{
		reflect.TypeOf((*context.Context)(nil)).Elem(),
		reflect.TypeOf((*map[string]any)(nil)).Elem(),
	}
	out := []reflect.Type{
		reflect.TypeOf(int64(0)),
		reflect.TypeOf((*error)(nil)).Elem(),
	}

	// create function signature
	variadic := false
	funcType := reflect.FuncOf(in, out, variadic)
  funcName := fmt.Sprintf("%s%s", "Count", commons.StringTitle(plural))
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
	deferedFuncs[funcName] = func(ctx context.Context, filter map[string]any) (int64, error) {
		db := "__undefined__"
		if isSystem {
			db = sysDb
		} else {
			account, ok := ctx.Value("account").(string)
			if ok && commons.StringIsNotBlank(account) {
				db = mc.GetName(account)
			}
		}

		return mc.Count(ctx, db, name, filter)
	}}
