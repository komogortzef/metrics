package main

import (
	"net/http"
	"sync"
)

type TelemetryProvider interface {
	Collect(*sync.WaitGroup)
	Send(*http.Client, *sync.WaitGroup)
}
