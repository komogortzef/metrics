package agent

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func mockHandler(w http.ResponseWriter, r *http.Request) {
	validURL := regexp.MustCompile(`^(http|https)://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(/.*)?$`)

	if !validURL.MatchString(r.URL.String()) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func TestReport(t *testing.T) {
	mockServ := httptest.NewServer(http.HandlerFunc(mockHandler))
	defer mockServ.Close()

	monitor := SelfMonitor{
		Mtx: &sync.RWMutex{},
	}
	address = mockServ.URL

	go monitor.Report()
	time.Sleep(time.Duration(reportInterval)*time.Second + 1)

	fmt.Println(successSend)
	assert.NotEqual(t, false, successSend)
}

func TestCollect(t *testing.T) {
	monitor := SelfMonitor{
		Mtx: &sync.RWMutex{},
	}

	go monitor.Collect()
	time.Sleep(time.Duration(pollInterval)*time.Second + 1)

	assert.NotZero(t, monitor.Alloc)
	assert.NotZero(t, monitor.pollCount)
}
