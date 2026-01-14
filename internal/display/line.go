package display

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"skaldenmet/internal/metrics"
	"skaldenmet/internal/proces"
	"sort"
	"syscall"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func IsProcessActive(pgid int32) bool {
	err := syscall.Kill(-int(pgid), 0)

	if err == nil {
		return true
	}

	if errors.Is(err, syscall.ESRCH) {
		return false
	}

	if errors.Is(err, syscall.EPERM) {
		return true
	}

	return false
}
func RenderTableCPU(data map[int32]metrics.CPUSummaryMetric) {
	table := tablewriter.NewWriter(os.Stdout)

	table.Header([]string{"PPID", "Process Name", "CPU % (AVG)", "MEM % (AVG)", "Status", "Duration"})

	keys := make([]int, 0, len(data))
	for k := range data {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	for _, pid := range keys {
		metric := data[int32(pid)]
		var status string
		var duration time.Duration
		if IsProcessActive(int32(pid)) {
			status = "Active"
			duration = time.Now().Sub(metric.Start)
		} else {
			status = "Finished"
			if !metric.End.IsZero() {
				duration = metric.End.Sub(metric.Start)
			}

		}
		row := []string{
			fmt.Sprintf("%d", pid),
			metric.Name,
			fmt.Sprintf("%.2f%%", metric.CPU),
			fmt.Sprintf("%.2f%%", metric.Memory),
			status,
			duration.Truncate(time.Second).String(),
		}
		table.Append(row)
	}

	table.Render()
}
func RenderTableGPU(data map[int32]metrics.GPUSummaryMetric) {
	table := tablewriter.NewWriter(os.Stdout)

	table.Header([]string{"PPID", "Process Name", "GPU Util (AVG)", "MEM (AVG)", "Total power", "Max Temp", "Status", "Duration"})

	keys := make([]int, 0, len(data))
	for k := range data {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	for _, pid := range keys {
		metric := data[int32(pid)]
		var status string
		var duration time.Duration
		if IsProcessActive(int32(pid)) {
			status = "Active"
			duration = time.Now().Sub(metric.Start)
		} else {
			status = "Finished"
			if !metric.End.IsZero() {
				duration = metric.End.Sub(metric.Start)
			}

		}
		row := []string{
			fmt.Sprintf("%d", pid),
			metric.Name,
			fmt.Sprintf("%.2f%%", metric.AvgUtil),
			fmt.Sprintf("%.2f GB", metric.AvgMemory),
			fmt.Sprintf("%.2f Wh", metric.Energy),
			fmt.Sprintf("%.2f C", metric.MaxTemp),
			status,
			duration.Truncate(time.Second).String(),
		}
		table.Append(row)
	}

	table.Render()
}

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "list the files",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		socketPath := "/tmp/skald_serve.socket"
		conn, err := net.Dial("unix", socketPath)
		if err != nil {
			log.Fatalf("could not connect to daemon: %v", err)
			return
		}
		defer conn.Close()
		var request proces.Request

		if args[0] == "cpu" {
			request = proces.Request{Type: "cpu"}
			err = json.NewEncoder(conn).Encode(request)
			if err != nil {
				log.Printf("Failed to encode: %v", err)
				return
			}

			var data map[int32]metrics.CPUSummaryMetric
			err = json.NewDecoder(conn).Decode(&data)
			if err != nil {
				log.Fatalf("failed to decode response: %v", err)
				return
			}
			RenderTableCPU(data)
		} else if args[0] == "gpu" {
			request = proces.Request{Type: "gpu"}
			err = json.NewEncoder(conn).Encode(request)
			if err != nil {
				log.Printf("Failed to encode: %v", err)
				return
			}

			var data map[int32]metrics.GPUSummaryMetric
			err = json.NewDecoder(conn).Decode(&data)
			if err != nil {
				log.Fatalf("failed to decode response: %v", err)
				return
			}
			RenderTableGPU(data)
		} else {
			log.Fatal("Unknown type of data!")
		}

	},
}
