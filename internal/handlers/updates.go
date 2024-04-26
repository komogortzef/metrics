package handlers

import (
	"net/http"
	"storage"
	"strings"
)

func Update(resp http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodGet {
		http.Error(resp, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uri := req.RequestURI[1:]
	req_elem := strings.Split(uri, "/")
	if len(req_elem) != 4 {
		http.Error(resp, "Bad Request", http.StatusBadRequest)
		return
	}

	metricType := req_elem[1]
	metricName := req_elem[2]
	metricVal := req_elem[3]

	err := storage.Mem.UpdateStorage(metricType, metricName, metricVal)
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
