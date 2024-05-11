package server

import (
	"strconv"
)

var WithAccInt64 = func(a []byte, b []byte) ([]byte, error) {
	astr := string(a)
	bstr := string(b)

	anum, _ := strconv.ParseInt(astr, 10, 64)
	bnum, err := strconv.ParseInt(bstr, 10, 64)

	return []byte(strconv.FormatInt(anum+bnum, 10)), err
}
