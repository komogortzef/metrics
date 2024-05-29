package server

import (
	"log"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSave(t *testing.T) {
	mem := MemStorage{
		Items: make(map[string]string, metricsNumber),
		Mtx:   &sync.RWMutex{},
	}

	tests := []struct {
		name    string
		argName string
		argVal  string
		err     error
	}{
		{
			name:    "to save gauge value",
			argName: "Alloc",
			argVal:  "hellofdc",
			err:     nil,
		},
		{
			name:    "to save counter value",
			argName: "PollCount",
			argVal:  "ssomefde",
			err:     nil,
		},
		{
			name:    "to save counter one more time",
			argName: "Frees",
			argVal:  "raaaaaaa",
			err:     nil,
		},
	}

	for _, test := range tests {
		log.Println("\n\nTEST:", test.name)
		err := mem.Update(test.argName, test.argVal)
		assert.Equal(t, err, test.err)
	}
}

func TestGet(t *testing.T) {
	mem := MemStorage{
		Items: map[string]string{
			"Alloc":     "1.44",
			"PollCount": "2",
		},
		Mtx: &sync.RWMutex{},
	}

	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{
			name: "get gauge",
			arg:  "Alloc",
			want: true,
		},
		{
			name: "get counter",
			arg:  "PollCount",
			want: true,
		},

		{
			name: "get unknown value",
			arg:  "SomeName",
			want: false,
		},
	}

	for _, test := range tests {
		log.Println("\n\nTES:", test.name)
		_, ok := mem.Get(test.arg)
		assert.Equal(t, ok, test.want)
	}
}
