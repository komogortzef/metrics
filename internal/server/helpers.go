package server

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	log "metrics/internal/logger"
	m "metrics/internal/models"

	"github.com/tidwall/gjson"
)

func addCounter(old []byte, input []byte) ([]byte, error) {
	var oldStruct m.Metrics
	if err := oldStruct.UnmarshalJSON(old); err != nil {
		return nil, fmt.Errorf("addCounter(): unmarshal error: %w", err)
	}
	num := gjson.GetBytes(input, m.Delta).Int()
	*oldStruct.Delta += num

	return oldStruct.MarshalJSON()
}

func getHelper(mtype string) helper {
	switch mtype {
	case m.Counter:
		return addCounter
	default:
		return nil
	}
}

func processURL(url string) (string, string, string) {
	strs := strings.Split(url, "/")[2:]
	for i, val := range strs {
		fmt.Println(i, ":", val)
	}
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
	log.Info("Dump starts")
	var buf []byte
	var mtx sync.RWMutex
	// объединение всех метрик в один байтовый срез(разделение с помощью '\n'):
	mtx.RLock()
	items, _ := store.List(ctx)
	for _, data := range items {
		data = append(data, byte('\n'))
		buf = append(buf, data...)
	}
	mtx.RUnlock()

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
