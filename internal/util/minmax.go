package util

// All the numbers types.
type Numbers interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

func Max[T Numbers](a, b T) T {
	if a > b {
		return a
	}
	return b
}
func Min[T Numbers](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Abs[T Numbers](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

func AbsMax[T Numbers](a, b T) T {
	if Abs(a) > Abs(b) {
		return a
	}
	return b
}
func AbsMin[T Numbers](a, b T) T {
	if Abs(a) < Abs(b) {
		return a
	}
	return b
}
