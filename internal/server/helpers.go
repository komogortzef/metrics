package server

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	l "metrics/internal/logger"
	m "metrics/internal/models"

	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

func (fs *FileStorage) restoreFromFile() (len int) {
	b, err := os.ReadFile(fs.FilePath)
	if err != nil {
		l.Warn("No file to restore!")
		return
	}
	buff := bytes.NewBuffer(b)
	scanner := bufio.NewScanner(buff)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		bytes := scanner.Bytes()
		name := gjson.GetBytes(bytes, m.ID).String()
		len, _ = fs.Put(name, bytes)
	}
	l.Info("number of metrics recovered from the file", zap.Int("len", len))
	return

}

func addCounter(old []byte, input []byte) ([]byte, error) {
	var oldStruct m.Metrics
	if err := oldStruct.UnmarshalJSON(old); err != nil {
		return nil, fmt.Errorf("addCounter(): unmarshal error: %w", err)
	}
	num := gjson.GetBytes(input, m.Delta).Int()
	*oldStruct.Delta += num

	return oldStruct.MarshalJSON()
}

func getList(storage Repository) [][]byte {
	switch s := storage.(type) {
	case *MemStorage:
		return s.listFromMem()
	case *FileStorage:
		return s.listFromMem()
	}
	return nil
}
