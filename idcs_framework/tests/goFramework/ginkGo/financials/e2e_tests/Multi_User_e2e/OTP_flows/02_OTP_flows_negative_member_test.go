package OTP_flows_test

import (
	"fmt"
	"goFramework/framework/service_api/financials"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("Member Invitation flows test - Negative Invalid invite code", Ordered, Label("Multi-User-e2e"), func() {
	// Revoke Invite before execution
	It("Generate OTP code for invite", func() {
		fmt.Println("base_url: ", base_url)
		code, body := financials.RemoveInvitation(base_url, userToken, place_holder_map["cloud_account_id"], userNameSU)
		Expect(code).To(Or(Equal(200), Equal(404)), "Failed to reject invite."+body)
	})

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

	It("Generates and send Member invitation", func() {
		fmt.Println("Generates and send Member invitation")
		invite_url := base_url + "/v1/cloudaccounts/invitations/create"
		code, body := financials.CreateExpiredInviteCode(invite_url, userToken, place_holder_map["cloud_account_id"], userNameSU)
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		Expect(code).To(Or(Equal(400)), "Failed to create invite")
	})

	It("Sends invite code to member", func() {
		fmt.Println("Sends invite code to member")
		payload := fmt.Sprintf(`{
			"adminAccountId": "%s",
			"memberEmail": "%s"
		}`, place_holder_map["cloud_account_id"], userNameSU)
		fmt.Println("Payload: ", payload)
		code, body := financials.SendInviteCode(base_url, userTokenSU, payload)
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		Expect(code).To(Or(Equal(200), Equal(404)), "Failed to send invite code")
	})

	It("Verify OTP for Member and join to group", func() {
		inviteCode := "12345"
		payload := fmt.Sprintf(`{
			"adminCloudAccountId": "%s",
			"inviteCode": "%s",
			"memberEmail": "%s"
		}`, place_holder_map["cloud_account_id"], inviteCode, userNameSU)
		fmt.Println("Payload: ", payload)
		code, body := financials.VerifyInviteCode(base_url, userTokenSU, payload)
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		validated := gjson.Get(body, "valid").Bool()
		otpState := gjson.Get(body, "invitationState").String()
		Expect(validated).To(BeFalse(), "Expected validated to be false.")
		Expect(otpState).To(Equal("INVITE_STATE_PENDING_ACCEPT"), "Expected invite status to be INVITE_STATE_PENDING_ACCEPT.")
	})
})

var _ = Describe("Member Invitation flows test - Negative Expired Invite", Ordered, Label("Multi-User-e2e"), func() {
	// Revoke Invite before execution
	It("Generate OTP code for invite", func() {
		fmt.Println("base_url: ", base_url)
		code, body := financials.RemoveInvitation(base_url, userToken, place_holder_map["cloud_account_id"], userNameSU)
		Expect(code).To(Or(Equal(200), Equal(404)), "Failed to reject invite."+body)
	})

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
})
