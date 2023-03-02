package utils

import "encoding/json"

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

func ToJSON(a any) string {
	jsonBytes, err := json.Marshal(a)
	if err != nil {
		return ""
	}

	return string(jsonBytes)
}
