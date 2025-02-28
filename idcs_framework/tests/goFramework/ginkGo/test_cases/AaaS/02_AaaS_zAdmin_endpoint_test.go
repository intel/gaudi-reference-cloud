package AaaS_test

import (
	"fmt"
	"goFramework/framework/service_api/financials"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AaaSAdminEndpoints", func() {
	It("Validate Ping of Authorization as a Service", func() {
		code, _ := financials.Ping(base_url, token)
		Expect(code).To(Or(Equal(200), Equal(403)), "Ping is not returning 200 or 403") // It could return 403 if not whitelisted
	})

	It("Assign Role", func() {
		payload := fmt.Sprintf(`{
			"cloudAccountId": "%s",
			"subject": "%s",
			"systemRole": "%s"
			}`, place_holder_map["cloud_account_id"], userNameSU, "cloud_account_member")
		fmt.Println("PAYLOAD", payload)
		code, body := financials.AssignRole(base_url, token, payload)
		fmt.Println("Endpoint Response: " + body)
		Expect(code).To(Or(Equal(200), Equal(409), Equal(500)), "Failed assigning role.")
	})

	It("Unassign Role", func() {
		payload := fmt.Sprintf(`{
			"cloudAccountId": "%s",
			"subject": "%s",
			"systemRole": "%s"
			}`, place_holder_map["cloud_account_id"], userNameSU, "cloud_account_member")
		fmt.Println("PAYLOAD", payload)
		code, body := financials.UnAssignRole(base_url, token, payload)
		fmt.Println("Endpoint Response: " + body)
		Expect(code).To(Equal(200), "Failed unassigning role.")
	})

	It("Get User Role Premium", func() {
		fmt.Println("Map", place_holder_map)
		fmt.Println("Map2", place_holder_map_su)
		code, body := financials.GetUserRoles(base_url, userToken, place_holder_map["cloud_account_id"])
		fmt.Println("Endpoint Response: " + body)
		Expect(code).To(Equal(200), "Failed to retrieve user roles.")
	})

	It("Get User Role Standard", func() {
		fmt.Println("Map", place_holder_map)
		fmt.Println("Map2", place_holder_map_su)
		code, body := financials.GetUserRoles(base_url, userTokenSU, place_holder_map_su["cloud_account_id"])
		fmt.Println("Endpoint Response: " + body)
		Expect(code).To(Equal(200), "Failed to retrieve user roles.")
	})

	It("Lookup Role", func() {
		payload := fmt.Sprintf(`{
			"action": "update",
			"cloudAccountId": "%s",
			"resourceIds": [
				"%s"
			],
			"resourceType": "instance"
		}`, place_holder_map["cloud_account_id"], "123445")
		fmt.Println("PAYLOAD", payload)
		code, body := financials.LookUp(base_url, token, payload)
		fmt.Println("Endpoint Response: " + body)
		Expect(code).To(Or(Equal(200), Equal(403), Equal(404)), "Failed to look up role")
	})

	It("Check Role access", func() {
		payload := fmt.Sprintf(`{
			"cloudAccountId": "%s",
			"path": "/v1/authorization/resources",
			"payload": {},
			"verb": "GET"
		}`, place_holder_map["cloud_account_id"])
		fmt.Println("PAYLOAD", payload)
		code, body := financials.Check(base_url, userNameSU, payload)
		fmt.Println("Endpoint Response: " + body)
		Expect(code).To(Or(Equal(200), Equal(403), Equal(404)), "Failed to look up role")
	})
})
