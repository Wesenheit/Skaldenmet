package comm

import (
	"encoding/json"
	"log"
	"net"
	"os"
)

type UnixSocketMonitor struct {
	SocketPath string
	listner    net.Listener
}

func (u *UnixSocketMonitor) Notify(info Process) error {
	conn, err := net.Dial("unix", u.SocketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	return json.NewEncoder(conn).Encode(info)
}
func (u *UnixSocketMonitor) Finalize() error {
	return u.listner.Close()
}

func (u *UnixSocketMonitor) StartListening(processChan chan<- Process) error {
	for {
		conn, err := u.listner.Accept()
		if err != nil {
			continue
		}
		var info Process
		if err := json.NewDecoder(conn).Decode(&info); err != nil {
			log.Printf("Error during decoding %s", err)
		}
		processChan <- info
		conn.Close()
	}
}

func Create() (*UnixSocketMonitor, error) {
	socketPath := "/tmp/skald.socket"
	if _, err := os.Stat(socketPath); err == nil {
		if err := os.Remove(socketPath); err != nil {
			return nil, err
		}
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}

	os.Chmod(socketPath, 0660)

	log.Printf("Daemon started. Listening on %s...\n", socketPath)

	return &UnixSocketMonitor{
		SocketPath: socketPath,
		listner:    listener,
	}, nil
}
