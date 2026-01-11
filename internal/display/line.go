package display

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"skaldenmet/internal/metrics"
	"sort"
	"syscall"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func IsProcessActive(pid int32) bool {
	process, err := os.FindProcess(int(pid))
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}
func RenderTable(data map[int32]metrics.CPUSummaryMetric) {
	table := tablewriter.NewWriter(os.Stdout)

	table.Header([]string{"PID", "Process Name", "CPU % (AVG)", "MEM % (AVG)", "Status"})

	keys := make([]int, 0, len(data))
	for k := range data {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	for _, pid := range keys {
		metric := data[int32(pid)]
		var status string
		if IsProcessActive(int32(pid)) {
			status = "Active"
		} else {
			status = "Finished"
		}
		row := []string{
			fmt.Sprintf("%d", pid),
			metric.Name,
			fmt.Sprintf("%.2f%%", metric.CPU),
			fmt.Sprintf("%.2f%%", metric.Memory),
			status,
		}
		table.Append(row)
	}

	// 5. Finally, render to terminal
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

		if args[0] == "cpu" {
			var data map[int32]metrics.CPUSummaryMetric
			err = json.NewDecoder(conn).Decode(&data)
			if err != nil {
				log.Fatalf("failed to decode response: %v", err)
				return
			}
			RenderTable(data)
		} else {
			log.Fatal("Unknown type of data!")
		}

	},
}
