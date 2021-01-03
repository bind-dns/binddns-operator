package utils

func CalcAscII(str string) (result int) {
	for i := range str {
		result += int(str[i])
	}
	return result
}
