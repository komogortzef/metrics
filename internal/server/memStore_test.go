package server

import (
	"log"
	"testing"
)

func TestWrite(t *testing.T) {

	tests := []struct {
		name string
		arg  []byte
		err  error
	}{}

	for _, test := range tests {
		log.Println("\n\nTEST:", test.name)
	}
}
