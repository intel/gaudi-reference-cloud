// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package authz

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"os"

	psqlwatcher "github.com/IguteChung/casbin-psql-watcher"
	"github.com/casbin/casbin/v2"
	config "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/authz/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

func syncFromCSV(e *casbin.Enforcer, csvFilePath string, fileType string) error {
	if e == nil {
		return errors.New("enforcer is required")
	}

	_, logger, span := obs.LogAndSpanFromContext(context.Background()).WithName("AdapterSynchronizer.SyncFromCSV").WithValues("fileType", fileType).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	// Checks if the file exists
	_, err := os.Stat(csvFilePath)
	if err != nil {
		logger.Error(err, "there is an error with the file", "csvFilePath", csvFilePath)
		return err
	}

	// Open the CSV file
	file, err := os.Open(csvFilePath)
	if err != nil {
		logger.Error(err, "error while opening the file", "csvFilePath", csvFilePath)
		return err
	}
	defer file.Close()

	// Read the CSV file
	csvReader := csv.NewReader(file)
	csvReader.Comma = ';'
	for {
		record, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				logger.Error(err, "error reading the csv file")
				os.Exit(1)
			}
		}

		if fileType == "policies" {
			// Add policy
			if !e.HasPolicy(record) {
				_, err := e.AddPolicy(record)
				if err != nil {
					logger.Error(err, "failed to add policy")
					return err
				}
			}
		} else {
			// Add group mapping
			if !e.HasGroupingPolicy(record) {
				_, err := e.AddGroupingPolicy(record)
				if err != nil {
					logger.Error(err, "failed to add grouping policy")
					return err
				}
			}
		}
	}

	return nil

}

func NewSyncedEnforcer(database *manageddb.ManagedDb, cfg *config.Config) (*casbin.Enforcer, error) {
	logger.Info("starting casbinEngine ...")

	if database == nil {
		err := errors.New("database is required")
		logger.Error(err, "database is required")
		return nil, err
	}
	if cfg == nil {
		err := errors.New("config is required")
		logger.Error(err, "config is required")
		return nil, err
	}

	adapter, err := NewAdapter(database.DatabaseURL)
	if err != nil {
		logger.Error(err, "failed to create new adapter", "databaseURL", database.DatabaseURL)
		return nil, err
	}

	// Initialize the Enforcer with the model and the Gorm adapter
	enforcer, err := casbin.NewEnforcer(cfg.ModelFilePath, adapter)
	if err != nil {
		logger.Error(err, "failed to create new enforcer", "modelPath", cfg.ModelFilePath)
		return nil, err
	}

	if cfg.Features.Watcher {
		// watcher to keep all casbin instances synchronized
		w, err := psqlwatcher.NewWatcherWithConnString(context.Background(), database.DatabaseURL.String(),
			psqlwatcher.Option{NotifySelf: false, Verbose: false})
		if err != nil {
			logger.Error(err, "failed to set NewWatcherWithConnString")
			return nil, err
		}

		// set the watcher for enforcer.
		err = enforcer.SetWatcher(w)
		if err != nil {
			logger.Error(err, "failed to set watcher to enforcer")
			return nil, err
		}

		// set the default callback to handle policy changes.
		err = w.SetUpdateCallback(func(string) {
			err := enforcer.LoadPolicy()
			if err != nil {
				logger.Error(err, "failed load policy watcher callback")
			}
		})
		if err != nil {
			logger.Error(err, "failed to SetUpdateCallback on watcher")
			return nil, err
		}
	}

	// Load the initial policies from the database
	err = enforcer.LoadPolicy()
	if err != nil {
		logger.Error(err, "failed to load policy")
		return nil, err
	}

	if cfg.Features.PoliciesStartupSync {
		// Synchronize the policies from the CSV file to the database
		err = syncFromCSV(enforcer, cfg.PolicyFilePath, "policies")
		if err != nil {
			logger.Error(err, "failed to sync policy from csv", "policiesCSVFilePath", cfg.PolicyFilePath)
			return nil, err
		} else {
			logger.Info("casbinEngine policies synced successfully")
		}

		// Synchronize the groups from the CSV file to the database
		err = syncFromCSV(enforcer, cfg.GroupFilePath, "groups")
		if err != nil {
			logger.Error(err, "failed to sync groups from csv", "groupsCSVFilePath", cfg.GroupFilePath)
			return nil, err

		} else {
			logger.Info("casbinEngine groups synced successfully")
		}
	}

	return enforcer, nil
}
