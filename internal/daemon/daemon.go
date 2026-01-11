package daemon

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"skaldenmet/internal/collectors"
	"skaldenmet/internal/comm"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Daemon struct {
	comm         comm.CommManager
	collectors   []collectors.Collector
	trackingList []comm.Process
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
	comm_handle, err := comm.Create()
	if err != nil {
		return nil, err
	}

	collectorList, err := getCollectors(v)
	if err != nil {
		return nil, err
	}
	daemon := &Daemon{
		comm:       comm_handle,
		collectors: collectorList,
	}
	return daemon, nil
}

func (d *Daemon) Finalize(ctx context.Context) {
	err := d.comm.Finalize()
	if err != nil {
		log.Printf("Error during finalization: %s", err)
	}

}

func (d *Daemon) Start(ctx context.Context) error {
	processChan := make(chan comm.Process, 100)
	go d.comm.StartListening(processChan)

	return nil
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
		defer daemon.Finalize(ctx)
		if err != nil {
			log.Printf("Deamon creation failed: %v", err)
			return
		}
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
