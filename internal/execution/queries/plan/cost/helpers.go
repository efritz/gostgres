package cost

func coalesce(value *int, defaultValue int) int {
	if value != nil {
		return *value
	}

	return defaultValue
}
