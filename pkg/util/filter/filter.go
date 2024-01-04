package filter

func Filter[T interface{}](vs []*T, f func(*T) bool) []*T {
	filtered := make([]*T, 0)
	for _, v := range vs {
		if f(v) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}
