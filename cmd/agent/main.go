package main

import "github.com/komogortzef/metrics/internal/telemetry"

func main() {
	selfMonitor := telemetry.NewSelfMonitor()
	selfMonitor.Run()
}
