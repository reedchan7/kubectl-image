package getter

import (
	"context"
	"fmt"
	"strings"

	"github.com/reedchan7/kubectl-image/src/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ImageGetter handles getting image information from Kubernetes resources
type ImageGetter struct {
	options *types.Options
}

// New creates a new ImageGetter
func New(options *types.Options) *ImageGetter {
	return &ImageGetter{
		options: options,
	}
}

// Get retrieves the image information of the specified resource
func (g *ImageGetter) Get() error {
	resourceType, exists := types.ValidResourceTypes[strings.ToLower(g.options.ResourceType)]
	if !exists {
		return fmt.Errorf("unsupported resource type: %s", g.options.ResourceType)
	}

	var image string
	var err error

	switch resourceType {
	case types.ResourceTypeDeployment:
		image, err = g.getDeploymentImage()
	case types.ResourceTypePod:
		image, err = g.getPodImage()
	default:
		return fmt.Errorf("unsupported resource type: %s", g.options.ResourceType)
	}

	if err != nil {
		return err
	}

	output := image
	if g.options.TagOnly {
		output = extractTag(image)
	}

	// Print just the image string to stdout
	if output != "" {
		fmt.Println(output)
	}
	return nil
}

// extractTag extracts the tag from a full image string.
// It returns an empty string if no tag is found.
func extractTag(image string) string {
	// Find the last colon, which separates the tag
	lastColon := strings.LastIndex(image, ":")
	if lastColon == -1 {
		return "" // No colon, no tag
	}

	// Ensure the part after the colon is a tag, not a port number in the repository
	if strings.Contains(image[lastColon+1:], "/") {
		return "" // Contains '/', so it's part of the path (e.g., my.repo:5000/image)
	}

	return image[lastColon+1:]
}

// getDeploymentImage gets the image of the first container in a deployment
func (g *ImageGetter) getDeploymentImage() (string, error) {
	ctx := context.TODO()
	deploymentsClient := g.options.Clientset.AppsV1().Deployments(g.options.Namespace)

	// Get the deployment
	deployment, err := deploymentsClient.Get(ctx, g.options.ResourceName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get deployment %s: %w", g.options.ResourceName, err)
	}

	if len(deployment.Spec.Template.Spec.Containers) == 0 {
		return "", fmt.Errorf("no containers found in deployment %s", g.options.ResourceName)
	}

	// Return the image of the first container
	return deployment.Spec.Template.Spec.Containers[0].Image, nil
}

// getPodImage gets the image of the first container in a pod
func (g *ImageGetter) getPodImage() (string, error) {
	ctx := context.TODO()
	podsClient := g.options.Clientset.CoreV1().Pods(g.options.Namespace)

	// Get the pod
	pod, err := podsClient.Get(ctx, g.options.ResourceName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get pod %s: %w", g.options.ResourceName, err)
	}

	if len(pod.Spec.Containers) == 0 {
		return "", fmt.Errorf("no containers found in pod %s", g.options.ResourceName)
	}

	// Return the image of the first container
	return pod.Spec.Containers[0].Image, nil
}
