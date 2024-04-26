package handlers

import (
	"net/http"
	"storage"
)

func GaugeUpdate(resp http.ResponseWriter, req *http.Request) {
	strg := storage.MemStorage{}
	strg.Gauges["first"] = 4.33

	if req.Method != http.MethodGet {
		http.Error(resp, "Method not allowed", http.StatusMethodNotAllowed)
		return
	} else {
		resp.WriteHeader(http.StatusOK)
		resp.Write([]byte("gauge"))
	}
}

func CounterUpdate(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(resp, "Method not allowed", http.StatusMethodNotAllowed)
		return
	} else {

		resp.WriteHeader(http.StatusOK)
		resp.Write([]byte("counter"))
	}
}
