package mongo

import (
	"encoding/base64"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func EncodeCursor(v any) string {
	var id string
	switch val := v.(type) {
	case bson.ObjectID:
		id = val.Hex()
	default:
		id = v.(string)
	}
	
	val := fmt.Sprintf("%s", id)

	return base64.RawURLEncoding.EncodeToString([]byte(val))
}

func DecodeCursor(okey string) (bson.ObjectID, error) {
	data, err := base64.StdEncoding.DecodeString(okey)

	if err != nil {
		return bson.NilObjectID, err 
	}

	oid, _ :=  bson.ObjectIDFromHex(string(data))

	return oid, nil
}


// func EncodeCursor(name string, v1 any, v2 any) string {
// 	var otherValue string
// 	var otherType string
	
// 	switch v1.(type) {
// 	case time.Time:
// 		otherType = "time.Time"
// 		otherValue = v1.(time.Time).Truncate(time.Millisecond).Format()
// 	case int:
// 		otherType = "time.Time"
// 		otherValue = v1.(time.Time).Truncate(time.Millisecond).Format()
// 	}

// 	var id string
// 	switch v2.(type) {
// 	case bson.ObjectID:
// 		id = v2.(bson.ObjectID).Hex()
// 	case string:
// 		id = v2.(string)
// 	default:
// 		id = v2.(string)
// 	}

// 	v := fmt.Sprintf("%v|%s", misc, id)

// 	fmt.Println(v, name)

// 	return base64.RawURLEncoding.EncodeToString([]byte(v))
// }

// func DecodeCursor(okey string) (name string, v1 any, v2 any) {
// 	data, err := base64.StdEncoding.DecodeString(okey)
// 	if err != nil {
// 		return "", nil, nil
// 	}

// 	parts := strings.Split(string(data), "|")
// 	if len(parts) != 3 {
// 		return "", nil, nil
// 	}
// }