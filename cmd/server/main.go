package main

import "metrics/internal/config"

func main() {
	serv, err := config.NewServer(config.WithEnv, config.WithCmd)
	if err != nil {
		panic(err)
	}

	if err = serv.ListenAndServe(); err != nil {
		panic(err)
	}
}
