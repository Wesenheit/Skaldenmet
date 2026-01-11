package storage

import (
	"context"
	metrics "skaldenmet/internal/metrics"
	"time"
)

type Storage interface {
	Store(ctx context.Context, batch []metrics.Metric) error
	QueryLastPids(ctx context.Context, how_many uint64) ([]metrics.SummaryMetric, error)
	QueryRange(ctx context.Context, metricName string, start, end time.Time) ([]metrics.Metric, error)
	Close() error
	Interval() time.Duration
}
