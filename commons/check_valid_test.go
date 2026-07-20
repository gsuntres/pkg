package commons

import (
	"testing"
)

const sampleSchema = `
{
	"type": "object",
	"properties": {
		"id": {
			"type": "string"
		},
		"name": {
			"type": "string"
		},
		"age": {
			"type": "number"
		}
	},
	"required": ["id", "name"]
}
`

func TestRequiredName (t *testing.T) {
	data := map[string]any {
		"id": "myid",
		"age": 100,
	}

	if err := Validate(sampleSchema, data); err == nil || err.Error() != `required/name: missing required field "name"` {
		t.Fatalf("Should throw `%s` but got %s", "required/name: missing required field \"name\"", err)
	}
}

func TestPassingValidate (t *testing.T) {
	data := map[string]any {
		"id": "myid",
		"name": "Hey",
		"age": 100,
	}

	if err := Validate(sampleSchema, data); err != nil {
		t.Fatal("Should have not thrown an error")
	}
}

