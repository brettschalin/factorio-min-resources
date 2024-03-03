//go:build go1.21

package slices

import "slices"

func Contains[S ~[]E, E comparable](s S, v E) bool {
	return slices.Contains(s, v)
}

func Index[S ~[]E, E comparable](s S, v E) int {
	return slices.Index(s, v)
}

func Reverse[S ~[]E, E any](s S) {
	slices.Reverse(s)
}
