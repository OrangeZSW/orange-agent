package utils

import "strconv"

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
