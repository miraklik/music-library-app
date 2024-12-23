package utils

import "strconv"

func ToInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 1
	}

	return val
}
