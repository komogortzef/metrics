package telemetry

type TelemetryProvider interface {
	Collect()
	Send()
	Run()
}
