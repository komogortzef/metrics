package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/komogortzef/metrics/internal/storage"
	"github.com/stretchr/testify/assert"
)

type MockStorage struct{}

func (m MockStorage) Save(data ...[]byte) error {
	tp := string(data[0])
	val := string(data[2])
	realNums := regexp.MustCompile(`^-?\d+(\.\d+)?$`)
	naturalNums := regexp.MustCompile(`^\d+$`)

	if tp == "counter" && !naturalNums.MatchString(val) {
		return errors.New("BadReq")
	}

	if !realNums.MatchString(val) {
		return errors.New("BadReq")
	}

	return nil
}

func (m MockStorage) Fetch(keys ...string) (any, error) {
	return nil, nil
}

func NewMock(store storage.Storage) *Handler {
	return &Handler{
		store,
	}
}

func TestSaveToMem(t *testing.T) {
	handler := NewMock(MockStorage{})

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
			url:         "http://localhost:8080/update/gauge/metric/44",
			expected:    http.StatusOK,
			description: "Sending a valid POST request with gauge type",
		},
		{
			name:        "Invalid HTTP method",
			method:      http.MethodGet,
			url:         "http://localhost:8080/update/gauge/name/44",
			expected:    http.StatusMethodNotAllowed,
			description: "Sending a GET request which is not allowed",
		},
		{
			name:        "Invalid URL format",
			method:      http.MethodPost,
			url:         "http://localhost:8080/update/gauge/44",
			expected:    http.StatusNotFound,
			description: "Sending a POST request with invalid URL format",
		},
		{
			name:        "Invalid value",
			method:      http.MethodPost,
			url:         "http://localhost:8080/update/gauge/metric/invalid",
			expected:    http.StatusBadRequest,
			description: "Sending a POST request with invalid value",
		},
		{
			name:        "Invalid counter",
			method:      http.MethodPost,
			url:         "http://localhost:8080/update/counter/metric/4.4",
			expected:    http.StatusBadRequest,
			description: "Sending a POST request with invalid counter's type",
		},
		{
			name:        "Valid float value for gauge type",
			method:      http.MethodPost,
			url:         "http://localhost:8080/update/gauge/metric/4.4",
			expected:    http.StatusOK,
			description: "Sending a valid POST request with gauge type",
		},
		{
			name:        "Valid value for counter",
			method:      http.MethodPost,
			url:         "http://localhost:8080/update/counter/metric/4",
			expected:    http.StatusOK,
			description: "Sending a valid POST request with counter type",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.url, nil)
			responseRecorder := httptest.NewRecorder()

			handler.SaveToMem(responseRecorder, request)

			response := responseRecorder.Result()
			defer response.Body.Close()

			assert.Equal(t, test.expected, response.StatusCode, test.description)
		})
	}
}
