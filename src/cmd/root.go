package main

import (
	"github.com/spf13/cobra"
)

// CreateImageCommand creates the main kubectl-image command with subcommands
func CreateImageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubectl-image",
		Short: "Manage Kubernetes resource images",
		Long: `Manage Kubernetes resource images with set and get operations.

This is a kubectl plugin. Install it and use as:
  kubectl image set deployment myapp nginx:1.20
  kubectl image get deployment myapp
`,
	}

	// Add subcommands
	cmd.AddCommand(createSetCommand())
	cmd.AddCommand(createGetCommand())
	cmd.AddCommand(createVersionCommand())

	return cmd
}

// Execute runs the command.
func Execute() error {
	rootCmd := CreateImageCommand()
	rootCmd.CompletionOptions = cobra.CompletionOptions{
		DisableDefaultCmd: true,
	}
	return rootCmd.Execute()
}
