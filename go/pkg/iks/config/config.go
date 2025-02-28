// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import (
	"encoding/json"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"golang.org/x/exp/maps"
	"os"
)

// Application configuration
type Config struct {
	ListenPort               uint16           `koanf:"listenPort"`
	Database                 manageddb.Config `koanf:"database"`
	UsernameRwFile           string           `koanf:"usernameRwFile"`
	PasswordRwFile           string           `koanf:"passwordRwFile"`
	EncryptionKeys           string           `koanf:"encryptionKeys"`
	AdminKey                 string           `koanf:"adminKey"`
	ComputeServerAddr        string           `koanf:"computeServerAddr"`
	ProductcatalogServerAddr string           `koanf:"productcatalogServerAddr"`
	DbSeed                   DbSeed           `koanf:"dbSeed"`
}

type DbSeed struct {
	Enabled bool `koanf:"enabled"`
	// vault secret data source
	DataFile string `koanf:"dataFile"`
	// clear text helm values data source
	Data map[string]interface{} `koanf:"data"`
}

// TemplateData returns prepared data
// for processing the SQLs templates
func (s *DbSeed) TemplateData() (map[string]interface{}, error) {
	// make sure initiate data map in case it wasn't
	// provided in the configuration
	if s.Data == nil {
		s.Data = map[string]interface{}{}
	}
	// adding enabled flag into the
	// data object for further processing
	// by the SQL templates
	s.Data["DbSeedEnabled"] = s.Enabled
	// if db seed is not enabled, no need to proceed
	if !s.Enabled {
		return s.Data, nil
	}
	// Db seed has two data sources
	// 1. clear text data came from values file
	// 2. sensitive data came from vault secret
	// here we need to read all the data and merge
	// it into singe data object for
	// further templates processing
	secretData, err := s.secretData()
	if err != nil {
		return nil, err
	}
	maps.Copy(s.Data, secretData)
	return s.Data, nil
}

// secretData read the template data from vault secret mounted as JSON file
func (s *DbSeed) secretData() (map[string]interface{}, error) {
	fileContent, err := os.ReadFile(s.DataFile)
	if err != nil {
		return nil, err
	}
	secretData := map[string]interface{}{}
	if err := json.Unmarshal(fileContent, &secretData); err != nil {
		return nil, err
	}
	return secretData, nil
}
