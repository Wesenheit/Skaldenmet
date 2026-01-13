package collectors

import (
	"errors"
	"log"
	"time"

	"skaldenmet/internal/metrics"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/spf13/viper"
)

type NVIDIAMonitor struct {
	timeout      time.Duration
	buffer       []metrics.Metric
	max_size     int
	device_count int16
}

func (c *NVIDIAMonitor) Name() string {
	return "NVIDIA Monitor"
}

func (c *NVIDIAMonitor) Interval() time.Duration {
	return c.timeout
}

type NVIDIADeviceState struct {
	Util        float64
	Memory      float64
	PowerW      float64
	Temperature float64
	Time        time.Time
}

func DeviceStateToMetric(device_state *NVIDIADeviceState, pid int32, ppid int32, device_id int) *metrics.GPUMetric {
	return &metrics.GPUMetric{
		Pid_id:      pid,
		PPid_id:     ppid,
		Util:        device_state.Util,
		Memory:      device_state.Memory,
		Device:      device_id,
		PowerW:      device_state.PowerW,
		Temperature: device_state.Temperature,
		Time:        device_state.Time,
	}
}

func (c *NVIDIAMonitor) MonitorDevice(device nvml.Device) (*NVIDIADeviceState, error) {
	// Memory
	memInfo, ret := nvml.DeviceGetMemoryInfo(device)
	if ret != nvml.SUCCESS {
		return nil, errors.New("Failed (mem)")
	}

	// Utilization
	utilization, ret := nvml.DeviceGetUtilizationRates(device)
	if ret != nvml.SUCCESS {
		return nil, errors.New("Failed (util)")
	}

	// Temperature
	temp, ret := nvml.DeviceGetTemperature(device, nvml.TEMPERATURE_GPU)
	if ret != nvml.SUCCESS {
		return nil, errors.New("Failed (temp)")
	}

	// Power
	power, ret := nvml.DeviceGetPowerUsage(device)
	if ret != nvml.SUCCESS {
		return nil, errors.New("Failed (power)")
	}
	metric := &NVIDIADeviceState{
		Memory:      float64(memInfo.Used) / (1024 * 1024 * 1024),
		Util:        float64(utilization.Gpu),
		Temperature: float64(temp),
		PowerW:      float64(power),
		Time:        time.Now(),
	}

	return metric, nil
}

func (c *NVIDIAMonitor) Collect(storage_chan chan []metrics.Metric, targets map[int32]int32) error {
	for device_id := 0; device_id < int(c.device_count); device_id++ {
		device, ret := nvml.DeviceGetHandleByIndex(device_id)
		if ret != nvml.SUCCESS {
			continue
		}

		computeProcs, ret := device.GetComputeRunningProcesses()
		if ret != nvml.SUCCESS {
			continue
		}

		dev_state, err := c.MonitorDevice(device)
		if err != nil {
			continue
		}

		for _, proc := range computeProcs {
			pid := int32(proc.Pid)

			if _, isTarget := targets[pid]; isTarget {
				metric := DeviceStateToMetric(dev_state, pid, targets[pid], device_id)
				c.buffer = append(c.buffer, metric)
			}
		}
	}

	if len(c.buffer) >= c.max_size {
		out := make([]metrics.Metric, len(c.buffer))
		copy(out, c.buffer)

		storage_chan <- out
		c.buffer = c.buffer[:0] // Reuse memory
	}

	return nil
}

func (c *NVIDIAMonitor) Finalize() error {
	if nvml.Shutdown() == nvml.SUCCESS {
		return nil
	} else {
		return errors.New("Failed to shut down NVML")
	}
}

func NewNVIDIAMonitor(v *viper.Viper) (*NVIDIAMonitor, error) {
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		return nil, errors.New("Failed to initalize")
	}
	deviceCount, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return nil, errors.New("Failed to get device count")
	}
	log.Printf("NVIDIA: Found %d GPU(s)", deviceCount)

	max_size := v.GetInt("nvidiaCollector.size")
	if max_size <= 0 {
		return nil, errors.New("Wrong size")
	}

	duration := v.GetDuration("nvidiaCollector.interval")
	if duration <= 0 {
		return nil, errors.New("Wrong interval in seconds")
	}
	return &NVIDIAMonitor{
		timeout:      duration,
		device_count: int16(deviceCount),
		max_size:     max_size,
		buffer:       []metrics.Metric{},
	}, nil
}
