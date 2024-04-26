package storage

type gauge map[string]float64
type counter map[string]int64

type MemStorage struct {
	Gauges   gauge
	Counters counter
}
