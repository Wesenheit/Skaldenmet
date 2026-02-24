package main

import (
	"github.com/Wesenheit/Skaldenmet/internal/cli"
	"github.com/Wesenheit/Skaldenmet/internal/daemon"
	"github.com/Wesenheit/Skaldenmet/internal/display"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{Use: "met", Short: "Simple single-node SLURM-like AI/HPC resource manager", Version: cli.GetVersion()}

	var runCobra = cli.RunCmd
	var daemonCobra = daemon.DaemonCmd
	var listCobra = display.ListCmd
	rootCmd.AddCommand(runCobra)
	rootCmd.AddCommand(daemonCobra)
	rootCmd.AddCommand(listCobra)

	rootCmd.Execute()
}
