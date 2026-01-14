package comm

import (
	"github.com/Wesenheit/Skaldenmet/internal/proces"
	"github.com/Wesenheit/Skaldenmet/internal/storage"
)

type CommManager interface {
	Notify(info proces.Process) error
	Finalize() error
	StartListening(processChan chan<- proces.Process) error
	ServeQueries(storage.Storage)
}
