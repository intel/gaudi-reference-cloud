// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import (
	"context"
	"errors"
	"os"
	"regexp"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
)

const (
	AWS_ACCOUNTID_DEFAULT    = "accountid"
	AWS_REGION_DEFAULT       = "us-west-2"
	AWS_ACCOUNTID_EXPRESSION = `(?m)^aws_account_id\s*=\s*(.*)$`
	AWS_CREDENTIALS_DEFAULT  = "/vault/secrets/aws_credentials"
)

type Config struct {
	ListenPort           uint16           `koanf:"listenPort"`
	TLS                  conf.TLSConfig   `koanf:"tls"`
	CloudAccountDatabase manageddb.Config `koanf:"database"`
	UserCredentials      struct {
		EmailExclusionPattern string `koanf:"emailExclusionPattern"`
		AWS                   struct {
			Region          string `koanf:"region"`
			CredentialsFile string `koanf:"credentialsFile"`
			AccountId       string `koanf:"accountId"`
			AccountIdFile   string `koanf:"accountIdFile"`
			UserPool        string `koanf:"userPool"`
			CustomScope     string `koanf:"customScope"`
			CognitoURL      string `koanf:"cognitoUrl"`
		} `koanf:"aws"`
	} `koanf:"usercredentials"`
}

var Cfg *Config

func NewDefaultConfig() *Config {
	if Cfg == nil {
		Cfg = &Config{
			ListenPort: 8443,
		}
	}
	return Cfg
}

func (config *Config) GetListenPort() uint16 {
	return config.ListenPort
}

func (config *Config) SetListenPort(port uint16) {
	config.ListenPort = port
}

func (config *Config) InitTestConfig() {
	cfg := NewDefaultConfig()
	cfg.UserCredentials.EmailExclusionPattern = "!#$%^&*()+=[]{};':\"\\|<>/?"
	Cfg = cfg
}

func (config *Config) Construct() {
	accountId, err := config.GetAWSAccountID()
	if err == nil {
		config.SetAWSAccountId(string(accountId))
	}
}

func (config *Config) GetAWSAccountID() (string, error) {
	logger := log.FromContext(context.Background()).WithName("Config.GetAWSAccountID")
	credentialsContent, err := os.ReadFile(config.GetAWSAccountIdFile())
	if err != nil {
		logger.Error(err, "failed to read credentials file")
		return "", err
	}
	re := regexp.MustCompile(AWS_ACCOUNTID_EXPRESSION)
	matchAccountId := re.FindStringSubmatch(string(credentialsContent))
	if len(matchAccountId) < 2 {
		logger.Error(errors.New("AWS account ID not found"), "credentialsContent", credentialsContent)
		return "", err
	}
	return matchAccountId[1], nil
}

func (config *Config) GetAWSCredentialsFile() string {
	awsCredentials, err := os.ReadFile(config.UserCredentials.AWS.CredentialsFile)
	if err == nil && !strings.EqualFold(string(awsCredentials), "") {
		return config.UserCredentials.AWS.CredentialsFile
	}
	return AWS_CREDENTIALS_DEFAULT
}

func (config *Config) GetAWSCognitoRegion() string {
	return config.UserCredentials.AWS.Region
}

func (config *Config) GetAWSAccountId() string {
	return config.UserCredentials.AWS.AccountId
}

func (config *Config) GetAWSAccountIdFile() string {
	return config.UserCredentials.AWS.AccountIdFile
}

func (config *Config) GetAWSUserPool() string {
	return config.UserCredentials.AWS.UserPool
}
func (config *Config) SetAWSUserPool(userPool string) {
	config.UserCredentials.AWS.UserPool = userPool
}

func (config *Config) GetCustomScope() string {
	return config.UserCredentials.AWS.CustomScope
}

func (config *Config) SetAWSAccountId(accountId string) {
	config.UserCredentials.AWS.AccountId = accountId
}

func (config *Config) SetAWSCognitoUrl(cognitoUrl string) {
	config.UserCredentials.AWS.CognitoURL = cognitoUrl
}

func (config *Config) GetAWSCognitoUrl() string {
	return config.UserCredentials.AWS.CognitoURL
}

func (config *Config) GetEmailExclusionPattern() string {
	return config.UserCredentials.EmailExclusionPattern
}
