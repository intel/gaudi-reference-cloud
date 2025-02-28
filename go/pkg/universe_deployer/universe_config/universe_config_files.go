// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package universe_config

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
	"github.com/jedib0t/go-pretty/v6/table"
)

// A type that represents a set of Universe Config files.
type UniverseConfigFiles struct {
	// Map from Universe Config filename to UniverseConfig.
	UniverseConfigMap map[string]*UniverseConfig
}

type GroupedByCommit struct {
	Commits map[string]*CommitInfo
}

func (g GroupedByCommit) Sorted() []string {
	commits := make([]string, 0, len(g.Commits))
	for commit := range g.Commits {
		commits = append(commits, commit)
	}
	sort.Strings(commits)
	return commits
}

type CommitInfo struct {
	Components []*FlatComponentInfo
}

type FlatComponentInfo struct {
	FileName             string
	Environment          string
	Region               string
	AvailabilityZone     string
	Component            string
	UniverseComponentRef *UniverseComponent
}

func ParseUniverseConfigFiles(ctx context.Context, filenames []string) (*UniverseConfigFiles, error) {
	universeConfigMap := map[string]*UniverseConfig{}
	for _, filename := range filenames {
		universeConfig, err := NewUniverseConfigFromFile(ctx, filename)
		if err != nil {
			return nil, err
		}
		universeConfigMap[filename] = universeConfig
	}
	return &UniverseConfigFiles{
		UniverseConfigMap: universeConfigMap,
	}, nil
}

func NewUniverseConfigFilesFromUniverseConfig(c *UniverseConfig) *UniverseConfigFiles {
	return &UniverseConfigFiles{
		UniverseConfigMap: map[string]*UniverseConfig{
			"": c,
		},
	}
}

func (u UniverseConfigFiles) WriteFiles(ctx context.Context) error {
	for filename, universeConfig := range u.UniverseConfigMap {
		if err := universeConfig.WriteFile(ctx, filename); err != nil {
			return err
		}
	}
	return nil
}

func (u UniverseConfigFiles) CheckUnchangedFiles(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("CheckUnchangedFiles")
	_ = log
	tempDir, err := os.MkdirTemp("", "universe_config_")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	for filename, universeConfig := range u.UniverseConfigMap {
		tempFileName := filepath.Join(tempDir, uuid.NewString()+".json")
		if err := universeConfig.WriteFile(ctx, tempFileName); err != nil {
			return err
		}
		equal, err := filesEqual(ctx, tempFileName, filename)
		if err != nil {
			return err
		}
		if !equal {
			return fmt.Errorf("File %s has not been annotated correctly. Run `make run-universe-config-annotate`.", filename)
		}
	}
	return nil
}

// Run a function on each component (all files, environments, regions, and availability zones).
func (u *UniverseConfigFiles) WalkComponents(fn func(FlatComponentInfo) error) error {
	apply := func(components map[string]*UniverseComponent, filename string, environment string, region string, availabilityZone string) error {
		for component, universeComponent := range components {
			flatComponentInfo := FlatComponentInfo{
				FileName:             filename,
				Environment:          environment,
				Region:               region,
				AvailabilityZone:     availabilityZone,
				Component:            component,
				UniverseComponentRef: universeComponent,
			}
			if err := fn(flatComponentInfo); err != nil {
				return err
			}
		}
		return nil
	}

	for filename, universeConfig := range u.UniverseConfigMap {
		for environment, universeEnvironment := range universeConfig.Environments {
			if err := apply(universeEnvironment.Components, filename, environment, "", ""); err != nil {
				return err
			}
			for region, universeRegion := range universeEnvironment.Regions {
				if err := apply(universeRegion.Components, filename, environment, region, ""); err != nil {
					return err
				}
				for availabilityZone, universeAvailabilityZone := range universeRegion.AvailabilityZones {
					if err := apply(universeAvailabilityZone.Components, filename, environment, region, availabilityZone); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// Group by Commit. This does not consider ConfigCommit.
func (u *UniverseConfigFiles) GroupByCommit(ctx context.Context) (*GroupedByCommit, error) {
	groupedByCommit := &GroupedByCommit{
		Commits: map[string]*CommitInfo{},
	}
	fn := func(flatComponentInfo FlatComponentInfo) error {
		commit := flatComponentInfo.UniverseComponentRef.Commit
		commitInfo := groupedByCommit.Commits[commit]
		if commitInfo == nil {
			commitInfo = &CommitInfo{}
			groupedByCommit.Commits[commit] = commitInfo
		}
		commitInfo.Components = append(commitInfo.Components, &flatComponentInfo)
		return nil
	}
	if err := u.WalkComponents(fn); err != nil {
		return nil, err
	}
	return groupedByCommit, nil
}

// Group by commits in Commit and ConfigCommit.
func (u *UniverseConfigFiles) GroupByCommitAndConfigCommit(ctx context.Context) (*GroupedByCommit, error) {
	groupedByCommit := &GroupedByCommit{
		Commits: map[string]*CommitInfo{},
	}
	fn := func(flatComponentInfo FlatComponentInfo) error {
		commits := []string{flatComponentInfo.UniverseComponentRef.Commit, flatComponentInfo.UniverseComponentRef.ConfigCommit}
		for _, commit := range commits {
			if commit != "" {
				commitInfo := groupedByCommit.Commits[commit]
				if commitInfo == nil {
					commitInfo = &CommitInfo{}
					groupedByCommit.Commits[commit] = commitInfo
				}
				commitInfo.Components = append(commitInfo.Components, &flatComponentInfo)
			}
		}
		return nil
	}
	if err := u.WalkComponents(fn); err != nil {
		return nil, err
	}
	return groupedByCommit, nil
}

// Normalize the structure of Universe Configs.
//   - If ConfigCommit is empty, copy Commit to it.
func (u *UniverseConfigFiles) Normalize(ctx context.Context) error {
	fn := func(comp FlatComponentInfo) error {
		if comp.UniverseComponentRef.ConfigCommit == "" {
			comp.UniverseComponentRef.ConfigCommit = comp.UniverseComponentRef.Commit
		}
		return nil
	}
	if err := u.WalkComponents(fn); err != nil {
		return err
	}
	return nil
}

func (u *UniverseConfigFiles) ResolveReferences(ctx context.Context, gitRepositoryDir string, gitRemote string) error {
	log := log.FromContext(ctx).WithName("ResolveReferences")

	groupedByCommit, err := u.GroupByCommitAndConfigCommit(ctx)
	if err != nil {
		return err
	}

	// Identify Git references that need to be resolved.
	refToCommitMap := map[string]string{}
	for commit := range groupedByCommit.Commits {
		if !util.IsGitCommit(commit) && commit != util.HEAD {
			refToCommitMap[commit] = ""
		}
	}

	if len(refToCommitMap) > 0 {
		allRefToCommitMap, err := util.GitLsRemote(ctx, gitRepositoryDir, gitRemote)
		if err != nil {
			return err
		}

		for ref := range refToCommitMap {
			commit, ok := allRefToCommitMap[ref]
			if !ok {
				return fmt.Errorf("unable to resolve reference '%s' using 'git ls-remote %s'", ref, gitRemote)
			}
			refToCommitMap[ref] = commit
		}
		log.Info("Git references resolved", "refToCommitMap", refToCommitMap)

		if _, err := u.ReplaceCommits(ctx, refToCommitMap); err != nil {
			return err
		}
		if _, err := u.ReplaceConfigCommits(ctx, refToCommitMap); err != nil {
			return err
		}
	}

	return nil
}

func (u *UniverseConfigFiles) ReplaceCommits(ctx context.Context, replacementMap map[string]string) (bool, error) {
	replaced := false
	fn := func(flatComponentInfo FlatComponentInfo) error {
		oldCommit := flatComponentInfo.UniverseComponentRef.Commit
		newCommit, ok := replacementMap[oldCommit]
		if ok {
			replaced = true
			flatComponentInfo.UniverseComponentRef.Commit = newCommit
		}
		return nil
	}
	if err := u.WalkComponents(fn); err != nil {
		return false, err
	}
	return replaced, nil
}

func (u *UniverseConfigFiles) ReplaceConfigCommits(ctx context.Context, replacementMap map[string]string) (bool, error) {
	replaced := false
	fn := func(flatComponentInfo FlatComponentInfo) error {
		oldConfigCommit := flatComponentInfo.UniverseComponentRef.ConfigCommit
		newConfigCommit, ok := replacementMap[oldConfigCommit]
		if ok {
			replaced = true
			flatComponentInfo.UniverseComponentRef.ConfigCommit = newConfigCommit
		}
		return nil
	}
	if err := u.WalkComponents(fn); err != nil {
		return false, err
	}
	return replaced, nil
}

func (u *UniverseConfigFiles) HasCommit(ctx context.Context, commit string) (bool, error) {
	return u.ReplaceCommits(ctx, map[string]string{commit: commit})
}

func (u *UniverseConfigFiles) HasConfigCommit(ctx context.Context, commit string) (bool, error) {
	return u.ReplaceConfigCommits(ctx, map[string]string{commit: commit})
}

// Ensure that all non-empty commits are in the form of a Git commit hash (40 hex digits).
func (u *UniverseConfigFiles) ValidateCommits(ctx context.Context) error {
	fn := func(flatComponentInfo FlatComponentInfo) error {
		commit := flatComponentInfo.UniverseComponentRef.Commit
		if commit != "" && !util.IsGitCommit(commit) {
			return fmt.Errorf("commit '%s' in Universe Config is not a valid Git hash", commit)
		}
		configCommit := flatComponentInfo.UniverseComponentRef.ConfigCommit
		if configCommit != "" && !util.IsGitCommit(configCommit) {
			return fmt.Errorf("commit '%s' in Universe Config is not a valid Git hash", configCommit)
		}
		return nil
	}
	return u.WalkComponents(fn)
}

type ComponentCommitsMode int

const (
	// Get distinct (component, commit, configCommit) tuples
	ComponentCommitsModeIncludeAll = iota
	// Get distinct (component, commit) tuples
	ComponentCommitsModeIncludeComponentCommit
	// Get distinct (component, configCommit) tuples
	ComponentCommitsModeIncludeComponentConfigCommit
)

// Returns the distinct sorted list of ComponentCommit.
func (u *UniverseConfigFiles) ComponentCommits(ctx context.Context, mode ComponentCommitsMode) ([]ComponentCommit, error) {
	// Use a map to store unique CommitComponent pairs
	m := make(map[ComponentCommit]struct{})

	// Populate the map.
	fn := func(flatComponentInfo FlatComponentInfo) error {
		item := ComponentCommit{
			Component: flatComponentInfo.Component,
		}
		if mode == ComponentCommitsModeIncludeAll || mode == ComponentCommitsModeIncludeComponentCommit {
			item.Commit = flatComponentInfo.UniverseComponentRef.Commit
		}
		if mode == ComponentCommitsModeIncludeAll || mode == ComponentCommitsModeIncludeComponentConfigCommit {
			item.ConfigCommit = flatComponentInfo.UniverseComponentRef.ConfigCommit
		}
		m[item] = struct{}{}
		return nil
	}
	if err := u.WalkComponents(fn); err != nil {
		return nil, err
	}

	// Create a slice to hold the distinct items.
	var items []ComponentCommit
	for item := range m {
		items = append(items, item)
	}

	// Sort the slice of items.
	sort.Slice(items, func(i, j int) bool {
		if items[i].Component == items[j].Component {
			if items[i].Commit == items[j].Commit {
				return items[i].ConfigCommit < items[j].ConfigCommit
			}
			return items[i].Commit < items[j].Commit
		}
		return items[i].Component < items[j].Component
	})

	return items, nil
}

func (u *UniverseConfigFiles) Merged(ctx context.Context) (*UniverseConfig, error) {
	universeConfigMerged := UniverseConfig{
		Environments: map[string]*UniverseEnvironment{},
	}
	for filename, universeConfig := range u.UniverseConfigMap {
		for environment, universeEnvironment := range universeConfig.Environments {
			_, ok := universeConfigMerged.Environments[environment]
			if ok {
				return nil, fmt.Errorf("environment %s is defined in multiple files: %s", environment, filename)
			}
			universeConfigMerged.Environments[environment] = universeEnvironment
		}
	}
	return &universeConfigMerged, nil
}

func (u *UniverseConfigFiles) annotate(ctx context.Context, gitRepositoryDir string, gitRemote string) error {
	log := log.FromContext(ctx).WithName("annotate")

	for _, universeConfig := range u.UniverseConfigMap {
		universeConfig.Doc2 = "To change deployed components, change the 'commit' value, then run 'make run-universe-config-annotate' to annotate and format this file."
	}

	groupedByCommit, err := u.GroupByCommit(ctx)
	if err != nil {
		return err
	}
	log.V(2).Info("groupedByCommit", "groupedByCommit", groupedByCommit)

	for commit, commitInfo := range groupedByCommit.Commits {
		if commit != util.HEAD {
			universeComponent, err := getGitCommitInfoWithFetch(ctx, commit, gitRepositoryDir, gitRemote)
			if err != nil {
				log.Error(err, "Unable to get git commit information", "commit", commit)
				// Continue with other commits.
			} else {
				for _, components := range commitInfo.Components {
					*components.UniverseComponentRef = universeComponent
				}
			}
		}
	}
	log.V(2).Info("annotated", "groupedByCommit", groupedByCommit)
	return nil
}

func AnnotateFiles(ctx context.Context, filenames []string, gitRepositoryDir string, gitRemote string, checkUnchanged bool) error {
	log := log.FromContext(ctx).WithName("AnnotateFiles")
	universeConfigFiles, err := ParseUniverseConfigFiles(ctx, filenames)
	if err != nil {
		return err
	}
	log.V(2).Info("universeConfigFiles", "universeConfigFiles", universeConfigFiles)
	if err := universeConfigFiles.annotate(ctx, gitRepositoryDir, gitRemote); err != nil {
		return err
	}
	if checkUnchanged {
		if err := universeConfigFiles.CheckUnchangedFiles(ctx); err != nil {
			return err
		}
	} else {
		if err := universeConfigFiles.WriteFiles(ctx); err != nil {
			return err
		}
	}
	return nil
}

type RenderMode int

const (
	RenderModePrettyTable = iota
	RenderModeCsv
)

var RenderModeIds = map[RenderMode][]string{
	RenderModePrettyTable: {"pretty"},
	RenderModeCsv:         {"csv"},
}

type SortBy int

const (
	SortByHierarchy = iota
	SortByAuthorDate
	SortByComponent
)

var SortByIds = map[SortBy][]string{
	SortByHierarchy:  {"hierarchy"},
	SortByAuthorDate: {"authorDate"},
	SortByComponent:  {"component"},
}

func PrintFiles(ctx context.Context, filenames []string, gitRepositoryDir string, gitRemote string, renderMode RenderMode, sortBy SortBy) error {
	universeConfigFiles, err := ParseUniverseConfigFiles(ctx, filenames)
	if err != nil {
		return err
	}
	if err := universeConfigFiles.annotate(ctx, gitRepositoryDir, gitRemote); err != nil {
		return err
	}
	groupedByCommit, err := universeConfigFiles.GroupByCommit(ctx)
	if err != nil {
		return err
	}
	uniqueCommitCount := len(groupedByCommit.Commits)

	componentsMap := map[string]bool{}

	type componentCommit struct {
		component, commit string
	}
	componentCommitsMap := map[componentCommit]bool{}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.Render()
	t.AppendHeader(table.Row{
		"env",
		"region",
		"az",
		"component",
		"authorDate",
		"commit",
		"authorName",
		"authorEmail",
		"committerDate",
		"committerName",
		"committerEmail",
		"subject",
	})

	addComponents := func(components map[string]*UniverseComponent, filename string, environment string, region string, availabilityZone string) {
		for component, universeComponent := range components {
			t.AppendRow(table.Row{
				environment,
				region,
				availabilityZone,
				component,
				universeComponent.AuthorDate,
				universeComponent.Commit,
				universeComponent.AuthorName,
				universeComponent.AuthorEmail,
				universeComponent.CommitterDate,
				universeComponent.CommitterName,
				universeComponent.CommitterEmail,
				universeComponent.Subject,
			})
			componentsMap[component] = true
			componentCommitsMap[componentCommit{component, universeComponent.Commit}] = true
		}
	}

	for filename, universeConfig := range universeConfigFiles.UniverseConfigMap {
		for environment, universeEnvironment := range universeConfig.Environments {
			addComponents(universeEnvironment.Components, filename, environment, "", "")
			for region, universeRegion := range universeEnvironment.Regions {
				addComponents(universeRegion.Components, filename, environment, region, "")
				for availabilityZone, universeAvailabilityZone := range universeRegion.AvailabilityZones {
					addComponents(universeAvailabilityZone.Components, filename, environment, region, availabilityZone)
				}
			}
		}
	}

	uniqueComponentCount := len(componentsMap)
	uniqueComponentCommitCount := len(componentCommitsMap)

	t.AppendFooter(table.Row{"", "", "", "", "Unique Commits", uniqueCommitCount})
	t.AppendFooter(table.Row{"", "", "", "", "Unique Components", uniqueComponentCount})
	t.AppendFooter(table.Row{"", "", "", "", "Unique Component-Commits", uniqueComponentCommitCount})

	t.SetStyle(table.StyleColoredBlackOnBlueWhite)

	if sortBy == SortByHierarchy {
		t.SortBy([]table.SortBy{
			{Name: "env", Mode: table.Asc},
			{Name: "region", Mode: table.Asc},
			{Name: "az", Mode: table.Asc},
			{Name: "component", Mode: table.Asc},
		})
	} else if sortBy == SortByAuthorDate {
		t.SortBy([]table.SortBy{
			{Name: "authorDate", Mode: table.Asc},
			{Name: "commit", Mode: table.Asc},
			{Name: "component", Mode: table.Asc},
			{Name: "env", Mode: table.Asc},
			{Name: "region", Mode: table.Asc},
			{Name: "az", Mode: table.Asc},
		})
	} else if sortBy == SortByComponent {
		t.SortBy([]table.SortBy{
			{Name: "component", Mode: table.Asc},
			{Name: "authorDate", Mode: table.Asc},
			{Name: "commit", Mode: table.Asc},
			{Name: "env", Mode: table.Asc},
			{Name: "region", Mode: table.Asc},
			{Name: "az", Mode: table.Asc},
		})
	}

	if renderMode == RenderModePrettyTable {
		t.SetColumnConfigs([]table.ColumnConfig{
			{Name: "authorName", Hidden: true},
			{Name: "authorEmail", Hidden: true},
			{Name: "committerDate", Hidden: true},
			{Name: "committerName", Hidden: true},
			{Name: "committerEmail", Hidden: true},
			{Name: "subject", Hidden: true},
		})
		t.Render()
	} else if renderMode == RenderModeCsv {
		t.RenderCSV()
	}
	return nil
}

func filesEqual(ctx context.Context, fileName1 string, fileName2 string) (bool, error) {
	bytes1, err := os.ReadFile(fileName1)
	if err != nil {
		return false, err
	}
	bytes2, err := os.ReadFile(fileName2)
	if err != nil {
		return false, err
	}
	return bytes.Equal(bytes1, bytes2), nil
}
