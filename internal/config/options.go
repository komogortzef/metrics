package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
)

const (
	defaultEndpoint       = "localhost:8080"
	defaultPollInterval   = 2
	defaultReportInterval = 10
	defaultStoreInterval  = 300
	defaultStorePath      = "/tmp/metrics-db.json"
	defaultRestore        = true
	defaultSendMode       = "text"
	noStorage             = ""
)

type config struct {
	Address         string `env:"ADDRESS" envDefault:"none"`
	FileStoragePath string
	DBAddress       string `env:"DATABASE_DSN" envDefault:"none"`
	StoreInterval   int    `env:"STORE_INTERVAL" envDefault:"-1"`
	PollInterval    int    `env:"POLL_INTERVAL" envDefault:"-1"`
	ReportInterval  int    `env:"REPORT_INTERVAL" envDefault:"-1"`
	Restore         bool   `env:"RESTORE" envDefault:"true"`
}

type Option func(*config) error

func WithEnv(cfg *config) (err error) {
	if err = env.Parse(cfg); err != nil {
		return fmt.Errorf("parse env err: %w", err)
	}
	return nil
}

func WithAgentFlags(cfg *config) (err error) {
	addr := flag.String("a", defaultEndpoint, "Endpoint arg: -a <host:port>")
	poll := flag.Int("p", defaultPollInterval, "Poll Interval arg: -p <sec>")
	rep := flag.Int("r", defaultReportInterval, "Report interval arg: -r <sec>")
	flag.Parse()
	if cfg.Address == "none" {
		cfg.Address = *addr
	}
	if cfg.PollInterval < 0 {
		cfg.PollInterval = *poll
	}
	if cfg.ReportInterval < 0 {
		cfg.ReportInterval = *rep
	}
	return
}

func WithServerFlags(cfg *config) (err error) {
	addr := flag.String("a", defaultEndpoint, "Endpoint arg: -a <host:port>")
	storeInterv := flag.Int("i", defaultStoreInterval, "Store interval arg: -i <sec>")
	filePath := flag.String("f", defaultStorePath, "File path arg: -f </path/to/file>")
	rest := flag.Bool("r", defaultRestore, "Restore storage arg: -r <true|false>")
	dbAddr := flag.String("d", noStorage, "DB address arg: -d <dbserver://username:password@host:port/db_name>")
	flag.Parse()
	if cfg.Address == "none" {
		cfg.Address = *addr
	}
	if cfg.StoreInterval < 0 {
		cfg.StoreInterval = *storeInterv
	}
	if filestore, ok := os.LookupEnv("FILE_STORAGE_PATH"); !ok {
		cfg.FileStoragePath = *filePath
	} else {
		cfg.FileStoragePath = filestore
	}
	if cfg.Restore {
		cfg.Restore = *rest
	}
	if cfg.DBAddress == "none" {
		cfg.DBAddress = *dbAddr
	}
	return
}
