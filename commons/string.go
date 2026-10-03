package commons

import (
	"fmt"
  "time"
	"reflect"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/sergi/go-diff/diffmatchpatch"
)

var caserTitle cases.Caser = cases.Title(language.English, cases.NoLower)

func StringTitle(s string) string {
	return caserTitle.String(s)
}

// StringNormalize will replace arbitrary lengths of consecutive white space with a single one.
func StringNormalize(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// StringIsBlank checks if a string is blank. Will trim the string before the check takes place.
func StringIsBlank(s string) bool {
  return len(strings.TrimSpace(s)) == 0
}

// StringIsNotBlank check if a string is NOT blank.
func StringIsNotBlank(s string) bool {
  return !StringIsBlank(s)
}

func StringMissing(a, b string) string {
	dmp := diffmatchpatch.New()
  diffs := dmp.DiffMain(a, b, false)

  var missing string

  for _, d := range diffs {
    if d.Type == diffmatchpatch.DiffDelete {
      missing += d.Text
    }
  }

  return missing
}

func StringBool(s string) bool {
  if s == "true" {
    return true
  } else {
    return false
  }
}

func Describe(value any, fieldsToOmit ...string) string {
  omit := make(map[string]struct{}, len(fieldsToOmit))

	for _, field := range fieldsToOmit {
		omit[field] = struct{}{}
	}

	var b strings.Builder
	describeValue(&b, reflect.ValueOf(value), 0, omit)

	return b.String()
}

func describeValue(
	b *strings.Builder,
	v reflect.Value,
	indent int,
	omit map[string]struct{},
) {
	if !v.IsValid() {
		return
	}

	if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}

		describeValue(b, v.Elem(), indent, omit)
		return
	}

	// Handle time.Time as a scalar value.
	if v.Type() == reflect.TypeOf(time.Time{}) {
		writeIndent(b, indent)
		fmt.Fprintf(b, "%s\n", v.Interface().(time.Time).Format(time.RFC3339))
		return
	}

	switch v.Kind() {
	case reflect.Map:
		for _, key := range v.MapKeys() {
			fieldName := fmt.Sprint(key.Interface())

			if shouldOmit(fieldName, omit) {
				continue
			}

			value := v.MapIndex(key)

			writeIndent(b, indent)
			fmt.Fprintf(b, "%s:", fieldName)

			if isComplex(value) && !isTimeValue(value) {
				b.WriteByte('\n')
				describeValue(b, value, indent+2, omit)
			} else {
				fmt.Fprintf(b, " %v\n", formatValue(value))
			}
		}

	case reflect.Struct:
		t := v.Type()

		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)

			if field.PkgPath != "" || shouldOmit(field.Name, omit) {
				continue
			}

			value := v.Field(i)

			writeIndent(b, indent)
			fmt.Fprintf(b, "%s:", field.Name)

			if isComplex(value) && !isTimeValue(value) {
				b.WriteByte('\n')
				describeValue(b, value, indent+2, omit)
			} else {
				fmt.Fprintf(b, " %v\n", formatValue(value))
			}
		}

	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			writeIndent(b, indent)
			b.WriteString("-\n")
			describeValue(b, v.Index(i), indent+2, omit)
		}

	default:
		writeIndent(b, indent)
		fmt.Fprintf(b, "%v\n", v.Interface())
	}
}


func shouldOmit(field string, omit map[string]struct{}) bool {
	if _, exists := omit[field]; exists {
		return true
	}

	return false
}

func isComplex(v reflect.Value) bool {
	if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}

	return v.Kind() == reflect.Map ||
		v.Kind() == reflect.Struct ||
		v.Kind() == reflect.Slice ||
		v.Kind() == reflect.Array
}

func writeIndent(b *strings.Builder, count int) {
	b.WriteString(strings.Repeat(" ", count))
}

func isTimeValue(v reflect.Value) bool {
	if v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}

	return v.IsValid() && v.Type() == reflect.TypeOf(time.Time{})
}

func formatValue(v reflect.Value) any {
	if v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if v.IsValid() && v.Type() == reflect.TypeOf(time.Time{}) {
		return v.Interface().(time.Time).Format(time.RFC3339)
	}

	return v.Interface()
}

func PrintDescribe(value any, fieldsToOmit ...string) {
	fmt.Print(Describe(value, fieldsToOmit...))
}