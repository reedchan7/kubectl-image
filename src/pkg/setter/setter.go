package setter

import (
	"context"
	"fmt"
	"strings"

	"github.com/reedchan7/kubectl-image/src/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ImageSetter handles image updates for Kubernetes resources
type ImageSetter struct {
	options *types.Options
}

// New creates a new ImageSetter
func New(options *types.Options) *ImageSetter {
	return &ImageSetter{
		options: options,
	}
}

// Set updates the image of the specified resource
func (s *ImageSetter) Set() error {
	resourceType, exists := types.ValidResourceTypes[strings.ToLower(s.options.ResourceType)]
	if !exists {
		return fmt.Errorf("unsupported resource type: %s", s.options.ResourceType)
	}

	switch resourceType {
	case types.ResourceTypeDeployment:
		return s.setDeploymentImage()
	case types.ResourceTypePod:
		return fmt.Errorf("direct pod image update is not supported - pods are immutable. Please update the deployment or other controller instead")
	default:
		return fmt.Errorf("unsupported resource type: %s", s.options.ResourceType)
	}
}

// setDeploymentImage updates the image of a deployment
func (s *ImageSetter) setDeploymentImage() error {
	ctx := context.TODO()
	deploymentsClient := s.options.Clientset.AppsV1().Deployments(s.options.Namespace)

	// Get the deployment
	deployment, err := deploymentsClient.Get(ctx, s.options.ResourceName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment %s: %v", s.options.ResourceName, err)
	}

	updated := false

	if s.options.ContainerName != "" {
		// Update specific container
		for i := range deployment.Spec.Template.Spec.Containers {
			container := &deployment.Spec.Template.Spec.Containers[i]
			if container.Name == s.options.ContainerName {
				newImage := s.getNewImageForContainer(container.Image)
				fmt.Printf("Updating container %s image from %s to %s\n", container.Name, container.Image, newImage)
				container.Image = newImage
				updated = true
				break
			}
		}

		if !updated {
			return fmt.Errorf("container %s not found in deployment %s", s.options.ContainerName, s.options.ResourceName)
		}
	} else {
		// Update first container only (default behavior)
		if len(deployment.Spec.Template.Spec.Containers) == 0 {
			return fmt.Errorf("no containers found in deployment %s", s.options.ResourceName)
		}

		container := &deployment.Spec.Template.Spec.Containers[0]
		newImage := s.getNewImageForContainer(container.Image)
		fmt.Printf("Updating container %s image from %s to %s\n", container.Name, container.Image, newImage)
		container.Image = newImage
		updated = true
	}

	// Update the deployment
	_, err = deploymentsClient.Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update deployment %s: %v", s.options.ResourceName, err)
	}

	fmt.Printf("deployment.apps/%s image updated\n", s.options.ResourceName)
	return nil
}

// getNewImageForContainer returns the new image name based on options
func (s *ImageSetter) getNewImageForContainer(currentImage string) string {
	if s.options.Tag != "" {
		// If using --tag flag, extract base image name
		var baseName string
		if s.options.Image != "" {
			// If image name is provided, use it as base
			baseName = s.options.Image
			if strings.Contains(baseName, ":") {
				// Remove existing tag
				parts := strings.Split(baseName, ":")
				baseName = parts[0]
			}
		} else {
			// Extract base image from current container image
			if strings.Contains(currentImage, ":") {
				parts := strings.Split(currentImage, ":")
				baseName = parts[0]
			} else {
				baseName = currentImage
			}
		}
		return baseName + ":" + s.options.Tag
	}

	// Direct image specification
	if s.options.Image != "" {
		return s.options.Image
	}

	return currentImage
}
