package config

import (
	"flag"
	"regexp"

	"github.com/caarlos0/env/v6"
)

const (
	addrArg   = 1 // 001
	pollArg   = 2 // 010
	reportArg = 4 // 100

	defaultAddr           = "localhost:8080"
	defaultPollInterval   = 2
	defaultReportInterval = 10
)

var isValidAddr = regexp.MustCompile(`^(.*):(\d+)$`).MatchString

type options struct {
	Address        string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	state          uint8
}

type Option func(*options)

var WithEnv = func(o *options) {
	err := env.Parse(o)
	if err != nil {
		panic(err)
	}

	if isValidAddr(o.Address) {
		o.state |= addrArg
	}
	if o.PollInterval > 0 {
		o.state |= pollArg
	}
	if o.ReportInterval > 0 {
		o.state |= reportArg
	}
}

var WithCmd = func(o *options) {
	addrFlag := flag.String("a", defaultAddr, "Input the endpoint Address: <host:port>")
	pollFlag := flag.Int("p", defaultPollInterval, "Input the poll interval: <sec>")
	repFlag := flag.Int("r", defaultReportInterval, "Input the report interval: <sec>")
	flag.Parse()

	if o.state&addrArg == 0 {
		if isValidAddr(*addrFlag) {
			o.Address = *addrFlag
			o.state |= addrArg
		}
	}
	if o.state&pollArg == 0 {
		if *pollFlag > 0 {
			o.PollInterval = *pollFlag
			o.state |= pollArg
		}
	}
	if o.state&reportArg == 0 {
		if *repFlag > 0 {
			o.ReportInterval = *repFlag
			o.state |= reportArg
		}
	}
}