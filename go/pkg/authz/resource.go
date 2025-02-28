// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package authz

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v2"
)

type Action struct {
	Name        string `json:"name" validate:"required"`
	Type        string `json:"type" validate:"required,oneof=collection resource"`
	Description string `json:"description"`
}

type Resource struct {
	Type           string   `json:"type" validate:"required"`
	Description    string   `json:"description"`
	AllowedActions []Action `json:"allowedActions" validate:"dive"`
}

type ResourceRepository struct {
	resources map[string]Resource
}

type ActionType string

const (
	ACTION_COLLECTION_TYPE ActionType = "collection"
	ACTION_RESOURCE_TYPE   ActionType = "resource"
)

func readResourcesFromFile(filePath string) (map[string]Resource, error) {
	// Checks if the file exists
	_, err := os.Stat(filePath)
	if err != nil {
		logger.Error(err, "resource yaml file doesn't exist", "filePath", filePath)
		return nil, err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		logger.Error(err, "failed to read file", "filePath", filePath)
		return nil, err
	}

	var resources []Resource
	err = yaml.Unmarshal(data, &resources)
	if err != nil {
		logger.Error(err, "failed to unmarshal yaml resources file", "filePath", filePath)
		return nil, err
	}

	resourceMap := make(map[string]Resource)
	for _, resource := range resources {
		validate := validator.New()
		err = validate.Struct(resource)
		if err != nil {
			logger.Error(err, "failed to validate resource", "resource", resource)
			return nil, err
		}
		if _, found := resourceMap[resource.Type]; found {
			errMsg := fmt.Sprintf("resource %s already defined in the resource file", resource.Type)
			err := fmt.Errorf(errMsg)
			logger.Error(err, errMsg)
			return nil, err
		}
		resourceMap[resource.Type] = resource
	}

	return resourceMap, nil
}

func NewResourceRepository(filePath string) (*ResourceRepository, error) {
	logger.Info("initializing resources repository from", "filepath", filePath)
	resources, err := readResourcesFromFile(filePath)
	if err != nil {
		logger.Error(err, "failed to load resources from file")
		return nil, err
	}

	return &ResourceRepository{resources: resources}, nil
}

func (r *ResourceRepository) Get(resourceType string) *Resource {
	if r == nil {
		return nil
	}
	if resource, found := r.resources[resourceType]; found {
		return &resource
	}
	return nil
}

func (rp *ResourceRepository) validateResource(resourceType string, actions ...string) error {
	if rp == nil {
		return fmt.Errorf("resource repository is required")
	}
	// check if resource exists in the definitions
	resource := rp.Get(resourceType)
	if resource == nil {
		errMsg := fmt.Sprintf("resource %s not found", resourceType)
		err := fmt.Errorf(errMsg)
		logger.Error(err, errMsg)
		return err
	}
	// validates if the actions are allowed for the resource
	for _, action := range actions {
		if !Contains(resource.GetActionNames(), action) {
			errMsg := fmt.Sprintf("action %s not allowed for resource %s, the allowed actions are %s", action, resourceType, strings.Join(resource.GetActionNames(), ", "))
			err := fmt.Errorf(errMsg)
			logger.Error(err, "validation error", "resourceType", resourceType, "action", action, "allowedActions", resource.AllowedActions)
			return err
		}
	}
	return nil
}

func (r *Resource) GetAction(actionName string) *Action {
	if r == nil {
		return nil
	}
	for _, action := range r.AllowedActions {
		if action.Name == actionName {
			return &action
		}
	}
	return nil
}

func (r *Resource) GetActionNames() []string {
	if r == nil {
		return nil
	}
	var actionNames []string
	for _, action := range r.AllowedActions {
		actionNames = append(actionNames, action.Name)
	}
	return actionNames
}
