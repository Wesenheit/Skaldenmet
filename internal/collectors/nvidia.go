package collectors

import (
	"errors"
	"log"
	"time"

	"github.com/Wesenheit/Skaldenmet/internal/metrics"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/spf13/viper"
)

type NVIDIAMonitor struct {
	timeout     time.Duration
	buffer      []metrics.Metric
	maxSize     int
	deviceCount int16
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

func DeviceStateToMetric(deviceState *NVIDIADeviceState, pid int32, ppid int32, deviceID int, totalProc int) *metrics.GPUMetric {
	return &metrics.GPUMetric{
		Pid_id:      pid,
		PPid_id:     ppid,
		Util:        deviceState.Util / float64(totalProc),
		Memory:      deviceState.Memory / float64(totalProc),
		Device:      deviceID,
		PowerW:      deviceState.PowerW / float64(totalProc),
		Temperature: deviceState.Temperature,
		Time:        deviceState.Time,
	}
}

func (c *NVIDIAMonitor) MonitorDevice(device nvml.Device) (*NVIDIADeviceState, error) {
	// Memory
	memInfo, ret := nvml.DeviceGetMemoryInfo(device)
	if ret != nvml.SUCCESS {
		return nil, errors.New("nvidia failed (mem)")
	}

	// Utilization
	utilization, ret := nvml.DeviceGetUtilizationRates(device)
	if ret != nvml.SUCCESS {
		return nil, errors.New("nvidia failed (util)")
	}

	// Temperature
	temp, ret := nvml.DeviceGetTemperature(device, nvml.TEMPERATURE_GPU)
	if ret != nvml.SUCCESS {
		return nil, errors.New("nvidia failed (temp)")
	}

	// Power
	power, ret := nvml.DeviceGetPowerUsage(device)
	if ret != nvml.SUCCESS {
		return nil, errors.New("nvidia failed (power)")
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

func (c *NVIDIAMonitor) Collect(storageChan chan []metrics.Metric, targets map[int32]int32) error {
	for deviceID := 0; deviceID < int(c.deviceCount); deviceID++ {
		device, ret := nvml.DeviceGetHandleByIndex(deviceID)
		if ret != nvml.SUCCESS {
			continue
		}

		computeProcs, ret := device.GetComputeRunningProcesses()
		if ret != nvml.SUCCESS {
			continue
		}

		devState, err := c.MonitorDevice(device)
		if err != nil {
			continue
		}

		for _, proc := range computeProcs {
			pid := int32(proc.Pid)

			if _, isTarget := targets[pid]; isTarget {
				metric := DeviceStateToMetric(devState, pid, targets[pid], deviceID, len(computeProcs))
				c.buffer = append(c.buffer, metric)
			}
		}
	}

	if len(c.buffer) >= c.maxSize {
		out := make([]metrics.Metric, len(c.buffer))
		copy(out, c.buffer)

		storageChan <- out
		c.buffer = c.buffer[:0] // Reuse memory
	}

	return nil
}

func (c *NVIDIAMonitor) Finalize() error {
	if nvml.Shutdown() == nvml.SUCCESS {
		return nil
	} else {
		return errors.New("failed to shut down NVML")
	}
}

func NewNVIDIAMonitor(v *viper.Viper) (*NVIDIAMonitor, error) {
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		return nil, errors.New("failed to initalize")
	}
	deviceCount, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return nil, errors.New("failed to get device count")
	}
	log.Printf("NVIDIA: Found %d GPU(s)", deviceCount)

	maxSize := v.GetInt("nvidiaCollector.size")
	if maxSize <= 0 {
		return nil, errors.New("wrong size")
	}

	duration := v.GetDuration("nvidiaCollector.interval")
	if duration <= 0 {
		return nil, errors.New("wrong interval in seconds")
	}
	return &NVIDIAMonitor{
		timeout:     duration,
		deviceCount: int16(deviceCount),
		maxSize:     maxSize,
		buffer:      []metrics.Metric{},
	}, nil
}
