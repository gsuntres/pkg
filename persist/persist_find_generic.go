package persist

import (
	"fmt"
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/gsuntres/pkg/mongo"
)

func BuildFindGeneric() {
	// prepare input arguments and return results
	in := []reflect.Type{
		reflect.TypeOf((*context.Context)(nil)).Elem(),
		reflect.TypeOf((*string)(nil)).Elem(),
		reflect.TypeOf((*string)(nil)).Elem(),
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
  funcName := fmt.Sprintf("%s", "Find")
	fields = append(fields, reflect.StructField{
		Name: funcName,
		Type: funcType,
	})

	mc := mongo.GetMongoClient()

	// we defer function's implementation until we create the actual struct
	deferedFuncs[funcName] = func(ctx context.Context, db string, name string, filter map[string]any, opts *mongo.FindOptions) (bson.M, error) {
		return mc.Find(ctx, db, name, filter, opts)
	}
}
