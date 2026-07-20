package commons

import (
  "testing"
)

func TestStringTitleCamel(t *testing.T) {
  v := "actionObject"

  vv := StringTitle(v)

  if vv != "ActionObject" {
    t.Fatalf("Failed to capitalize %s", vv)
  }
}

func TestStringTitle(t *testing.T) {
  v := "string to make title"

  vv := StringTitle(v)

  if vv != "String To Make Title" {
    t.Fatalf("Failed to capitalize %s", vv)
  }
}

func TestStringNormalize(t *testing.T) {
  v := "    some text     here   "

  vv := StringNormalize(v)
  if vv != "some text here" {
    t.Fatal("Text should have been normalized")
  }
}

func TestStringIsBlank(t *testing.T) {
  if !StringIsBlank("") {
    t.Fatal("Failed to check string is blank")
  }
}

func TestStringIsNotBlank(t *testing.T) {
  if !StringIsNotBlank("fasdfdas") {
    t.Fatal("Failed to check string is not blank")
  }
}

func TestStringBool(t *testing.T) {
  v := StringBool("tr")

  if v != false {
    t.Fatal("should have been false")  
  }
}

func TestStringBool_true(t *testing.T) {
  v := StringBool("true")

  if v != true {
    t.Fatal("should have been true")
  }
}