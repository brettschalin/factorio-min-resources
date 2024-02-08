//go:build !go1.21

// package shims, and its subdirectories, provides implementations for functions added to Go after 1.18 that I find useful.
// This file contains reimplementations of these functions, and shims.go simply wraps the stdlib

package shims

// from package cmp in 1.21+
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}

func Min[T Ordered](x T, y ...T) T {
	min := x

	for _, v := range y {
		if v < x {
			min = v
		}
	}

	return min
}

func Max[T Ordered](x T, y ...T) T {
	max := x

	for _, v := range y {
		if v > x {
			max = v
		}
	}

	return max
}
