package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:           "migy",
	Short:         "A standalone database migration helper for MySQL",
	Long:          "",
	SilenceErrors: true,
	SilenceUsage:  true,
}

var (
	targetDir string
)

func init() {
	cmd.PersistentFlags().StringVarP(&targetDir, "dir", "d", ".", "directory with migration files")
}

func main() {
	err := cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err.Error())
		os.Exit(-1)
	}
}

func warning(msg string) {
	fmt.Fprintln(os.Stderr, "warning:", msg)
}
