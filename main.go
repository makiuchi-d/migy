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
	quit      bool
)

func init() {
	cmd.PersistentFlags().StringVarP(&targetDir, "dir", "d", ".", "directory with migration files")
	cmd.PersistentFlags().BoolVarP(&quit, "quit", "q", false, "quit stdout")
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

func info(a ...any) {
	if quit {
		return
	}
	fmt.Fprintln(os.Stdout, a...)
}
