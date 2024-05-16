package main

import "metrics/internal/config"

func main() {
	server, err := config.NewServer(config.WithEnv, config.WithCmd)
	if err != nil {
		panic(err)
	}

	if err = server.ListenAndServe(); err != nil {
		panic(err)
	}
}
