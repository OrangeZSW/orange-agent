package utils

// GetStringFromMap
func GetStringFromMap(m map[string]interface{}, key string) string {
	if value, ok := m[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func GetIntFromMap(m map[string]interface{}, key string) int {
	if value, ok := m[key]; ok {
		if str, ok := value.(int); ok {
			return str
		}
	}
	return 0
}
