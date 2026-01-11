package main

import (
	run "skaldenmet/internal/cli"
	"skaldenmet/internal/daemon"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{Use: "Phylax"}

	var runCobra = run.RunCmd
	var daemonCobra = daemon.DaemonCmd
	rootCmd.AddCommand(runCobra)
	rootCmd.AddCommand(daemonCobra)
	rootCmd.Execute()
}
