package main

import (
	"metrics/internal/config"
)

func main() {
	agent, _ := config.NewAgent(config.WithEnv, config.WithCmd)
	agent.Run()
}
