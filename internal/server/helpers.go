package server

import (
	"strconv"
)

func countAdd(val string, delta string) string {
	numVal, _ := strconv.ParseInt(val, 10, 64)
	numDelta, _ := strconv.ParseInt(delta, 10, 64)
	numVal += numDelta

	return strconv.FormatInt(numVal, 10)
}
