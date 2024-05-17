package main

import "metrics/internal/config"

func main() {
	agent, err := config.NewAgent(config.WithEnv, config.WithCmd)
	if err != nil {
		panic(err)
	}

	agent.Run()
}
