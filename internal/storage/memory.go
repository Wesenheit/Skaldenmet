// Package storage implements various storage options to store aggregated metrics
package storage

import (
	"context"
	"errors"
	"github.com/Wesenheit/Skaldenmet/internal/metrics"
	"github.com/Wesenheit/Skaldenmet/internal/proces"
	"github.com/spf13/viper"
	"log"
	"maps"
	"sync"
	"time"
)

type MemoryStorage struct {
	storageCPU map[int32]metrics.CPUSummaryMetric
	storageGPU map[int32]metrics.GPUSummaryMetric
	mu         sync.RWMutex
	interval   time.Duration
	maxSize    uint32
}

func NewMemoryStorage(v *viper.Viper) (*MemoryStorage, error) {

	maxSize := v.GetInt("storage.size")
	if maxSize <= 0 {
		return nil, errors.New("wrong memory storage size")
	}

	duration := v.GetDuration("storage.interval")
	if duration <= 0 {
		return nil, errors.New("wrong memory interval in seconds")
	}

	return &MemoryStorage{
		storageCPU: make(map[int32]metrics.CPUSummaryMetric),
		storageGPU: make(map[int32]metrics.GPUSummaryMetric),
		maxSize:    uint32(maxSize),
		interval:   duration,
	}, nil
}
func (m *MemoryStorage) Store(ctx context.Context, procChan chan proces.Process, metChan chan []metrics.Metric) error {
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
			m.storageCPU[proc.PGID] = metrics.CPUSummaryMetric{
				Start: proc.StartTime,
				Name:  proc.Name,
			}
			m.storageGPU[proc.PGID] = metrics.GPUSummaryMetric{
				Start: proc.StartTime,
				Name:  proc.Name,
			}
			m.mu.Unlock()

		case batch := <-metChan:
			pendingMetrics = append(pendingMetrics, batch...)
		}
	}
}
func (m *MemoryStorage) Close() error {
	return nil
}

func AggregateAny[S any, T metrics.Metric, V any](
	metList []metrics.Metric,
	storage map[int32]S,
	aggregator func(S, []V) S,
	convert func(T) V,
) {
	toAggregate := make(map[int32][]V)
	for _, met := range metList {
		if specific, ok := met.(T); ok {
			ppid := specific.PPid()
			toAggregate[ppid] = append(toAggregate[ppid], convert(specific))
		}
	}
	for ppid, list := range toAggregate {
		storage[ppid] = aggregator(storage[ppid], list)
	}
}

func (m *MemoryStorage) AggregateBatch(metList []metrics.Metric) {
	m.mu.Lock()
	defer m.mu.Unlock()
	AggregateAny(metList, m.storageCPU, metrics.AggregateUniqueCPU, func(ptr *metrics.CPUMetric) metrics.CPUMetric { return *ptr })
	AggregateAny(metList, m.storageGPU, metrics.AggregateUniqueGPU, func(ptr *metrics.GPUMetric) metrics.GPUMetric { return *ptr })
}

func GetSnapshot[T any](storage map[int32]T, mu *sync.RWMutex) map[int32]T {
	mu.RLock()
	defer mu.RUnlock()
	snapshot := maps.Clone(storage)
	return snapshot
}

func (m *MemoryStorage) GetCPUSnapshot() map[int32]metrics.CPUSummaryMetric {
	return GetSnapshot(m.storageCPU, &m.mu)
}

func (m *MemoryStorage) GetGPUSnapshot() map[int32]metrics.GPUSummaryMetric {
	return GetSnapshot(m.storageGPU, &m.mu)
}

func (m *MemoryStorage) Interval() time.Duration {
	return m.interval
}
