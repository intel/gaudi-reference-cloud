package OTP_flows_test

import (
	"fmt"
	"goFramework/framework/service_api/financials"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Member Invitation flows test - Member rejects Invite", Ordered, Label("Multi-User-e2e"), func() {
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
		fmt.Println("Creating invitations...")
		fmt.Println(place_holder_map["cloud_account_id"])
		invite_url := base_url + "/v1/cloudaccounts/invitations/create"
		code, body := financials.CreateInviteCode(invite_url, userToken, place_holder_map["cloud_account_id"], userNameSU)
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		Expect(code).To(Or(Equal(200)), "Failed to create invite...")
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

	It("Member Rejects Invitation", func() {
		fmt.Println("Member Rejects invitation...")
		fmt.Println("base_url: ", base_url)
		code, _ := financials.Rejectinvitation(base_url, userTokenSU, place_holder_map["cloud_account_id"], "INVITE_STATE_PENDING_ACCEPT", userNameSU)
		Expect(code).To(Equal(200), "Failed to reject invite.")
	})
})

var _ = Describe("Member Invitation flows test - Admin revokes invite", Ordered, Label("Multi-User-e2e"), func() {
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

	It("Admin Revokes invitation", func() {
		fmt.Println("Admin Revokes invitation...")
		fmt.Println("base_url: ", base_url)
		code, _ := financials.RevokeInvitation(base_url, userToken, place_holder_map["cloud_account_id"], userNameSU)
		Expect(code).To(Or(Equal(200), Equal(404)), "Failed to revoke invite.")
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

var _ = Describe("Member Invitation flows test - Invite Member with no cloudaccount", Ordered, Label("Multi-User-e2e"), func() {
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
		}`, place_holder_map["cloud_account_id"], "testemail@test.com")
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
		}`, place_holder_map["cloud_account_id"], "testemail@test.com", OTPCode)
		Expect(OTPCode).To(Not(BeNil()), "OTP code is nil.")
		fmt.Println("Payload: ", payload)
		code, body := financials.VerifyOTP(otp_url, userToken, payload)
		Expect(code).NotTo(BeNil(), "OTP Code cannot be nil.")
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		Expect(code).To(Or(Equal(200), Equal(404)), "Failed to verify OTP code.")
	})
})

var _ = Describe("Member Invitation flows test - Positive", Ordered, Label("Multi-User-e2e"), func() {
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

	It("Resend Invitations", func() {
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
		Expect(code).NotTo(Equal(200), "Admin User should not resend accepted invitations.")
	})

	It("Member Should not be able to create and send invites", func() {
		fmt.Println(place_holder_map["cloud_account_id"])
		invite_url := base_url + "/v1/cloudaccounts/invitations/create"
		code, body := financials.CreateInviteCode(invite_url, userTokenSU, place_holder_map_su["cloud_account_id"], userName)
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		Expect(code).NotTo(Or(Equal(200)), "Members should not create and send invites.")
	})

	It("Admin Removes member...", func() {
		fmt.Println("base_url: ", base_url)
		code, body := financials.RemoveInvitation(base_url, userToken, place_holder_map["cloud_account_id"], userNameSU)
		Expect(code).To(Or(Equal(200), Equal(404)), "Failed to reject invite."+body)
	})

	It("Read Admin invitations", func() {
		fmt.Println("Reading Admin invitations...s")
		code, body := financials.ReadInvitations(base_url, userToken, place_holder_map["cloud_account_id"], "")
		fmt.Println("Response code: ", code)
		fmt.Println("Response body: ", body)
		Expect(code).To(Or(Equal(200), Equal(404)), "Failed to read admin invites.")
	})

})

/*

var _ = Describe("Multi User - Member User Instance Launch Without credits", Ordered, Label("Multi-User-e2e"), func() {
	var vnet_created string
	var ssh_publickey_name_created string
	var create_response_status int
	var create_response_body string
	var instance_id_created string
	//var meta_data_map = make(map[string]string)
	//var resourceInfo testsetup.ResourcesInfo

	cloud_account_created = place_holder_map_su["cloud_account_id"]

	//CREATE VNET
	It("Create vnet with name", func() {
		// fmt.Println("Starting the VNet Creation via API...")
		// // form the endpoint and payload
		vnet_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/vnets"
		vnet_name := compute_utils.GetVnetName()
		vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
		// hit the api
		fmt.Println("Vnet end point ", vnet_endpoint)
		vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, userTokenSU, vnet_payload)
		vnet_created = gjson.Get(vnet_creation_body, "metadata.name").String()
		Expect(vnet_creation_status).To(Equal(200), "Failed to create VNet")
		Expect(vnet_name).To(Equal(vnet_created), "Failed to create Vnet, response validation failed")
	})

	// CREATE SSH KEY
	It("Create ssh public key with name", func() {
		logger.Logf.Info("Starting the SSH-Public-Key Creation via API...")
		// form the endpoint and payload
		ssh_publickey_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/" + "sshpublickeys"
		fmt.Println("SSH key is" + sshPublicKey)
		sshkey_name := "autossh-" + utils.GenerateSSHKeyName(4)
		fmt.Println("SSH  end point ", ssh_publickey_endpoint)
		ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)
		// hit the api
		sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, userTokenSU, ssh_publickey_payload)
		Expect(sshkey_creation_status).To(Equal(200), "Failed to create SSH Public key")
		ssh_publickey_name_created = gjson.Get(sshkey_creation_body, "metadata.name").String()
		Expect(sshkey_name).To(Equal(ssh_publickey_name_created), "Failed to create SSH Public key, response validation failed")
	})

	// LAUNCH INSTANCE WITHOUT CREDITS
	It("LAUNCH INSTANCE WITHOUT CREDITS - Admin", func() {
		logger.Logf.Info("Starting the Instance Creation via API...")
		// form the endpoint and payload
		instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		bm_name_iu := "autobm-" + utils.GenerateSSHKeyName(4)
		instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), bm_name_iu, "vm-spr-sml", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
		fmt.Println("instance_payload", instance_payload)
		place_holder_map["instance_type"] = "vm-spr-sml"
		// hit the api
		create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, userTokenSU, instance_payload)
		Expect(create_response_status).To(Equal(403), "Expected response code on paid instance for su user should be 403 when cloud credits are zero")
		Expect(strings.Contains(create_response_body, `"message":"paid service not allowed"`)).To(BeTrue(), "Failed to validate ")
	})

	// ADD CREDITS TO ACCOUNT
	It("Create Cloud credits for Standard user by redeeming coupons", func() {
		fmt.Println("Starting Cloud Credits Creation...")
		create_coupon_endpoint := base_url + "/v1/cloudcredits/coupons"
		coupon_payload := financials_utils.EnrichCreateCouponPayload(financials_utils.GetCreateCouponPayloadStandard(), 100, "idc_billing@intel.com", 1)
		fmt.Println("Payload", coupon_payload)
		coupon_creation_status, coupon_creation_body := financials.CreateCoupon(create_coupon_endpoint, token, coupon_payload)
		Expect(coupon_creation_status).To(Equal(200), "Failed to create coupon")

		// Redeem coupon
		redeem_coupon_endpoint := base_url + "/v1/cloudcredits/coupons/redeem"
		redeem_payload := financials_utils.EnrichRedeemCouponPayload(financials_utils.GetRedeemCouponPayload(), gjson.Get(coupon_creation_body, "code").String(), place_holder_map_su["cloud_account_id"])
		fmt.Println("Payload", redeem_payload)
		coupon_redeem_status, _ := financials.RedeemCoupon(redeem_coupon_endpoint, userTokenSU, redeem_payload)
		Expect(coupon_redeem_status).To(Equal(200), "Standard users should  redeem coupons")
	})

	// CREATE NEW INSTANCE
	It("Create paid test instance", func() {
		fmt.Println("Starting the Instance Creation via API...")
		// form the endpoint and payload
		instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		bm_name_iu := "autobm-" + utils.GenerateSSHKeyName(4)
		instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), bm_name_iu, "vm-spr-sml", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
		fmt.Println("instance_payload", instance_payload)
		place_holder_map["instance_type"] = "vm-spr-sml"
		// hit the api
		create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, userTokenSU, instance_payload)
		BMName_iu := gjson.Get(create_response_body, "metadata.name").String()
		Expect(create_response_status).To(Equal(200), "Failed to create VM instance")
		Expect(bm_name_iu).To(Equal(BMName_iu), "Failed to create BM instance, resposne validation failed")
		place_holder_map["instance_creation_time"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	})

	// GET THE INSTANCE
	It("Get the created instance and validate", func() {
		instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		// Adding a sleep because it seems to take some time to reflect the creation status
		time.Sleep(180 * time.Second)
		instance_id_created = gjson.Get(create_response_body, "metadata.resourceId").String()
		response_status, response_body := frisby.GetInstanceById(instance_endpoint, userTokenSU, instance_id_created)
		Expect(response_status).To(Equal(200), "Failed to retrieve VM instance")
		Expect(strings.Contains(response_body, `"phase":"Ready"`)).To(BeTrue(), "Validation failed on instance retrieval")
		place_holder_map["resource_id"] = instance_id_created
		place_holder_map["machine_ip"] = gjson.Get(response_body, "status.interfaces[0].addresses[0]").String()
		fmt.Println("IP Address is :" + place_holder_map["machine_ip"])
	})

	//DELETE THE INSTANCE CREATED
	It("Delete the created instance", func() {
		instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		time.Sleep(10 * time.Second)
		// delete the instance created
		delete_response_status, _ := frisby.DeleteInstanceById(instance_endpoint, userTokenSU, instance_id_created)
		Expect(delete_response_status).To(Equal(200), "Failed to delete VM instance")
		time.Sleep(5 * time.Second)
		// validate the deletion
		// Adding a sleep because it seems to take some time to reflect the deletion status
		time.Sleep(1 * time.Minute)
		get_response_status, _ := frisby.GetInstanceById(instance_endpoint, token, instance_id_created)
		Expect(get_response_status).To(Equal(404), "Resource shouldn't be found")
		place_holder_map["instance_deletion_time"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	})

	// DELETE SSH KEYS
	It("Delete the SSH key created", func() {
		logger.Logf.Info("Delete SSH keys...")
		fmt.Println("Delete the SSH-Public-Key Created above...")
		ssh_publickey_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/" + "sshpublickeys"
		delete_response_byname_status, _ := frisby.DeleteSSHKeyByName(ssh_publickey_endpoint, userTokenSU, ssh_publickey_name_created)
		Expect(delete_response_byname_status).To(Equal(200), "assert ssh-public-key deletion response code")
	})
})

var _ = Describe("Multi User - Admin User Instance Launch Without credits", Ordered, Label("Multi-User-e2e"), func() {
	var vnet_created string
	var ssh_publickey_name_created string
	var create_response_status int
	var create_response_body string
	var instance_id_created string
	//var meta_data_map = make(map[string]string)
	//var resourceInfo testsetup.ResourcesInfo

	cloud_account_created = place_holder_map["cloud_account_id"]
	//CREATE VNET
	It("Create vnet with name", func() {
		// fmt.Println("Starting the VNet Creation via API...")
		// // form the endpoint and payload
		vnet_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/vnets"
		vnet_name := compute_utils.GetVnetName()
		vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
		// hit the api
		fmt.Println("Vnet end point ", vnet_endpoint)
		vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, userToken, vnet_payload)
		vnet_created = gjson.Get(vnet_creation_body, "metadata.name").String()
		Expect(vnet_creation_status).To(Equal(200), "Failed to create VNet")
		Expect(vnet_name).To(Equal(vnet_created), "Failed to create Vnet, response validation failed")
	})

	// CREATE SSH KEY
	It("Create ssh public key with name", func() {
		logger.Logf.Info("Starting the SSH-Public-Key Creation via API...")
		// form the endpoint and payload
		ssh_publickey_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/" + "sshpublickeys"
		fmt.Println("SSH key is" + sshPublicKey)
		sshkey_name := "autossh-" + utils.GenerateSSHKeyName(4)
		fmt.Println("SSH  end point ", ssh_publickey_endpoint)
		ssh_publickey_payload := compute_utils.EnrichSSHKeyPayload(compute_utils.GetSSHPayload(), sshkey_name, sshPublicKey)
		// hit the api
		sshkey_creation_status, sshkey_creation_body := frisby.CreateSSHKey(ssh_publickey_endpoint, userToken, ssh_publickey_payload)
		Expect(sshkey_creation_status).To(Equal(200), "Failed to create SSH Public key")
		ssh_publickey_name_created = gjson.Get(sshkey_creation_body, "metadata.name").String()
		Expect(sshkey_name).To(Equal(ssh_publickey_name_created), "Failed to create SSH Public key, response validation failed")
	})

	// LAUNCH INSTANCE WITHOUT CREDITS
	It("LAUNCH INSTANCE WITHOUT CREDITS - Admin", func() {
		logger.Logf.Info("Starting the Instance Creation via API...")
		// form the endpoint and payload
		instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		bm_name_iu := "autobm-" + utils.GenerateSSHKeyName(4)
		instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), bm_name_iu, "vm-spr-sml", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
		fmt.Println("instance_payload", instance_payload)
		place_holder_map["instance_type"] = "vm-spr-sml"
		// hit the api
		create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
		Expect(create_response_status).To(Or(Equal(200), Equal(403)), "Expected response code on paid instance for iu user should be 403 when cloud credits are zero")
	})

	// ADD CREDITS TO ACCOUNT
	It("Create Cloud credits for intel user by redeeming coupons", func() {
		logger.Logf.Info("Starting Cloud Credits Creation...")
		create_coupon_endpoint := base_url + "/v1/cloudcredits/coupons"
		coupon_payload := financials_utils.EnrichCreateCouponPayload(financials_utils.GetCreateCouponPayload(), 100, "idc_billing@intel.com", 1)
		coupon_creation_status, coupon_creation_body := financials.CreateCoupon(create_coupon_endpoint, token, coupon_payload)
		Expect(coupon_creation_status).To(Equal(200), "Failed to create coupon")

		logger.Logf.Info("Redeem credits to current user...")
		redeem_coupon_endpoint := base_url + "/v1/cloudcredits/coupons/redeem"
		redeem_payload := financials_utils.EnrichRedeemCouponPayload(financials_utils.GetRedeemCouponPayload(), gjson.Get(coupon_creation_body, "code").String(), place_holder_map["cloud_account_id"])
		coupon_redeem_status, _ := financials.RedeemCoupon(redeem_coupon_endpoint, userToken, redeem_payload)
		fmt.Println("Payload", redeem_payload)
		Expect(coupon_redeem_status).To(Equal(200), "Failed to redeem coupon")
	})

	// CREATE NEW INSTANCE
	It("Create paid test instance", func() {
		fmt.Println("Starting the Instance Creation via API...")
		// form the endpoint and payload
		instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		bm_name_iu := "autobm-" + utils.GenerateSSHKeyName(4)
		instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), bm_name_iu, "vm-spr-sml", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
		fmt.Println("instance_payload", instance_payload)
		place_holder_map["instance_type"] = "vm-spr-sml"
		// hit the api
		create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
		BMName_iu := gjson.Get(create_response_body, "metadata.name").String()
		Expect(create_response_status).To(Equal(200), "Failed to create VM instance")
		Expect(bm_name_iu).To(Equal(BMName_iu), "Failed to create BM instance, resposne validation failed")
		place_holder_map["instance_creation_time"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	})

	// GET THE INSTANCE
	It("Get the created instance and validate", func() {
		instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		// Adding a sleep because it seems to take some time to reflect the creation status
		time.Sleep(180 * time.Second)
		instance_id_created = gjson.Get(create_response_body, "metadata.resourceId").String()
		response_status, response_body := frisby.GetInstanceById(instance_endpoint, userToken, instance_id_created)
		Expect(response_status).To(Equal(200), "Failed to retrieve VM instance")
		Expect(strings.Contains(response_body, `"phase":"Ready"`)).To(BeTrue(), "Validation failed on instance retrieval")
		place_holder_map["resource_id"] = instance_id_created
		place_holder_map["machine_ip"] = gjson.Get(response_body, "status.interfaces[0].addresses[0]").String()
		fmt.Println("IP Address is :" + place_holder_map["machine_ip"])
	})

	It("Wait for usages to show up", func() {
		time.Sleep(20 * time.Minute)
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
			fmt.Println("PROD ARRR", arr.Array())
			if len(arr.Array()) > 0 {
				return true
			}
			fmt.Println("Waiting 40 more minutes to get products usages...")
			return false
		}, 8*time.Hour, 15*time.Minute).Should(BeTrue())
		arr.ForEach(func(key, value gjson.Result) bool {
			data := value.String()
			logger.Logf.Infof("Usage Data : %s", data)
			product := gjson.Get(data, "productType").String()
			logger.Logf.Infof("Product Data : %s", product)
			Amount := gjson.Get(data, "amount").String()
			actualAMount, _ := strconv.ParseFloat(Amount, 64)
			Expect(actualAMount).Should(BeNumerically(">", float64(0)), "Failed to get positive usage")

			return true // keep iterating
		})
	})

	It("Validate Credits", func() {
		Eventually(func() bool {
			token_response, _ := auth.Get_Azure_Bearer_Token(userName)
			userToken = "Bearer " + token_response
			fmt.Println("TOKEN...", userToken)
			baseUrl := base_url + "/v1/cloudcredits/credit"
			response_status, responseBody := financials.GetCredits(baseUrl, userToken, place_holder_map["cloud_account_id"])
			Expect(response_status).To(Equal(200), "Failed to retrieve Billing Account Cloud Credits")
			usedAmount := gjson.Get(responseBody, "totalUsedAmount").Float()
			usedAmount = testsetup.RoundFloat(usedAmount, 0)
			if usedAmount > float64(float64(0)) {
				return true
			}
			fmt.Println("Waiting 30 more minutes to get credit depletion...")
			return false
		}, 8*time.Hour, 30*time.Minute).Should(BeTrue())
	})

	//DELETE THE INSTANCE CREATED
	It("Delete the created instance", func() {
		instance_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/instances"
		time.Sleep(10 * time.Second)
		// delete the instance created
		delete_response_status, _ := frisby.DeleteInstanceById(instance_endpoint, token, instance_id_created)
		Expect(delete_response_status).To(Equal(200), "Failed to delete VM instance")
		time.Sleep(5 * time.Second)
		// validate the deletion
		// Adding a sleep because it seems to take some time to reflect the deletion status
		time.Sleep(1 * time.Minute)
		get_response_status, _ := frisby.GetInstanceById(instance_endpoint, token, instance_id_created)
		Expect(get_response_status).To(Equal(404), "Resource shouldn't be found")
		place_holder_map["instance_deletion_time"] = strconv.FormatInt(time.Now().UnixMilli(), 10)
	})

	// DELETE SSH KEYS
	It("Delete the SSH key created", func() {
		logger.Logf.Info("Delete SSH keys...")
		fmt.Println("Delete the SSH-Public-Key Created above...")
		ssh_publickey_endpoint := compute_url + "/v1/cloudaccounts/" + cloud_account_created + "/" + "sshpublickeys"
		delete_response_byname_status, _ := frisby.DeleteSSHKeyByName(ssh_publickey_endpoint, token, ssh_publickey_name_created)
		Expect(delete_response_byname_status).To(Equal(200), "assert ssh-public-key deletion response code")
	})
})
*/
