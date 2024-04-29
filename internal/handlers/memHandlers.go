package handlers

import (
	"net/http"
	"storage"
	"strings"
)

func SaveToMem(resp http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		http.Error(resp, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uri := req.RequestURI[1:]
	reqElem := strings.Split(uri, "/")
	if len(reqElem) != 4 {
		http.Error(resp, "Not Found", http.StatusNotFound)
		return
	}

	tp := []byte(reqElem[1])
	name := []byte(reqElem[2])
	val := []byte(reqElem[3])

	err := storage.Mem.Save(tp, name, val)
	if err != nil {
		switch err.Error() {
		case "NotFound":
			http.Error(resp, "Not Found", http.StatusNotFound)
			return
		case "BadReq":
			http.Error(resp, "Bad Request", http.StatusBadRequest)
		}
	}
}
