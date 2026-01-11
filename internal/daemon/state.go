package daemon

import (
	"context"
	"errors"
	"log"
	"skaldenmet/internal/comm"
	"sync"
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
		return nil, errors.New("Wrong interval in seconds")
	}
	return &StateManager{
		refresh:  duration,
		rootPIDs: make(map[int32]struct{}),
		fullTree: make(map[int32]int32),
	}, nil
}
func runDispatcher(processChan <-chan comm.Process, pidChan chan<- int32, stateChan chan<- comm.Process) {
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

	aliveOnSystem := make(map[int32]struct{})
	parentMap := make(map[int32]int32)
	for _, p := range allProcs {
		aliveOnSystem[p.Pid] = struct{}{}
		ppid, _ := p.Ppid()
		parentMap[p.Pid] = ppid
	}

	s.Lock()
	defer s.Unlock()

	for root := range s.rootPIDs {
		if _, alive := aliveOnSystem[root]; !alive {
			delete(s.rootPIDs, root)
		}
	}

	s.fullTree = make(map[int32]int32)

	for pid := range parentMap {
		curr := pid
		for curr > 1 {
			if _, isRoot := s.rootPIDs[curr]; isRoot {
				s.fullTree[pid] = curr
				break
			}

			next, exists := parentMap[curr]
			if !exists || next == curr {
				break
			}
			curr = next
		}
	}
}

func (s *StateManager) GetSnapshot() map[int32]int32 {
	s.RLock()
	defer s.RUnlock()
	copyMap := make(map[int32]int32)
	for k, v := range s.fullTree {
		copyMap[k] = v
	}
	return copyMap
}

func ExpandTargets(seedPIDs []int32) map[int32]int32 {
	expanded := make(map[int32]int32)

	for _, pid := range seedPIDs {
		addChildRecursive(pid, expanded)
	}

	return expanded
}

func addChildRecursive(pid int32, targets map[int32]int32) {
	if _, exists := targets[pid]; exists {
		return
	}

	p, err := process.NewProcess(pid)
	if err != nil {
		return
	}

	ppid, err := p.Ppid()
	if err != nil {
		return
	}
	targets[pid] = ppid

	children, err := p.Children()
	if err != nil {
		return
	}

	for _, child := range children {
		addChildRecursive(child.Pid, targets)
	}
}
