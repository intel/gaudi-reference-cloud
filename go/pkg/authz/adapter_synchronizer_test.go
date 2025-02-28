// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package authz

import (
	"testing"

	config "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authz/config"
)

func TestNewSyncedEnforcerMissingFiles(t *testing.T) {
	_ = test.ClientConn()

	// Given a configuration with no files
	cfg := &config.Config{}
	cfg.Features.PoliciesStartupSync = true
	cfg.Features.Watcher = false

	// When creating a new synced enforcer with no files
	_, err := NewSyncedEnforcer(test.mdb, cfg)

	// Expect errors
	if err == nil {
		t.Fatalf("an error was expected. not possible to start without the model.conf file")
	}

	// Given that model file exists
	cfg.ModelFilePath = "test/data/model.conf"

	// When creating a new synced enforcer
	_, err = NewSyncedEnforcer(test.mdb, cfg)

	// Expect errors
	if err == nil {
		t.Fatalf("an error was expected. not possible to start without the policies file")
	}

	// Given that model and policies files exist
	cfg.ModelFilePath = "test/data/model.conf"
	cfg.PolicyFilePath = "test/data/policy.csv"

	// When creating a new synced enforcer
	_, err = NewSyncedEnforcer(test.mdb, cfg)

	// Expect errors
	if err == nil {
		t.Fatalf("an error was expected. not possible to start without the groups file")
	}

	// Given that model, policies, and groups files exist
	cfg.ModelFilePath = "test/data/model.conf"
	cfg.PolicyFilePath = "test/data/policy.csv"
	cfg.GroupFilePath = "test/data/groups.csv"

	// When creating a new synced enforcer without resources file
	_, err = NewSyncedEnforcer(test.mdb, cfg)

	// Expect no errors
	if err != nil {
		t.Fatalf("we should not expect errors. resources file not required to create a synced enforcer")
	}

}

func TestNewSyncedEnforcer(t *testing.T) {
	_ = test.ClientConn()

	// Given the correct configuration
	cfg := config.NewDefaultConfig()
	cfg.Features.PoliciesStartupSync = true
	cfg.Features.Watcher = false
	cfg.ModelFilePath = "test/data/model.conf"
	cfg.PolicyFilePath = "test/data/policy.csv"
	cfg.GroupFilePath = "test/data/groups.csv"
	cfg.ResourcesFilePath = "test/data/resources.yaml"

	// When creating a new synced enforcer
	_, err := NewSyncedEnforcer(test.mdb, cfg)

	// Expect no errors
	if err != nil {
		t.Fatalf("we should not expect errors")
	}
}
