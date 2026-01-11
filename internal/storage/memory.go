package storage

import (
	"context"
	"errors"
	"log"
	"skaldenmet/internal/comm"
	"skaldenmet/internal/metrics"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type MemoryStorage struct {
	storage_CPU map[int32]metrics.CPUSummaryMetric
	storage_GPU map[int32]metrics.CPUSummaryMetric
	mu          sync.RWMutex
	interval    time.Duration
	maxSize     uint32
}

func NewMemoryStorage(v *viper.Viper) (*MemoryStorage, error) {

	maxSize := v.GetInt("storage.size")
	if maxSize <= 0 {
		return nil, errors.New("Wrong memory storage size")
	}

	duration := v.GetDuration("storage.interval")
	if duration <= 0 {
		return nil, errors.New("Wrong memory interval in seconds")
	}

	return &MemoryStorage{
		storage_CPU: make(map[int32]metrics.CPUSummaryMetric),
		maxSize:     uint32(maxSize),
		interval:    duration,
	}, nil
}
func (m *MemoryStorage) Store(ctx context.Context, procChan chan comm.Process, metChan chan []metrics.Metric) error {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	var pendingMetrics []metrics.Metric

	for {
		select {
		case <-ctx.Done():
			log.Print("Finalizing Storage")
			m.AggregateBatch(pendingMetrics)
			return m.Close()

		case <-ticker.C:
			m.AggregateBatch(pendingMetrics)
			pendingMetrics = nil

		case proc := <-procChan:
			m.mu.Lock()
			m.storage_CPU[proc.PGID] = metrics.CPUSummaryMetric{
				Start: proc.StartTime,
				Name:  proc.Name,
			}
			log.Printf("Process %s,%d", proc.Name, proc.PGID)
			m.mu.Unlock()

		case batch := <-metChan:
			pendingMetrics = append(pendingMetrics, batch...)
		}
	}
}
func (m *MemoryStorage) Close() error {
	return nil
}

func (m *MemoryStorage) AggregateBatch(met_list []metrics.Metric) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var toAggregateCPU map[int32][]metrics.CPUMetric
	for _, metric := range met_list {
		cpuMetric, ok := metric.(*metrics.CPUMetric)
		if ok {
			PPID := metric.PPid()
			toAggregateCPU[PPID] = append(toAggregateCPU[PPID], *cpuMetric)
		}
	}
	for ppid, aggregated := range toAggregateCPU {
		m.storage_CPU[ppid] = metrics.AggregateUniqueCPU(m.storage_CPU[ppid], aggregated)
	}
}

func (m *MemoryStorage) Interval() time.Duration {
	return m.interval
}
