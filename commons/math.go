package commons

// MathMin will compare two integers and return the one with the smaller value.
func MathMin[T ~int | ~int8 | ~int16 | ~int32 | ~int64 |
         ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
         ~float32 | ~float64 | ~string](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// MathMax will compare two integers and return the one with the smaller value.
func MathMax[T ~int | ~int8 | ~int16 | ~int32 | ~int64 |
         ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
         ~float32 | ~float64 | ~string](a, b T) T {
	if a > b {
		return a
	}
	return b
}