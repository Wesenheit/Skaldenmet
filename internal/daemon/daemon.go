package daemon

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"skaldenmet/internal/collectors"
	"skaldenmet/internal/comm"
	"skaldenmet/internal/metrics"
	"skaldenmet/internal/proces"
	"skaldenmet/internal/storage"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Daemon struct {
	reciver    comm.CommManager
	server     comm.CommManager
	collectors []collectors.Collector
	storage    storage.Storage
	wg         sync.WaitGroup
	manager    *StateManager
}

var NameFunMapping = map[string]func(v *viper.Viper) (collectors.Collector, error){
	"cpuCollector": func(v *viper.Viper) (collectors.Collector, error) {
		return collectors.NewCpuBaseCollector(v)
	},
	"nvidiaCollector": func(v *viper.Viper) (collectors.Collector, error) {
		return collectors.NewNVIDIAMonitor(v)
	},
}

func startCollector(v *viper.Viper, name string, start func(v *viper.Viper) (collectors.Collector, error), collectorList []collectors.Collector) []collectors.Collector {
	mapViper := v.GetStringMapString(name)
	if len(mapViper) > 0 {
		Collector, err := start(v)
		if err != nil {
			log.Printf("Error %s: %v", Collector.Name(), err)
		} else {
			collectorList = append(collectorList, Collector)
			log.Printf("Enabled %s", Collector.Name())
		}
	}
	return collectorList
}
func getCollectors(v *viper.Viper) ([]collectors.Collector, error) {
	collectorList := []collectors.Collector{}

	for name, function := range NameFunMapping {
		collectorList = startCollector(v, name, function, collectorList)
	}

	if len(collectorList) == 0 {
		return nil, errors.New("No collectors")
	}
	return collectorList, nil
}

func NewDaemon(v *viper.Viper) (*Daemon, error) {
	reciver_handle, err := comm.Create("/tmp/skald.socket")
	if err != nil {
		return nil, err
	}

	server_handle, err := comm.Create("/tmp/skald_serve.socket")
	if err != nil {
		return nil, err
	}
	collectorList, err := getCollectors(v)
	if err != nil {
		return nil, err
	}

	state, err := NewState(v)
	if err != nil {
		return nil, err
	}

	store, err := storage.NewMemoryStorage(v)
	if err != nil {
		return nil, err
	}
	daemon := &Daemon{
		reciver:    reciver_handle,
		server:     server_handle,
		collectors: collectorList,
		manager:    state,
		storage:    store,
	}
	return daemon, nil
}

func (d *Daemon) Finalize(ctx context.Context) {
	err := d.reciver.Finalize()
	if err != nil {
		log.Printf("Error during finalization: %s", err)
	}
	err = d.server.Finalize()
	if err != nil {
		log.Printf("Error during finalization: %s", err)
	}

}

func (d *Daemon) Start(ctx context.Context) error {
	processChan := make(chan proces.Process, 100)
	procStoreChan := make(chan proces.Process, 100)
	pidChan := make(chan int32, 100)
	storageChan := make(chan []metrics.Metric, 100)

	go d.reciver.StartListening(processChan)
	go d.server.ServeQueries(d.storage)

	go d.manager.Start(ctx, pidChan)
	go runDispatcher(processChan, pidChan, procStoreChan)
	go d.storage.Store(ctx, procStoreChan, storageChan)

	for _, collector := range d.collectors {
		d.wg.Add(1)
		go d.RunCollector(ctx, collector, storageChan)
	}

	<-ctx.Done()
	d.wg.Wait()
	return nil
}

func (d *Daemon) RunCollector(ctx context.Context, collector collectors.Collector,
	storageChan chan []metrics.Metric,
) error {
	interval := collector.Interval()
	name := collector.Name()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	defer d.wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Stopping collector: %s", name)
			return nil
		case <-ticker.C:
			err := collector.Collect(storageChan, d.manager.GetSnapshot())
			if err != nil {
				return err
			}
		}
	}
}

var DaemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run the monitoring daemon",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		if configFile == "" {
			log.Fatal("Please provide a config file with --config")
		}
		v := viper.New()
		v.SetConfigType("yaml")
		v.SetConfigFile(configFile)

		err := v.ReadInConfig()
		if err != nil {
			log.Fatalf("Error reading config file: %v", err)
		}

		daemon, err := NewDaemon(v)
		if err != nil {
			log.Printf("Deamon creation failed: %v", err)
			return
		}
		defer daemon.Finalize(ctx)
		if err := daemon.Start(ctx); err != nil {
			log.Printf("Daemon failed: %v", err)
		}
		log.Printf("Exiting")
	},
}

var configFile string

func init() {
	DaemonCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file path")
}
