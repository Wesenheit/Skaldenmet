package comm

import "time"

type Process struct {
	PID       int       `json:"pid"`
	Command   string    `json:"command"`
	LogPath   string    `json:"log_path"`
	StartTime time.Time `json:"start_time"`
}

type CommManager interface {
	Notify(info Process) error
	Finalize() error
	StartListening(processChan chan<- Process) error
}
