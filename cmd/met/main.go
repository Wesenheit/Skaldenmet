package main

import (
	run "github.com/Wesenheit/Skaldenmet/internal/cli"
	"github.com/Wesenheit/Skaldenmet/internal/daemon"
	"github.com/Wesenheit/Skaldenmet/internal/display"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{Use: "met"}

	var runCobra = run.RunCmd
	var daemonCobra = daemon.DaemonCmd
	var listCobra = display.ListCmd
	rootCmd.AddCommand(runCobra)
	rootCmd.AddCommand(daemonCobra)
	rootCmd.AddCommand(listCobra)

	rootCmd.Execute()
}
