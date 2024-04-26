package routes

import (
	h "handlers"
	"net/http"
)

var Mux = http.NewServeMux()

func init() {
	Mux.HandleFunc("/update/gauge/", h.GaugeUpdate)
	Mux.HandleFunc("/update/counter/", h.CounterUpdate)
}
