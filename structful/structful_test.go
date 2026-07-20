package structful

// import (
// 	"testing"
// )

// func TestRequiredFields(t *testing.T) {
// 	op := &Operator{}

// 	data := map[string]any{
// 		"id": "1",
// 		"_group": "agroup",
// 		"_name": "myname",
// 		"_version": "1",
// 		"foo": "bar",
// 	}

// 	if err := op.Save(data); err != nil {
// 		t.Fatal(err)	
// 	}	
// }

// func TestRequiredFields_RequireName(t *testing.T) {
// 	op := &Operator{}

// 	data := map[string]any{
// 		"id": "1",
// 		"_group": "agroup",
// 		"_version": "1",
// 		"foo": "bar",
// 	}

// 	if err := op.Save(data); err != nil && err.Error() != `required/_name: missing required field "_name"`  {
// 		t.Fatal("should have required name", err)	
// 	}	
// }

// func TestRequiredFields_RequireGroup(t *testing.T) {
// 	op := &Operator{}

// 	data := map[string]any{
// 		"id": "1",
// 		"_version": "1",
// 		"_name": "myname",
// 		"foo": "bar",
// 	}

// 	if err := op.Save(data); err != nil && err.Error() != `required/_group: missing required field "_group"`  {
// 		t.Fatal("should have required group", err)	
// 	}	
// }

// func TestRequiredFields_RequireVersion(t *testing.T) {
// 	op := &Operator{}

// 	data := map[string]any{
// 		"id": "1",
// 		"_group": "agroup",
// 		"_name": "myname",
// 		"foo": "bar",
// 	}

// 	if err := op.Save(data); err != nil && err.Error() != `required/_version: missing required field "_version"`  {
// 		t.Fatal("should have required _version", err)
// 	}	
// }