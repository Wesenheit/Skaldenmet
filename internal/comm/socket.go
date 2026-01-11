package comm

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"skaldenmet/internal/proces"
	"skaldenmet/internal/storage"
)

type UnixSocketMonitor struct {
	SocketPath string
	listner    net.Listener
}

func (u *UnixSocketMonitor) Notify(info proces.Process) error {
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

func (u *UnixSocketMonitor) StartListening(processChan chan<- proces.Process) error {
	for {
		conn, err := u.listner.Accept()
		if err != nil {
			continue
		}
		var info proces.Process
		if err := json.NewDecoder(conn).Decode(&info); err != nil {
			log.Printf("Error during decoding %s", err)
		}
		processChan <- info
		conn.Close()
	}
}

func Create(socketPath string) (*UnixSocketMonitor, error) {
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
func (u *UnixSocketMonitor) ServeQueries(stor storage.Storage) {
	for {
		conn, err := u.listner.Accept()
		if err != nil {
			continue
		}
		go func(c net.Conn) {
			defer c.Close()
			data := stor.GetCPUSnapshot()
			err := json.NewEncoder(c).Encode(data)
			if err != nil {
				log.Printf("Failed to encode and send: %v", err)
			}
		}(conn)
	}
}
