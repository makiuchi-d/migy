package main

import (
	"runtime/debug"

	"github.com/spf13/cobra"
)

var cmdVersion = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Long:  `Show version`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return showVersion()
	},
}

func init() {
	cmd.AddCommand(cmdVersion)
}

func showVersion() error {
	ver := "(devel)"
	if bi, ok := debug.ReadBuildInfo(); ok {
		ver = bi.Main.Version
	}
	info(ver)
	return nil
}
