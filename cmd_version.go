package main

import (
	"fmt"
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
	fmt.Println("migy version", getVersion())
	return nil
}

func getVersion() string {
	ver := "(devel)"
	if bi, ok := debug.ReadBuildInfo(); ok {
		ver = bi.Main.Version
	}
	return ver
}
