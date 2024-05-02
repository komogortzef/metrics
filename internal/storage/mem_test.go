package storage

import (
	"testing"
)

func TestSave(t *testing.T) {
	mem := MemStorage{}

	tests := []struct {
		name string
		args [][]byte
		want error
	}{
		{
			name: "to save correct gauge value",
			args: [][]byte{[]byte("gauge"), []byte("someGauge"), []byte("100.01")},
			want: nil,
		},
		{
			name: "to save incorrect gauge value",
			args: [][]byte{[]byte("gauge"), []byte("someGauge"), []byte("invalid")},
			want: StoreError{"Invalid gauge value"},
		},
	}

	for _, test := range tests {
		go t.Run(test.name, func(t *testing.T) {
			if err := mem.Save(test.args...); err != test.want {
				t.Errorf("Result: %v\tWant: %v", err, test.want)
			}
		})
	}
}
