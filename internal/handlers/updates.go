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

	err := storage.Mem.Save(reqElem[1:]...)
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
