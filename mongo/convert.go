package mongo

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	
  // "github.com/go-viper/mapstructure/v2"
)

func ToMap(data any) (bson.M, error) {
  // 1. Marshal the struct to BSON bytes
  b, err := bson.Marshal(data)
  if err != nil {
    return nil, err
  }

  // 2. Unmarshal the bytes back into a bson.M
  var res bson.M
  err = bson.Unmarshal(b, &res)

  return res, err
}

func ToStruct(m bson.M, target any) error {
  // 1. Convert map to BSON bytes
  data, err := bson.Marshal(m)
  if err != nil {
    return err
  }

  // 2. Unmarshal bytes into the target struct
  return bson.Unmarshal(data, target)
}

// func ToMap(input any) (bson.M, error) {
//   var result bson.M
//   config := &mapstructure.DecoderConfig{
//     TagName: "bson", // Use bson tags instead of field names
//     Result:  &result,
//   }

//   decoder, err := mapstructure.NewDecoder(config)
//   if err != nil {
//     return nil, err
//   }

//   err = decoder.Decode(input)
//   return result, err
// }

// func BsonToStructHook() mapstructure.DecodeHookFunc {
//   return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
//     // 1. Convert primitive.ObjectID -> string
//     if f == reflect.TypeOf(primitive.ObjectID{}) && t.Kind() == reflect.String {
//       return data.(primitive.ObjectID).Hex(), nil
//     }

//     // 2. Convert primitive.DateTime -> time.Time
//     if f == reflect.TypeOf(primitive.DateTime(0)) && t == reflect.TypeOf(time.Time{}) {
//       return data.(primitive.DateTime).Time(), nil
//     }

//     return data, nil
//   }
// }