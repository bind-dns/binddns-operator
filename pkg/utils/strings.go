package utils

import "unsafe"

func CalcAscII(str string) (result int) {
	for i := range str {
		result += int(str[i])
	}
	return result
}

// BytesToString convert []byte to string
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StringToBytes convert string to []byte
func StringToBytes(s string) []byte {
	sh := (*[2]uintptr)(unsafe.Pointer(&s))
	bh := [3]uintptr{sh[0], sh[1], sh[1]}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func SubString(source string, start, end int) string {
	r := []rune(source)
	length := len(r)

	if start < 0 || end > length || start > end {
		return ""
	}
	if start == 0 && end == length {
		return source
	}
	return string(r[start:end])
}

func SliceContain(slice []string, target string) bool {
	for i := range slice {
		if slice[i] == target {
			return true
		}
	}
	return false
}
