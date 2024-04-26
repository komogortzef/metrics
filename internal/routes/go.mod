module routes

go 1.22.1

replace handlers => ../handlers

replace storage => ../storage

require handlers v0.0.0-00010101000000-000000000000

require storage v0.0.0-00010101000000-000000000000 // indirect
