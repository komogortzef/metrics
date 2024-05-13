package agent

import (
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"runtime"
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

func newMockMonitor() SelfMonitor {
	var st runtime.MemStats
	runtime.ReadMemStats(&st)

	return SelfMonitor{
		MemStats:    st,
		Mtx:         &sync.Mutex{},
		successSend: true,
	}
}

func TestReport(t *testing.T) {
	mockServ := httptest.NewServer(http.HandlerFunc(mockHandler))
	defer mockServ.Close()

	monitor := newMockMonitor()
	ENDPOINT = mockServ.URL

	go monitor.Report()
	time.Sleep(12 * time.Second)

	log.Println("success flag:", monitor.successSend)
	assert.Equal(t, monitor.successSend, true)
}

func TestCollect(t *testing.T) {
	monitor := newMockMonitor()

	go monitor.Collect()
	time.Sleep(3 * time.Second)

	assert.NotZero(t, monitor.Alloc)
	assert.NotZero(t, monitor.pollCount)
}
