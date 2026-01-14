package storage

import (
	"context"
	"github.com/Wesenheit/Skaldenmet/internal/metrics"
	"github.com/Wesenheit/Skaldenmet/internal/proces"
	"time"
)

type Storage interface {
	Store(context.Context, chan proces.Process, chan []metrics.Metric) error
	Close() error
	Interval() time.Duration
	GetCPUSnapshot() map[int32]metrics.CPUSummaryMetric
	GetGPUSnapshot() map[int32]metrics.GPUSummaryMetric
}
