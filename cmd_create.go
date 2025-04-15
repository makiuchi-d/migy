package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// cmdCreate represents the create command
var cmdCreate = &cobra.Command{
	Use:   "create",
	Short: "Create a new pair of up/down SQL migration files",
	Long: `Create a new pair of SQL migration files for applying and reverting database changes.

The up file defines the forward migration. The down file contains the corresponding rollback, ensuring changes can be reversed.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("create!")
	},
}

func init() {
	cmd.AddCommand(cmdCreate)
}
