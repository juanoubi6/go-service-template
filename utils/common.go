package utils

func ListContains[k comparable](list []k, value k) bool {
	for _, listVal := range list {
		if listVal == value {
			return true
		}
	}

	return false
}

func ToPointer[K any](value K) *K { //nolint
	return &value
}
