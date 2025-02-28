// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package universe_config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/env_config/types"
)

// A type to parse Universe Config files.
type UniverseConfig struct {
	Doc          string                          `json:"_doc,omitempty"`
	Doc2         string                          `json:"_doc2,omitempty"`
	Environments map[string]*UniverseEnvironment `json:"environments,omitempty"`
}

type UniverseEnvironment struct {
	Components         map[string]*UniverseComponent `json:"components,omitempty"`
	ForceAllComponents bool                          `json:"forceAllComponents,omitempty"`
	Regions            map[string]*UniverseRegion    `json:"regions,omitempty"`
}

type UniverseRegion struct {
	Components        map[string]*UniverseComponent        `json:"components,omitempty"`
	AvailabilityZones map[string]*UniverseAvailabilityZone `json:"availabilityZones,omitempty"`
}

type UniverseAvailabilityZone struct {
	Components map[string]*UniverseComponent `json:"components,omitempty"`
}

type UniverseComponent struct {
	Commit         string     `json:"commit,omitempty"`
	AuthorDate     *time.Time `json:"authorDate,omitempty"`
	AuthorEmail    string     `json:"authorEmail,omitempty"`
	AuthorName     string     `json:"authorName,omitempty"`
	CommitterDate  *time.Time `json:"committerDate,omitempty"`
	CommitterEmail string     `json:"committerEmail,omitempty"`
	CommitterName  string     `json:"committerName,omitempty"`
	Subject        string     `json:"subject,omitempty"`
	// If provided, environment-specific configuration files will come from this commit.
	// Otherwise, environment-specific configuration files will come from Commit.
	ConfigCommit string `json:"configCommit,omitempty"`
}

type ComponentCommit struct {
	Component    string
	Commit       string
	ConfigCommit string
}

// Load a dedicated Universe Config file (e.g. prod.json).
func NewUniverseConfigFromFile(ctx context.Context, filename string) (*UniverseConfig, error) {
	universeConfigBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", filename, err)
	}
	var universeConfig UniverseConfig
	if err := json.Unmarshal(universeConfigBytes, &universeConfig); err != nil {
		return nil, fmt.Errorf("unmarshaling file %s: %v", filename, err)
	}
	return &universeConfig, nil
}

// Generate a Universe Config from the Components sections in EnvConfig.
// This can be used for development environments.
// This should not be used for production environments because the EnvConfig
// can be affected by changes to deployment/helmfile/defaults.yaml.gotmpl.
func NewUniverseConfigFromEnvConfig(ctx context.Context, envConfig types.EnvConfig) (*UniverseConfig, error) {
	universeConfig := &UniverseConfig{
		Environments: map[string]*UniverseEnvironment{},
	}
	universeEnvironment := &UniverseEnvironment{
		Components: newComponentsFromEnvConfig(ctx, envConfig.Values.Global.Components),
		Regions:    map[string]*UniverseRegion{},
	}
	for region, envConfigRegion := range envConfig.Values.Regions {
		universeRegion := &UniverseRegion{
			Components:        newComponentsFromEnvConfig(ctx, envConfigRegion.Components),
			AvailabilityZones: map[string]*UniverseAvailabilityZone{},
		}
		for availabilityZone, envConfigAvailabilityZone := range envConfigRegion.AvailabilityZones {
			universeAvailabilityZone := &UniverseAvailabilityZone{
				Components: newComponentsFromEnvConfig(ctx, envConfigAvailabilityZone.Components),
			}
			universeRegion.AvailabilityZones[availabilityZone] = universeAvailabilityZone
		}
		universeEnvironment.Regions[region] = universeRegion
	}
	universeConfig.Environments[envConfig.Environment.Name] = universeEnvironment
	return universeConfig, nil
}

func newComponentsFromEnvConfig(ctx context.Context, components map[string]*types.Component) map[string]*UniverseComponent {
	universeComponents := map[string]*UniverseComponent{}
	for component, envConfigComponent := range components {
		if envConfigComponent != nil && envConfigComponent.Enabled && envConfigComponent.Commit != "" {
			universeComponents[component] = &UniverseComponent{
				Commit:       envConfigComponent.Commit,
				ConfigCommit: envConfigComponent.ConfigCommit,
			}
		}
	}
	return universeComponents
}

func (c UniverseConfig) WriteFile(ctx context.Context, filename string) error {
	universeConfigBytes, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return nil
	}
	return os.WriteFile(filename, universeConfigBytes, 0640)
}

func (in *UniverseConfig) DeepCopy() *UniverseConfig {
	b, err := json.Marshal(in)
	if err != nil {
		return nil
	}
	var out UniverseConfig
	if err := json.Unmarshal(b, &out); err != nil {
		return nil
	}
	return &out
}

// Trim a Universe Config so that it only includes the specified commit.
// Containing objects with no commits will be removed from the returned value.
// This returns a new object with references to the current object.
func (c UniverseConfig) Trimmed(ctx context.Context, commit string) *UniverseConfig {
	pred := func(component string, universeComponent UniverseComponent) bool {
		return universeComponent.Commit == commit
	}
	return c.Filtered(ctx, pred)
}

// Filter a Universe Config by an arbitrary predicate.
// Containing objects that do not satisfy the predicate will be removed from the returned value.
// This returns a new object with references to the current object.
func (c UniverseConfig) Filtered(ctx context.Context, pred func(string, UniverseComponent) bool) *UniverseConfig {
	trimmed := &UniverseConfig{
		Environments: map[string]*UniverseEnvironment{},
	}
	for idcEnv, universeEnvironment := range c.Environments {
		trimmedEnvironment := &UniverseEnvironment{
			Components: map[string]*UniverseComponent{},
			Regions:    map[string]*UniverseRegion{},
		}
		for component, universeComponent := range universeEnvironment.Components {
			if pred(component, *universeComponent) {
				trimmedEnvironment.Components[component] = universeComponent
			}
		}
		for region, universeRegion := range universeEnvironment.Regions {
			trimmedRegion := &UniverseRegion{
				Components:        map[string]*UniverseComponent{},
				AvailabilityZones: map[string]*UniverseAvailabilityZone{},
			}
			for component, universeComponent := range universeRegion.Components {
				if pred(component, *universeComponent) {
					trimmedRegion.Components[component] = universeComponent
				}
			}
			for availabilityZone, universeAvailabilityZone := range universeRegion.AvailabilityZones {
				trimmedAvailabilityZone := &UniverseAvailabilityZone{
					Components: map[string]*UniverseComponent{},
				}
				for component, universeComponent := range universeAvailabilityZone.Components {
					if pred(component, *universeComponent) {
						trimmedAvailabilityZone.Components[component] = universeComponent
					}
				}
				if len(trimmedAvailabilityZone.Components) > 0 {
					trimmedRegion.AvailabilityZones[availabilityZone] = trimmedAvailabilityZone
				}
			}
			if len(trimmedRegion.Components) > 0 || len(trimmedRegion.AvailabilityZones) > 0 {
				trimmedEnvironment.Regions[region] = trimmedRegion
			}
		}
		if len(trimmedEnvironment.Components) > 0 || len(trimmedEnvironment.Regions) > 0 {
			trimmed.Environments[idcEnv] = trimmedEnvironment
		}
	}
	return trimmed
}

func (c *UniverseConfig) Normalize(ctx context.Context) error {
	return NewUniverseConfigFilesFromUniverseConfig(c).Normalize(ctx)
}

func (c *UniverseConfig) ResolveReferences(ctx context.Context, gitRepositoryDir string, gitRemote string) error {
	return NewUniverseConfigFilesFromUniverseConfig(c).ResolveReferences(ctx, gitRepositoryDir, gitRemote)
}

func (c *UniverseConfig) ReplaceCommits(ctx context.Context, replacementMap map[string]string) (bool, error) {
	return NewUniverseConfigFilesFromUniverseConfig(c).ReplaceCommits(ctx, replacementMap)
}

func (c *UniverseConfig) ReplaceConfigCommits(ctx context.Context, replacementMap map[string]string) (bool, error) {
	return NewUniverseConfigFilesFromUniverseConfig(c).ReplaceConfigCommits(ctx, replacementMap)
}

func (c *UniverseConfig) HasCommit(ctx context.Context, commit string) (bool, error) {
	return NewUniverseConfigFilesFromUniverseConfig(c).HasCommit(ctx, commit)
}

func (c *UniverseConfig) HasConfigCommit(ctx context.Context, commit string) (bool, error) {
	return NewUniverseConfigFilesFromUniverseConfig(c).HasConfigCommit(ctx, commit)
}

func (c *UniverseConfig) ValidateCommits(ctx context.Context) error {
	return NewUniverseConfigFilesFromUniverseConfig(c).ValidateCommits(ctx)
}

func (c *UniverseConfig) GroupByCommit(ctx context.Context) (*GroupedByCommit, error) {
	return NewUniverseConfigFilesFromUniverseConfig(c).GroupByCommit(ctx)
}

func (c *UniverseConfig) ComponentCommits(ctx context.Context, mode ComponentCommitsMode) ([]ComponentCommit, error) {
	return NewUniverseConfigFilesFromUniverseConfig(c).ComponentCommits(ctx, mode)
}
