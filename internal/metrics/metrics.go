package metrics

import (
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

func IsRunning(pid uint64) bool {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return false
	}
	return p != nil
}

type SummaryMetric struct {
	Pid     uint64
	Running bool
	Name    string
}

type Metric interface {
	Name() string
	Pid() uint64
	Tags() map[string]string
	Fields() map[string]float64
	Timestamp() time.Time
}

type CPUMetric struct {
	Pid_id uint64
	CPU    float64
	Memory float64
	Time   time.Time
}

func (m *CPUMetric) Name() string {
	return "CPU"
}

func (m *CPUMetric) Pid() uint64 {
	return m.Pid_id
}

func (m *CPUMetric) Tags() map[string]string {
	return map[string]string{}
}

func (m *CPUMetric) Fields() map[string]float64 {
	return map[string]float64{
		"CPU":    m.CPU,
		"Memory": m.Memory,
	}
}

func (m *CPUMetric) Timestamp() time.Time {
	return m.Time
}

func Summarize(metric Metric) SummaryMetric {
	return SummaryMetric{
		Pid:     metric.Pid(),
		Running: IsRunning(metric.Pid()),
		Name:    metric.Name(),
	}
}
