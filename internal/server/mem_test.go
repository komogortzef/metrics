package server

import (
	"log"
	"sync"
	"testing"

	"metrics/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestSave(t *testing.T) {
	mem := MemStorage{
		Items: make(map[string][]byte, models.MetricsNumber),
		Mtx:   &sync.RWMutex{},
	}

	tests := []struct {
		name    string
		argName string
		argVal  []byte
		err     error
	}{
		{
			name:    "to save gauge value",
			argName: "Alloc",
			argVal:  []byte("hellofdc"),
			err:     nil,
		},
		{
			name:    "to save counter value",
			argName: "PollCount",
			argVal:  []byte("ssomefde"),
			err:     nil,
		},
		{
			name:    "to save counter one more time",
			argName: "Frees",
			argVal:  []byte("raaaaaaa"),
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
		Items: map[string][]byte{
			"Alloc":     []byte("1.44"),
			"PollCount": []byte("2"),
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
