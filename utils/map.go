package utils

func GetIntFromMap(m map[string]interface{}, key string) int {
	if value, ok := m[key]; ok {
		if str, ok := value.(int); ok {
			return str
		}
	}
	return 0
}
