// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
)

const (
	AWS_QUEUE_ACCOUNTID_DEFAULT = "accountid"
	AWS_QUEUE_REGION_DEFAULT    = "us-west-2"
	TAX_GROUP_ID                = "SERVICES"
	TAX_GROUP                   = 7
	FileStorageServiceType      = "FileStorageAsAService"
	ObjectStorageServiceType    = "ObjectStorageAsAService"
)

type Config struct {
	ListenPort                                 uint16           `koanf:"listenPort"`
	Database                                   manageddb.Config `koanf:"database"`
	SyncInterval                               int              `koanf:"syncInterval"`
	UsageInterval                              int              `koanf:"usageInterval"`
	ClientIdPrefix                             string           `koanf:"clientIdPrefix"`
	ClientAcctGroupId                          string           `koanf:"clientAcctGroupId"`
	EnterpriseAcctLinkSchedulerInterval        int              `koanf:"enterpriseAcctLinkSchedulerInterval"`
	ReportUsageSchedulerInterval               int              `koanf:"reportUsageSchedulerInterval"`
	ReportProductUsageSchedulerInterval        int              `koanf:"reportProductUsageSchedulerInterval"`
	PaidServicesDeactivationControllerInterval int              `koanf:"paidServicesDeactivationControllerInterval"`
	InstanceSearchWindow                       int              `koanf:"instanceSearchWindow"`
	PremiumDefaultCreditAmount                 float64          `koanf:"premiumDefaultCreditAmount"`
	PremiumDefaultCreditExpirationDays         int              `koanf:"premiumDefaultCreditExpirationDays"`
	EntDefaultCreditAmount                     float64          `koanf:"entDefaultCreditAmount"`
	EntDefaultCreditExpirationDays             int              `koanf:"entDefaultCreditExpirationDays"`
	AcctPlanActiveInterval                     int              `koanf:"acctPlanActiveInterval"`
	TestProfile                                bool
	AuthorizationEnabled                       bool     `koanf:"authorizationEnabled"`
	StorageInstanceTypes                       []string `koanf:"storageInstanceTypes"`
	DB                                         struct {
		User           string `koanf:"user"`
		Password       string `koanf:"password"`
		Host           string `koanf:"host"`
		Port           int    `koanf:"port"`
		Name           string `koanf:"name"`
		connectTimeout int    `koanf:"connectTimeout"`
	} `koanf:"db"`
	AriaSystem struct {
		ClientNoFile           string `koanf:"clientNoFile"`
		ClientNo               int64  `koanf:"clientNo"`
		AuthKeyFile            string `koanf:"authKeyFile"`
		AuthKey                string `koanf:"authKey"`
		ApiCrtFile             string `koanf:"apiCrtFile"`
		ApiKeyFile             string `koanf:"apiKeyFile"`
		InsecureSsl            bool   `koanf:"insecureSsl"`
		ReleaseVersion         string `koanf:"releaseVersion"`
		CoreApiSuffix          string `koanf:"coreApiSuffix"`
		DirectPostUrl          string `koanf:"directPostUrl"`
		FunctionMode           string `koanf:"functionMode"`
		TaxGroup               int    `koanf:"taxGroup"`
		ClientTaxGroupId       string `koanf:"clientTaxGroupId"`
		StorageUsageUnitType   string `koanf:"storageUsageUnitType"`
		TokenUsageUnitType     string `koanf:"tokenUsageUnitType"`
		InferenceUsageUnitType string `koanf:"inferenceUsageUnitType"`
		Server                 struct {
			CoreApiUrl   string `koanf:"coreApiUrl"`
			AdminApiUrl  string `koanf:"adminApiUrl"`
			ObjectApiUrl string `koanf:"objectApiUrl"`
		} `koanf:"server"`
	} `koanf:"ariasystem"`
	AWS struct {
		SQS struct {
			Region          string `koanf:"region"`
			AccountId       string `koanf:"accountId"`
			AccountIdFile   string `koanf:"accountIdFile"`
			QueueUrl        string `koanf:"queueUrl"`
			CredentialsFile string `koanf:"credentialsFile"`
		} `koanf:"sqs"`
	} `koanf:"aws"`
	BillingDriverAria struct {
		Features struct {
			DeactivationScheduler       bool `koanf:"deactivationScheduler"`
			ReportUsageScheduler        bool `koanf:"reportUsageScheduler"`
			ReportProductUsageScheduler bool `koanf:"reportProductUsageScheduler"`
			SyncStoragePlan             bool `koanf:"syncStoragePlan"`
			SyncInferencePlan           bool `koanf:"syncInferencePlan"`
			SyncTokenPlan               bool `koanf:"syncTokenPlan"`
		} `koanf:"features"`
		ProductCatalog struct {
			StorageUsageUnitType   string `koanf:"storageUsageUnitType"`
			TokenUsageUnitType     string `koanf:"tokenUsageUnitType"`
			InferenceUsageUnitType string `koanf:"inferenceUsageUnitType"`
			ApiVersion             string `koanf:"apiVersion"`
		} `koanf:"productcatalog"`
	} `koanf:"billingDriverAria"`
}

var Cfg *Config

func NewDefaultConfig() *Config {
	if Cfg == nil {
		Cfg = &Config{
			ListenPort:                                 8443,
			SyncInterval:                               14400,
			UsageInterval:                              14400,
			ClientAcctGroupId:                          "Chase",
			AcctPlanActiveInterval:                     14400,
			EnterpriseAcctLinkSchedulerInterval:        14400,
			ReportUsageSchedulerInterval:               14400,
			ReportProductUsageSchedulerInterval:        14400,
			PaidServicesDeactivationControllerInterval: 14400,
			PremiumDefaultCreditAmount:                 2500,
			PremiumDefaultCreditExpirationDays:         35,
			EntDefaultCreditAmount:                     2500,
			EntDefaultCreditExpirationDays:             35,
			InstanceSearchWindow:                       2,
		}
		Cfg.AriaSystem.ReleaseVersion = "47"
		Cfg.AriaSystem.CoreApiSuffix = "/v1/core"
		Cfg.AriaSystem.Server.CoreApiUrl = "https://api.future.stage.ariasystems.net"
		Cfg.AriaSystem.Server.AdminApiUrl = "https://api.future.stage.ariasystems.net/AdminTools.php/Dispatcher"
		Cfg.AriaSystem.DirectPostUrl = "https://api.future.stage.ariasystems.net/api/direct_post_eom.php"
		Cfg.AriaSystem.FunctionMode = "direct_post_reg"
		Cfg.AriaSystem.InsecureSsl = false
		Cfg.AriaSystem.TaxGroup = TAX_GROUP
		Cfg.AriaSystem.ClientTaxGroupId = TAX_GROUP_ID
		Cfg.StorageInstanceTypes = []string{"storage-file", "storage-object"}
		Cfg.BillingDriverAria.ProductCatalog.StorageUsageUnitType = "terabyte hour"
		Cfg.BillingDriverAria.ProductCatalog.InferenceUsageUnitType = "inference"
		Cfg.BillingDriverAria.ProductCatalog.TokenUsageUnitType = "million tokens"
	}
	return Cfg
}

func InitTestConfig() error {
	cfg := NewDefaultConfig()
	cfg.AWS.SQS.QueueUrl, _ = os.LookupEnv("AWS_QUEUE_URL")
	cfg.AWS.SQS.Region, _ = os.LookupEnv("AWS_QUEUE_REGION")
	cfg.AWS.SQS.AccountId, _ = os.LookupEnv("AWS_QUEUE_ACCOUNTID")
	cfg.AriaSystem.AuthKey, _ = os.LookupEnv("IDC_ARIA_AUTH_KEY")
	if cfg.AriaSystem.AuthKey == "" {
		return fmt.Errorf("auth key has not been set")
	}
	clientNumber, _foundClientNo := os.LookupEnv("IDC_ARIA_CLIENT_NO")
	if !_foundClientNo {
		return fmt.Errorf("client number has not been set")
	}
	var err error
	cfg.AriaSystem.ClientNo, err = strconv.ParseInt(clientNumber, 10, 64)
	if err != nil {
		return fmt.Errorf("error parsing int value %v", err)
	}
	cfg.ClientIdPrefix = GetTestPrefix()
	cfg.TestProfile = true
	cfg.AriaSystem.ApiCrtFile, _ = os.LookupEnv("IDC_ARIA_API_CERT_FILE")
	cfg.AriaSystem.ApiKeyFile, _ = os.LookupEnv("IDC_ARIA_API_KEY_FILE")
	Cfg = cfg
	Cfg.AriaSystem.InsecureSsl = false
	Cfg.AriaSystem.TaxGroup = TAX_GROUP
	Cfg.AriaSystem.ClientTaxGroupId = TAX_GROUP_ID
	Cfg.AriaSystem.StorageUsageUnitType = "count"
	Cfg.BillingDriverAria.ProductCatalog.StorageUsageUnitType = "terabyte hour"
	Cfg.StorageInstanceTypes = []string{"storage-file", "storage-object"}
	return nil
}

func GetTestPrefix() string {
	user, err := user.Current()
	if err != nil {
		panic(fmt.Errorf("can't get current user: %w", err))
	}
	return user.Username
}

func (config *Config) Construct() {
	if config.ClientIdPrefix == "$(ARIA_CLIENT_ID_PREFIX)" {
		config.ClientIdPrefix = GetTestPrefix()
	}

	authKey, err := os.ReadFile(config.GetAriaSystemAuthKeyFile())
	if err == nil {
		config.SetAriaSystemAuthKey(string(authKey))
	}

	clientNoString, err := os.ReadFile(config.GetAriaSystemClientNoFile())
	if err == nil {
		clientNoInt, err := strconv.Atoi(string(clientNoString))
		if err == nil {
			config.SetAriaSystemClientNo(int64(clientNoInt))
		}
	}
	accountId, err := os.ReadFile(config.GetAWSSQSAccountIdFile())
	if err == nil {
		config.SetAWSSQSAccountId(string(accountId))
	}
}

func (config *Config) GetListenPort() uint16 {
	return config.ListenPort
}

func (config *Config) SetListenPort(port uint16) {
	config.ListenPort = port
}

func (config *Config) GetSyncInterval() int {
	return config.SyncInterval
}

func (config *Config) SetSyncInterval(syncInterval int) {
	config.SyncInterval = syncInterval
}

func (config *Config) GetUsageInterval() int {
	return config.UsageInterval
}

func (config *Config) SetUsageInterval(usageInterval int) {
	config.UsageInterval = usageInterval
}

func (config *Config) GetClientIdPrefix() string {
	return config.ClientIdPrefix
}

func (config *Config) SetClientIdPrefix(clientIdPrefix string) {
	config.ClientIdPrefix = clientIdPrefix
}

func (config *Config) GetDBHost() string {
	return config.DB.Host
}

func (config *Config) GetDBPort() int {
	return config.DB.Port
}

func (config *Config) GetDBConnectTimeout() int {
	return config.DB.connectTimeout
}

func (config *Config) GetDBUser() string {
	return config.DB.User
}

func (config *Config) GetDBName() string {
	return config.DB.Name
}

func (config *Config) GetAriaSystemAuthKeyFile() string {
	return config.AriaSystem.AuthKeyFile
}

func (config *Config) GetAriaSystemClientNoFile() string {
	return config.AriaSystem.ClientNoFile
}

func (config *Config) SetAriaSystemAuthKey(authKey string) {
	config.AriaSystem.AuthKey = authKey
}

func (config *Config) SetAriaSystemClientNo(clientNo int64) {
	config.AriaSystem.ClientNo = clientNo
}

func (config *Config) SetAriaSystemDirectPostUrl(directPostUrl string) {
	config.AriaSystem.DirectPostUrl = directPostUrl
}

func (config *Config) SetAriaSystemFunctionMode(functionMode string) {
	config.AriaSystem.FunctionMode = functionMode
}

func (config *Config) SetAriaSystemTaxGroup(taxGroup int) {
	config.AriaSystem.TaxGroup = taxGroup
}

func (config *Config) SetAriaSystemClientTaxGroupId(clientTaxGroupId string) {
	config.AriaSystem.ClientTaxGroupId = clientTaxGroupId
}

func (config *Config) SetProductCatalogStorageUsageUnitType(storageUsageUnitType string) {
	config.BillingDriverAria.ProductCatalog.StorageUsageUnitType = storageUsageUnitType
}

func (config *Config) SetStorageInstanceTypes(storageInstanceTypes []string) {
	config.StorageInstanceTypes = storageInstanceTypes
}

func (config *Config) GetAriaSystemAuthKey() string {
	return config.AriaSystem.AuthKey
}

func (config *Config) GetAriaSystemClientNo() int64 {
	return config.AriaSystem.ClientNo
}

func (config *Config) GetAriaSystemReleaseVersion() string {
	return config.AriaSystem.ReleaseVersion
}

func (config *Config) GetAriaSystemCoreApiSuffix() string {
	return config.AriaSystem.CoreApiSuffix
}

func (config *Config) GetAriaSystemServerUrlCoreApi() string {
	return config.AriaSystem.Server.CoreApiUrl
}

func (config *Config) GetAriaSystemServerUrlAdminToolsApi() string {
	return config.AriaSystem.Server.AdminApiUrl
}

func (config *Config) GetAriaSystemServerUrlObjectQueryApi() string {
	return config.AriaSystem.Server.ObjectApiUrl
}

func (config *Config) GetPremiumDefaultCreditAmount() float64 {
	return config.PremiumDefaultCreditAmount
}

func (config *Config) GetPremiumDefaultCreditExpirationDays() int {
	return config.PremiumDefaultCreditExpirationDays
}

func (config *Config) GetAriaSystemDirectPostUrl() string {
	return config.AriaSystem.DirectPostUrl
}

func (config *Config) GetAriaSystemFunctionMode() string {
	return config.AriaSystem.FunctionMode
}

func (config *Config) GetAriaSystemTaxGroup() int {
	return config.AriaSystem.TaxGroup
}

func (config *Config) GetAriaSystemClientTaxGroupId() string {
	return config.AriaSystem.ClientTaxGroupId
}

func (config *Config) GetAWSSQSQueueUrl() string {
	if !strings.Contains(config.AWS.SQS.QueueUrl, config.AWS.SQS.Region) {
		config.AWS.SQS.QueueUrl = strings.Replace(config.AWS.SQS.QueueUrl, AWS_QUEUE_REGION_DEFAULT, config.AWS.SQS.Region, -1)
	}
	if !strings.Contains(config.AWS.SQS.QueueUrl, config.AWS.SQS.AccountId) {
		config.AWS.SQS.QueueUrl = strings.Replace(config.AWS.SQS.QueueUrl, AWS_QUEUE_ACCOUNTID_DEFAULT, config.AWS.SQS.AccountId, -1)
	}
	return config.AWS.SQS.QueueUrl
}

func (config *Config) GetAWSSQSRegion() string {
	return config.AWS.SQS.Region
}

func (config *Config) GetAWSSQSAccountId() string {
	return config.AWS.SQS.AccountId
}

func (config *Config) GetAWSSQSAccountIdFile() string {
	return config.AWS.SQS.AccountIdFile
}

func (config *Config) SetAWSSQSAccountId(accountId string) {
	config.AWS.SQS.AccountId = accountId
}

func (config *Config) GetAWSSQSCredentialsFile() string {
	awsCredentials, err := os.ReadFile(config.AWS.SQS.CredentialsFile)
	if err == nil && !strings.EqualFold(string(awsCredentials), "") {
		return config.AWS.SQS.CredentialsFile
	}
	return ""
}

func (config *Config) GetFeaturesDeactivationScheduler() bool {
	return config.BillingDriverAria.Features.DeactivationScheduler
}

func (config *Config) GetAriaSystemApiCrtFile() string {
	return config.AriaSystem.ApiCrtFile
}

func (config *Config) GetAriaSystemApiKeyFile() string {
	return config.AriaSystem.ApiKeyFile
}

func (config *Config) GetAriaSystemInsecureSsl() bool {
	return config.AriaSystem.InsecureSsl
}

func (config *Config) GetProductCatalogStorageUsageUnitType() string {
	return config.BillingDriverAria.ProductCatalog.StorageUsageUnitType
}

func (config *Config) GetFeaturesSyncStoragePlan() bool {
	return config.BillingDriverAria.Features.SyncStoragePlan
}

func (config *Config) GetStorageInstanceTypes() []string {
	return config.StorageInstanceTypes
}

func (config *Config) GetAriaSystemStorageUsageUnitType() string {
	return config.AriaSystem.StorageUsageUnitType
}

func (config *Config) GetProductCatalogInferenceUsageUnitType() string {
	return config.BillingDriverAria.ProductCatalog.InferenceUsageUnitType
}

func (config *Config) GetFeaturesSyncInferencePlan() bool {
	return config.BillingDriverAria.Features.SyncInferencePlan
}

func (config *Config) GetAriaSystemInferenceUsageUnitType() string {
	return config.AriaSystem.InferenceUsageUnitType
}

func (config *Config) GetProductCatalogTokenUsageUnitType() string {
	return config.BillingDriverAria.ProductCatalog.TokenUsageUnitType
}

func (config *Config) GetFeaturesSyncTokenPlan() bool {
	return config.BillingDriverAria.Features.SyncTokenPlan
}

func (config *Config) GetAriaSystemTokenUsageUnitType() string {
	return config.AriaSystem.TokenUsageUnitType
}

func (config *Config) GetProductCatalogApiVersion() string {
	return config.BillingDriverAria.ProductCatalog.ApiVersion
}
