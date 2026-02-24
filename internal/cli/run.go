// Package cli start the job from the command line
package cli

import (
	"github.com/Wesenheit/Skaldenmet/internal/comm"
	"github.com/Wesenheit/Skaldenmet/internal/proces"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

func getLogFiles(name string) (*os.File, *os.File, error) {
	outFile, errOut := os.Create(name + ".out")
	if errOut != nil {
		return nil, nil, errOut
	}
	errFile, errErr := os.Create(name + ".err")
	if errErr != nil {
		return nil, nil, errErr
	}
	return outFile, errFile, nil
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
		fileOut, fileErr, err := getLogFiles(name)
		if err != nil {
			log.Print("Failed to create files")
		}

		Cmd := exec.Command("sh", "-c", userCommand)
		Cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		Cmd.Env = os.Environ()
		Cmd.Stdout = fileOut
		Cmd.Stderr = fileErr

		err = Cmd.Start()
		if err != nil {
			log.Printf("failed to execute command: %s", err)
		}
		go func() {
			defer fileOut.Close()
			defer fileErr.Close()

			Cmd.Wait()
		}()
		pgid, _ := syscall.Getpgid(Cmd.Process.Pid)
		log.Printf("Started command %s with PPID %d", name, pgid)
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
