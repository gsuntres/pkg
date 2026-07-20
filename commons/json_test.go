package commons

import (
	"log"
	"fmt"
	"bytes"
	"encoding/json"
	"testing"	
)

func TestJsonPrintError(t *testing.T) {

	jsonstr := `
	{
		"name": "George",
		"age": 87,
		active": "false",
		"properties": {
			"one": "two",
			"three": "four"
		}
	}
	`

	var buf bytes.Buffer
	log.SetOutput(&buf)

	content := []byte(jsonstr)
	var data map[string]any
	
	err := json.Unmarshal(content, &data)

	JsonPrintError(err, content)

	errOut := fmt.Sprintf("%s", buf.String())

	if errOut == "" {
		t.Fatal("should have printed the error")
	}
} 

func TestJsonSprintError(t *testing.T) {
	jsonstr := `
	{
		"name": "George",
		"age": 87,
		active": "false",
		"properties": {
			"one": "two",
			"three": "four"
		}
	}
	`

	var buf bytes.Buffer
	log.SetOutput(&buf)

	content := []byte(jsonstr)
	var data map[string]any
	
	err := json.Unmarshal(content, &data)

	_, s := JsonSprintError(err, content)

	if s == "" {
		t.Fatal("should have printed the error")
	}
}

func TestJsonPrint(t *testing.T) {
	m := map[string]any{
		"name": "George",
		"age": 87,
		"active": false,
		"properties": map[string]any{
			"one": "two",
			"three": "four",
		},
	}

	var buf bytes.Buffer
	log.SetFlags(0)
	log.SetOutput(&buf)

	JsonPrint(m)

	out := fmt.Sprintf("%s", buf.String())

	expected := `{"active":false,"age":87,"name":"George","properties":{"one":"two","three":"four"}}`

  missing := StringMissing(expected, out)

	if missing != "" {
		t.Fatalf("Not expected %s", missing)
	}
}