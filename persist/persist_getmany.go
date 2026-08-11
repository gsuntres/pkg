package persist

import (
	"fmt"
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/gsuntres/pkg/commons"
	"github.com/gsuntres/pkg/mongo"
)

func BuildGetMany(col map[string]any, report *InitReport) {
	name := col["_name"].(string)
	plural := col["plural"].(string)

	// prepare input arguments and return results
	in := []reflect.Type{
		reflect.TypeOf((*context.Context)(nil)).Elem(),
		reflect.TypeOf((*[]string)(nil)).Elem(),
	}
	out := []reflect.Type{
		reflect.TypeOf(bson.A{}),
		reflect.TypeOf((*error)(nil)).Elem(),
	}

	// create function signature
	variadic := false
	funcType := reflect.FuncOf(in, out, variadic)
  	funcName := fmt.Sprintf("%s%s", "Get", commons.StringTitle(plural))
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
	deferedFuncs[funcName] = func(ctx context.Context, ids []string) (bson.A, error) {
		db := "__undefined__"
		if isSystem {
			db = sysDb
		} else {
			account := ctx.Value("account").(string)
			if account != "" {
				db = mc.GetName(account)
			}
		}

    	return mc.GetByIds(ctx, db, name, ids)
	}
}
