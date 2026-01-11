package run

import (
	"log"
	"os"
	"os/exec"
	"skaldenmet/internal/comm"
	"skaldenmet/internal/proces"
	"strings"
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
	Run: func(cmd *cobra.Command, args []string) {
		dashIndex := cmd.ArgsLenAtDash()
		var userCommand string
		if dashIndex == -1 {
			userCommand = strings.Join(args, " ")
		} else {
			userCommand = strings.Join(args[dashIndex:], " ")
		}

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

		Cmd := exec.Command("sh", "-c", userCommand)
		Cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		Cmd.Env = os.Environ()
		Cmd.Stdout = file_out
		Cmd.Stderr = file_err
		defer file_err.Close()
		defer file_out.Close()

		err = Cmd.Start()
		if err != nil {
			log.Printf("failed to execute command: %s", err)
		}
		pgid, _ := syscall.Getpgid(Cmd.Process.Pid)
		info := proces.Process{
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
