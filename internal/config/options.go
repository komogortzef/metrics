package config

import (
	"flag"
	"fmt"
	"metrics/internal/logger"
	"os"
	"regexp"
	"strconv"

	"github.com/caarlos0/env/v6"
)

const (
	addrArg       = 1  // 000001
	pollArg       = 2  // 000010
	reportArg     = 4  // 000100
	storIntervArg = 8  // 001000
	storPathArg   = 16 // 010000
	restoreArg    = 32 // 100000
	fullConfig    = 63 // 111111

	defaultAddr           = "localhost:8080"
	defaultPollInterval   = 2
	defaultReportInterval = 10
	defStorPath           = "/tmp/metrics-db.json"
	defStorInterv         = 300
)

var (
	isValidAddr = regexp.MustCompile(`^(.*):(\d+)$`).MatchString
	isValidPath = regexp.MustCompile(`^(/[^/\0]+)+/?$`).MatchString
)

type options struct {
	Address        string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	storeInterval  int
	fileStorage    string
	restore        bool

	state uint8
}

type Option func(*options)

func (o *options) setAddr(addr ...string) {
	if len(addr) == 0 {
		err := env.Parse(o)
		if err != nil {
			logger.Warn("Coulnd't parse env")
		}
		if isValidAddr(o.Address) {
			o.state |= addrArg
		}
		return
	}

	if o.state&addrArg == 0 {
		if isValidAddr(addr[0]) {
			o.Address = addr[0]
		}
	}
}

var WithEnvAg = func(o *options) {
	o.setAddr()

	if o.PollInterval > 0 {
		o.state |= pollArg
	}
	if o.ReportInterval > 0 {
		o.state |= reportArg
	}
}

var WithEnvSrv = func(o *options) {
	fmt.Printf("With env begin: %b\n", o.state)
	o.setAddr()

	if val, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		num, err := strconv.Atoi(val)
		if err != nil {
			logger.Warn("Couldn't parse STORE_INTERVAL")
		} else {
			o.storeInterval = num
			o.state |= storIntervArg
		}
	}

	if val, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		if isValidPath(val) {
			o.fileStorage = val
			o.state |= storPathArg
		} else {
			logger.Warn("Couldn't parse FILE_STORAGE_PATH")
		}
	}

	if val, ok := os.LookupEnv("RESTORE"); ok {
		yesno, err := strconv.ParseBool(val)
		if err != nil {
			logger.Warn("Couldn't parse RESTORE")
		} else {
			o.restore = yesno
			o.state |= restoreArg
		}
	}

	fmt.Printf("With env ends: %b\n", o.state)
}

var WithCmdAg = func(o *options) {
	addrFlag := flag.String("a", defaultAddr, "Input the endpoint Address: <host:port>")
	pollFlag := flag.Int("p", defaultPollInterval, "Input the poll interval: <sec>")
	repFlag := flag.Int("r", defaultReportInterval, "Input the report interval: <sec>")
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
	fmt.Printf("from cmd begin: %b\n", o.state)

	addrFlag := flag.String("a", defaultAddr, "Input the endpoint Address: <host:port>")
	storInterFlag := flag.Int("i", defStorInterv, "Input the store interval: <sec>")
	filePathFlag := flag.String("f", defStorPath, "Input file storage path: </path/to/file")
	restoreFlag := flag.String("r", "true", "Input restore flag: <true|false")
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
			logger.Warn("Couldn't parse RESTORE")
		} else {
			o.restore = yesno
		}
	}
}
