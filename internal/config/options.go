package config

import (
	ctx "context"
	"flag"
	"fmt"
	"os"

	s "metrics/internal/service"

	"github.com/caarlos0/env/v11"
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

type Option func(ctx.Context, *config) error

func EnvFlagsAgent(_ ctx.Context, cfg *config) (err error) {
	err = env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("env parse error: %w", err)
	}
	addr := flag.String("a", s.DefaultEndpoint, "Endpoint arg: -a <host:port>")
	poll := flag.Int("p", s.DefaultPollInterval, "Poll Interval arg: -p <sec>")
	rep := flag.Int("r", s.DefaultReportInterval, "Report interval arg: -r <sec>")
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

func EnvFlagsServer(_ ctx.Context, cfg *config) (err error) {
	err = env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("env parse error: %w", err)
	}
	addr := flag.String("a", s.DefaultEndpoint, "Endpoint arg: -a <host:port>")
	storeInterv := flag.Int("i", s.DefaultStoreInterval, "Store interval arg: -i <sec>")
	filePath := flag.String("f", s.DefaultStorePath, "File path arg: -f </path/to/file>")
	rest := flag.Bool("r", s.DefaultRestore, "Restore storage arg: -r <true|false>")
	dbAddr := flag.String("d", s.NoStorage, "DB address arg: -d <dbserver://username:password@host:port/db_name>")
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
