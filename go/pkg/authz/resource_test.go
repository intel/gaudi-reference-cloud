// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package authz

import (
	"os"
	"reflect"
	"testing"
)

// helper function to create a temporary YAML file
func createTempResourcesFile(contents string) (string, error) {
	tmpfile, err := os.CreateTemp("", "resources")
	if err != nil {
		return "", err
	}
	if _, err := tmpfile.Write([]byte(contents)); err != nil {
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		return "", err
	}
	return tmpfile.Name(), nil
}

func TestReadResourcesFromFile(t *testing.T) {
	fileContent := `
- type: "service1"
  description: "resource description"
  allowedactions:
    - name: "read"
      type: "resource"
      description: "action description"
    - name: "write"
      type: "collection"
- type: "service2"
  allowedactions:
    - name: "execute"
      type: "resource"
`
	filePath, err := createTempResourcesFile(fileContent)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(filePath) // clean up

	resources, err := readResourcesFromFile(filePath)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Define what we expect to see
	expected := map[string]Resource{
		"service1": {
			Type:           "service1",
			Description:    "resource description",
			AllowedActions: []Action{{Name: "read", Type: "resource", Description: "action description"}, {Name: "write", Type: "collection"}},
		},
		"service2": {
			Type:           "service2",
			AllowedActions: []Action{{Name: "execute", Type: "resource"}},
		},
	}

	if !reflect.DeepEqual(resources, expected) {
		t.Errorf("Expected resources to be %v, got %v", expected, resources)
	}
}

func TestNewResourceRepository(t *testing.T) {
	fileContent := `
- type: "service1"
  allowedactions:
    - name: "read"
      type: "resource"
    - name: "write"
      type: "collection"
`
	filePath, err := createTempResourcesFile(fileContent)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(filePath) // clean up

	repo, _ := NewResourceRepository(filePath)
	if repo == nil {
		t.Errorf("Expected non-nil repository")
	}

	if len(repo.resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(repo.resources))
	}
}

func TestNewResourceRepositoryDuplicatedResource(t *testing.T) {
	fileContent := `
- type: "service1"
  allowedactions:
    - name: "read"
      type: "resource"
    - name: "write"
      type: "collection"
- type: "service1"
  allowedactions:
    - name: "read"
      type: "resource"
    - name: "write"
      type: "collection"
`
	filePath, err := createTempResourcesFile(fileContent)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(filePath) // clean up

	repo, err := NewResourceRepository(filePath)
	if repo != nil {
		t.Errorf("Expected nil repository")
	}

	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	expected := "resource service1 already defined in the resource file"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestResourceRepository_Get(t *testing.T) {
	fileContent := `
- type: "service1"
  allowedactions:
    - name: "read"
      type: "resource"
    - name: "write"
      type: "collection"
`
	filePath, err := createTempResourcesFile(fileContent)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(filePath) // clean up

	repo, _ := NewResourceRepository(filePath)

	resource := repo.Get("service1")
	if resource == nil {
		t.Errorf("Expected non-nil resource")
	}
	if resource.Type != "service1" {
		t.Errorf("Expected resource type to be 'service1', got '%s'", resource.Type)
	}

	resource = repo.Get("non-existent")
	if resource != nil {
		t.Errorf("Expected nil resource for non-existent type")
	}
}
