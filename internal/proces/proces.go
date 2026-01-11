package proces

import "time"

type Process struct {
	PGID      int32     `json:"pid"`
	Name      string    `json:"name"`
	Command   string    `json:"command"`
	LogPath   string    `json:"log_path"`
	StartTime time.Time `json:"start_time"`
}
