package slice

func RemoveDuplicate[T comparable](slice []T) []T {
	allKeys := make(map[T]bool)
	uniqueSlice := []T{}
	for _, item := range slice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			uniqueSlice = append(uniqueSlice, item)
		}
	}
	return uniqueSlice
}

func Union[T comparable](slice1, slice2 []T) []T {
	allKeys := make(map[T]struct{})

	for _, item := range slice1 {
		allKeys[item] = struct{}{}
	}

	for _, item := range slice2 {
		allKeys[item] = struct{}{}
	}

	unionSlice := make([]T, 0, len(allKeys))
	for key := range allKeys {
		unionSlice = append(unionSlice, key)
	}

	return unionSlice
}
