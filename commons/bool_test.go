package commons

import "testing"

func TestBoolString_true(t *testing.T) {
	v := BoolString(true)

	if v != "true" {
		t.Fatal("should have been true")
	}
}

func TestBoolString_false(t *testing.T) {
	v := BoolString(false)

	if v != "false" {
		t.Fatal("should have been false")
	}
}