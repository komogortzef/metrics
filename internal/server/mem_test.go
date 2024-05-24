package server

import (
	"log"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSave(t *testing.T) {
	mem := MemStorage{
		Items: make(map[string][]byte),
		Mtx:   &sync.RWMutex{},
	}

	tests := []struct {
		name    string
		argName string
		argVal  []byte
		argOper Operation
		want    map[string][]byte
	}{
		{
			name:    "to save gauge value",
			argName: "Gauge",
			argVal:  []byte("1.44"),
			argOper: nil,
			want: map[string][]byte{
				"Gauge": []byte("1.44"),
			},
		},
		{
			name:    "to save counter value",
			argName: "Counter",
			argVal:  []byte("1"),
			argOper: withAccInt64,
			want: map[string][]byte{
				"Gauge":   []byte("1.44"),
				"Counter": []byte("1"),
			},
		},
		{
			name:    "to save counter one more time",
			argName: "Counter",
			argVal:  []byte("1"),
			argOper: withAccInt64,
			want: map[string][]byte{
				"Gauge":   []byte("1.44"),
				"Counter": []byte("2"),
			},
		},
	}

	for _, test := range tests {
		log.Println("\n\nTEST:", test.name)
		mem.Save(test.argName, test.argVal, test.argOper)
		log.Println("mem:", mem.Items)
		log.Println("testMem:", test.want)
		assert.Equal(t, mem.Items, test.want)
	}
}

func TestGet(t *testing.T) {
	mem := MemStorage{
		Items: map[string][]byte{
			"Gauge":   []byte("1.44"),
			"Counter": []byte("2"),
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
			arg:  "Gauge",
			want: true,
		},
		{
			name: "get counter",
			arg:  "Counter",
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
