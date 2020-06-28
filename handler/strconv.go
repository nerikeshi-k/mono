package handler

import "strconv"

func atoi(s string) int {
	num, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0
	}
	return int(num)
}

func atoiPos(s string) int {
	num := atoi(s)
	if num < 0 {
		return 0
	}
	return num
}
