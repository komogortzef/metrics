package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	l "metrics/internal/logger"
	m "metrics/internal/models"

	"github.com/caarlos0/env/v6"
)

type (
	Service interface {
		Run() error
	}

	agentEnv struct {
		Address        string `env:"ADDRESS"`
		PollInterval   int    `env:"POLL_INTERVAL"`
		ReportInterval int    `env:"REPORT_INTERVAL"`
	}

	serverEnv struct {
		Address         string `env:"ADDRESS"`
		StoreInterval   int    `env:"STORE_INTERVAL"`
		FileStoragePath string `env:"FILE_STORAGE_PATH"`
		Restore         bool   `env:"RESTORE"`
	}
)

func (o *options) setAddr(addr ...string) {
	if len(addr) == 0 {
		err := env.Parse(o)
		if err != nil {
			l.Warn("Coulnd't parse env")
		}
		if m.IsValidAddr(o.Address) {
			o.state |= addrArg
		}
		return
	}

	if o.state&addrArg == 0 {
		if m.IsValidAddr(addr[0]) {
			o.Address = addr[0]
		}
	}
}

var WithEnvAg = func(o *options) {
	o.state |= seniorBitMask
	o.setAddr()

	if o.PollInterval > 0 {
		o.state |= pollArg
	}
	if o.ReportInterval > 0 {
		o.state |= reportArg
	}
}

var WithEnvSrv = func(o *options) {
	o.setAddr()

	if val, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		num, err := strconv.Atoi(val)
		if err != nil {
			l.Warn("Couldn't parse STORE_INTERVAL")
		} else {
			o.storeInterval = num
			o.state |= storIntervArg
		}
	}

	if val, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		if m.IsValidPath(val) {
			o.fileStorage = val
			o.state |= storPathArg
		} else {
			l.Warn("Couldn't parse FILE_STORAGE_PATH")
		}
	}

	if val, ok := os.LookupEnv("RESTORE"); ok {
		yesno, err := strconv.ParseBool(val)
		if err != nil {
			l.Warn("Couldn't parse RESTORE")
		} else {
			o.restore = yesno
			o.state |= restoreArg
		}
	}
}

var WithCmdAg = func(o *options) {
	if o.state == fullAgentConfig {
		return
	}
	o.state |= seniorBitMask

	addrFlag := flag.String(
		"a", m.DefaultEndpoint, "Input the endpoint Address: <host:port>")

	pollFlag := flag.Int(
		"p", m.DefaultPollInterval, "Input the poll interval: <sec>")

	repFlag := flag.Int(
		"r", m.DefaultReportInterval, "Input the report interval: <sec>")

	flag.Parse()

	o.setAddr(*addrFlag)

	if o.state&pollArg == 0 {
		if *pollFlag > 0 {
			o.PollInterval = *pollFlag
		}
	}
	if o.state&reportArg == 0 {
		if *repFlag > 0 {
			o.ReportInterval = *repFlag
		}
	}
}

var WithCmdSrv = func(o *options) {
	if o.state == fullServConfig {
		return
	}

	addrFlag := flag.String(
		"a", m.DefaultEndpoint, "Input the endpoint Address: <host:port>")

	storInterFlag := flag.Int(
		"i", m.DefaultStoreInterval, "Input the store interval: <sec>")

	filePathFlag := flag.String(
		"f", m.DefaultStorePath, "Input file storage path: </path/to/file")

	restoreFlag := flag.String(
		"r", m.DefaultRestore, "Input restore flag: <true|false")

	flag.Parse()

	o.setAddr(*addrFlag)

	if o.state&storIntervArg == 0 {
		o.storeInterval = *storInterFlag
	}

	if o.state&storPathArg == 0 {
		o.fileStorage = *filePathFlag

	}

	if o.state&restoreArg == 0 {
		yesno, err := strconv.ParseBool(*restoreFlag)
		if err != nil {
			l.Warn("Couldn't parse RESTORE")
		} else {
			o.restore = yesno
		}
	}
}

func Configure(opts ...func(*options)) (Service, error) {
	err := l.InitLog()
	if err != nil {
		return nil, fmt.Errorf("init logger error: %w", err)
	}
	var options options
	for _, opt := range opts {
		opt(&options)
	}

	var service Service
	if options.state>>7 == 1 {
		service, err = newAgent(&options)
	} else {
		service, err = newServer(&options)
	}

	return service, err
}
