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

	previousDuration := before.End.Sub(before.Start).Seconds()
	if before.End.IsZero() {
		previousDuration = 0
	}
	accumulatedCPU := before.CPU * previousDuration
	accumulatedMemory := before.Memory * previousDuration

	grouped := make(map[int32][]CPUMetric)
	for _, metric := range metrics {
		grouped[metric.Pid()] = append(grouped[metric.Pid()], metric)
	}

	var latestTime time.Time
	startTime := before.End
	if startTime.IsZero() {
		startTime = before.Start
	}

	for _, metricGroup := range grouped {
		if len(metricGroup) == 0 {
			continue
		}

		for i, metric := range metricGroup {
			if metric.Time.After(latestTime) {
				latestTime = metric.Time
			}

			var timeDelta float64
			if i == 0 {
				timeDelta = metric.Time.Sub(startTime).Seconds()
			} else {
				timeDelta = metric.Time.Sub(metricGroup[i-1].Time).Seconds()
			}

			accumulatedCPU += metric.CPU * timeDelta
			accumulatedMemory += metric.Memory * timeDelta
		}
	}

	totalDuration := latestTime.Sub(before.Start).Seconds()
	avgCPU := accumulatedCPU / totalDuration
	avgMemory := accumulatedMemory / totalDuration
	return CPUSummaryMetric{
		Start:  before.Start,
		End:    latestTime,
		CPU:    avgCPU,
		Memory: avgMemory,
		Name:   before.Name,
	}
}
