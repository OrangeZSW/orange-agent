package utils

import (
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
