package comm

import (
	"skaldenmet/internal/proces"
	"skaldenmet/internal/storage"
)

type CommManager interface {
	Notify(info proces.Process) error
	Finalize() error
	StartListening(processChan chan<- proces.Process) error
	ServeQueries(stor storage.Storage)
}
