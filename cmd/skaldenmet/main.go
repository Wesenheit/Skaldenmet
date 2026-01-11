package main

import (
	run "skaldenmet/internal/cli"
	"skaldenmet/internal/daemon"
	"skaldenmet/internal/display"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{Use: "Phylax"}

	var runCobra = run.RunCmd
	var daemonCobra = daemon.DaemonCmd
	var listCobra = display.ListCmd
	rootCmd.AddCommand(runCobra)
	rootCmd.AddCommand(daemonCobra)
	rootCmd.AddCommand(listCobra)

	rootCmd.Execute()
}
