package main

import (
	"net/http"
	"sync"
)

type TelemetryProvider interface {
	Collect(*sync.WaitGroup)
	Send(*http.Client, *sync.WaitGroup)
}

func RunAgent(m TelemetryProvider) {
	client := &http.Client{}
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go m.Collect(wg)
	wg.Add(1)
	go m.Send(client, wg)

	wg.Wait()
}
