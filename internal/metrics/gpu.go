package metrics

import (
	"strconv"
	"time"
)

type GPUMetric struct {
	Pid_id      uint64
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

func (m *GPUMetric) Pid() uint64 {
	return m.Pid_id
}

func (m *GPUMetric) Tags() map[string]string {
	return map[string]string{
		"device": strconv.Itoa(int(m.Device)),
	}
}

func (m *GPUMetric) Fields() map[string]float64 {
	return map[string]float64{
		"UtilGPU":     m.Util,
		"Memory":      m.Memory,
		"Power":       m.PowerW,
		"Temperature": m.Temperature,
	}
}

func (m *GPUMetric) Timestamp() time.Time {
	return m.Time
}
