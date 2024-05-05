package telemetry

import (
	"net/http"
	"net/http/httptest"
	"regexp"
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

func TestSend(t *testing.T) {
	mockServ := httptest.NewServer(http.HandlerFunc(mockHandler))
	defer mockServ.Close()

	monitor := NewSelfMonitor()
	monitor.serverAddr = mockServ.URL

	go monitor.Send()
	time.Sleep(12 * time.Second)

	assert.NotEmpty(t, monitor.sendReports)
}

func TestCollect(t *testing.T) {
	monitor := NewSelfMonitor()

	go monitor.Collect()
	time.Sleep(3 * time.Second)

	assert.NotZero(t, monitor.Alloc)
	assert.NotZero(t, monitor.pollCount)
}
