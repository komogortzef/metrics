package agent

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"runtime"
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
		ReportInterval: 10,
		PollInterval:   2,
		MemStats:       st,
	}
}

func TestSend(t *testing.T) {
	mockServ := httptest.NewServer(http.HandlerFunc(mockHandler))
	defer mockServ.Close()

	monitor := newMockMonitor()
	monitor.Endpoint = mockServ.URL

	go monitor.Send()
	time.Sleep(12 * time.Second)

	assert.NotEmpty(t, monitor.sendReports)
}

func TestCollect(t *testing.T) {
	monitor := newMockMonitor()

	go monitor.Collect()
	time.Sleep(3 * time.Second)

	assert.NotZero(t, monitor.Alloc)
	assert.NotZero(t, monitor.pollCount)
}
