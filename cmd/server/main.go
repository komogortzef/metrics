package main

import "github.com/komogortzef/metrics/internal/config"

func main() {

	if err := config.Run(); err != nil {
		panic(err)
	}
}
