package OPA_Security_Checks_test

import (
	"fmt"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Member Invitation flows test", Ordered, Label("MultiUser-OPA-Checks"), func() {
	It("Generate OTP code for invite", func() {
		fmt.Println("Creating OTP from endpoint...")
		otp_url := base_url + "/v1/otp/create"
		payload := fmt.Sprintf(`{
			"cloudAccountId": "%s",
			"memberEmail": "%s"
		}`, place_holder_map["cloud_account_id"], userNameSU)
		fmt.Println("Payload: ", payload)
		code, body := financials.CreateOTP(otp_url, userToken, payload)
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		Expect(code).To(Equal(200), "Failed to create OTP")
	})

	// Refactor usernames and tokens to diferentiate users
	It("Resend OTP code for invite", func() {
		fmt.Println("Resend OTP test...")
		otp_url := base_url + "/v1/otp/resend"
		payload := fmt.Sprintf(`{
			"cloudAccountId": "%s",
			"memberEmail": "%s"
		}`, place_holder_map["cloud_account_id"], userNameSU)
		fmt.Println("Payload: ", payload)
		code, body := financials.ResendOTP(otp_url, userToken, payload)
		fmt.Println("Payload: ", payload)
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		Expect(code).To(Equal(200), "Failed to resend OTP code")
	})

	It("Verify OTP code for invite", func() {
		fmt.Println("Verify OTP...")
		otp_url := base_url
		OTPCode := financials.GetMailFromMailInbox(inboxIdPremium, `([0-9]{6})<\/h3>`, apiKey)
		payload := fmt.Sprintf(`{
			"cloudAccountId": "%s",
			"memberEmail": "%s",
			"otpCode": "%s"
		}`, place_holder_map["cloud_account_id"], userNameSU, OTPCode)
		Expect(OTPCode).To(Not(BeNil()), "OTP code is nil.")
		fmt.Println("Payload: ", payload)
		code, body := financials.VerifyOTP(otp_url, userToken, payload)
		Expect(code).NotTo(BeNil(), "OTP Code cannot be nil.")
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		Expect(code).To(Or(Equal(200), Equal(404)), "Failed to verify OTP code.")
	})

	//Validate Invite code member

	var inviteCode string
	It("Generates and send Member invitation", func() {
		fmt.Println("Creating invitations...")
		fmt.Println(place_holder_map["cloud_account_id"])
		invite_url := base_url + "/v1/cloudaccounts/invitations/create"
		code, body := financials.CreateInviteCode(invite_url, userToken, place_holder_map["cloud_account_id"], userNameSU)
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		Expect(code).To(Or(Equal(200)), "Failed to create invite...")
	})

	It("Verify OTP code for invite", func() {
		fmt.Println("Verify OTP...")
		time.Sleep(10 ^ time.Second)
		pattern := "([0-9]{6})</h3>"
		OTPCode := financials.GetMailFromMailInbox(inboxIdPremium, pattern, apiKey)
		Expect(OTPCode).NotTo(BeNil(), "OTP Code cannot be nil.")
	})

	It("Sends invite code to member", func() {
		fmt.Println("Sending invite code to member...")
		fmt.Println(place_holder_map["cloud_account_id"])
		payload := fmt.Sprintf(`{
			"adminAccountId": "%s",
			"memberEmail": "%s"
		}`, place_holder_map["cloud_account_id"], userNameSU)
		fmt.Println("Payload: ", payload)
		code, body := financials.SendInviteCode(base_url, userTokenSU, payload)
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		Expect(code).To(Equal(200), "Failed to send invite code")
	})

	It("Retrieve OTP Code for Invited Member (Standard User)", func() {
		fmt.Println("Retrieving invite code from MailSlurp...")
		time.Sleep(20 * time.Second)
		inviteCode = financials.GetMailFromMailInbox(inboxIdStandard, "([0-9]{8})", apiKey)
		Expect(inviteCode).To(Not(BeNil()), "Failed to create invite, OTP code not found")
	})

	It("Verify OTP for Member and join to group", func() {
		fmt.Println("Verifying invite code...")
		payload := fmt.Sprintf(`{
				"adminCloudAccountId": "%s",
				"inviteCode": "%s",
				"memberEmail": "%s"
			}`, place_holder_map["cloud_account_id"], inviteCode, userNameSU)
		fmt.Println("Payload: ", payload)
		code, body := financials.VerifyInviteCode(base_url, userTokenSU, payload)
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		Expect(code).To(Or(Equal(200)), "Failed to verify member invite code.")
	})

	It("Verify OTP for Member and join to group for second time", func() {
		fmt.Println("Verifying invite code...")
		payload := fmt.Sprintf(`{
				"adminCloudAccountId": "%s",
				"inviteCode": "%s",
				"memberEmail": "%s"
			}`, place_holder_map["cloud_account_id"], inviteCode, userNameSU)
		fmt.Println("Payload: ", payload)
		code, body := financials.VerifyInviteCode(base_url, userTokenSU, payload)
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		Expect(code).To(Not(Equal(200)), "User should not use twice the invitation code.")
	})
})

var _ = Describe("Validate Member / Admin access", Ordered, Label("MultiUser-OPA-Checks"), func() {
	It("Validate Member cannot access admin billing information", func() {
		usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + place_holder_map["cloud_account_id"]
		var usage_response_status int
		var usage_response_body string

		usage_response_status, usage_response_body = financials.GetUsage(usage_url, userTokenSU)
		fmt.Println("usage_response_body", usage_response_body)
		fmt.Println("usage_response_status", usage_response_status)
		Expect(usage_response_status).To(Equal(403), "Failed to validate usage_response_status")
	})
	It("Validate Admin can access admin billing information", func() {
		usage_url := base_url + "/v1/billing/usages?cloudAccountId=" + place_holder_map["cloud_account_id"]
		var usage_response_status int
		var usage_response_body string

		usage_response_status, usage_response_body = financials.GetUsage(usage_url, userToken)
		fmt.Println("usage_response_body", usage_response_body)
		fmt.Println("usage_response_status", usage_response_status)
		Expect(usage_response_status).To(Equal(200), "Failed to validate usage_response_status")
	})
	It("Validate Admin cannot access member instances", func() {
		instance_endpoint := compute_url + "/v1/cloudaccounts/" + place_holder_map_su["cloud_account_id"] + "/instances"

		delete_response_status, delete_response_body := frisby.GetAllInstance(instance_endpoint, userToken)
		fmt.Println("delete_response_body", delete_response_body)
		fmt.Println("delete_response_status", delete_response_status)
		Expect(delete_response_status).To(Equal(403), "Failed to validate usage_response_status")
	})
})
