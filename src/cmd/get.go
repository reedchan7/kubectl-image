package main

import (
	"fmt"

	"github.com/reedchan7/kubectl-image/src/pkg/client"
	"github.com/reedchan7/kubectl-image/src/pkg/getter"
	"github.com/reedchan7/kubectl-image/src/pkg/types"
	"github.com/reedchan7/kubectl-image/src/pkg/validator"
	"github.com/spf13/cobra"
)

// createGetCommand creates the 'get' subcommand
func createGetCommand() *cobra.Command {
	var options types.Options

	cmd := &cobra.Command{
		Use:   "get RESOURCE_TYPE NAME",
		Short: "Get the images of a Kubernetes resource",
		Long: `Get the images of a Kubernetes resource such as deployment or pod.

Examples:
  # Get deployment images
  kubectl image get deployment myapp
  
  # Get only the tag of the deployment's first image
  kubectl image get deploy myapp --tag
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetCommand(&options, args)
		},
	}

	cmd.Flags().BoolVarP(&options.TagOnly, "tag", "t", false, "Return only the image tag")

	return cmd
}

// runGetCommand handles the get command execution
func runGetCommand(options *types.Options, args []string) error {
	// Parse arguments
	options.ResourceType = args[0]
	options.ResourceName = args[1]

	return executeGetCommand(options)
}

// executeGetCommand executes the image get command
func executeGetCommand(options *types.Options) error {
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
	if err := v.ValidateGet(); err != nil {
		return err
	}

	// Get images
	g := getter.New(options)
	return g.Get()
}
