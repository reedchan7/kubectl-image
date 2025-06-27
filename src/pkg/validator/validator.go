package validator

import (
	"fmt"
	"strings"

	"github.com/reedchan7/kubectl-image/src/pkg/types"
)

// Validator handles input validation
type Validator struct {
	options *types.Options
}

// New creates a new Validator
func New(options *types.Options) *Validator {
	return &Validator{
		options: options,
	}
}

// ValidateSet validates the input options for set command
func (v *Validator) ValidateSet() error {
	if err := v.validateResourceType(); err != nil {
		return err
	}

	if err := v.validateResourceName(); err != nil {
		return err
	}

	if err := v.validateImageOptions(); err != nil {
		return err
	}

	return nil
}

// ValidateGet validates the input options for get command
func (v *Validator) ValidateGet() error {
	if err := v.validateResourceType(); err != nil {
		return err
	}

	if err := v.validateResourceName(); err != nil {
		return err
	}

	return nil
}

// validateResourceType validates the resource type
func (v *Validator) validateResourceType() error {
	if v.options.ResourceType == "" {
		return fmt.Errorf("resource type is required")
	}

	if _, exists := types.ValidResourceTypes[strings.ToLower(v.options.ResourceType)]; !exists {
		return fmt.Errorf("unsupported resource type: %s", v.options.ResourceType)
	}

	return nil
}

// validateResourceName validates the resource name
func (v *Validator) validateResourceName() error {
	if v.options.ResourceName == "" {
		return fmt.Errorf("resource name is required")
	}

	return nil
}

// validateImageOptions validates image and tag options
func (v *Validator) validateImageOptions() error {
	if v.options.Image == "" && v.options.Tag == "" {
		return fmt.Errorf("either image name or --tag must be specified")
	}

	return nil
}
