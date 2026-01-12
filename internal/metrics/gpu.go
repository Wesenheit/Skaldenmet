package metrics

import (
	"time"
)

type GPUMetric struct {
	Pid_id      int32
	PPid_id     int32
	Util        float64
	Memory      float64
	Device      int
	PowerW      float64
	Temperature float64
	Time        time.Time
}

func (m *GPUMetric) Name() string {
	return "NVIDIA_GPU"
}

func (m *GPUMetric) Pid() int32 {
	return m.Pid_id
}

func (m *GPUMetric) PPid() int32 {
	return m.PPid_id
}

func (m *GPUMetric) Timestamp() time.Time {
	return m.Time
}

type GPUSummaryMetric struct {
	Start     time.Time
	End       time.Time
	AvgUtil   float64
	AvgMemory float64
	Energy    float64
	MaxTemp   float64
	Name      string
}

func AggregateUniqueGPU(before GPUSummaryMetric, metrics []GPUMetric) GPUSummaryMetric {
	if len(metrics) == 0 {
		return before
	}

	previousDuration := before.End.Sub(before.Start).Seconds()
	if before.End.IsZero() {
		previousDuration = 0
	}
	accumulatedUtil := before.AvgUtil * previousDuration
	accumulatedMemory := before.AvgMemory * previousDuration
	maxTemp := before.MaxTemp
	totalPower := before.Energy

	grouped := make(map[int32][]GPUMetric)
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
			if metric.Temperature > maxTemp {
				maxTemp = metric.Temperature
			}

			accumulatedUtil += metric.Util * timeDelta
			accumulatedMemory += metric.Memory * timeDelta
			totalPower += metric.PowerW * timeDelta / (3600 * 1000) //MW and s to W and h conversion
		}
	}

	totalDuration := latestTime.Sub(before.Start).Seconds()
	avgUtil := accumulatedUtil / totalDuration
	avgMemory := accumulatedMemory / totalDuration
	return GPUSummaryMetric{
		Start:     before.Start,
		End:       latestTime,
		AvgUtil:   avgUtil,
		AvgMemory: avgMemory,
		MaxTemp:   maxTemp,
		Energy:    totalPower,
		Name:      before.Name,
	}
}
