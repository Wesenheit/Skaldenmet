package daemon

import (
	"context"
	"errors"
	"log"
	"maps"
	"skaldenmet/internal/proces"
	"sync"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/spf13/viper"
)

type StateManager struct {
	sync.RWMutex
	refresh  time.Duration
	rootPIDs map[int32]struct{}
	fullTree map[int32]int32
}

func NewState(v *viper.Viper) (*StateManager, error) {
	duration := v.GetDuration("state.interval")
	if duration <= 0 {
		return nil, errors.New("wrong interval in seconds")
	}
	return &StateManager{
		refresh:  duration,
		rootPIDs: make(map[int32]struct{}),
		fullTree: make(map[int32]int32),
	}, nil
}
func runDispatcher(processChan <-chan proces.Process, pidChan chan<- int32, stateChan chan<- proces.Process) {
	for proc := range processChan {
		pidChan <- proc.PGID
		if stateChan != nil {
			stateChan <- proc
		}
	}
}
func (s *StateManager) Start(ctx context.Context, pidchan chan int32) {
	ticker := time.NewTicker(s.refresh)

	for {
		select {
		case <-ctx.Done():
			log.Print("Finalizing State managment")
			return
		case <-ticker.C:
			s.RefreshTree()

		case pid := <-pidchan:
			s.AddRoot(pid)
			s.RefreshTree()
		}
	}
}

func (s *StateManager) AddRoot(pid int32) {
	s.Lock()
	defer s.Unlock()
	s.rootPIDs[pid] = struct{}{}
}

func (s *StateManager) RefreshTree() {
	allProcs, err := process.Processes()
	if err != nil {
		return
	}

	s.Lock()
	defer s.Unlock()

	if len(s.rootPIDs) == 0 {
		s.fullTree = make(map[int32]int32)
		return
	}

	leaderAlive := make(map[int32]bool)
	for pgid := range s.rootPIDs {
		leaderAlive[pgid] = false
	}

	newFullTree := make(map[int32]int32)

	for _, p := range allProcs {
		pgidInt, err := syscall.Getpgid(int(p.Pid))
		pgid := int32(pgidInt)
		if err != nil {
			continue
		}
		if _, monitored := s.rootPIDs[pgid]; monitored {
			newFullTree[p.Pid] = pgid

			if p.Pid == pgid {
				leaderAlive[pgid] = true
			}
		}
	}

	for pgid := range s.rootPIDs {
		if !leaderAlive[pgid] {
			hasOrphans := false
			for _, root := range newFullTree {
				if root == pgid {
					hasOrphans = true
					break
				}
			}

			if hasOrphans {
				log.Printf("Warning: Job PGID %d is orphaned (Leader dead, children active)\n", pgid)

			} else {
				delete(s.rootPIDs, pgid)
			}
		}
	}

	s.fullTree = newFullTree
}

func (s *StateManager) GetSnapshot() map[int32]int32 {
	s.RLock()
	defer s.RUnlock()
	copyMap := make(map[int32]int32)
	maps.Copy(copyMap, s.fullTree)
	return copyMap
}
