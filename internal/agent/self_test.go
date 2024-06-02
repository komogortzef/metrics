package agent

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"
	"time"
)

func mockHandler(w http.ResponseWriter, r *http.Request) {
	validURL := regexp.MustCompile(
		`^(http|https)://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(/.*)?$`)

	if !validURL.MatchString(r.URL.String()) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func TestReport(t *testing.T) {
	mockServ := httptest.NewServer(http.HandlerFunc(mockHandler))
	defer mockServ.Close()

	address = mockServ.URL

}

func TestCollect(t *testing.T) {
	monitor := SelfMonitor{
		Mtx: &sync.RWMutex{},
	}

	go monitor.Collect()
	time.Sleep(time.Duration(pollInterval)*time.Second + 1)
}
