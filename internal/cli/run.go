package run

import (
	"log"
	"os"
	"os/exec"
	"skaldenmet/internal/comm"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

func getLogFiles(name string) (*os.File, *os.File, error) {
	out_file, err_out := os.Create(name + ".out")
	if err_out != nil {
		return nil, nil, err_out
	}
	err_file, err_err := os.Create(name + ".err")
	if err_out != nil {
		return nil, nil, err_err
	}
	return out_file, err_file, nil
}

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "run the command",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		userCommand := args[0]

		var name string
		if varName == "" {
			name = "local"
		} else {
			name = varName
		}
		file_out, file_err, err := getLogFiles(name)
		if err != nil {
			log.Print("Failed to create files")
		}

		externalCmd := exec.Command(userCommand)
		externalCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		externalCmd.Stdout = file_out
		externalCmd.Stderr = file_err
		defer file_err.Close()
		defer file_out.Close()

		err = externalCmd.Start()
		if err != nil {
			log.Printf("failed to execute command: %s", err)
		}
		pgid, _ := syscall.Getpgid(externalCmd.Process.Pid)
		info := comm.Process{
			PGID:      int32(pgid),
			Command:   args[0],
			StartTime: time.Now(),
			Name:      name,
		}

		manager := comm.UnixSocketMonitor{SocketPath: "/tmp/skald.socket"}
		err = manager.Notify(info)
		if err != nil {
			log.Printf("failed to notify: %s", err)
		}
	},
}

var varName string

func init() {
	RunCmd.Flags().StringVarP(&varName, "name", "n", "", "name of the job")
}
