package main

import "github.com/spf13/cobra"

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
		panic(err)
	}
}
