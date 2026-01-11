package storage

import (
	"context"
	"skaldenmet/internal/metrics"
	"skaldenmet/internal/proces"
	"time"
)

type Storage interface {
	Store(context.Context, chan proces.Process, chan []metrics.Metric) error
	Close() error
	Interval() time.Duration
	GetCPUSnapshot() map[int32]metrics.CPUSummaryMetric
}
