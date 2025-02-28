package OTP_flows_test

import (
	"fmt"
	"goFramework/framework/library/auth"
	"goFramework/framework/service_api/financials"
	"goFramework/testsetup"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

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
		Expect(code).To(Or(Equal(200), Equal(404), Equal(500)), "Failed to send invite code")
	})

	It("Resend Invitations", func() {
		fmt.Println("Resend Invitations...")
		fmt.Println("Wait for invitation to be expired.")
		payload := fmt.Sprintf(`{
			"adminAccountId": "%s",
			"memberEmail": "%s"
		}`, place_holder_map["cloud_account_id"], userNameSU)
		fmt.Println("Payload: ", payload)
		resend_url := base_url + "/v1/cloudaccounts/invitations/resend"
		code, body := financials.Resendinvitation(resend_url, userToken, payload)
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		Expect(code).NotTo(Equal(200), "Expired invite should not be sent.")
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

	It("Generates and send Member invitation", func() {
		fmt.Println("Generates and send Member invitation")
		invite_url := base_url + "/v1/cloudaccounts/invitations/create"
		code, body := financials.CreateExpiredInviteCode(invite_url, userToken, place_holder_map["cloud_account_id"], userNameSU)
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		Expect(code).To(Or(Equal(400)), "Failed to create invite")
	})
})

var _ = Describe("Member Invitation flows test - Resend invite for deleted user", Ordered, Label("Multi-User-e2e"), func() {
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
		Expect(code).To(Or(Equal(200), Equal(404)), "Failed to send invite code")
	})

	It("Resend Invitations", func() {
		fmt.Println("Deleting Member CloudAccount...")
		url := base_url + "/v1/cloudaccounts"
		userNameSU = auth.Get_UserName("Enterprise")
		cloudAccId, err := testsetup.GetCloudAccountId(userNameSU, base_url, token)
		if err == nil {
			financials.DeleteCloudAccountById(url, token, cloudAccId)
		}
		fmt.Println("Resend Invitations...")
		payload := fmt.Sprintf(`{
			"adminAccountId": "%s",
			"memberEmail": "%s"
		}`, place_holder_map["cloud_account_id"], userNameSU)
		fmt.Println("Payload: ", payload)
		resend_url := base_url + "/v1/cloudaccounts/invitations/resend"
		code, body := financials.Resendinvitation(resend_url, userToken, payload)
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		Expect(code).To(Or(Equal(200), Equal(404)), "Failed to send invite code")
	})
})

var _ = Describe("Member Invitation flows test - Resend expired invite for deleted user", Ordered, Label("Multi-User-e2e"), func() {
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
		Expect(code).To(Or(Equal(200), Equal(400)), "Failed to create invite")
	})

	It("Resend Invitations", func() {
		fmt.Println("Resend Invitations...")
		fmt.Println("Wait for invitation to be expired.")
		payload := fmt.Sprintf(`{
			"adminAccountId": "%s",
			"memberEmail": "%s"
		}`, place_holder_map["cloud_account_id"], userNameSU)
		fmt.Println("Payload: ", payload)
		resend_url := base_url + "/v1/cloudaccounts/invitations/resend"
		code, body := financials.Resendinvitation(resend_url, userToken, payload)
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		Expect(code).NotTo(Equal(200), "Expired invite should not be sent.")
	})
})

// Verify OTP and invite code cannot be redeemed twice

var _ = Describe("Member Invitation flows test - Admin cannot Invite self", Ordered, Label("Multi-User-e2e"), func() {
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
		}`, place_holder_map["cloud_account_id"], userName)
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
		}`, place_holder_map["cloud_account_id"], userName, OTPCode)
		Expect(OTPCode).To(Not(BeNil()), "OTP code is nil.")
		fmt.Println("Payload: ", payload)
		code, body := financials.VerifyOTP(otp_url, userToken, payload)
		Expect(code).NotTo(BeNil(), "OTP Code cannot be nil.")
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		Expect(code).To(Or(Equal(200), Equal(404)), "Failed to verify OTP code.")
	})

	It("Verify OTP code for invite", func() {
		fmt.Println("Verify OTP...")
		otp_url := base_url
		OTPCode := financials.GetMailFromMailInbox(inboxIdPremium, `([0-9]{6})<\/h3>`, apiKey)
		payload := fmt.Sprintf(`{
			"cloudAccountId": "%s",
			"memberEmail": "%s",
			"otpCode": "%s"
		}`, place_holder_map["cloud_account_id"], userName, OTPCode)
		Expect(OTPCode).To(Not(BeNil()), "OTP code is nil.")
		fmt.Println("Payload: ", payload)
		code, body := financials.VerifyOTP(otp_url, userToken, payload)
		Expect(code).NotTo(BeNil(), "OTP Code cannot be nil.")
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		validated := gjson.Get(body, "validated").Bool()
		otpState := gjson.Get(body, "otpState").String()
		Expect(validated).To(BeFalse(), "Expected validated to be false.")
		Expect(otpState).To(Equal("OTP_STATE_PENDING"), "Expected otpState to be OTP_STATE_PENDING.")
	})
})

var _ = Describe("Member Invitation flows test - Admin cannot Invite self", Ordered, Label("Multi-User-e2e"), func() {
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
		}`, place_holder_map["cloud_account_id"], userName)
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
		}`, place_holder_map["cloud_account_id"], userName, OTPCode)
		Expect(OTPCode).To(Not(BeNil()), "OTP code is nil.")
		fmt.Println("Payload: ", payload)
		code, body := financials.VerifyOTP(otp_url, userToken, payload)
		Expect(code).NotTo(BeNil(), "OTP Code cannot be nil.")
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		Expect(code).To(Or(Equal(200), Equal(404)), "Failed to verify OTP code.")
	})

	It("Sends invite code to member", func() {
		fmt.Println("Sending invite code to member...")
		fmt.Println(place_holder_map["cloud_account_id"])
		payload := fmt.Sprintf(`{
			"adminAccountId": "%s",
			"memberEmail": "%s"
		}`, place_holder_map["cloud_account_id"], userName)
		fmt.Println("Payload: ", payload)
		code, body := financials.SendInviteCode(base_url, userToken, payload)
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		Expect(code).To(Not(Equal(200)), "Admin user should not invite self")
	})
})

var _ = Describe("Member Invitation flows test - Wrong otp retries for 3 times check blocked value in response", Ordered, Label("Multi-User-e2e"), func() {
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

	It("Verify OTP API for Blocked value at 3rd try of wrong OTP ", func() {
		fmt.Println("Verify OTP with wrong code...")

		otp_url := base_url
		OTPCode := "12345"
		retries := 2

		for i := 0; i < retries; i++ {
			payload := fmt.Sprintf(`{
				"cloudAccountId": "%s",
				"memberEmail": "%s",
				"otpCode": "%s"
			}`, place_holder_map["cloud_account_id"], userNameSU, OTPCode)

			fmt.Printf("Attempt #%d with payload: %s\n", i+1, payload)

			code, body := financials.VerifyOTP(otp_url, userToken, payload)

			fmt.Println("Response code: ", code)
			fmt.Println("Response body: ", body)

			Expect(code).NotTo(BeNil(), "OTP Code cannot be nil.")
			Expect(code).To(Equal(200), "Expected a 200 response for OTP verification.")

			if i == retries-1 { // On 3rd failed attempt
				validated := gjson.Get(body, "validated").Bool()
				blocked := gjson.Get(body, "blocked").Bool()
				otpState := gjson.Get(body, "otpState").String()
				retryLeft := gjson.Get(body, "retryLeft").Int()
				message := gjson.Get(body, "message").String()
				Expect(validated).To(BeFalse(), "validated to be false.")
				Expect(blocked).To(Equal(true), "blocked to be true after 3 retries.")
				Expect(otpState).To(Equal("OTP_STATE_PENDING"), "Expected otpState to be OTP_STATE_PENDING.")
				Expect(retryLeft).To(Equal(0), "Expected retryLeft to be 0")
				Expect(message).To(ContainSubstring("max verification limit reached, retry after 1 minute"), "Expected the max verification limit message.")
			}
		}
	})
})
