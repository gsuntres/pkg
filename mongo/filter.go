package mongo

import (
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Query struct {
	Filter 							  map[string]any `json:"filter"`
	WithDiscriminator 		bool
	DiscriminatorField    string
}

type Filter struct {
	Name string `json:"name"`
	Op string `json:"op"`
	Value string `json:"value"`
}

func makeFilter(name string, value any) *Filter {
	var op string

	var v string

	switch val := value.(type) {
	case string:
		op = "eq"
		v = val
	case bson.M:
		for kk, vv := range val {
			op = kk

			switch vvt := vv.(type) {
			case string:
				v = vvt
			case time.Time:
				v = vvt.UTC().Format(time.RFC3339)
			}
			
			break
		}
		
	case map[string]any:
		for kk, vv := range val {
			op = kk
			v = vv.(string)

			break
		}
		
	default:
		op = "eq"
	}

	o := &Filter{
		Name: name,
		Op: op,
		Value: v,
	}

	return o
}

func Mongofy(q *Query) map[string]any {

	alwaysAnd := make([]map[string]interface{}, 0)

	conditions := make([]map[string]interface{}, 0)

	logic := "and"

	for k, v := range q.Filter {
		if k == "_logic" {
			logic = v.(string)

			continue
		}

		f := makeFilter(k, v)

		mongoOp := mapToOp(f.Op)

		// check discriminator field
		if q.WithDiscriminator && q.DiscriminatorField == k {
			alwaysAnd = append(alwaysAnd, map[string]any{ k: map[string]any{ mongoOp: f.Value }})
		} else {
			conditions = append(conditions, map[string]any{ k: map[string]any{ mongoOp: f.Value }})	
		}		
	}

	if len(alwaysAnd) == 0 && len(conditions) == 0 {
		return map[string]any{}
	}

	if logic == "or" {
		if len(alwaysAnd) != 0 {
			wrapped := make([]map[string]interface{}, 0)
			wrapped = append(wrapped, alwaysAnd...)
			wrapped = append(wrapped, map[string]any{
				"$or": conditions,
			})
			
			return map[string]any {
				"$and": wrapped,
			}
		}

		return map[string]any{
			"$or": conditions,
		}
	}

	wrapped := make([]map[string]interface{}, 0)
	wrapped = append(wrapped, alwaysAnd...)
	wrapped = append(wrapped, conditions...)

	return map[string]any{
		"$and": wrapped,
	}
}

func mapToOp(v string) string {
	switch v {
	case "eq":
		return "$eq"
	case "le":
		return "$le"
	case "includes": 
		return "$regex"
	default:
		log.Printf("WARN: no conversion for %v", v)
		return v
	}
}