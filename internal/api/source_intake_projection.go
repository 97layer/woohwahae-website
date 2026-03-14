package api

func sourceIntakeSliceLimit[T any](items []T, limit int) []T {
	if limit <= 0 || len(items) <= limit {
		return append([]T{}, items...)
	}
	return append([]T{}, items[:limit]...)
}
