// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package iks_integration_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"
	"testing"

	auth_admin "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks_commons/auth"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher/lasso/pkg/log"
)

var createClusterUrl, getNodeGroupUrl, getNodeGroupByIdUrl, putNodeGroupByIdUrl, getClusterUrl, getClustersUrl,
	createNodeGroupUrl, deleteClusterUrl, createVIPSUrl, getVIPSUrl, getVIPByIDurl, getKubeConfigUrl, getInstanceTypesUrl,
	deleteNodeGroupByIdUrl, deleteVIPByIdUrl, updateSecRuleUrl, getSecRuleUrl, deleteSecRuleUrl string
var getAdminClustersURL, getAdminClusterByUUIDURL, getAdminVIPSUrl string
var getUserCloudAccount string
var getMethod, deleteMethod, postMethod, putMethod string
var hostConfig HostConfig
var bearerToken, adminToken, authConfigPath, tokenOutputPath, accountEmail, environment string
var bearerTokenExpiry, adminTokenExpiry int64
var err error

func TestIksIntegrationTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IksIntegrationTest Suite")
}

var cloudAccount string
var pipelineRun bool
var testEnv string
var region string
var userEmail string
var sshKeyName string
var k8sVersion string
var runStorageTests bool
var noAdminAPIUserEmail string
var runSecRulesTests bool

func init() {
	flag.BoolVar(&pipelineRun, "pipelineRun", false, "Set to true for Jenkins runs")
	flag.StringVar(&cloudAccount, "cloudAccount", "", "Target cloud account")
	flag.StringVar(&testEnv, "testEnv", "staging", "Target environment: staging or production")
	flag.StringVar(&region, "region", "us-staging-1", "Target region")
	flag.StringVar(&userEmail, "userEmail", "", "User's email")
	flag.StringVar(&sshKeyName, "sshKeyName", "compute-key", "SSH key name to use for node group creation")
	flag.StringVar(&k8sVersion, "k8sVersion", "1.28", "K8s version to use for node group creation")
	flag.BoolVar(&runStorageTests, "runStorageTests", false, "Set to true to run storage tests")
	flag.StringVar(&noAdminAPIUserEmail, "noAdminAPIUserEmail", "", "User account with no admin API access")
	flag.BoolVar(&runSecRulesTests, "runSecRulesTests", false, "Set to true to run security rules tests")
}

var _ = BeforeSuite(func() {
	flag.Parse()
	log.Infof("*******************************IKS E2E Integration Test Started******************************")

	log.Infof("============Started Config Setup=========")
	hostConfig, err = ReadHostConfig("config/config.yaml")
	Expect(err).To(BeNil())

	accountId := ""
	host := ""

	if pipelineRun {
		log.Infof("Non-interactive Pipeline run")
		log.Infof("Account: %s, Email: %s, Env: %s, Region: %s", cloudAccount, userEmail, testEnv, region)

		// For pipeline runs we want all tests to run
		hostConfig.AddNodeToNodeGroup = true
		hostConfig.CreateCluster = true
		hostConfig.CreateNodeGroup = true
		hostConfig.CreateVIP = true
		hostConfig.DeleteCluster = true
		hostConfig.DeleteNodeFromNodeGroup = true
		hostConfig.DeleteSpecificNodeGroup = true
		hostConfig.DeleteSpecificVIP = true
		hostConfig.DownloadKubeConfig = true
		hostConfig.RunStorageTests = runStorageTests
		hostConfig.RunSecRulesTests = runSecRulesTests

		environment = testEnv
		region = strings.TrimSpace(region)

		if environment == "staging" {
			regionNumber := region[len(region)-1:]
			host = "https://staging-idc-us-" + regionNumber + ".eglb.intel.com"
		} else if environment == "production" {
			host = "https://compute-" + region + "-api.cloud.intel.com"
		} else if environment == "qa1" {
			regionNumber := region[len(region)-1:]
			host = "https://qa1-idc-us-" + regionNumber + ".eglb.intel.com"
		}

		accountEmail = userEmail
		accountId = cloudAccount

		//load auth config file for token generation
		authConfigPath = "config/auth_config.json"
		auth_admin.Get_config_file_data(authConfigPath)

		adminToken = ""
		adminTokenExpiry = 0

		if userEmail == noAdminAPIUserEmail {
			log.Infof("Skipping admin token retrieval since account does not have admin permissions")
		} else {
			log.Infof("Retrieving admin token...")
			adminToken, adminTokenExpiry = auth_admin.Get_Azure_Admin_Bearer_Token(testEnv)
			log.Infof("Admin token: %s", adminToken)
		}

		log.Infof("Retrieving user token...")
		bearerToken, bearerTokenExpiry, err = auth_admin.Get_Azure_Bearer_Token(userEmail)
		if err != nil {
			log.Infof("err: %s", err.Error())
		}
		Expect(err).To(BeNil())
		log.Infof("Bearer Token: %s", bearerToken)

	} else {
		// Interactive run
		bearerToken = hostConfig.BearerToken
		if bearerToken == "" {
			log.Errorf("Bearer Token cannot be Empty. Please add Valid Bearer Token in config file")
			Expect(bearerToken).ToNot(Equal(""))
		}

		adminToken = hostConfig.AdminToken
		if adminToken == "" {
			log.Errorf("Bearer Token cannot be Empty. Please add Valid Bearer Token in config file")
			Expect(adminToken).ToNot(Equal(""))
		}

		bearerTokenExpiry = hostConfig.BearerTokenExpiry
		if bearerTokenExpiry == 0 {
			log.Errorf("Bearer Token Expiry cannot be Empty. Please add Valid Bearer Token Expiry in config file")
			Expect(bearerTokenExpiry).ToNot(Equal(0))
		}

		adminTokenExpiry = hostConfig.AdminTokenExpiry
		if adminTokenExpiry == 0 {
			log.Errorf("Admin Token Expiry cannot be Empty. Please add Valid Admin Token Expiry in config file")
			Expect(adminTokenExpiry).ToNot(Equal(0))
		}

		environment = hostConfig.Environment
		if environment == "" {
			log.Errorf("Environment cannot be Empty. Please add Valid Environment in config file")
			Expect(environment).ToNot(Equal(""))
		}

		host = hostConfig.Host
		if host == "" {
			log.Errorf("Host cannot be Empty. Please add Valid Host Token in config file")
			Expect(host).ToNot(Equal(""))
		}
		log.Infof("Read Host from config file and current host is: %s", host)

		globalHost := hostConfig.GlobalHost
		if globalHost == "" {
			log.Errorf("Global Host cannot be Empty. Please add Valid Global Host Token in config file")
			Expect(globalHost).ToNot(Equal(""))
		}
		log.Infof("Read Global Host from config file and current host is: %s", globalHost)

		accountEmail = hostConfig.DefaultAccount
		if accountEmail == "" {
			log.Errorf("Account Email cannot be Empty. Please add Valid Account in config file")
			Expect(accountEmail).ToNot(Equal(""))
		}
		log.Infof("Read Account Email from config file and current account id is: %s", accountEmail)

		// Get Cloud AccountID from CloudAccount URL
		var cloudAccountDetails CloudAccountDetails
		getUserCloudAccount = fmt.Sprintf("%s/v1/cloudaccounts/name/%s", globalHost, accountEmail)
		//get Cloud Account Details
		data, err := SendHttpRequest(getUserCloudAccount, getMethod, nil, adminToken)
		Expect(err).To(BeNil())
		Expect(data).ToNot(BeNil())
		err = json.Unmarshal(data, &cloudAccountDetails)
		Expect(err).To(BeNil())

		accountId = cloudAccountDetails.AccountID

		if accountId == "" {
			log.Errorf("Account ID cannot be Empty. Please add Valid Account in config file")
			Expect(accountId).ToNot(Equal(""))
		}
		log.Infof("Read Account ID from Cloud Account API and current account id is: %s", accountId)
	}

	// IKS URL's
	createClusterUrl = fmt.Sprintf("%s/v1/cloudaccounts/%s/iks/clusters", host, accountId)
	getClustersUrl = fmt.Sprintf("%s/v1/cloudaccounts/%s/iks/clusters", host, accountId)
	getClusterUrl = fmt.Sprintf("%s/v1/cloudaccounts/%s/iks/clusters/%s", host, accountId, "%s")
	deleteClusterUrl = fmt.Sprintf("%s/v1/cloudaccounts/%s/iks/clusters/%s", host, accountId, "%s")
	createNodeGroupUrl = fmt.Sprintf("%s/v1/cloudaccounts/%s/iks/clusters/%s/nodegroups", host, accountId, "%s")
	getNodeGroupUrl = fmt.Sprintf("%s/v1/cloudaccounts/%s/iks/clusters/%s/nodegroups?nodes=true", host, accountId, "%s")
	getNodeGroupByIdUrl = fmt.Sprintf("%s/v1/cloudaccounts/%s/iks/clusters/%s/nodegroups/%s?nodes=true", host, accountId, "%s", "%s")
	putNodeGroupByIdUrl = fmt.Sprintf("%s/v1/cloudaccounts/%s/iks/clusters/%s/nodegroups/%s?nodes=true", host, accountId, "%s", "%s")
	deleteNodeGroupByIdUrl = fmt.Sprintf("%s/v1/cloudaccounts/%s/iks/clusters/%s/nodegroups/%s", host, accountId, "%s", "%s")
	createVIPSUrl = fmt.Sprintf("%s/v1/cloudaccounts/%s/iks/clusters/%s/vips", host, accountId, "%s")
	getVIPSUrl = fmt.Sprintf("%s/v1/cloudaccounts/%s/iks/clusters/%s/vips", host, accountId, "%s")
	getVIPByIDurl = fmt.Sprintf("%s/v1/cloudaccounts/%s/iks/clusters/%s/vips/%s", host, accountId, "%s", "%s")
	deleteVIPByIdUrl = fmt.Sprintf("%s/v1/cloudaccounts/%s/iks/clusters/%s/vips/%s", host, accountId, "%s", "%s")
	getKubeConfigUrl = fmt.Sprintf("%s/v1/cloudaccounts/%s/iks/clusters/%s/kubeconfig", host, accountId, "%s")
	getInstanceTypesUrl = fmt.Sprintf("%s/v1/cloudaccounts/%s/iks/metadata/instancetypes", host, accountId)
	updateSecRuleUrl = fmt.Sprintf("%s/v1/cloudaccounts/%s/iks/clusters/%s/security?t=%s", host, accountId, "%s", "%d")
	getSecRuleUrl = fmt.Sprintf("%s/v1/cloudaccounts/%s/iks/clusters/%s/security?t=%s", host, accountId, "%s", "%d")
	deleteSecRuleUrl = fmt.Sprintf("%s/v1/cloudaccounts/%s/iks/clusters/%s/security/%s?t=%s", host, accountId, "%s", "%d", "%d")

	//IKS Admin URL's
	getAdminClustersURL = fmt.Sprintf("%s/v1/iks/admin/clusters", host)
	getAdminClusterByUUIDURL = fmt.Sprintf("%s/v1/iks/admin/clusters/%s", host, "%s")
	getAdminVIPSUrl = fmt.Sprintf("%s/v1/iks/admin/clusters/%s/ilbs", host, "%s")

	//REST API's Method Type
	postMethod = "POST"
	getMethod = "GET"
	deleteMethod = "DELETE"
	putMethod = "PUT"

	log.Infof("============Ended Config Setup==========")
})

var _ = AfterSuite(func() {
	// We can do any cleanup or log any teardown logic here
	log.Infof("****************************IKS E2E Integration Test Ended*************************")
})

var _ = ReportAfterSuite("custom report", func(report Report) {
	ReportGeneration(report)
})
