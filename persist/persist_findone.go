package persist

import (
	"fmt"
	"context"
	"reflect"
	
	"go.mongodb.org/mongo-driver/v2/bson"
	v2options "go.mongodb.org/mongo-driver/v2/mongo/options"
	
	"github.com/gsuntres/pkg/options"
	"github.com/gsuntres/pkg/commons"
	"github.com/gsuntres/pkg/mongo"
)

func BuildFindOne(col map[string]any, report *InitReport) {
	name := col["_name"].(string)
	singular := col["singular"].(string)

	// prepare input arguments and return results
	optionsType := reflect.TypeOf((*options.FindOneOptionsBuilder)(nil)).Elem()

	in := []reflect.Type{
		reflect.TypeOf((*context.Context)(nil)).Elem(),
		reflect.TypeOf((*map[string]any)(nil)).Elem(),
		reflect.SliceOf(optionsType),
	}

	out := []reflect.Type{
		reflect.TypeOf(bson.M{}),
		reflect.TypeOf((*error)(nil)).Elem(),
	}

	// create function signature
	variadic := true
	funcType := reflect.FuncOf(in, out, variadic)
	funcName := fmt.Sprintf("%s%s", "Find", commons.StringTitle(singular))
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
	deferedFuncs[funcName] = func(ctx context.Context, filter map[string]any, opts... options.FindOneOptionsBuilder) (bson.M, error) {
		db := "__undefined__"
		if isSystem {
			db = sysDb
		} else {
			var account string
			accountAny := ctx.Value("account")
			if accountAny != nil {
				account = accountAny.(string)
			}
			if account != "" {
				db = mc.GetName(account)
			}
		}

		var op options.FindOneOptions
		if len(opts) != 0 {
			for _, optfn := range opts[0].Opts {
		  	err := optfn(&op)
		    if err != nil {
		      return nil, fmt.Errorf("failed to run options: %w", err)
		    }
		  }
		}		
		
		mOptions := v2options.FindOne()
	  
	  if op.Projection != nil {
	  	mOptions.SetProjection(op.Projection)
	  }

		return mc.FindOne(ctx, db, name, filter, mOptions)
	}
}
