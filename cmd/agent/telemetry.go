package main

type Telemetry interface {
	CollectTelemetry()
	SendTelemetry() error
}
