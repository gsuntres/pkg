package commons

import (
	"log"
	"reflect"
	"strings"
	"bytes"
	"encoding/json"

	"github.com/go-viper/mapstructure/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// StructHasStringValue given any struct will check if the requested property has
// a non-blank value.
func StructHasStringValue(o any, field string) bool {
	v := reflect.ValueOf(o).Elem().FieldByName(field)

	return v.String() != ""
}

// StructHasNotStringValue given any struct will check if the requested property has
// a blank value.
func StructHasNotStringValue(o any, field string) bool {
	return !StructHasStringValue(o, field)
}

// StructHasProperty cheks if a struct has the requested property.
func StructHasProperty(value interface{}, name string) bool {
	vo := reflect.ValueOf(value).Elem()
	
	typeOfValue := vo.Type()

	has := false
	
	for i:= 0; i < vo.NumField(); i++ {
		if typeOfValue.Field(i).Name == name {
			has = true
			break
		}
	}

	return has
}

func MapPick[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	out := make(map[K]V, len(keys))
	for _, k := range keys {
		if v, ok := m[k]; ok {
			out[k] = v
		}
	}

	return out
}

func MapOmit[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	out := make(map[K]V, len(m))

	omit := make(map[K]struct{}, len(keys))
	for _, k := range keys {
		omit[k] = struct{}{}
	}

	for k, v := range m {
		if _, ok := omit[k]; !ok {
			out[k] = v
		}
	}

	return out
}

// MapMerge merges maps into one.
func MapMerge[T ~map[string]any](maps...T) T {
	out := make(T)

	for _, m := range maps {
		if m != nil {
			for k, v := range m {
				out[k] = v
			}	
		}		
	}

	return out
}

// StructMustTomap given any struct return the equivalent map[string]any or nil.
// Will never throw. Will also work for map[string]any.
func StructMustToMap(data any) map[string]any {
	b, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed marshal %v", err)

		return nil
	}

	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()

	var res map[string]any
	err = dec.Decode(&res)
	if err != nil {
		log.Printf("Failed to unmarshal %v", err)

		return nil
	}

	return res
}

// StructSetValue will update the value of the given field of struct o.
func StructSetValue(o any, field string, value any) {
	ref := reflect.ValueOf(o).Elem()

  if ref.Kind() == reflect.Ptr {
    ref = reflect.Indirect(ref)
	}

	if ref.Kind() == reflect.Interface {
    ref = ref.Elem()
	}

	if ref.Kind() == reflect.Struct {
		f := ref.FieldByName(field)

		if f.IsValid() && f.CanSet() {
			f.Set(reflect.ValueOf(value))
		}
	}
}

func StructToStruct(source any, target any) {
	mapstructure.Decode(source, &target)
}

// StructToMapRecursive given a struct or a primitive will return the equivalent
// map[string]any of the struct or the primitive as is.
func StructToMapRecursive(obj any) any {
	v := reflect.ValueOf(obj)

	// Handle pointers by getting the underlying element
	if v.Kind() == reflect.Ptr {
		if v.IsNil() { return nil }
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		result := make(map[string]any)
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			// Skip unexported fields (private fields)
			if field.PkgPath != "" { continue }

			// Recurse into the field's value
			result[field.Name] = StructToMapRecursive(v.Field(i).Interface())
		}

		return result

	case reflect.Slice, reflect.Array:
		result := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			result[i] = StructToMapRecursive(v.Index(i).Interface())
		}

		return result

	default:
		// Return basic types (int, string, etc.) as is
		return obj
	}
}

func BsonClone(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}

	out := make(map[string]any, len(m))

	for k, v := range m {
		out[k] = deepCopy(v)
	}

	return out
}

func deepCopy(v any) any {
	switch val := v.(type) {
	case bson.M:
		return BsonClone(map[string]any(val))
	case bson.A:
		arr := make([]any, len(val))
		for i, item := range val {
			arr[i] = deepCopy(item)
		}
		return arr
	case bson.D:
		d := make(bson.D, len(val))
		for i, elem := range val {
			d[i] = bson.E{
				Key:   elem.Key,
				Value: deepCopy(elem.Value),
			}
		}
		return d
	case map[string]any: // bson.M
		return BsonClone(val)
	case []any: // bson.A
		arr := make([]any, len(val))
		for i, item := range val {
			arr[i] = deepCopy(item)
		}
		return arr

	default:
		// primitives (string, int, bool, etc.)
		return val
	}
}

func BsonDFromSlice(in [][]any) bson.D {
	if in == nil {
		return bson.D{}
	}

	var doc bson.D

  	for _, pair := range in {
      if len(pair) != 2 {
          continue
      }

      key, ok := pair[0].(string)
      if !ok {
          continue // or error
      }

      doc = append(doc, bson.E{
          Key:   key,
          Value: pair[1], // already `any`
      })
  	}

  	return doc
}

func BsonDGetAny(d bson.D, key string) (any, bool) {
	for _, e := range d {
		if e.Key == key {
			return e.Value, true
		}
	}

	return nil, false
}

func BsonDGet[T any](d bson.D, key string) (T, bool) {
	var zero T

	for _, e := range d {
		if e.Key == key {
			val, ok := e.Value.(T)
			if !ok {
				return zero, false
			}

			return val, true
		}
	}

	return zero, false
}

func BsonDGetString(d bson.D, key string) (string, bool) {
	return BsonDGet[string](d, key)
}

func BsonToStruct(m bson.M, o any) error {
	b, err := bson.Marshal(m)
	if err != nil {
		log.Printf("Failed marshal %v", err)

		return err
	}

	err = bson.Unmarshal(b, o)
	if err != nil {
		log.Printf("Failed to unmarshal %v", err)

		return err
	}

	return nil
}

// MapToStruct will convert a map[string]any to a struct.
func MapToStruct(m map[string]any, o any) error {
	b, err := json.Marshal(m)
	if err != nil {
		log.Printf("Failed marshal %v", err)

		return err
	}

	err = json.Unmarshal(b, o)
	if err != nil {
		log.Printf("Failed to unmarshal %v", err)

		return err
	}

	return nil
}

func BsonAnyToMap(v any) any {
	switch val := v.(type) {
	case bson.D:
		m := map[string]any{}
		for _, elem := range val {
			m[elem.Key] = BsonAnyToMap(elem.Value)
		}
		return m
	case bson.M:
		m := map[string]any{}
		for k, v2 := range val {
			m[k] = BsonAnyToMap(v2)
		}
		return m
	case map[string]any:
		m := map[string]any{}
		for i, v2 := range val {
			m[i] = BsonAnyToMap(v2)
		}
		return m
	case bson.A:
		arr := make([]any, len(val))
		for i, v2 := range val {
			arr[i] = BsonAnyToMap(v2)
		}
		return arr

	default:
		return v
	}

	return v
}

func BsonToMap(b bson.M) map[string]any {
	result := map[string]any{}

	for k, v := range b {
		switch val := v.(type) {
		case bson.M:
			result[k] = BsonToMap(val)
		case []interface{}:
			arr := make([]any, len(val))
			for i, item := range val {
				if nested, ok := item.(bson.M); ok {
					arr[i] = BsonToMap(nested)
				} else {
					arr[i] = item
				}
			}
			result[k] = arr
		default:
			result[k] = val
		}
	}

	return result
}

type WrapA struct {
	Items []map[string]any
}

// BsonAToSlice will convert a bson.A to []map[string]any.
func BsonAToSlice(m any) ([]map[string]any, error) {

	v := bson.M{
		"items": m.(bson.A),
	}

	b, err := bson.Marshal(v)
	if err != nil {
		log.Printf("Failed marshal %v", err)

		return nil, err
	}

	var o WrapA
	err = bson.Unmarshal(b, &o)
	if err != nil {
		log.Printf("Failed to unmarshal %v", err)

		return nil, err
	}

	return o.Items, nil
}

func MapFromString(s string) (map[string]any, error) {
	var b map[string]any
	if err := bson.Unmarshal([]byte(s), &b); err != nil {
		return nil, err
	}

	return b, nil
}

func MapFromStringExt(s string) (map[string]any, error) {
	var b map[string]any
	if err := bson.UnmarshalExtJSON([]byte(s), true, &b); err != nil {
		return nil, err
	}

	return b, nil
}

// MapIsSubset given two map[string]any m1 and m2 will determine if m1 is a subset of m2.
// Only fields' name is evaluated not their values.
func MapIsSubset(subset, superset any) bool {
	sub := subset.(map[string]any)
	sup := superset.(map[string]any)

	isSubset := true

	for k, _ := range sub {
		_, ok := sup[k]
		if !ok {
				isSubset = false
		}
	}

	return isSubset
}

// MapIsSubsetOfStruct given a map[string]any and a struct will determine if the map is a subset of the struct.
// Only fields' name is evaluated not their values.
func MapIsSubsetOfStruct(m map[string]any, s any) bool {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem() // Dereference if it's a pointer
	}

	// Ensure we are working with a struct
	if v.Kind() != reflect.Struct {
			return false
	}

	isSubset := true

	t := v.Type()
	for key := range m {
		// FieldByName only finds exported fields
		if _, found := t.FieldByName(key); !found {
			isSubset = false

			break
		}
	}

	return isSubset
}

// StructHasJsonName determines if a struct has the given json tag name.
func StructHasJsonName(s any, targetName string) bool {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")

		// The name is the first part before any commas (e.g., "user_id,omitempty")
		name := strings.Split(tag, ",")[0]

		if name == targetName {
			return true
		}
	}

	return false
}

// StructHasJsonName determines if a struct has the given bson tag name.
func StructHasBsonName(s any, targetName string) bool {
  t := reflect.TypeOf(s)
  if t.Kind() == reflect.Ptr {
    t = t.Elem()
  }

  for i := 0; i < t.NumField(); i++ {
    field := t.Field(i)
    tag := field.Tag.Get("bson")

    // The name is the first part before any commas (e.g., "user_id,omitempty")
    name := strings.Split(tag, ",")[0]

    if name == targetName {
      return true
    }
  }

  return false
}

// StructCopyMatching copies the fields of one struct to another only if they have the same name and type.
func StructCopyMatching(source, target any) {
  sVal := reflect.ValueOf(source).Elem()
  tVal := reflect.ValueOf(target).Elem()

  for i := 0; i < sVal.NumField(); i++ {
    sField := sVal.Type().Field(i)
    tField, ok := tVal.Type().FieldByName(sField.Name)

    if ok && sField.Type == tField.Type {
      tVal.FieldByName(sField.Name).Set(sVal.Field(i))
    }
  }
}

// GetWithDefault
func GetWithDefault(m map[string]any, path string, def any) any {
	keys := strings.Split(path, ".")

	var current any = m

	for _, k := range keys {
		if m2, ok := current.(map[string]any); ok {
			if val, exists := m2[k]; exists {
				current = val
			} else {
				return def
			}
		} else {
			return def
		}
	}

	return current
}

func GetStringWithDefault(m map[string]any, path, def string) string {
	return GetGenericWithDefault[string](m, path, def)
}

func GetString(m map[string]any, path string) string {
	var zero string
	a, ok := GetGeneric[string](m, path)

	if !ok {
		return zero
	}

	return a
}

func GetGenericWithDefault[T any](m map[string]any, path string, def T) T {
	v, ok := GetGeneric[T](m, path)
	if !ok {
		return def
	}
	return v
}

func GetGeneric[T any](m map[string]any, path string) (T, bool) {
	var zero T

	keys := strings.Split(path, ".")

	var current any = m

	for _, k := range keys {
		m2, ok := current.(map[string]any)
		if !ok {
			return zero, false
		}

		val, exists := m2[k]
		if !exists {
			return zero, false
		}

		current = val
	}

	v, ok := current.(T)
	if !ok {
		return zero, false
	}

	return v, true
}

func Get(m map[string]any, path string) any {
	keys := strings.Split(path, ".")

	var current any = m

	for _, k := range keys {
		if m2, ok := current.(map[string]any); ok {
			if val, exists := m2[k]; exists {
				current = val
			} else {
				return nil
			}
		} else {
			return nil
		}
	}

	return current
}

func Set(data map[string]interface{}, path string, value interface{}) {
  parts := strings.Split(path, ".")

  current := data

  for i, part := range parts {
    // last key
    if i == len(parts)-1 {
      current[part] = value
      return
    }

    // existing map
    next, ok := current[part].(map[string]interface{})
    if !ok {
      // create missing map
      next = make(map[string]interface{})
      current[part] = next
    }

    current = next
  }
}

func UnsetSimple(data map[string]interface{}, path string) {
  parts := strings.Split(path, ".")

  var current interface{} = data

  for i, part := range parts {
    m, ok := current.(map[string]interface{})
    if !ok {
       return
    }

    // last key → delete it
    if i == len(parts)-1 {
      delete(m, part)
      return
    }

    // go deeper
    next, ok := m[part]
    if !ok {
       return
    }

    current = next
  }
}

func Unset(doc any, path string) {
    parts := strings.Split(path, ".")

    unsetRecursive(doc, parts)
}

func unsetRecursive(node any, parts []string) {
  if len(parts) == 0 {
    return
  }

  key := parts[0]

  switch v := node.(type) {

  // ✅ BSON / map case
  case bson.M:
    if len(parts) == 1 {
      delete(v, key)
      return
    }

    next, ok := v[key]
    if !ok {
      return
    }

    unsetRecursive(next, parts[1:])

  // also support plain Go maps
  case map[string]any:
    if len(parts) == 1 {
      delete(v, key)
      return
    }

    next, ok := v[key]
    if !ok {
      return
    }

    unsetRecursive(next, parts[1:])

  // ✅ array case
  case bson.A:
    // support numeric index
    for _, item := range v {
      unsetRecursive(item, parts)
    }
  }
}

func StructListFields(value any) {
	if value == nil {
		return
	}

	v := reflect.ValueOf(value)

	// Dereference pointers and interfaces.
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		log.Printf("expected struct, got %s", v.Kind())
		return
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		log.Printf("%s: ", field.Name)

		if fieldValue.CanInterface() {
			log.Printf("%v", fieldValue.Interface())
		} else {
			log.Println("<unexported>")
		}
	}
}

func MapListKeys(o map[string]any) {
	for k, _ := range o {
		log.Printf("• %s", k)
	}
}