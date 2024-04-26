module github.com/komogortzef/metrics

go 1.21.9

replace routes => ./internal/routes

replace handlers => ./internal/handlers

replace storage => ./internal/storage

require routes v0.0.0-00010101000000-000000000000

require (
	handlers v0.0.0-00010101000000-000000000000 // indirect
	storage v0.0.0-00010101000000-000000000000 // indirect
)
