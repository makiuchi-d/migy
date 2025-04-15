package main

import "github.com/spf13/cobra"

var cmd = &cobra.Command{
	Use:   "migy",
	Short: "A standalone database migration helper for MySQL",
	Long:  "",
}

func main() {
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
