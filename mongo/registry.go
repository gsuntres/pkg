package mongo

import (
	"reflect"
	"time"
	
	"github.com/go-viper/mapstructure/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TimeToStringHook(f reflect.Type, t reflect.Type, data any) (any, error) {

	// 1. Target must be a string
  if t.Kind() != reflect.String {
  	return data, nil
  }

  // 2. Check if source is time.Time OR *time.Time
  isTime := f == reflect.TypeOf(time.Time{})
  isTimePtr := f == reflect.TypeOf(&time.Time{})

  if isTime || isTimePtr {
	   // Handle pointer vs value during type assertion
	  if isTimePtr {
	    if ptr, ok := data.(*time.Time); ok && ptr != nil {
	  
	      return ptr.Format(time.RFC3339), nil
	    }
	
	    return "", nil
	  }
		
		return data.(time.Time).Format(time.RFC3339), nil
	}

  return data, nil
}

func StructToMap(input any) (map[string]any, error) {
  var result map[string]any
  
  config := &mapstructure.DecoderConfig{
    DecodeHook: TimeToStringHook,
    Result:     &result,
    TagName:    "bson",
  }

  decoder, err := mapstructure.NewDecoder(config)
  if err != nil {
     return nil, err
  }

  err = decoder.Decode(input)

  return result, err
}

func TruncatingTimeEncoder(ec bson.EncodeContext, vw bson.ValueWriter, val reflect.Value) error {
	// 1. Check if the type is exactly time.Time
	if val.Type() != reflect.TypeOf(time.Time{}) {
		// Fallback: Use the encoder from the current context's registry
		enc, err := ec.Registry.LookupEncoder(val.Type())
		if err != nil {
			return err
		}

		return enc.EncodeValue(ec, vw, val)
	}

	// 2. Perform the truncation logic
	t := val.Interface().(time.Time)
	truncated := t.Truncate(time.Millisecond)

	// 3. Write as a standard BSON DateTime
	return vw.WriteDateTime(truncated.UnixMilli())
}

func GetCustomRegistry() *bson.Registry {
	reg := bson.NewRegistry()

	// reg.RegisterTypeEncoder(
	// 	reflect.TypeOf(time.Time{}),
	// 	bson.ValueEncoderFunc(TruncatingTimeEncoder),
	// )
	// reg.SetRegistry(mgocompat.NewRegistry())

	return reg
}