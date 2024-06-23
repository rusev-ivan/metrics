package main

import (
	"log"
	"net/http"
	"strconv"
	"sync"
)

const (
	CounterMetricType = "counter"
	GaugeMetricType   = "gauge"
)

type InMemStorage struct {
	counters   map[string]int64
	countersMu sync.Mutex

	gauges   map[string]float64
	gaugesMu sync.Mutex
}

func NewInMemStorage() InMemStorage {
	return InMemStorage{
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
	}
}

func (s *InMemStorage) UpdateCounter(name string, value int64) error {
	s.countersMu.Lock()
	defer s.countersMu.Unlock()

	s.counters[name] += value
	return nil
}

func (s *InMemStorage) UpdateGauge(name string, value float64) error {
	s.gaugesMu.Lock()
	defer s.gaugesMu.Unlock()

	s.gauges[name] = value
	return nil
}

func main() {
	storage := NewInMemStorage()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /update/{type}/{name}/{value}", func(res http.ResponseWriter, req *http.Request) {
		metricType := req.PathValue("type")
		metricName := req.PathValue("name")
		metricValue := req.PathValue("value")

		if metricName == "" {
			http.Error(res, "Not found", http.StatusNotFound)
			return
		}

		switch metricType {
		case CounterMetricType:
			if value, err := strconv.ParseInt(metricValue, 10, 64); err == nil {
				storage.UpdateCounter(metricName, value)
			} else {
				http.Error(res, "Bad request", http.StatusBadRequest)
			}
		case GaugeMetricType:
			if value, err := strconv.ParseFloat(metricValue, 64); err == nil {
				storage.UpdateGauge(metricName, value)
			} else {
				http.Error(res, "Bad request", http.StatusBadRequest)
			}
		default:
			http.Error(res, "Bad request", http.StatusBadRequest)
		}
	})

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}

}
