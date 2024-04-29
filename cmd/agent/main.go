package main

import "telemetry"

func main() {

	selfMonitor := telemetry.NewSelfMonitor()
	selfMonitor.Run()
}
