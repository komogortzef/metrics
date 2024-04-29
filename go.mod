module github.com/komogortzef/metrics

go 1.22.1

replace routes => ./internal/routes

replace handlers => ./internal/handlers

replace storage => ./internal/storage

replace telemetry => ./internal/telemetry

require (
	routes v0.0.0-00010101000000-000000000000
	telemetry v0.0.0-00010101000000-000000000000
)

require (
	handlers v0.0.0-00010101000000-000000000000 // indirect
	storage v0.0.0-00010101000000-000000000000 // indirect
)
