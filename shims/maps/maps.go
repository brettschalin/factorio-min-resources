//go:build go1.21

package maps

import (
	"maps"
)

func Equal[M1, M2 ~map[K]V, K, V comparable](m1 M1, m2 M2) bool {
	return maps.Equal(m1, m2)
}
