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
