package commons

import (
  "fmt"
  "testing"
)

func TestMathMin(t *testing.T) {
  v := MathMin[int64](int64(15), int64(98))

  if v != 15 {
    t.Fatalf("Should have been 15 not %v", v)
  }
}

func TestMathMin_Swap(t *testing.T) {
  v := MathMin[int64](int64(100), int64(12))

  if v != 12 {
    t.Fatalf("Should have been 12 not %v", v)
  }
}

func TestMathMin_String(t *testing.T) {
  v := MathMin[string]("15", "152")

  if v != "15" {
    t.Fatalf("Should have been 15 not %v", v)
  }
}

func ExampleMathMin() {
   fmt.Printf("%d", MathMin[int64](int64(10), int64(20)))
   // Output:
   // 10
}
