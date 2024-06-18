package server

import (
	"log"
	"testing"

	m "metrics/internal/models"
)

func TestWrite(t *testing.T) {
	met, _ := m.NewMetric(m.Gauge, "someGauge", 4.44)
	arg1, _ := met.MarshalJSON()
	met2, _ := m.NewMetric(m.Counter, "someCount", 4)
	arg2, _ := met2.MarshalJSON()
	met3, _ := m.NewMetric(m.Counter, "anotherCount", 1)
	arg3, _ := met3.MarshalJSON()

	tests := []struct {
		name string
		arg  []byte
		err  error
	}{
		{
			name: "to save gauge value",
			arg:  arg1,
			err:  nil,
		},
		{
			name: "to save counter value",
			arg:  arg2,
			err:  nil,
		},
		{
			name: "to save counter one more time",
			arg:  arg3,
			err:  nil,
		},
	}

	for _, test := range tests {
		log.Println("\n\nTEST:", test.name)
	}
}
