package setter

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

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
		err := s.setDeploymentImage()
		if err != nil {
			return err
		}

		// If wait flag is set, wait for rollout to complete
		if s.options.Wait {
			return s.waitForDeploymentRollout()
		}
		return nil
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
		// Check if the tag actually contains a full image name (common user mistake)
		if strings.Contains(s.options.Tag, "/") || strings.Contains(s.options.Tag, ":") {
			fmt.Printf("tag should only contain the version/tag part (e.g., 'v1.0.1', '7eeb161'), not a full image name. Use the image argument instead for full image names\n")
			os.Exit(1)
		}

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

// waitForDeploymentRollout waits for the deployment rollout to complete
func (s *ImageSetter) waitForDeploymentRollout() error {
	ctx := context.TODO()
	deploymentsClient := s.options.Clientset.AppsV1().Deployments(s.options.Namespace)

	fmt.Printf("Waiting for deployment %s rollout to complete...\n", s.options.ResourceName)

	// Create a ticker for status updates
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Create a timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	startTime := time.Now()
	var deploymentReadyTime time.Time

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for deployment %s rollout to complete", s.options.ResourceName)
		case <-ticker.C:
			// Get the deployment
			deployment, err := deploymentsClient.Get(ctx, s.options.ResourceName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get deployment %s: %v", s.options.ResourceName, err)
			}

			// Check if deployment is ready using the standard Kubernetes deployment conditions
			deploymentReady := false
			for _, condition := range deployment.Status.Conditions {
				if condition.Type == "Progressing" && condition.Status == "True" && condition.Reason == "NewReplicaSetAvailable" {
					deploymentReady = true
					break
				}
			}

			// Also check numeric status fields
			if !deploymentReady {
				deploymentReady = deployment.Status.UpdatedReplicas == deployment.Status.Replicas &&
					deployment.Status.ReadyReplicas == deployment.Status.Replicas &&
					deployment.Status.AvailableReplicas == deployment.Status.Replicas &&
					deployment.Status.ObservedGeneration >= deployment.Generation
			}

			// Get pods for status information
			podList, err := s.options.Clientset.CoreV1().Pods(s.options.Namespace).List(ctx, metav1.ListOptions{
				LabelSelector: metav1.FormatLabelSelector(deployment.Spec.Selector),
			})
			if err != nil {
				return fmt.Errorf("failed to list pods for deployment %s: %v", s.options.ResourceName, err)
			}

			// Count pods by status
			runningPods := 0
			pendingPods := 0
			terminatingPods := 0
			otherPods := 0

			for _, pod := range podList.Items {
				if pod.DeletionTimestamp != nil {
					terminatingPods++
				} else {
					switch pod.Status.Phase {
					case "Running":
						// Check if all containers are ready
						allContainersReady := true
						for _, containerStatus := range pod.Status.ContainerStatuses {
							if !containerStatus.Ready {
								allContainersReady = false
								break
							}
						}
						if allContainersReady {
							runningPods++
						} else {
							pendingPods++
						}
					case "Pending":
						pendingPods++
					default:
						otherPods++
					}
				}
			}

			// Check if rollout is complete
			rolloutComplete := deploymentReady && runningPods == int(deployment.Status.Replicas) && terminatingPods == 0

			// If deployment is ready but there are still terminating pods, check if we should wait longer
			if deploymentReady && runningPods == int(deployment.Status.Replicas) && terminatingPods > 0 {
				// Record when deployment became ready
				if deploymentReadyTime.IsZero() {
					deploymentReadyTime = time.Now()
					duration := deploymentReadyTime.Sub(startTime).Round(time.Millisecond)
					fmt.Printf(" ✅  New pods are ready (took %v), waiting for old pods cleanup...\n", duration)
				}

				// If we've been waiting for cleanup for more than 60 seconds, consider it done
				if time.Since(deploymentReadyTime) > 60*time.Second {
					totalDuration := time.Since(startTime).Round(time.Millisecond)
					cleanupDuration := time.Since(deploymentReadyTime).Round(time.Millisecond)
					fmt.Printf(" ⚠️  Old pods cleanup taking longer than expected, but deployment is ready\n")
					fmt.Printf(" ✅  Deployment %s successfully rolled out (took %v total, cleanup %v ongoing)\n",
						s.options.ResourceName, totalDuration, cleanupDuration)
					return nil
				}
			}

			// If rollout is complete, exit
			if rolloutComplete {
				totalDuration := time.Since(startTime).Round(time.Millisecond)
				if deploymentReadyTime.IsZero() {
					fmt.Printf(" ✅  Deployment %s successfully rolled out (took %v)\n", s.options.ResourceName, totalDuration)
				} else {
					cleanupDuration := time.Since(deploymentReadyTime).Round(time.Millisecond)
					fmt.Printf(" ✅  Deployment %s successfully rolled out (took %v total, cleanup %v)\n",
						s.options.ResourceName, totalDuration, cleanupDuration)
				}
				return nil
			}

			// Print progress
			fmt.Printf(" ⏳  Waiting for rollout to finish: %d/%d pods ready, %d pending, %d terminating\n",
				runningPods, deployment.Status.Replicas, pendingPods, terminatingPods)

			// Print details for problematic pods
			for _, pod := range podList.Items {
				if pod.Status.Phase != "Running" || pod.DeletionTimestamp != nil {
					status := string(pod.Status.Phase)
					if pod.DeletionTimestamp != nil {
						status = "Terminating"
					}

					fmt.Printf(" 🔍  Pod %s status: %s\n", pod.Name, status)

					// Print container statuses for more details
					for _, containerStatus := range pod.Status.ContainerStatuses {
						if !containerStatus.Ready {
							if containerStatus.State.Waiting != nil {
								fmt.Printf("     Container %s is waiting: %s - %s\n",
									containerStatus.Name,
									containerStatus.State.Waiting.Reason,
									containerStatus.State.Waiting.Message)
							} else if containerStatus.State.Terminated != nil {
								fmt.Printf("     Container %s terminated: %s - %s\n",
									containerStatus.Name,
									containerStatus.State.Terminated.Reason,
									containerStatus.State.Terminated.Message)
							}
						}
					}
				}
			}
		}
	}
}
