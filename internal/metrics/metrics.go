package metrics

import (
	"time"
)

type Metric interface {
	Pid() int32
	PPid() int32
	Timestamp() time.Time
}

type CPUMetric struct {
	Pid_id int32
	PPID   int32
	CPU    float64
	Memory float64
	Time   time.Time
}

func (m *CPUMetric) Pid() int32 {
	return m.Pid_id
}
func (m *CPUMetric) PPid() int32 {
	return m.PPID
}

func (m *CPUMetric) Timestamp() time.Time {
	return m.Time
}

type CPUSummaryMetric struct {
	Start  time.Time
	End    time.Time
	CPU    float64
	Memory float64
	Name   string
}

func AggregateUniqueCPU(before CPUSummaryMetric, metrics []CPUMetric) CPUSummaryMetric {
	if len(metrics) == 0 {
		return before
	}
	var CPU float64
	var Memory float64
	previous := before.End
	var end_time time.Time
	if previous.IsZero() {
		end_time = before.Start
	} else {
		end_time = before.End
	}

	for _, metric := range metrics {
		CPU += metric.CPU
		Memory += metric.Memory
	}
	previousCPU := before.CPU
	previousMemory := before.Memory

	recorder_time := metrics[0].Time
	previous_duration := end_time.Sub(before.Start).Seconds()
	current_duration := recorder_time.Sub(end_time).Seconds()

	currentCPU := (previousCPU*previous_duration + CPU*current_duration) / (current_duration + previous_duration)
	currentMemory := (previousMemory*previous_duration + Memory*current_duration) / (current_duration + previous_duration)
	return CPUSummaryMetric{Start: before.Start, End: recorder_time, CPU: currentCPU, Memory: currentMemory, Name: before.Name}
}
