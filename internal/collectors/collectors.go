package collectors

import (
	"errors"
	"github.com/Wesenheit/Skaldenmet/internal/metrics"
	"time"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/spf13/viper"
)

type Collector interface {
	Name() string
	Collect(storage_chan chan []metrics.Metric, targets map[int32]int32) error
	Interval() time.Duration
	Finalize() error
}

type CpuBaseCollector struct {
	timout time.Duration
	buffer []metrics.Metric
	size   int
}

func NewCpuBaseCollector(v *viper.Viper) (*CpuBaseCollector, error) {
	duration := v.GetDuration("cpuCollector.interval")
	if duration <= 0 {
		return nil, errors.New("Wrong interval in seconds")
	}

	size := v.GetInt("cpuCollector.size")
	if size <= 0 {
		return nil, errors.New("Wrong size")
	}

	return &CpuBaseCollector{
		timout: duration,
		buffer: []metrics.Metric{},
		size:   size,
	}, nil

}

func (c *CpuBaseCollector) Collect(storage_chan chan []metrics.Metric, targets map[int32]int32) error {

	for pid, ppid := range targets {
		p, err := process.NewProcess(pid)
		if err != nil {
			continue
		}

		cpuPer, err := p.CPUPercent()
		if err != nil {
			continue
		}

		memPer, err := p.MemoryPercent()
		if err != nil {
			continue
		}

		newMetric := &metrics.CPUMetric{
			Pid_id: pid,
			PPID:   ppid,
			CPU:    cpuPer,
			Memory: float64(memPer),
			Time:   time.Now(),
		}

		c.buffer = append(c.buffer, newMetric)
	}

	if len(c.buffer) >= c.size {
		out := make([]metrics.Metric, len(c.buffer))
		copy(out, c.buffer)

		storage_chan <- out
		c.buffer = c.buffer[:0]
	}

	return nil
}

func (c *CpuBaseCollector) Name() string {
	return "BaseCPU"
}

func (c *CpuBaseCollector) Interval() time.Duration {
	return c.timout
}

func (c *CpuBaseCollector) Finalize() error {
	return nil
}
