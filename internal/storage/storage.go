package storage

import (
	"context"
	"skaldenmet/internal/comm"
	"skaldenmet/internal/metrics"
	"time"
)

type Storage interface {
	Store(context.Context, chan comm.Process, chan []metrics.Metric) error
	Close() error
	Interval() time.Duration
}
