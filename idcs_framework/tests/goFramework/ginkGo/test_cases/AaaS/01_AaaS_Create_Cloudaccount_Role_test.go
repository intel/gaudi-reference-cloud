package AaaS_test

import (
	"fmt"
	"goFramework/framework/common/logger"
	"goFramework/framework/service_api/compute/frisby"
	"goFramework/framework/service_api/financials"
	"goFramework/ginkGo/compute/compute_utils"
	"goFramework/ginkGo/financials/financials_utils"
	"goFramework/utils"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("AaaS cloudaccount flow tests", Ordered, func() {

	var _ = Describe("Member Invitation flows test - Positive", Ordered, Label("Multi-User-e2e"), func() {
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
	})

	var _ = Describe("AaaS cloudaccount roles tests", Ordered, func() {

		It("Whitelist STaaS products", func() {
			logger.Logf.Info("Checking user cloudAccount")
			admin_name := "idc_billing@intel.com"
			code, _ := financials.Whitelist_Cloud_Account_STaaS(base_url, token, place_holder_map["cloud_account_id"], admin_name, "ObjectStorageAsAService")
			Expect(code).To(Or(Equal(200), Equal(500)), "Failed to Whitelist Object Storage")
			code, _ = financials.Whitelist_Cloud_Account_STaaS(base_url, token, place_holder_map["cloud_account_id"], admin_name, "FileStorageAsAService")
			Expect(code).To(Or(Equal(200), Equal(500)), "Failed to Whitelist File Storage")
			logger.Log.Info("STaaS Products Whitelisted succesfully.")
		})

		//CREATE VNET
		It("Create vnet with name", func() {
			// fmt.Println("Starting the VNet Creation via API...")
			// // form the endpoint and payload
			vnet_endpoint := compute_url + "/v1/cloudaccounts/" + place_holder_map["cloud_account_id"] + "/vnets"
			vnet_name := compute_utils.GetVnetName()
			vnet_payload := compute_utils.EnrichVnetPayload(compute_utils.GetVnetPayload(), vnet_name)
			// hit the api
			fmt.Println("Vnet end point ", vnet_endpoint)
			fmt.Println("Vnet name ", vnet_name)
			fmt.Println("vnet_payload ", vnet_payload)
			vnet_creation_status, vnet_creation_body := frisby.CreateVnet(vnet_endpoint, userToken, vnet_payload)
			vnet_created = gjson.Get(vnet_creation_body, "metadata.name").String()
			Expect(vnet_creation_status).To(Equal(200), "Failed to create VNet")
			Expect(vnet_name).To(Equal(vnet_created), "Failed to create Vnet, response validation failed")
		})

		// CREATE SSH KEY
		It("Create ssh public key with name", func() {
			logger.Logf.Info("Starting the SSH-Public-Key Creation via API...")
			// form the endpoint and payload
			ssh_publickey_endpoint := compute_url + "/v1/cloudaccounts/" + place_holder_map["cloud_account_id"] + "/" + "sshpublickeys"
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
		It("LAUNCH INSTANCE WITHOUT CREDITS - Admin - In this case as is premium it already has credits", func() {
			logger.Logf.Info("Starting the Instance Creation via API...")
			// form the endpoint and payload
			instance_endpoint := compute_url + "/v1/cloudaccounts/" + place_holder_map["cloud_account_id"] + "/instances"
			bm_name_iu := "autobm-" + utils.GenerateSSHKeyName(4)
			instance_payload := compute_utils.EnrichInstancePayload(compute_utils.GetInstancePayload(), bm_name_iu, "vm-spr-sml", "ubuntu-2204-jammy-v20230122", ssh_publickey_name_created, vnet_created)
			fmt.Println("instance_payload", instance_payload)
			place_holder_map["instance_type"] = "vm-spr-sml"
			// hit the api
			create_response_status, create_response_body = frisby.CreateInstance(instance_endpoint, userToken, instance_payload)
			Expect(create_response_status).To(Equal(200), "Expected response code on paid instance for pu user should be 200 as for upgrade it adds credits")
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
		It("Create STaaS Instance", func() {
			fmt.Println("Current ENV: ", os.Getenv("IDC_ENV"))
			if os.Getenv("IDC_ENV") != "staging" && os.Getenv("IDC_ENV") != "qa1" && os.Getenv("IDC_ENV") != "dev3" { // this will be replaced with environments with STaaS enabled
				fmt.Println("Skipping tests as this is not an env with STaaS enabled")
				Skip("Skipping test as the environment does not have STaaS enabled for testing.")
			}
			fmt.Println("Starting the Instance Creation via STaaS API...")
			prefix := "aaastest"
			suffix := utils.GenerateInt(2)
			volume_name := prefix + suffix
			size := "5TB"
			staas_payload := fmt.Sprintf(`{
				"metadata":{
					"name":"%s",
					"description":"test"
				},
				"spec":{
					"availabilityZone":"us-staging-1a",
					"filesystemType":"ComputeGeneral",
					"instanceType":"storage-file",
					"request":{
						"storage":"%s",
						"instanceType":""
					},
					"storageClass":"GeneralPurpose"
				}
			}`, volume_name, size)

			if strings.Contains(base_url, "staging") || strings.Contains(base_url, "qa1") || strings.Contains(base_url, "dev3") {
				fmt.Println("User TOk", userToken)
				fmt.Println("User ca", place_holder_map["cloud_account_id"])
				fmt.Println("Payload", staas_payload)
				create_response_code, create_response_body := financials.CreateFileSystem(compute_url, userToken, staas_payload, place_holder_map["cloud_account_id"])
				fmt.Println("STaaS Response", create_response_body)
				instance_id_created = gjson.Get(create_response_body, "metadata.resourceId").String()
				fmt.Println("resourceId: ", instance_id_created)
				Expect(create_response_code).To(Equal(200), "Failed to create FileSystem")
			}
		})

		// GET THE INSTANCE
		It("Get the created instance and validate", func() {
			fmt.Println("Cloud Account id: ", cloud_account_created)
			fmt.Println("Created instance id: ", instance_id_created)
			if os.Getenv("IDC_ENV") != "staging" && os.Getenv("IDC_ENV") != "qa1" && os.Getenv("IDC_ENV") != "dev3" { // this will be replaced with environments with STaaS enabled
				fmt.Println("Skipping tests as this is not an env with STaaS enabled")
				Skip("Skipping test as the environment does not have STaaS enabled for testing.")
			}
			fmt.Println("Starting the Instance retrieval via STaaS API...")
			if strings.Contains(base_url, "staging") || strings.Contains(base_url, "qa1") || strings.Contains(base_url, "dev3") {
				response_status, response_body := financials.GetFileSystemStatusByResourceId(compute_url, userToken, place_holder_map["cloud_account_id"], instance_id_created)
				fmt.Print("Response body: ", response_body)
				Expect(response_status).To(Equal(200), "Failed to retrieve FileSystem")
				Eventually(func() bool {
					response_status, response_body := financials.GetFileSystemStatusByResourceId(compute_url, userToken, place_holder_map["cloud_account_id"], instance_id_created)
					Expect(response_status).To(Equal(200), "Failed to retrieve FileSystem")
					fmt.Print("Response body: ", response_body)
					if strings.Contains(response_body, `"phase":"FSReady"`) || strings.Contains(response_body, `"phase":"FSFailed`) {
						return true
					}
					fmt.Println("Waiting for instance to be ready...")
					return false
				}, 1*time.Hour, 3*time.Minute).Should(BeTrue(), "Validation failed on instance retrieval, fileSystem is not in ready phase")
				place_holder_map["resource_id_fileSystem"] = instance_id_created
				fmt.Println("Resource ID FS", place_holder_map["resource_id_fileSystem"])
				place_holder_map["file_size"] = gjson.Get(response_body, "spec.request.storage").String()
				Expect(place_holder_map["file_size"]).To(Equal("5TB"), "Validation failed on file size retrieval file Size: "+place_holder_map["file_size"])
			}
		})

		It("Create Object Storage", func() {
			if os.Getenv("IDC_ENV") != "staging" && os.Getenv("IDC_ENV") != "qa1" && os.Getenv("IDC_ENV") != "dev3" { // this will be replaced with environments with STaaS enabled
				fmt.Println("Skipping tests as this is not an env with STaaS enabled")
				Skip("Skipping test as the environment does not have STaaS enabled for testing.")
			}
			fmt.Println("Starting the Instance Creation via STaaS API...")
			prefix := "aaastest"
			suffix := utils.GenerateInt(2)
			volume_name := prefix + suffix
			versioned := true
			staas_payload := fmt.Sprintf(`{
				"metadata":{
					"name":"%s",
					"description":"test"
				},
				"spec":{
					"availabilityZone":"us-staging-1a",
					"versioned": %t,
					"instanceType":"storage-object"
				}
			}`, volume_name, versioned)
			if strings.Contains(base_url, "staging") || strings.Contains(base_url, "qa1") || strings.Contains(base_url, "dev3") {
				fmt.Println("User TOk", userToken)
				fmt.Println("User ca", place_holder_map["cloud_account_id"])
				fmt.Println("Payload", staas_payload)
				create_response_code, create_response_body := financials.CreateObjectStorage(compute_url, userToken, staas_payload, place_holder_map["cloud_account_id"])
				fmt.Println("STaaS Response", create_response_body)
				instance_id_created = gjson.Get(create_response_body, "metadata.resourceId").String()
				fmt.Println("resourceId: ", instance_id_created)
				Expect(create_response_code).To(Equal(200), "Failed to create Object Storage")
			}
		})

		// GET THE INSTANCE
		It("Get the created Object Storage and validate", func() {
			if os.Getenv("IDC_ENV") != "staging" && os.Getenv("IDC_ENV") != "qa1" && os.Getenv("IDC_ENV") != "dev3" { // this will be replaced with environments with STaaS enabled
				fmt.Println("Skipping tests as this is not an env with STaaS enabled")
				Skip("Skipping test as the environment does not have STaaS enabled for testing.")
			}
			fmt.Println("Cloud Account id: ", cloud_account_created)
			fmt.Println("Created instance id: ", instance_id_created)
			fmt.Println("Starting the Instance retrieval via STaaS API...")
			if strings.Contains(base_url, "staging") || strings.Contains(base_url, "qa1") || strings.Contains(base_url, "dev3") {
				response_status, response_body := financials.GetObjectStorageStatusByResourceId(compute_url, userToken, place_holder_map["cloud_account_id"], instance_id_created)
				fmt.Print("Response body: ", response_body)
				Expect(response_status).To(Equal(200), "Failed to retrieve Object Storage")
				Eventually(func() bool {
					response_status, response_body := financials.GetObjectStorageStatusByResourceId(compute_url, userToken, place_holder_map["cloud_account_id"], instance_id_created)
					Expect(response_status).To(Equal(200), "Failed to retrieve FileSystem")
					fmt.Print("Response body: ", response_body)
					if strings.Contains(response_body, `"phase":"BucketReady"`) || strings.Contains(response_body, `"phase":"BucketProvisioning"`) {
						return true
					}
					fmt.Println("Waiting for instance to be ready...")
					return false
				}, 1*time.Hour, 3*time.Minute).Should(BeTrue(), "Validation failed on instance retrieval, fileSystem is not in ready phase")
				instance_id_created = gjson.Get(response_body, "metadata.resourceId").String()
				fmt.Println("Resource ID OS:", instance_id_created)
				place_holder_map["resource_id_objectStorage"] = instance_id_created
				fmt.Println("Resource ID OS:", place_holder_map["resource_id_objectStorage"])
				place_holder_map["file_size"] = gjson.Get(response_body, "spec.request.size").String()
				Expect(place_holder_map["file_size"]).To(Equal("5TB"), "Validation failed on file size retrieval file Size: "+place_holder_map["file_size"])
			}
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
			Expect(code).To(Or(Equal(200), Equal(409)), "Failed assigning role.")
		})

		It("Create Role", func() {
			if !strings.Contains(base_url, "staging") && !strings.Contains(base_url, "dev.api") && !strings.Contains(base_url, "qa1") && !strings.Contains(base_url, "dev3") {
				fmt.Println("Creating dummy resource ID as environment does not have STaaS enabled...")
				place_holder_map["resource_id_fileSystem"] = uuid.New().String()
			}

			fmt.Println("Resource id 1:", place_holder_map["resource_id_fileSystem"])
			role_alias := "aaastest" + compute_utils.GetRandomString()
			payload := fmt.Sprintf(`{
				"alias":"%s",
				"effect":"allow",
				"actions":[
					"get"
				],
				"permissions":[
					{
						"actions":[
							"get"
						],
						"resourceId":"%s",
						"resourceType":"filestorage"
					}
				],
				"users":[
					"%s"
				]
			}`, role_alias, place_holder_map["resource_id_fileSystem"], userNameSU)
			fmt.Println("PAYLOAD", payload)
			code, body := financials.CreateUserRole(base_url, userToken, payload, place_holder_map["cloud_account_id"])
			Expect(code).To(Or(Equal(200)), "Failed creating role "+body)
		})

		It("Get User Roles", func() {
			code, body := financials.GetUserRoles(base_url, userToken, place_holder_map["cloud_account_id"])
			fmt.Println("Endpoint Response: " + body)

			// Get the number of roles in the array
			numRoles := gjson.Get(body, "cloudAccountRoles.#").Int()

			// Check if there is at least one role
			if numRoles > 0 {
				// Access the last role using the length of the array
				user_role_id = gjson.Get(body, fmt.Sprintf("cloudAccountRoles.%d.id", numRoles-1)).String()
				fmt.Println("Role Id", user_role_id)
			} else {
				fmt.Println("No roles found in the response.")
			}

			Expect(code).To(Equal(200), "Failed to retrieve user roles.")
		})

		It("Get User Role SU", func() {
			code, body := financials.GetUserRoles(base_url, userTokenSU, place_holder_map_su["cloud_account_id"])
			fmt.Println("Endpoint Response: " + body)

			// Get the number of roles in the array
			numRoles := gjson.Get(body, "cloudAccountRoles.#").Int()

			// Check if there is at least one role
			if numRoles > 0 {
				// Access the last role using the length of the array
				user_role_id_su = gjson.Get(body, fmt.Sprintf("cloudAccountRoles.%d.id", numRoles-1)).String()
				fmt.Println("Role Id", user_role_id_su)
			} else {
				fmt.Println("No roles found in the response.")
			}

			Expect(code).To(Equal(200), "Failed to retrieve user roles.")
		})

		// GET THE INSTANCE SU
		It("Get the created Object Storage and validate", func() {
			fmt.Println("Cloud Account id: ", cloud_account_created)
			fmt.Println("Created instance id OS: ", place_holder_map["resource_id_objectStorage"])
			fmt.Println("Starting the Instance retrieval via STaaS API...")
			if strings.Contains(base_url, "staging") || strings.Contains(base_url, "qa1") || strings.Contains(base_url, "dev3") {
				response_status, response_body := financials.GetObjectStorageStatusByResourceId(compute_url, userTokenSU, place_holder_map["cloud_account_id"], place_holder_map["resource_id_objectStorage"])
				fmt.Print("Response body: ", response_body)
				Expect(response_status).To(Equal(403), "User should not have access to recently created OS as the role asigned only gives access to FS")
				Eventually(func() bool {
					response_status, response_body := financials.GetObjectStorageStatusByResourceId(compute_url, userToken, place_holder_map["cloud_account_id"], place_holder_map["resource_id_objectStorage"])
					Expect(response_status).To(Equal(200), "Failed to retrieve FileSystem")
					fmt.Print("Response body: ", response_body)
					place_holder_map["resource_id_objectStorage"] = instance_id_created
					place_holder_map["file_size"] = gjson.Get(response_body, "spec.request.size").String()
					if strings.Contains(response_body, `"phase":"BucketReady"`) {
						return true
					}
					fmt.Println("Waiting for instance to be ready...")
					return false
				}, 1*time.Hour, 3*time.Minute).Should(BeTrue(), "Validation failed on instance retrieval, Object Storage is not in provisioning state")
				Expect(place_holder_map["file_size"]).To(Equal("5TB"), "Validation failed on file size retrieval file Size: "+place_holder_map["file_size"])
			}
		})

		// Once role is updated try to access staas resource with guest cloudaccount SU
		It("Update User Role", func() {
			if !strings.Contains(base_url, "staging") && !strings.Contains(base_url, "qa1") && !strings.Contains(base_url, "dev3") {
				fmt.Println("Creating dummy resource ID as environment does not have STaaS enabled...")
				place_holder_map["resource_id_objectStorage"] = uuid.New().String()
			}
			role_alias := "aaastest" + compute_utils.GetRandomString()
			payload := fmt.Sprintf(`{
				"alias":"%s",
				"effect":"allow",
				"actions":[
					"update"
				],
				"permissions":[
					{
						"actions":[
							"get"
						],
						"resourceId":"%s",
						"resourceType":"filestorage"
					}
				],
				"users":[
					"%s"
				]
			}`, role_alias, place_holder_map["resource_id_objectStorage"], userNameSU)
			code, body := financials.UpdateUserRole(base_url, userToken, payload, place_holder_map["cloud_account_id"], user_role_id)
			fmt.Println("Endpoint Response: " + body)
			Expect(code).To(Equal(200), "Failed to update user role.")
		})

		It("Get User Role", func() {
			code, body := financials.GetUserRoles(base_url, userToken, place_holder_map["cloud_account_id"])
			fmt.Println("Endpoint Response: " + body)
			Expect(code).To(Equal(200), "Failed to retrieve user roles.")
		})

		It("Delete User Role", func() {
			code, body := financials.DeleteUserRole(base_url, userToken, place_holder_map["cloud_account_id"], user_role_id)
			fmt.Println("Endpoint Response: " + body)
			fmt.Println("Role Id", user_role_id)
			fmt.Println("Endpoint Response: " + body)
			Expect(code).To(Equal(200), "Failed to retrieve user roles.")
		})

		It("Create second Role", func() {
			if !strings.Contains(base_url, "staging") && !strings.Contains(base_url, "dev.api") && !strings.Contains(base_url, "qa1") && !strings.Contains(base_url, "dev3") {
				fmt.Println("Creating dummy resource ID as environment does not have STaaS enabled...")
				place_holder_map["resource_id_fileSystem"] = uuid.New().String()
			}

			fmt.Println("Resource id 1:", place_holder_map["resource_id_fileSystem"])
			role_alias := "aaastest" + compute_utils.GetRandomString()
			payload := fmt.Sprintf(`{
				"alias":"%s",
				"effect":"allow",
				"actions":[
					"get"
				],
				"permissions":[
					{
						"actions":[
							"get"
						],
						"resourceId":"%s",
						"resourceType":"filestorage"
					}
				],
				"users":[
					"%s"
				]
			}`, role_alias, place_holder_map["resource_id_fileSystem"], userNameSU)
			fmt.Println("PAYLOAD", payload)
			code, body := financials.CreateUserRole(base_url, userToken, payload, place_holder_map["cloud_account_id"])
			Expect(code).To(Or(Equal(200)), "Failed creating role "+body)
		})

		It("Get User Role", func() {
			code, body := financials.GetUserRoles(base_url, userToken, place_holder_map["cloud_account_id"])
			fmt.Println("Endpoint Response: " + body)
			Expect(code).To(Equal(200), "Failed to retrieve user roles.")
		})

	})
	var _ = Describe("AaaS Role permission tests", Ordered, func() {

		It("Get User Roles", func() {
			code, body := financials.GetUserRoles(base_url, userToken, place_holder_map["cloud_account_id"])
			fmt.Println("Endpoint Response ROLES BEFORE PERMISSION ADDED: " + body)
			// Get the number of roles in the array
			numRoles := gjson.Get(body, "cloudAccountRoles.#").Int()
			// Check if there is at least one role
			if numRoles > 0 {
				// Access the last role using the length of the array
				user_role_id = gjson.Get(body, fmt.Sprintf("cloudAccountRoles.%d.id", numRoles-1)).String()
				fmt.Println("Role Id", user_role_id_su)
			} else {
				fmt.Println("No roles found in the response.")
			}
			fmt.Println("Role Id", user_role_id)
			fmt.Println("Endpoint Response: " + body)
			Expect(code).To(Equal(200), "Failed to retrieve user roles.")
		})

		It("Assign Role", func() {
			payload := fmt.Sprintf(`{
				"cloudAccountRoleIds": [
					"%s"
				]
			}`, user_role_id)
			fmt.Println("PAYLOAD", payload)
			code, body := financials.AssignRoleUser(base_url, userToken, payload, place_holder_map["cloud_account_id"], userNameSU)
			fmt.Println("Endpoint Response: " + body)
			Expect(code).To(Or(Equal(200), Equal(409)), "Failed assigning role.")
		})

		It("Get User Roles", func() {
			code, body := financials.GetUserRoles(base_url, userToken, place_holder_map["cloud_account_id"])
			fmt.Println("Endpoint Response ROLES AFTER ROLE ADDED: " + body)
			// Get the number of roles in the array
			numRoles := gjson.Get(body, "cloudAccountRoles.#").Int()
			// Check if there is at least one role
			if numRoles > 0 {
				// Access the last role using the length of the array
				user_role_id = gjson.Get(body, fmt.Sprintf("cloudAccountRoles.%d.id", numRoles-1)).String()
				fmt.Println("Role Id", user_role_id_su)
			} else {
				fmt.Println("No roles found in the response.")
			}
			permission_role_id = gjson.Get(body, "cloudAccountRoles.0.permissions.0.id").String()
			fmt.Println("Role Id", user_role_id)
			fmt.Println("permission_role_id", permission_role_id)
			fmt.Println("Endpoint Response: " + body)
			Expect(code).To(Equal(200), "Failed to retrieve user roles.")
		})

		It("Add Role Permissions", func() {
			fmt.Println("Role: ", user_role_id)
			payload := fmt.Sprintf(`{
				"permission": {
					"actions": [
                        "get",
                        "delete",
                        "search"
					],
					"resourceId": "%s",
					"resourceType": "filestorage"
				}
			}`, uuid.New())
			fmt.Println("PAYLOAD", payload)
			code, body := financials.CreateUserRolePermission(base_url, userToken, payload, place_holder_map["cloud_account_id"], user_role_id)
			fmt.Println("Endpoint Response: " + body)
			Expect(code).To(Equal(200), "Failed assigning a role.")
		})

		It("Get User Roles", func() {
			code, body := financials.GetUserRoles(base_url, userToken, place_holder_map["cloud_account_id"])
			fmt.Println("Endpoint Response ROLES AFTER PERMISSION ADDED: " + body)
			// Get the number of roles in the array
			numRoles := gjson.Get(body, "cloudAccountRoles.#").Int()
			// Check if there is at least one role
			if numRoles > 0 {
				// Access the last role using the length of the array
				user_role_id = gjson.Get(body, fmt.Sprintf("cloudAccountRoles.%d.id", numRoles-1)).String()
				fmt.Println("Role Id", user_role_id_su)
			} else {
				fmt.Println("No roles found in the response.")
			}
			permission_role_id = gjson.Get(body, "cloudAccountRoles.0.permissions.0.id").String()
			fmt.Println("Role Id", user_role_id)
			fmt.Println("permission_role_id", permission_role_id)
			fmt.Println("Endpoint Response: " + body)
			Expect(code).To(Equal(200), "Failed to retrieve user roles.")
		})

		It("Update Role Permissions", func() {
			fmt.Println("Role: ", user_role_id)
			payload := fmt.Sprintf(`{
				"permission": {
					"actions": [
					"get", "create"
					],
					"resourceId": "%s",
					"resourceType": "filestorage"
				}
			}`, uuid.New())
			fmt.Println("PAYLOAD", payload)
			fmt.Println("Permission: ", permission_role_id)
			code, body := financials.UpdateUserRolePermission(base_url, userToken, payload, place_holder_map["cloud_account_id"], user_role_id, permission_role_id)
			fmt.Println("Endpoint Response: " + body)
			Expect(code).To(Or(Equal(200), Equal(404)), "Failed updating role permission")
		})

		It("Get User Roles", func() {
			code, body := financials.GetUserRoles(base_url, userToken, place_holder_map["cloud_account_id"])
			fmt.Println("Endpoint Response ROLES AFTER PERMISSION UPDATED: " + body)
			// Get the number of roles in the array
			numRoles := gjson.Get(body, "cloudAccountRoles.#").Int()
			// Check if there is at least one role
			if numRoles > 0 {
				// Access the last role using the length of the array
				user_role_id = gjson.Get(body, fmt.Sprintf("cloudAccountRoles.%d.id", numRoles-1)).String()
				fmt.Println("Role Id", user_role_id_su)
			} else {
				fmt.Println("No roles found in the response.")
			}
			fmt.Println("Role Id", user_role_id)
			fmt.Println("Endpoint Response: " + body)
			Expect(code).To(Equal(200), "Failed to retrieve user roles.")
		})

		It("Delete Role Permission", func() {
			code, body := financials.DeleteUserRolePermission(base_url, userToken, place_holder_map["cloud_account_id"], user_role_id, permission_role_id)
			fmt.Println("Endpoint Response: " + body)
			fmt.Println("Role Id", user_role_id)
			fmt.Println("Endpoint Response: " + body)
			Expect(code).To(Equal(200), "Failed to retrieve user roles.")
		})

		It("Check Role exists", func() {
			payload := fmt.Sprintf(`{
				"cloudAccountId": "%s",
				"subject": "%s",
				"systemRole": "%s"
			}`, place_holder_map["cloud_account_id"], place_holder_map_su["cloud_account_id"], user_role_id)
			fmt.Println("PAYLOAD", payload)
			code, body := financials.SystemRoleExists(base_url, token, payload)
			fmt.Println("Endpoint Response: " + body)
			Expect(code).To(Equal(404), "Failed, role was not deleted.")
		})

		It("Check Role actions", func() {
			payload := fmt.Sprintf(`{
				"cloudAccountId": "%s",
				"resourceId": "%s",
				"resourceType": "instance"
			}`, place_holder_map["cloud_account_id"], place_holder_map["resource_id_2"])
			fmt.Println("PAYLOAD", payload)
			code, body := financials.Actions(base_url, token, payload)
			fmt.Println("Endpoint Response: " + body)
			Expect(code).To(Or(Equal(200), Equal(403)), "Failed checking role actions") // This could return 403 as this endpoint may become internal only
		})
	})

})
