package IKS_Usage_VM_2nd_NodeGroup_PU_test

import (
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/library/auth"
	"goFramework/framework/service_api/financials"
	"goFramework/framework/service_api/iks"
	"goFramework/ginkGo/financials/financials_utils"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("Check IKS NodeGroup OutOfCredits 2nd intel user", Ordered, Label("IKS"), func() {
	productList := []string{}
	var found bool

	It("Validate Enterprise cloudAccount", func() {
		logger.Logf.Info("Checking user cloudAccount")
		url := base_url + "/v1/cloudaccounts"
		code, body := financials.GetCloudAccountByName(url, token, userName)
		Expect(code).To(Equal(200), "Failed to retrieve CloudAccount Details")
		cloudAccountType = gjson.Get(body, "type").String()
		Expect(cloudAccountType).NotTo(BeNil(), "Failed to retrieve CloudAccount Type")
		cloudAccIdid = gjson.Get(body, "id").String()
		Expect(cloudAccIdid).NotTo(BeNil(), "Failed to retrieve CloudAccount ID")
		logger.Log.Info("CloudAccount Validated")
	})

	It("Create Cloud credits for enterprise user by redeeming coupons", func() {
		fmt.Println("Starting Cloud Credits Creation...")
		create_coupon_endpoint := base_url + "/v1/cloudcredits/coupons"
		coupon_payload := financials_utils.EnrichCreateCouponPayload(financials_utils.GetCreateCouponPayload(), 10, "idc_billing@intel.com", 1)
		coupon_creation_status, coupon_creation_body := financials.CreateCoupon(create_coupon_endpoint, token, coupon_payload)
		Expect(coupon_creation_status).To(Equal(200), "Failed to create coupon")

		// Redeem coupon
		redeem_coupon_endpoint := base_url + "/v1/cloudcredits/coupons/redeem"
		redeem_payload := financials_utils.EnrichRedeemCouponPayload(financials_utils.GetRedeemCouponPayload(), gjson.Get(coupon_creation_body, "code").String(), place_holder_map["cloud_account_id"])
		coupon_redeem_status, _ := financials.RedeemCoupon(redeem_coupon_endpoint, token, redeem_payload)
		Expect(coupon_redeem_status).To(Equal(200), "Failed to redeem coupon")
	})

	It("Get IKS Versions", func() {
		logger.Logf.Info("Getting all IKS versions")
		baseUrl := compute_url + "/v1/cloudaccounts/" + place_holder_map["cloud_account_id"] + "/iks/metadata/runtimes"
		fmt.Println("IKS Versions URL...", baseUrl)
		response_status, responseBody := iks.GetVersions(baseUrl, token)
		iks_version = gjson.Get(responseBody, "runtimes.0.k8sversionname.0").String()
		Expect(response_status).To(Equal(200), "Failed to retrieve IKS version")
	})
	//1.24
	It("Create IKS cluster", func() {
		logger.Logf.Info("Creating IKS Cluster")
		baseUrl := compute_url + "/v1/cloudaccounts/" + place_holder_map["cloud_account_id"] + "/iks/clusters"
		payload := `{"instanceType": "vm-spr-sml", "description": "test", "k8sversionname": "` + iks_version + `", "name": "test", "runtimename": "", "network": {"region": "us", "enableloadbalancer": true, "clusterdns": "0.0.0.0", "clustercidr": "0.0.0.0"}, "tags": []}`
		logger.Logf.Info("Payload: ", payload)
		response_status, responseBody := iks.CreateCluster(baseUrl, token, payload)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).NotTo(Equal(200), "Failed creating IKS cluster")
	})

	It("Create IKS second cluster", func() {
		logger.Logf.Info("Creating IKS Cluster")
		baseUrl := compute_url + "/v1/cloudaccounts/" + place_holder_map["cloud_account_id"] + "/iks/clusters"
		payload := `{"instanceType": "vm-spr-sml", "description": "test", "k8sversionname": "` + iks_version + `", "name": "test", "runtimename": "", "network": {"region": "us", "enableloadbalancer": true, "clusterdns": "0.0.0.0", "clustercidr": "0.0.0.0"}, "tags": []}`
		logger.Logf.Info("Payload: ", payload)
		response_status, responseBody := iks.CreateCluster(baseUrl, token, payload)
		logger.Logf.Info("Response body", responseBody)
		Expect(response_status).NotTo(Equal(200), "Failed creating second IKS cluster")
	})

	It("Validate Usages showing up for all products", func() {
		usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + place_holder_map["cloud_account_id"]
		var usage_response_status int
		var usage_response_body string
		arr := gjson.Result{}

		Eventually(func() bool {
			token = "Bearer " + auth.Get_Azure_Admin_Bearer_Token()
			fmt.Println("TOKEN...", token)
			usage_response_status, usage_response_body = financials.GetUsage(usage_url, token)
			fmt.Println("usage_response_body", usage_response_body)
			fmt.Println("usage_response_status", usage_response_status)
			Expect(usage_response_status).To(Equal(200), "Failed to validate usage_response_status")
			logger.Logf.Info("usage_response_body: %s ", usage_response_body)
			result := gjson.Parse(usage_response_body)
			arr = gjson.Get(result.String(), "usages")
			fmt.Println("products usages", arr)
			fmt.Println("products array", arr.Array())
			if len(arr.Array()) > 0 {
				return true
			}
			fmt.Println("Waiting 40 more minutes to get products usages...")
			return false
		}, 8*time.Hour, 15*time.Minute).Should(BeTrue())
		for _, s := range productList {
			found = false
			fmt.Println("Product Name", s)
			arr.ForEach(func(key, value gjson.Result) bool {
				data := value.String()
				logger.Logf.Infof("Usage Data : %s", data)
				product := gjson.Get(data, "productType").String()
				logger.Logf.Infof("Product Data : %s", product)
				if strings.Contains(product, "sml") {
					found = true
				} else {
					found = true
					Amount := gjson.Get(data, "amount").String()
					actualAMount, _ := strconv.ParseFloat(Amount, 64)
					Expect(actualAMount).Should(BeNumerically(">", float64(0)), "Failed to get positive usage")
				}
				return true // keep iterating
			})
			Expect(found).To(Equal(true), "Usage not found for the product", s)
		}
	})
})
