package commons

import (
	"log"
	"reflect"
	"encoding/json"
)

// JsonSprintError returns the Marshal/Unmarshal error with some helpful information.
func JsonSprintError(err error, content []byte) (error, string) {
	if syntaxErr, ok := err.(*json.SyntaxError); ok {
		totalLen := int64(len(content))
		
		min := MathMin(int64(50), syntaxErr.Offset)

		from := MathMax(syntaxErr.Offset - min, 0)
		to := MathMin(syntaxErr.Offset + min, totalLen)

		sliceBuf := content[from:to]

		return err, string(sliceBuf)
	} else {
		
		return err, ""
	}
}

func JsonPrintError(err error, content []byte) {
	e, s := JsonSprintError(err, content)
	
	log.Printf("Failed to unmarshal %#v \n %s", e, s)
} 

// JsonEqual check if two json string representations are equal.
func JsonEqual(a, b string) bool {
	var o1 any
	var o2 any

	if err := json.Unmarshal([]byte(a), &o1); err != nil {
		return false
	}

	if err := json.Unmarshal([]byte(b), &o2); err != nil {
		return false
	}

	return reflect.DeepEqual(o1, o2)
}

// JsonPrint bson.M
func JsonSprint(a any) string {
	raw, err := json.Marshal(a)
	if err != nil {
		log.Printf("Failed to read json: %s", err)

		return ""
	}

	return string(raw)	
}

func JsonPrint(a any) {
	log.Printf("%s", JsonSprint(a))
}

func JsonToMap(s string) (map[string]any, error) {
	b := []byte(s)

	var schema map[string]any
	if err := json.Unmarshal(b, &schema); err != nil {
		log.Printf("Failed to parse json: %v", err)

		return nil, err
	}

	return schema, nil
}