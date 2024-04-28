package main

import (
	"log"
	"net/http"
	"sync"
)

func main() {

	selfMonitor := NewSelf()

	RunAgent(&selfMonitor)
}

func RunAgent(monitor TelemetryProvider) {
	client := &http.Client{}
	wg := &sync.WaitGroup{}

	go monitor.Collect(wg)
	wg.Add(1)
	go monitor.Send(client, wg)
	wg.Add(1)

	wg.Wait()
	log.Println("Горутины завершили выполнение")
}
