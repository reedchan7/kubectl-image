package types

import "k8s.io/client-go/kubernetes"

// Options holds the command line options
type Options struct {
	ResourceType  string
	ResourceName  string
	Image         string
	Tag           string
	ContainerName string
	Namespace     string
	TagOnly       bool

	Clientset kubernetes.Interface
}

// ResourceType represents supported Kubernetes resource types
type ResourceType string

const (
	ResourceTypeDeployment ResourceType = "deployment"
	ResourceTypePod        ResourceType = "pod"
)

// ValidResourceTypes returns a list of supported resource types and their aliases
var ValidResourceTypes = map[string]ResourceType{
	"deployment":  ResourceTypeDeployment,
	"deployments": ResourceTypeDeployment,
	"deploy":      ResourceTypeDeployment,
	"pod":         ResourceTypePod,
	"pods":        ResourceTypePod,
	"po":          ResourceTypePod,
}
