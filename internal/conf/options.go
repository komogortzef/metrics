package conf

import (
	"flag"
	"fmt"
	"os"

	m "metrics/internal/models"

	"github.com/caarlos0/env/v11"
)

func WithEnvCmd(cfg config) error {
	err := env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("env parse error: %w", err)
	}

	addr := flag.String("a", m.DefaultEndpoint, "Endpoint arg: -a <host:port>")
	switch c := cfg.(type) {
	case *agentConfig:
		poll := flag.Int("p", m.DefaultPollInterval, "Poll Interval arg: -p <sec>")
		rep := flag.Int("r", m.DefaultReportInterval, "Report interval arg: -r <sec>")
		flag.Parse()
		if c.Address == "none" {
			c.Address = *addr
		}
		if c.PollInterval < 0 {
			c.PollInterval = *poll
		}
		if c.ReportInterval < 0 {
			c.ReportInterval = *rep
		}
	case *serverConfig:
		storeInterv := flag.Int("i", m.DefaultStoreInterval, "Store interval arg: -i <sec>")
		filePath := flag.String("f", m.DefaultStorePath, "File path arg: -f </path/to/file>")
		rest := flag.Bool("r", m.DefaultRestore, "Restore storage arg: -r <true|false>")
		dbAddr := flag.String("d", "", "DB address arg: -d <dbserver://username:password@host:port/db_name>")
		flag.Parse()
		if c.Address == "none" {
			c.Address = *addr
		}
		if c.StoreInterval < 0 {
			c.StoreInterval = *storeInterv
		}
		if filestore, ok := os.LookupEnv("FILE_STORAGE_PATH"); !ok {
			c.FileStoragePath = *filePath
		} else {
			c.FileStoragePath = filestore
		}

		if c.Restore {
			c.Restore = *rest
		}
		if c.DBAddress == "none" {
			c.DBAddress = *dbAddr
		}
	}

	return err
}
