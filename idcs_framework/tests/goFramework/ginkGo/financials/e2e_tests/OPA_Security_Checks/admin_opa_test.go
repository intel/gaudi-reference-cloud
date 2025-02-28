package OPA_Security_Checks_test

import (
	"crypto/tls"
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var baseURL string

type TestResult struct {
	Endpoint     string
	Method       string
	Passed       bool
	Error        error
	ResponseCode int
	ResponseBody string
}

var _ = Describe("AdminOpaChecks", func() {
	var swagger *financials_utils.Swagger
	swaggerPath := "../../../../../../../go/pkg/pb/swagger/all.swagger.json"

	var admin_billing_endpoints = []string{"BillingUsageService_Read", "BillingCouponService_Read", "BillingCouponService_Create", "BillingCouponService_Disable", "BillingOptionService_Read", "BillingCreditService_Read", "BillingCreditService_Create", "BillingCreditService_ReadUnappliedCreditBalance", "BillingDeactivateInstancesService_GetDeactivateInstances"}
	var admin_cloudaccount_endpoints = []string{"CloudAccountService_GetByName", "CloudAccountService_Search", "CloudAccountService_Delete", "CloudAccountService_GetById", "CloudAccountService_Update", "CloudAccountService_Create", "CloudAccountService_Ensure"}
	var admin_cloudcredits_endpoints = []string{"CloudCreditsCouponService_Read", "CloudCreditsCouponService_Create", "CloudCreditsCouponService_ReadCredits", "CloudCreditsCouponService_Disable", "CloudCreditsCreditService_CreditMigrate"}
	var admin_fleetadmin_endpoints = []string{"FleetAdminService_Ping", "FleetAdminUIService_SearchNodes", "FleetAdminUIService_SearchComputeNodePoolsForPoolAccountManager", "FleetAdminUIService_SearchComputeNodePoolsForNodeAdmin", "FleetAdminUIService_PutComputeNodePool", "FleetAdminUIService_SearchCloudAccountsForComputeNodePool", "FleetAdminUIService_DeleteCloudAccountFromComputeNodePool", "FleetAdminUIService_UpdateNode", "FleetAdminUIService_SearchComputeNodePoolsForInstanceScheduling", "FleetAdminUIService_AddCloudAccountToComputeNodePool"}
	var admin_iks_endpoints = []string{"IksAdmin_AuthenticateIKSAdminUser", "IksAdmin_ClusterRecreate", "IksAdmin_ClusterSnapshot", "IksAdmin_CreateIMI", "IksAdmin_CreateInstanceTypes", "IksAdmin_CreateK8SVersion", "IksAdmin_CreateNewAddOn", "IksAdmin_DeleteAddOn", "IksAdmin_DeleteIMI", "IksAdmin_DeleteInstanceType", "IksAdmin_DeleteK8SVersion", "IksAdmin_DeleteLoadBalancer", "IksAdmin_GetAddOn", "IksAdmin_GetAddOns", "IksAdmin_GetCloudAccountApproveList", "IksAdmin_GetCluster", "IksAdmin_GetClusters", "IksAdmin_GetControlPlaneSSHKeys", "IksAdmin_GetEvents", "IksAdmin_GetIMI", "IksAdmin_GetIMIs", "IksAdmin_GetIMIsInfo", "IksAdmin_GetInstanceType", "IksAdmin_GetInstanceTypeInfo", "IksAdmin_GetInstanceTypes", "IksAdmin_GetK8SVersion", "IksAdmin_GetLoadBalancer", "IksAdmin_GetLoadBalancers", "IksAdmin_PostCloudAccountApproveList", "IksAdmin_PostLoadBalancer", "IksAdmin_PutAddOn", "IksAdmin_PutCPNodegroup", "IksAdmin_PutCloudAccountApproveList", "IksAdmin_PutIMI", "IksAdmin_PutInstanceType", "IksAdmin_PutK8SVersion", "IksAdmin_PutLoadBalancer", "IksAdmin_UpdateIMIInstanceTypeToK8sCompatibility", "IksAdmin_UpdateInstanceTypeIMIToK8sCompatibility", "IksAdmin_UpgradeClusterControlPlane"}
	var admin_instance_endpoints = []string{"InstanceService_Search", "InstanceService_Search2", "InstanceService_Delete", "InstanceService_Delete2", "InstanceGroupService_Search", "InstanceGroupService_Delete"}
	var admin_metering_endpoints = []string{"MeteringService_Update", "MeteringService_Search", "MeteringService_Create", "MeteringService_Create", "MeteringService_FindPrevious", "MeteringService_CreateInvalidRecords"}
	var admin_storage_endpoints = []string{"StorageAdminService_GetResourceUsage", "StorageAdminService_InsertStorageQuotaByAccount", "StorageAdminService_UpdateStorageQuotaByAccount", "StorageAdminService_DeleteStorageQuotaByAccount", "StorageAdminService_GetStorageQuotaByAccount", "StorageAdminService_GetAllStorageQuota"}
	var admin_product_endpoints = []string{"ProductAccessService_ReadAccess", "ProductAccessService_CheckProductAccess", "ProductAccessService_AddAccess", "ProductAccessService_RemoveAccess", "ProductCatalogService_AdminRead", "ProductCatalogService_UserRead"}
	var user_endpoints = []string{"SshPublicKeyService_Search", "VNetService_Search"}

	var testResults []TestResult

	BeforeEach(func() {
		var err error
		swagger, err = financials_utils.ParseSwaggerFile(swaggerPath)
		Expect(err).To(BeNil())
		base_url = compute_utils.GetBaseUrl()
		compute_url = compute_utils.GetComputeUrl()

	})

	runTests := func(endpoints []string, shouldPass bool, passed_url string) {
		baseURL = passed_url
		extractedEndpoints := financials_utils.ExtractEndpointsByOperationID(swagger, endpoints)
		for endpoint, method := range extractedEndpoints {
			fmt.Printf("Testing URL: %s %s\n", baseURL, endpoint)
			fmt.Printf("Testing endpoint: %s %s\n", method, endpoint)

			// Replace placeholder with cloudaccount value
			endpoint = strings.Replace(endpoint, "{metadata.cloudAccountId}", place_holder_map_su["cloud_account_id"], -1)

			test_url := baseURL + endpoint
			var reqBody io.Reader
			if method == "POST" || method == "PUT" {
				// Example JSON body, adjust as needed
				jsonBody := `{"key": "value"}`
				reqBody = strings.NewReader(jsonBody)
			}

			req, err := http.NewRequest(strings.ToUpper(method), test_url, reqBody)
			Expect(err).NotTo(HaveOccurred())

			// Add headers
			req.Header.Add("Authorization", token)
			req.Header.Add("Content-Type", "application/json")

			// Set proxy on flex envs
			var client *http.Client
			idcEnv := os.Getenv("IDC_ENV")
			if idcEnv == "staging" || idcEnv == "dev3" || idcEnv == "qa1" {
				proxyURL := os.Getenv("https_proxy")
				if proxyURL != "" {
					proxy, err := url.Parse(proxyURL)
					Expect(err).NotTo(HaveOccurred())
					client = &http.Client{
						Transport: &http.Transport{
							Proxy:           http.ProxyURL(proxy),
							TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
						},
					}
				} else {
					client = &http.Client{
						Transport: &http.Transport{
							TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
						},
					}
				}
			} else {
				client = &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
					},
				}
			}
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())

			if shouldPass {
				Expect(resp.StatusCode).To(Not(Equal(403)), fmt.Sprintf("Endpoint: %s %s\nResponse Code: %d\nResponse Body: %s\n", method, test_url, resp.StatusCode, string(body)))
			} else {
				Expect(resp.StatusCode).To(Equal(403), fmt.Sprintf("Endpoint: %s %s\nResponse Code: %d\nResponse Body: %s\n", method, test_url, resp.StatusCode, string(body)))
			}

			testResults = append(testResults, TestResult{Endpoint: test_url, Method: method, Passed: true, Error: nil, ResponseCode: resp.StatusCode, ResponseBody: string(body)})

			fmt.Printf("Response Code: %d\n", resp.StatusCode)
			fmt.Printf("Response Body: %s\n", string(body))
		}
	}

	It("Admin user can access Admin billing Endpoints", func() {
		runTests(admin_billing_endpoints, true, base_url)
	})

	It("Admin user can access Admin Cloudaccount endpoints", func() {
		runTests(admin_cloudaccount_endpoints, true, base_url)
	})

	It("Admin user can access Admin Cloudcredits endpoints", func() {
		runTests(admin_cloudcredits_endpoints, true, base_url)
	})

	It("Admin user can access Admin Fleetadmin endpoints", func() {
		runTests(admin_fleetadmin_endpoints, true, base_url)
	})

	It("Admin user can access Admin IKS endpoints", func() {
		runTests(admin_iks_endpoints, true, base_url)
	})

	It("Admin user can access Admin Instance endpoints", func() {
		runTests(admin_instance_endpoints, true, base_url)
	})

	It("Admin user can access Admin metering endpoints", func() {
		runTests(admin_metering_endpoints, true, base_url)
	})

	It("Admin user can access Admin storage endpoints", func() {
		runTests(admin_storage_endpoints, true, base_url)
	})

	It("Admin user can access Admin Product endpoints", func() {
		runTests(admin_product_endpoints, true, base_url)
	})

	It("Admin user cannot access user endpoints", func() {
		runTests(user_endpoints, false, compute_url)
	})

	AfterEach(func() {
		fmt.Println("Test Summary:")
		passedCount := 0
		failedCount := 0
		for _, result := range testResults {
			status := "PASSED"
			if !result.Passed {
				status = "FAILED"
				logger.Log.Info("failed")
				fmt.Println("failed")
				failedCount++
			} else {
				logger.Log.Info("passed")
				fmt.Println("passed")
				passedCount++
			}
			fmt.Printf("Endpoint: %s %s - %s\n", result.Method, result.Endpoint, status)
			if result.Error != nil {
				fmt.Printf("Error: %v\n", result.Error)
			}
			fmt.Printf("Response Code: %d\n", result.ResponseCode)
			fmt.Printf("Response Body: %s\n", result.ResponseBody)
		}
		fmt.Printf("Total Passed: %d\n", passedCount)
		fmt.Printf("Total Failed: %d\n", failedCount)
	})
})
