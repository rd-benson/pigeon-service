package common

// MapDiff returns additions & removals between maps A and B
func MapDiffSlice[K comparable, V comparable](A map[K][]V, B map[K][]V) (
	additions map[K][]V, removals map[K][]V) {
	additions = MapChanges(A, B)
	removals = MapChanges(B, A)
	return additions, removals
}

func MapChanges[K comparable, V comparable](A map[K][]V, B map[K][]V) (changes map[K][]V) {
	changes = make(map[K][]V)
	for k, s := range B {
		change := []V{}
		for _, v := range s {
			if !SliceContains(v, A[k]) {
				change = append(change, v)
			}
		}
		if len(change) != 0 {
			changes[k] = change
		}
	}
	return changes
}

func SliceContains[T comparable](elem T, slice []T) bool {
	for _, v := range slice {
		if elem == v {
			return true
		}
	}
	return false
}
