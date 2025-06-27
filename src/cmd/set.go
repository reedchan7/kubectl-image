package main

import (
	"fmt"

	"github.com/reedchan7/kubectl-image/src/pkg/client"
	"github.com/reedchan7/kubectl-image/src/pkg/setter"
	"github.com/reedchan7/kubectl-image/src/pkg/types"
	"github.com/reedchan7/kubectl-image/src/pkg/validator"
	"github.com/spf13/cobra"
)

// createSetCommand creates the 'set' subcommand
func createSetCommand() *cobra.Command {
	var options types.Options

	cmd := &cobra.Command{
		Use:   "set RESOURCE_TYPE NAME [IMAGE]",
		Short: "Set the image of a Kubernetes resource",
		Long: `Set the image of a Kubernetes resource such as deployment.

Updates the first container by default for safety.

Examples:
  # Set deployment image directly
  kubectl image set deployment myapp nginx:1.20
  kubectl image set deploy myapp nginx:1.21
  
  # Set using tag only (preserves base image name)
  kubectl image set deployment myapp --tag v1.0.1
  kubectl image set deploy myapp -t v1.0.2
  
  # Set specific container
  kubectl image set deployment myapp --tag v1.0.1 --container app-container
`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetCommand(&options, args)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&options.Tag, "tag", "t", "", "Image tag to set")
	cmd.Flags().StringVarP(&options.ContainerName, "container", "c", "", "Container name to update (if not specified, updates first container)")

	return cmd
}

// runSetCommand handles the set command execution
func runSetCommand(options *types.Options, args []string) error {
	// Parse arguments
	options.ResourceType = args[0]
	options.ResourceName = args[1]
	if len(args) >= 3 {
		options.Image = args[2]
	}

	return executeSetCommand(options)
}

// executeSetCommand executes the image set command
func executeSetCommand(options *types.Options) error {
	// Get current namespace from kubectl context
	if ns, err := client.GetCurrentNamespace(); err == nil {
		options.Namespace = ns
	} else {
		options.Namespace = "default"
	}

	// Create Kubernetes client
	clientset, err := client.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}
	options.Clientset = clientset

	// Validate input
	v := validator.New(options)
	if err := v.ValidateSet(); err != nil {
		return err
	}

	// Set image
	s := setter.New(options)
	return s.Set()
}
