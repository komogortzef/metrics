package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/komogortzef/metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T,
	ts *httptest.Server,
	method, path string) (*http.Response, string) {
	log.Println("test request starts")
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	log.Println("sending request!")
	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	log.Println("sending ends")

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestSaveToMem(t *testing.T) {
	r := chi.NewRouter()

	srv := Config{ // определил для тестов
		Store: storage.MemStorage{},
	}

	r.Get("/", srv.ShowAll)
	r.Route("/", func(r chi.Router) {
		r.Get("/value/{tp}/{name}", srv.GetMetric)
		r.Post("/update/{tp}/{name}/{val}", srv.SaveToMem)
	})

	tserv := httptest.NewServer(r)
	defer tserv.Close()

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
		{
			name:        "get all",
			method:      http.MethodGet,
			url:         "/",
			expected:    http.StatusOK,
			description: "Sending a GET request",
		},
		{
			name:        "non-existent value",
			method:      http.MethodGet,
			url:         "/value/gauge/name",
			expected:    http.StatusNotFound,
			description: "Sending a GET request with non-existent value",
		},
		{
			name:        "existent value",
			method:      http.MethodGet,
			url:         "/value/gauge/metric",
			expected:    http.StatusOK,
			description: "Sending a GET request with existent value",
		},
	}

	for _, test := range tests {
		resp, get := testRequest(t, tserv, test.method, test.url)
		defer resp.Body.Close()
		assert.Equal(t, test.expected, resp.StatusCode)

		if test.method == http.MethodGet {
			fmt.Println(get)
		}
	}
}
