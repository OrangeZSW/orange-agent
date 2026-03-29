package utils

import (
	"encoding/json"
	"strconv"
	"strings"
)

func Int64ToUint(i int64) uint {
	return uint(i)
}

// unit -> str
func UintToStr(i uint) string {
	return strconv.Itoa(int(i))
}

// str -> unit
func StrToUint(s string) uint {
	i, _ := strconv.ParseUint(s, 10, 64)
	return uint(i)
}

// bool -> str
func BoolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// []str -> str
func StrArrToStr(arr []string) string {
	return "[" + strings.Join(arr, ",") + "]"
}

// str -> struct
func StrToStruct(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}

// str -> map
func StrToMap(s string) (map[string]interface{}, error) {
	var m map[string]interface{}
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// unit -> int64
func UintToInt64(i uint) int64 {
	return int64(i)
}
