package server

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func testRequest(
	t *testing.T,
	ts *httptest.Server,
	method, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, http.NoBody)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestHandlers(t *testing.T) {

	tests := []struct {
		name        string
		method      string
		url         string
		expected    int
		description string
	}{
		{
			name:        "Valid POST request",
			method:      http.MethodPost,
			url:         "/update/gauge/metric/44",
			expected:    http.StatusOK,
			description: "Sending a valid POST request with gauge type",
		},
		{
			name:        "Invalid URL format",
			method:      http.MethodPost,
			url:         "/update/gauge/44",
			expected:    http.StatusNotFound,
			description: "Sending a POST request with invalid URL format",
		},
		{
			name:        "Invalid value",
			method:      http.MethodPost,
			url:         "/update/gauge/metric/invalid",
			expected:    http.StatusBadRequest,
			description: "Sending a POST request with invalid value",
		},
		{
			name:        "Invalid counter",
			method:      http.MethodPost,
			url:         "/update/counter/metric/4.4",
			expected:    http.StatusBadRequest,
			description: "Sending a POST request with invalid counter's type",
		},
		{
			name:        "Valid float value for gauge type",
			method:      http.MethodPost,
			url:         "/update/gauge/metric/4.4",
			expected:    http.StatusOK,
			description: "Sending a valid POST request with gauge type",
		},
		{
			name:        "Valid value for counter",
			method:      http.MethodPost,
			url:         "/update/counter/metricCount/4",
			expected:    http.StatusOK,
			description: "Sending a valid POST request with counter type",
		},
	}

	for _, test := range tests {
		log.Println("\n\nTEST NAME:", test.name)

	}
}
