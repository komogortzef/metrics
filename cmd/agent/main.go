package main

import "github.com/komogortzef/metrics/internal/agent"

func main() {
	agent, err := agent.GetConfig()
	if err != nil {
		panic(err)
	}

	agent.ShowConfig()

	agent.Perform()
}
