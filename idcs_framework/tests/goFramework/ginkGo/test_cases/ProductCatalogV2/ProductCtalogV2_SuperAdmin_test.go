package ProductCatalogV2_test

import (
	"fmt"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"os"
	"strconv"

	"math/rand"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var vendorId string
var familyName string
var regionName string
var vendorName string
var rateId string
var rateSetId string
var metadataSetId string
var metadataId string
var serviceName string
var serviceLocation string

var _ = Describe("Product Catalog V2 Super Admin tests - Vendors", func() {

	It("Get Vendors", func() {
		fmt.Println("TOKEN: ", token)
		code, body := financials.GetVendorsV2(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting vendors.")
	})

	It("Get Vendors Admin", func() {
		code, body := financials.GetVendorsAdmin(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting admin user vendors.")
	})

	It("Create Vendor", func() {
		vendorName = "automation-vendor" + compute_utils.GetRandomString()
		payload := fmt.Sprintf(`{
			"organizationName": "automation-test",
			"description": "test vendor",
			"name": "%s"
		}`, vendorName)
		fmt.Println("Payload: ", payload)
		code, body := financials.CreateVendors(base_url, token, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed creating vendor.")
	})

	It("Get Vendors by Name", func() {
		code, body := financials.GetVendorsAdminByName(base_url, token, vendorName)
		fmt.Println("Response 1: ", body)
		Expect(code).To(Equal(200), "Failed getting vendors.")
		vendorId = gjson.Get(body, "vendors.0.name").String()
		fmt.Println("VendorId: ", vendorId)
		Expect(vendorId).NotTo(BeNil(), "VendorId value is nil.")
	})

	It("Update Vendor", func() {
		payload := fmt.Sprintf(`{
			"description": "test vendor updated + %s"
		}`, vendorName)
		fmt.Println("Payload: ", payload)
		code, body := financials.UpdateVendor(base_url, token, vendorId, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed updating vendor.")
	})

	It("Disable Vendor", func() {
		payload := fmt.Sprintf(`{
			"enabled": %t
		}`, false)
		fmt.Println("Payload: ", payload)
		code, body := financials.DeleteVendor(base_url, token, vendorId)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed disabling vendor.")
	})

	It("Get Vendors", func() {
		code, body := financials.GetVendorsAdmin(base_url, token)
		fmt.Println("Response 1: ", body)
		Expect(code).To(Equal(200), "Failed getting vendors.")
		vendorId = gjson.Get(body, "vendors.0.name").String()
		fmt.Println("VendorId: ", vendorId)
		Expect(vendorId).NotTo(BeNil(), "VendorId value is nil.")
	})
})

var _ = Describe("Product Catalog V2 Super Admin tests - Regions", func() {
	var regionId string

	It("Get Regions", func() {
		fmt.Println("uSER TOKEN", token)
		fmt.Println("vendor", vendorName)
		code, body := financials.GetRegions(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting regions.")
	})

	It("Get Regions Admin", func() {
		code, body := financials.GetRegionsAdmin(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting regions admin.")
	})

	It("Create Region", func() {
		regionName = "us-region-" + strconv.Itoa(rand.Intn(10000))
		payload := fmt.Sprintf(`{
				"description": "test",
				"name": "%s"
			}`, regionName)
		fmt.Println("Payload: ", payload)
		code, body := financials.CreateRegion(base_url, token, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed creating region.")
	})

	It("Get Regions Admin By Name", func() {
		code, body := financials.GetRegionsAdminByName(base_url, token, regionName)
		fmt.Println("Response 2: ", body)
		Expect(code).To(Equal(200), "Failed getting regions.")
		regionId = gjson.Get(body, "regions.0.name").String()
		Expect(regionId).NotTo(BeNil(), "RegionId value is nil.")
	})

	It("Update Region", func() {
		regionName = "automation-region-updated" + compute_utils.GetRandomString()
		payload := fmt.Sprintf(`{
				"description": "test region updated %s"
			}`, regionName)
		fmt.Println("Payload: ", payload)
		code, body := financials.UpdateRegion(base_url, token, regionId, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed updating region.")
	})

	It("Disable Region", func() {
		code, body := financials.DeleteRegion(base_url, token, regionName)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed disabling region.")
	})

	It("Get Regions Admin", func() {
		code, body := financials.GetRegions(base_url, token)
		fmt.Println("Response 2: ", body)
		Expect(code).To(Equal(200), "Failed getting regions.")
		regionId = gjson.Get(body, "regions.0.name").String()
		Expect(regionId).NotTo(BeNil(), "RegionId value is nil.")
	})
})

var _ = Describe("Product Catalog V2 Super Admin tests - Families", func() {
	var familyId string

	It("Get Families", func() {
		fmt.Println("uSER TOKEN", token)
		fmt.Println("vendor", vendorName)
		code, body := financials.GetFamilies(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting families.")
	})

	It("Get Families Admin", func() {
		code, body := financials.GetFamilies(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting families admin.")
	})

	It("Get Vendors", func() {
		code, body := financials.GetVendorsAdmin(base_url, token)
		fmt.Println("Response 1: ", body)
		Expect(code).To(Equal(200), "Failed getting vendors.")
		vendorId = gjson.Get(body, "vendors.0.name").String()
		fmt.Println("VendorId: ", vendorId)
		Expect(vendorId).NotTo(BeNil(), "VendorId value is nil.")
	})

	It("Create Family", func() {
		familyName = "automationtest"
		payload := fmt.Sprintf(`{
			"description": "Testing as a Service",
			"name": "%s",
			"vendorName": "%s"
		}`, familyName, vendorId)
		fmt.Println("Payload: ", payload)
		code, body := financials.CreateFamily(base_url, token, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed creating family.")
	})

	// TODO: Add test to filter by Id

	It("Get Families By Name", func() {
		code, body := financials.GetFamiliesAdminByName(base_url, token, familyName)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting families.")
		familyId = gjson.Get(body, "product_families.0.id").String()
		Expect(familyId).NotTo(BeNil(), "familyId value is nil.")
	})

	It("Update Family", func() {
		familyName := "automation-family-updated" + compute_utils.GetRandomString()
		payload := fmt.Sprintf(`{
			"description": "updated family %s"
		}`, familyName)
		fmt.Println("Payload: ", payload)
		code, body := financials.UpdateFamily(base_url, token, familyName, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed disabling family.")
	})

	It("Disable Family", func() {
		payload := fmt.Sprintf(`{
			"enabled": %t
		}`, false)
		fmt.Println("Payload: ", payload)
		code, body := financials.DeleteFamily(base_url, token, familyName)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed disabling family.")
	})

	It("Get Families", func() {
		code, body := financials.GetFamilies(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting families.")
		familyId = gjson.Get(body, "product_families.0.id").String()
		Expect(familyId).NotTo(BeNil(), "familyId value is nil.")
	})
})

var _ = Describe("Product Catalog V2 Super Admin tests - Rates", func() {

	It("Get Rate sets Admin", func() {
		code, body := financials.GetRates(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting rate sets.")
		rateSetId = gjson.Get(body, "rate_sets.0.id").String()
		Expect(rateId).NotTo(BeNil(), "rateSetId value is nil.")
	})

	It("Get Rates Admin", func() {
		code, body := financials.GetRates(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting rates.")
		rateId = gjson.Get(body, "rates.0.id").String()
		Expect(rateId).NotTo(BeNil(), "RateId value is nil.")
	})

	It("Create rate set Admin", func() {
		rateSetName := "automation-" + strconv.Itoa(rand.Intn(100))
		payload := fmt.Sprintf(`{
			"name": "%s"
		}`, rateSetName)
		fmt.Println("Payload: ", payload)
		code, body := financials.CreateRateSet(base_url, token, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed creating rate set.")
		fmt.Println("Creating duplicate...", body)
		code, body = financials.CreateRateSet(base_url, token, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Not(Equal(200)), "Rate names should not be duplicated.")
	})

	It("Get Rate sets Admin", func() {
		code, body := financials.GetRateSets(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting rate sets.")

		// Get the count of rate sets
		rateSetsCount := gjson.Get(body, "rate_sets.#").Int()
		Expect(rateSetsCount).To(BeNumerically(">", 0), "No rate sets found.")

		// Get the latest rate set id
		rateSetId = gjson.Get(body, fmt.Sprintf("rate_sets.%d.id", rateSetsCount-1)).String()
		fmt.Println("Latest rate set: ", rateSetId)
		Expect(rateSetId).NotTo(BeEmpty(), "rateSetId value is nil or empty string.")
	})

	It("Create rate Admin", func() {
		payload := fmt.Sprintf(`{
			"accountType": "%s",
			"byWhom": "%s",
			"rate": 10.3854,
			"rateSetId": 1,
			"usageUnitType": "dollarsPerMinute"
		}`, "premium", "test-automation")
		fmt.Println("Payload: ", payload)
		code, body := financials.CreateRate(base_url, token, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed creating rate.")
	})

	It("Get Rates Admin", func() {
		code, body := financials.GetRates(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting rates.")

		// Get the count of rates
		ratesCount := gjson.Get(body, "rates.#").Int()
		Expect(ratesCount).To(BeNumerically(">", 0), "No rates found.")

		// Get the latest rate id
		rateId = gjson.Get(body, fmt.Sprintf("rates.%d.id", ratesCount-1)).String()
		Expect(rateId).NotTo(BeEmpty(), "RateId value is nil or empty string.")
	})

	It("Update rate Admin", func() {
		payload := fmt.Sprintf(`{
			"accountType": "%s",
			"byWhom": "%s",
			"rate": 10,
			"rateSetId": %s,
			"usageUnitType": "dollarsPerMinute"
		}`, "standard", "test-automation-updated", rateSetId)
		fmt.Println("Payload: ", payload)
		code, body := financials.UpdateRate(base_url, token, payload, rateId)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed updating rate.")
	})

	It("Update rate - rate set id Admin", func() {
		updated_name := "test-automation-updated" + compute_utils.GetRandomString()
		payload := fmt.Sprintf(`{
			"name" : "%s"
		}`, updated_name)
		fmt.Println("Payload: ", payload)
		code, body := financials.UpdateRateSet(base_url, token, payload, rateSetId)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed updating rate set.")
	})

	It("Delete rate Admin", func() {
		fmt.Println("Rate ID: ", rateId)
		code, body := financials.DeleteRate(base_url, token, rateId)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed removing rate.")
	})

	It("Delete  rate set Admin", func() {
		fmt.Println("Rate Set ID: ", rateSetId)
		code, body := financials.DeleteRateSet(base_url, token, rateSetId)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed removing rate set.")
	})
})

var _ = Describe("Product Catalog V2 Super Admin tests - Metadata", func() {

	It("Get MetadataSet Admin", func() {
		code, body := financials.GetMetadataSet(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting metadata sets.")
		metadataSetId = gjson.Get(body, "metadata_sets.0.id").String()
		fmt.Println("metadataSetId", metadataSetId)
		Expect(metadataSetId).NotTo(BeEmpty(), "MetadataSetId value is nil.")
	})

	It("Create MetadataSet Admin", func() {
		metadata_name := "metadatasettest" + compute_utils.GetRandomString()
		payload := fmt.Sprintf(`{
			"context": "product",
			"name": "%s"
		}`, metadata_name)
		code, body := financials.CreateMetadataSet(base_url, token, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed creating metadata set.")
	})

	It("Get MetadataSet Admin", func() {
		code, body := financials.GetMetadataSet(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting metadata sets.")

		// Get the count of metadata sets
		metadataSetsCount := gjson.Get(body, "metadata_sets.#").Int()
		Expect(metadataSetsCount).To(BeNumerically(">", 0), "No metadata sets found.")

		// Get the latest metadata set id
		metadataSetId = gjson.Get(body, fmt.Sprintf("metadata_sets.%d.id", metadataSetsCount-1)).String()
		fmt.Println("metadataSetId", metadataSetId)
		Expect(metadataSetId).NotTo(BeEmpty(), "MetadataSetId value is nil or empty string.")
	})

	It("Update MetadataSet Admin", func() {
		payload := fmt.Sprintf(`{
			"context": "%s"
		}`, "test-automation-updated")
		code, body := financials.UpdateMetadataSet(base_url, token, metadataSetId, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed updating metadata set.")
	})

	It("Delete Metadata Set Admin", func() {
		code, body := financials.DeleteMetadataSet(base_url, token, metadataSetId)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed deleting metadata set.")
	})

	var metadataId string
	It("Get Metadata Admin", func() {
		code, body := financials.GetMetadata(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting metadatas.")
		metadataId = gjson.Get(body, "metadata.0.id").String()
		fmt.Println("metadataId", metadataId)
		Expect(metadataId).NotTo(BeNil(), "MetadataId value is nil.")
	})

	It("Create Metadata Admin", func() {
		payload := fmt.Sprintf(`{
		"context": "%s",
		"key": "test",
		"metadataSetId": 1,
		"type": "test",
		"value": "test"
		}`, "test-automation")
		code, body := financials.CreateMetadata(base_url, token, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed creating metadata.")
	})

	It("Get Metadata Admin", func() {
		code, body := financials.GetMetadata(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting metadatas.")

		// Get the count of metadata
		metadataCount := gjson.Get(body, "metadata.#").Int()
		Expect(metadataCount).To(BeNumerically(">", 0), "No metadata found.")

		// Get the latest metadata id
		metadataId = gjson.Get(body, fmt.Sprintf("metadata.%d.id", metadataCount-1)).String()
		fmt.Println("metadataId", metadataId)
		Expect(metadataId).NotTo(BeEmpty(), "MetadataId value is nil or empty string.")
	})

	It("Update Metadata Admin", func() {
		payload := fmt.Sprintf(`{
		"context": "%s",
		"key": "test",
		"metadataSetId": 1,
		"type": "test updated",
		"value": "test"
		}`, "test-automation-updated")
		code, body := financials.UpdateMetadata(base_url, token, metadataId, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed updated metadata.")
	})

	It("Delete Metadata Admin", func() {
		code, body := financials.DeleteMetadata(base_url, token, metadataId)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed deleting metadata.")
	})
})

var _ = Describe("Product Catalog V2 Super Admin tests - ServiceType", func() {

	It("Get ServiceType Admin", func() {
		code, body := financials.GetServiceType(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting services.")
		serviceName = gjson.Get(body, "services.1.name").String()
		serviceLocation = gjson.Get(body, "services.1.location").String()
		fmt.Println("Service Name: ", serviceName)
		fmt.Println("Service sLocation: ", serviceLocation)
		Expect(serviceName).NotTo(BeEmpty(), "serviceName value is nil.")
		Expect(serviceLocation).NotTo(BeEmpty(), "serviceLocation value is nil.")
	})

	It("Get MetadataSet Admin", func() {
		code, body := financials.GetMetadataSet(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting metadata sets.")

		// Get the count of metadata sets
		metadataSetsCount := gjson.Get(body, "metadata_sets.#").Int()
		Expect(metadataSetsCount).To(BeNumerically(">", 0), "No metadata sets found.")

		// Get the latest metadata set id
		metadataSetId = gjson.Get(body, fmt.Sprintf("metadata_sets.%d.id", metadataSetsCount-1)).String()
		fmt.Println("metadataSetId", metadataSetId)
		Expect(metadataSetId).NotTo(BeEmpty(), "MetadataSetId value is nil or empty string.")
	})

	It("Create ServiceType Admin", func() {
		random_service_name := "test" + compute_utils.GetRandomString()
		random_service_type := "test" + compute_utils.GetRandomString()
		fmt.Println("metadataSetId", metadataSetId)
		payload := fmt.Sprintf(`{
			"name": "%s",
			"location": "%s",
			"type": "%s",
			"metadataSetId": %s
		}`, random_service_name, "test", random_service_type, metadataSetId)
		fmt.Print("Payload: ", payload)
		code, body := financials.CreateServiceType(base_url, token, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed creating services.")
		//{"name":"Bare Metal","location":"global","type":"ComputeAsAService","metadata_set_id":1,"created_at":"2025-01-22T18:59:05.711218Z","updated_at":"2025-01-22T18:59:15.805684Z"}
	})

	It("Get ServiceType Admin", func() {
		code, body := financials.GetServiceType(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting services.")

		// Get the count of services
		servicesCount := gjson.Get(body, "services.#").Int()
		Expect(servicesCount).To(BeNumerically(">", 0), "No services found.")

		// Get the latest service name and location
		serviceName = gjson.Get(body, fmt.Sprintf("services.%d.name", servicesCount-1)).String()
		serviceLocation = gjson.Get(body, fmt.Sprintf("services.%d.location", servicesCount-1)).String()
		fmt.Println("Service Name: ", serviceName)
		fmt.Println("Service Location: ", serviceLocation)
		Expect(serviceName).NotTo(BeEmpty(), "serviceName value is nil or empty string.")
		Expect(serviceLocation).NotTo(BeEmpty(), "serviceLocation value is nil or empty string.")
	})

	It("Update ServiceType Admin", func() {
		random_service_name := "test"
		payload := fmt.Sprintf(`{
			"type": "%s"
			}`, random_service_name)
		fmt.Print("Payload: ", payload)
		fmt.Print("Token: ", token)
		code, body := financials.UpdateServiceType(base_url, token, serviceName, serviceLocation, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed updating services.")
	})

	It("Delete ServiceType Admin", func() {
		fmt.Println("serviceLocation", serviceLocation)
		fmt.Println("serviceName", serviceName)
		code, body := financials.DeleteServiceType(base_url, token, serviceName, serviceLocation)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed deleting services.")
	})

})

var _ = Describe("Product Catalog V2 Super Admin tests - Products", func() {
	var regionId string
	var product_name string
	var regionName string

	It("Get Regions Admin", func() {
		code, body := financials.GetRegions(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting regions.")
		regionId = gjson.Get(body, "regions.0.id").String()
		Expect(regionId).NotTo(BeNil(), "RegionId value is nil.")
	})

	It("Get Products - User", func() {
		payload := fmt.Sprintf(`{
			"cloudaccountId": "%s",
			"regionId":"%s"
		}`, cloud_account_created, regionId)
		fmt.Println("Payload: ", payload)
		code, body := financials.GetProductsV2(base_url, token, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Or(Equal(200), Equal(404), Equal(500)), "Failed Getting user products.")
	})

	It("Get Products - Admin", func() {
		payload := fmt.Sprintf(`{
		}`)
		fmt.Println("Payload: ", payload)
		code, body := financials.GetProductsAdmin(base_url, token, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting admin products.")
	})

	It("Get Products API ", func() {
		payload := fmt.Sprintf(`{
				"cloudaccountId": "%s",
	  			"productFilter": {
				}
			}`, place_holder_map["cloud_account_id"])
		fmt.Println("payload", payload)
		fmt.Println("USER TOKEN", userToken)
		fmt.Println("Cloudaccount", place_holder_map["cloud_account_id"])
		fmt.Println("cloudaccountType", place_holder_map["cloud_account_type"])
		code, body := financials.GetProductsAPI(base_url, userToken, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting from products clustered API")
	})

	It("Get MetadataSet Admin", func() {
		code, body := financials.GetMetadataSet(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting metadata sets.")

		// Get the count of metadata sets
		metadataSetsCount := gjson.Get(body, "metadata_sets.#").Int()
		Expect(metadataSetsCount).To(BeNumerically(">", 0), "No metadata sets found.")

		// Get the latest metadata set id
		metadataSetId = gjson.Get(body, fmt.Sprintf("metadata_sets.%d.id", metadataSetsCount-1)).String()
		fmt.Println("metadataSetId", metadataSetId)
		Expect(metadataSetId).NotTo(BeEmpty(), "MetadataSetId value is nil or empty string.")
	})

	It("Get Rate sets Admin", func() {
		code, body := financials.GetRateSets(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting rate sets.")

		// Get the count of rate sets
		rateSetsCount := gjson.Get(body, "rate_sets.#").Int()
		Expect(rateSetsCount).To(BeNumerically(">", 0), "No rate sets found.")

		// Get the latest rate set id
		rateSetId = gjson.Get(body, fmt.Sprintf("rate_sets.%d.id", rateSetsCount-1)).String()
		fmt.Println("Latest rate set: ", rateSetId)
		Expect(rateSetId).NotTo(BeEmpty(), "rateSetId value is nil or empty string.")
	})

	It("Get Regions", func() {
		fmt.Println("USER TOKEN", token)
		fmt.Println("vendor", vendorName)
		code, body := financials.GetRegions(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting regions.")

		// Get the count of regions
		regionsCount := gjson.Get(body, "regions.#").Int()
		Expect(regionsCount).To(BeNumerically(">", 0), "No regions found.")

		regionName = gjson.Get(body, fmt.Sprintf("regions.%d.name", regionsCount-1)).String()
		fmt.Println("Region Name: ", regionName)
		Expect(regionName).NotTo(BeEmpty(), "regionName value is nil or empty string.")
	})

	It("Get ServiceType Admin", func() {
		code, body := financials.GetServiceType(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting services.")

		// Get the count of services
		servicesCount := gjson.Get(body, "services.#").Int()
		Expect(servicesCount).To(BeNumerically(">", 0), "No services found.")

		// Get the latest service name and location
		serviceName = gjson.Get(body, fmt.Sprintf("services.%d.name", servicesCount-1)).String()
		serviceLocation = gjson.Get(body, fmt.Sprintf("services.%d.location", servicesCount-1)).String()
		fmt.Println("Service Name: ", serviceName)
		fmt.Println("Service Location: ", serviceLocation)
		Expect(serviceName).NotTo(BeEmpty(), "serviceName value is nil or empty string.")
		Expect(serviceLocation).NotTo(BeEmpty(), "serviceLocation value is nil or empty string.")
	})

	It("Get Families Admin", func() {
		fmt.Println("USER TOKEN", token)
		fmt.Println("vendor", vendorName)
		code, body := financials.GetFamilies(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting families admin.")

		// Get the count of families
		familiesCount := gjson.Get(body, "product_families.#").Int()
		Expect(familiesCount).To(BeNumerically(">", 0), "No families found.")

		familyName = gjson.Get(body, fmt.Sprintf("product_families.%d.name", familiesCount-1)).String()
		fmt.Println("Family Name: ", familyName)
		Expect(familyName).NotTo(BeEmpty(), "familyName value is nil or empty string.")
	})

	It("Add product", func() {
		product_name := "test" + compute_utils.GetRandomString()
		fmt.Println("Family Name: ", familyName)
		payload := fmt.Sprintf(`{
			"metadataSetId": %s,
			"name": "%s",
			"productFamilyName": "%s",
			"rateSetId": %s,
			"regionName": "%s",
			"serviceName": "%s",
			"usage": "%s"
		}`, metadataSetId, product_name, familyName, rateSetId, regionName, serviceName, "dollarsPerMinute")
		fmt.Println("payload", payload)
		code, body := financials.AddProduct(base_url, token, payload)
		fmt.Println("Response: ", body)

		env := os.Getenv("IDC_ENV")
		if (env != "staging" && env != "qa1") && code == 500 {
			expectedMessage := `{"code":13,"message":"failed to create GTS plan, product saved with status error: rpc error: code = Internal desc = failed to create gts plan","details":[]}`
			if body == expectedMessage {
				fmt.Println("Test passed due to specific error message in non-staging and non-qa1 environment.")
				return
			}
		}

		Expect(code).To(Equal(200), "Failed adding product.")
	})

	It("Get Products - Admin", func() {
		payload := fmt.Sprintf(`{}`)
		fmt.Println("Payload: ", payload)
		code, body := financials.GetProductsAdmin(base_url, token, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting admin products.")

		// Get the count of products
		productsCount := gjson.Get(body, "products.#").Int()
		Expect(productsCount).To(BeNumerically(">", 0), "No products found.")

		// Retrieve productName and regionName
		product_name = gjson.Get(body, fmt.Sprintf("products.%d.name", productsCount-1)).String()
		regionName = gjson.Get(body, fmt.Sprintf("products.%d.metadata.region", productsCount-1)).String()

		fmt.Println("Product Name: ", product_name)
		fmt.Println("Region Name: ", regionName)

		Expect(product_name).NotTo(BeEmpty(), "productName value is nil or empty string.")
		Expect(regionName).NotTo(BeEmpty(), "regionName value is nil or empty string.")
	})

	It("Update product ", func() {
		payload := fmt.Sprintf(`{
			"metadataSetId": %s,
			"productFamilyName": "%s",
			"rateSetId": %s,
			"serviceName": "%s"
		}`, "1", familyName, rateSetId, serviceName)
		fmt.Println("payload", payload)
		code, body := financials.UpdateProduct(base_url, token, payload, product_name, regionName)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed updating product")
	})

	It("Delete product ", func() {
		code, body := financials.DeleteProduct(base_url, token, product_name, regionName)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed deleting product")
	})

	/*It("create new product - Admin", func() {
		code, body := financials.GetFamilies(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting families.")
		familyId := gjson.Get(body, "families.0.id").String()
		//familyName := "automation-family-updated" + compute_utils.GetRandomString()
		vendorName := "automation-vendor-updated" + compute_utils.GetRandomString()
		product_id := "test-automation-product-1"
		product_payload := fmt.Sprintf(`{
			    "product": {
			    "accountTypesBlocked": "",
			    "accountWhitelist": false,
			    "billingEnable": true,
			    "description": "automation-test-updated",
			    "displayCategory": "automation-test-updated",
			    "displayDesc": "automation-test-updated",
			    "displayHighlight": "automation-test-updated",
			    "displayInfo": "automation-test-updated",
			    "displayName": "automation-test-updated",
			    "eccn": "",
			    "enabled": true,
			    "familyId": "` + familyId + `",
			    "id": "` + product_id + `",
			    "instanceCategory": "singlenode",
			    "instanceMode": "",
			    "instanceType": "bm-spr-pvc-1550-8",
			    "matchExpr": "",
			    "metadata": "",
			    "name": "automation-test",
			    "pcq": "19513",
			    "releaseStatus": "",
			    "serviceName": "Virtual Machine",
			    "serviceType": "Virtual Machine",
			    "status": "",
			    "usecase": "GPU",
			    "vendorId": "` + vendorName + `"
			  },
			  "rates": [
			    {
			      "accountType": "ACCOUNT_TYPE_PREMIUM",
			      "productId": "bc41bc72-c59f-4447-87e0-6b0ed6478790",
			      "rate": "0.00",
			      "regionId": "global",
			      "unit": "per Minute",
			      "usageExpr": "time – previous.time"
			    }
	  		]
		}`)
		fmt.Println("Payload: ", product_payload)
		code, body = financials.CreateChangeRequest(base_url, token, product_payload)
		Expect(code).To(Equal(200), "Failed create a Change Request.")
		code, body = financials.EnableChangeRequest(base_url, token, product_id)
		Expect(code).To(Equal(200), "Failed enable a Change Request.")
		code, body = financials.ApproveChangeRequest(base_url, token, product_id)
		Expect(code).To(Equal(200), "Approve a Change Request.")
		Expect(code).To(Equal(200), "Failed create admin product.")
	})

	It("Update Existing Product - Admin", func() {
		code, body := financials.GetFamilies(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting families.")
		familyId := gjson.Get(body, "families.0.id").String()
		//familyName := "automation-family-updated" + compute_utils.GetRandomString()
		vendorName := "automation-vendor-updated" + compute_utils.GetRandomString()
		product_id := "test-automation-product-1"
		//Same process, If we use an existing product Id, it will change the information.
		product_payload := fmt.Sprintf(`{
			    "product": {
			    "accountTypesBlocked": "",
			    "accountWhitelist": false,
			    "billingEnable": true,
			    "description": "automation-test-updated",
			    "displayCategory": "automation-test-updated",
			    "displayDesc": "automation-test-updated",
			    "displayHighlight": "automation-test-updated",
			    "displayInfo": "automation-test-updated",
			    "displayName": "automation-test-updated",
			    "eccn": "",
			    "enabled": true,
			    "familyId": "` + familyId + `",
			    "id": "` + product_id + `",
			    "instanceCategory": "singlenode",
			    "instanceMode": "",
			    "instanceType": "bm-spr-pvc-1550-8",
			    "matchExpr": "",
			    "metadata": "",
			    "name": "automation-test",
			    "pcq": "19513",
			    "releaseStatus": "",
			    "serviceName": "Virtual Machine",
			    "serviceType": "Virtual Machine",
			    "status": "",
			    "usecase": "GPU",
			    "vendorId": "` + vendorName + `"
			  },
			  "rates": [
			    {
			      "accountType": "ACCOUNT_TYPE_PREMIUM",
			      "productId": "bc41bc72-c59f-4447-87e0-6b0ed6478790",
			      "rate": "0.00",
			      "regionId": "global",
			      "unit": "per Minute",
			      "usageExpr": "time – previous.time"
			    }
	  		]
		}`)
		fmt.Println("Payload: ", product_payload)
		code, body = financials.CreateChangeRequest(base_url, token, product_payload)
		Expect(code).To(Equal(200), "Failed create a Change Request.")
		code, body = financials.EnableChangeRequest(base_url, token, product_id)
		Expect(code).To(Equal(200), "Failed enable a Change Request.")
		code, body = financials.ApproveChangeRequest(base_url, token, product_id)
		Expect(code).To(Equal(200), "Failed to approve a Change Request.")
		Expect(code).To(Equal(200), "Failed to update admin product.")
	})

	It("Update Region Product Pricing - Admin", func() {
		code, body := financials.GetFamilies(base_url, token)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting families.")
		familyId := gjson.Get(body, "families.0.id").String()
		//familyName := "automation-family-updated" + compute_utils.GetRandomString()
		vendorName := "automation-vendor-updated" + compute_utils.GetRandomString()
		product_id := "test-automation-product-1"
		//Same process, If we use an existing product Id, it will change the information.
		product_payload := fmt.Sprintf(`{
			    "product": {
			    "accountTypesBlocked": "",
			    "accountWhitelist": false,
			    "billingEnable": true,
			    "description": "automation-test-updated",
			    "displayCategory": "automation-test-updated",
			    "displayDesc": "automation-test-updated",
			    "displayHighlight": "automation-test-updated",
			    "displayInfo": "automation-test-updated",
			    "displayName": "automation-test-updated",
			    "eccn": "",
			    "enabled": true,
			    "familyId": "` + familyId + `",
			    "id": "` + product_id + `",
			    "instanceCategory": "singlenode",
			    "instanceMode": "",
			    "instanceType": "bm-spr-pvc-1550-8",
			    "matchExpr": "",
			    "metadata": "",
			    "name": "automation-test",
			    "pcq": "19513",
			    "releaseStatus": "",
			    "serviceName": "Virtual Machine",
			    "serviceType": "Virtual Machine",
			    "status": "",
			    "usecase": "GPU",
			    "vendorId": "` + vendorName + `"
			  },
			  "rates": [
			    {
			      "accountType": "ACCOUNT_TYPE_PREMIUM",
			      "productId": "bc41bc72-c59f-4447-87e0-6b0ed6478790",
			      "rate": "100000.00",
			      "regionId": "global-test",
			      "unit": "per Minute",
			      "usageExpr": "time – previous.time"
			    }
	  		]
		}`)
		fmt.Println("Payload: ", product_payload)
		code, body = financials.CreateChangeRequest(base_url, token, product_payload)
		Expect(code).To(Equal(200), "Failed create a Change Request.")
		code, body = financials.EnableChangeRequest(base_url, token, product_id)
		Expect(code).To(Equal(200), "Failed enable a Change Request.")
		code, body = financials.ApproveChangeRequest(base_url, token, product_id)
		Expect(code).To(Equal(200), "Failed to approve a Change Request.")
		Expect(code).To(Equal(200), "Failed update admin product.")
		Expect(code).To(Equal(200), "Failed update admin rate in product.")
	}) */

	// This TC needs to be fixed
	/* It("Get existing interest for any product - Admin", func() {
		code, _ := financials.GetInterests(base_url, token, cloud_account_created, userNameSU, cloudaccount_email, product_id, region_id)
		Expect(code).To(Equal(200), "Failed to get existing interest for any product.")
	})

	// There could be only 1 product for each filter that was created and updated.
	It("Use Region for whitelisting", func() {
		payload := fmt.Sprintf(`{
			"region": "global-test"
		}`)
		fmt.Println("Payload: ", payload)
		code, body := financials.GetProductsAdmin(base_url, token, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting admin products.")
		products := gjson.Get(body, "products").String()
		Expect(products).NotTo(Equal("[]"), "Failed filtering admin products.")
	})
	It("Use Product for whitelisting", func() {
		payload := fmt.Sprintf(`{
			"name": "automation-test"
		}`)
		fmt.Println("Payload: ", payload)
		code, body := financials.GetProductsAdmin(base_url, token, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting admin products.")
		products := gjson.Get(body, "products").String()
		Expect(products).NotTo(Equal("[]"), "Failed filtering admin products.")
	})
	It("Use Account type for  whitelisting", func() {
		payload := fmt.Sprintf(`{
			"accountType": "ACCOUNT_TYPE_PREMIUM",
		}`)
		fmt.Println("Payload: ", payload)
		code, body := financials.GetProductsAdmin(base_url, token, payload)
		fmt.Println("Response: ", body)
		Expect(code).To(Equal(200), "Failed getting admin products.")
		products := gjson.Get(body, "products").String()
		Expect(products).NotTo(Equal("[]"), "Failed filtering admin products.")
	})*/
})
