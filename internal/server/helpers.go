package server

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	log "metrics/internal/logger"
	"metrics/internal/service"

	"github.com/tidwall/gjson"
)

func addCounter(old []byte, input []byte) ([]byte, error) {
	var oldStruct service.Metrics
	if err := oldStruct.UnmarshalJSON(old); err != nil {
		return nil, fmt.Errorf("addCounter(): unmarshal error: %w", err)
	}
	num := gjson.GetBytes(input, service.Delta).Int()
	*oldStruct.Delta += num

	return oldStruct.MarshalJSON()
}

func getHelper(mtype string) helper {
	switch mtype {
	case service.Counter:
		return addCounter
	default:
		return nil
	}
}

func processURL(url string) (string, string, string) {
	strs := strings.Split(url, "/")[2:]
	switch len(strs) {
	case 0:
		return "", "", ""
	case 1:
		return strs[0], "", ""
	case 2:
		return strs[0], strs[1], ""
	}
	return strs[0], strs[1], strs[2]
}

func dump(ctx context.Context, path string, store Repository) error {
	log.Info("Dump to file...")
	var buf []byte
	// объединение всех метрик в один байтовый срез(разделение с помощью '\n'):
	items, _ := store.List(ctx)
	for _, data := range items {
		data = append(data, byte('\n'))
		buf = append(buf, data...)
	}

	return os.WriteFile(path, buf, 0666)
}

func dumpWait(ctx context.Context, store Repository, path string, interval int) {
	log.Info("fs.DumpWait run...")
	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := dump(ctx, path, store); err != nil {
					log.Warn("fs.dumpWithinterval(): Couldn't save data to file")
					return
				}
			case <-ctx.Done():
				log.Info("DumpWait end...")
				return
			}
		}
	}()
}
