//go:build go1.21

package shims

import (
	"cmp"

	"slices"
)

type Ordered = cmp.Ordered

func Min[T Ordered](x T, y ...T) T {
	if len(y) == 0 {
		return x
	}

	return min(x, slices.Min(y))
}

func Max[T Ordered](x T, y ...T) T {
	if len(y) == 0 {
		return x
	}

	return max(x, slices.Max(y))
}
