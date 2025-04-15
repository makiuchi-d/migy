package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// cmdInit represents the init command
var cmdInit = &cobra.Command{
	Use:   "init",
	Short: "Generate the initial migration SQL file",
	Long: `The init command generates the first migration SQL file.

This file sets up the initial state of the database, including the _migrations table used to track applied migrations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("init!")
	},
}

func init() {
	cmd.AddCommand(cmdInit)
	cmdInit.Flags().BoolP("force", "f", false, "Override the output file if it exists")
}
