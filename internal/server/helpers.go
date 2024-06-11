package server

import (
	"fmt"

	log "metrics/internal/logger"
	m "metrics/internal/models"

	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

func addCounter(old []byte, input []byte) ([]byte, error) {
	var oldStruct m.Metrics
	if err := oldStruct.UnmarshalJSON(old); err != nil {
		return nil, fmt.Errorf("addCounter(): unmarshal error: %w", err)
	}
	num := gjson.GetBytes(input, m.Delta).Int()
	*oldStruct.Delta += num
	log.Info("add count", zap.Int64("old", *oldStruct.Delta))

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
