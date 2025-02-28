package gRPC_Internal_test

import (
	"fmt"
	"goFramework/framework/library/financials/metering"
	"goFramework/framework/library/grpc/authorization"
	"goFramework/framework/service_api/financials"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
)

var _ = Describe("CouponService", func() {
	var token string
	It("Validate Cognito JWT retrieval.", func() {
		fmt.Println("Client: ", cognitoClientId)
		code, body := financials.GetJwtCognito(cognitoUrl, "client_credentials", cognitoClientId, clientSecret, "idc-staging-staging-internal/coupons:view")
		Expect(code).To(Equal(200), "Failed to create cognito JWT token.")
		Expect(body).NotTo(BeNil(), "Failed to retrieve cognito JWT token.")
		token = gjson.Get(body, "access_token").String()
		Expect(token).NotTo(BeNil(), "Failed to retrieve cognito JWT token.")
	})

	It("Validate Cognito JWT retrieval.", func() {
		fmt.Println("Client: ", cognitoClientId)
		code, body := financials.GetJwtCognito(cognitoUrl, "client_credentials", cognitoClientId, clientSecret, "idc-staging-staging-internal/coupons:create")
		Expect(code).To(Equal(200), "Failed to create cognito JWT token.")
		Expect(body).NotTo(BeNil(), "Failed to retrieve cognito JWT token.")
		token = gjson.Get(body, "access_token").String()
		Expect(token).NotTo(BeNil(), "Failed to retrieve cognito JWT token.")
	})

	It("Validate global gRPC call", func() {
		fmt.Println("Global grpc: ", grpcGlobalUrl)
		result := authorization.GetCoupons(grpcGlobalUrl)
		Expect(result).To(Equal(true), "Failed to retrieve coupons.")
	})

	It("Validate user can read coupons", func() {
		fmt.Println("Global grpc: ", grpcGlobalUrl)
		result := authorization.GetCoupons(grpcGlobalUrl)
		Expect(result).To(Equal(true), "Failed to retrieve coupons.")
	})

	It("Validate user can read coupon without create scope", func() {
		code, _ := financials.GetJwtCognito(cognitoUrl, "client_credentials", cognitoClientId, clientSecret, "idc-staging-staging-internal/coupons")
		Expect(code).To(Equal(200), "Failed to create cognito JWT token.")
		fmt.Println("Global grpc: ", grpcGlobalUrl)
		fmt.Println("Create metering record test")
		result, err := metering.Create_Usage_Record("validPayload", 200)
		fmt.Println("Metering Response", result)
		fmt.Println("Metering Response code", err)
		Expect(result).To(Equal(true), "Failed to create metering record without gRPC scope")
	})

	It("Validate user can read coupon with create scope", func() {
		code, _ := financials.GetJwtCognito(cognitoUrl, "client_credentials", cognitoClientId, clientSecret, "idc-staging-staging-internal/coupons:create")
		Expect(code).To(Equal(200), "Failed to create cognito JWT token.")
		fmt.Println("Global grpc: ", grpcGlobalUrl)
		fmt.Println("Create metering record test")
		result, err := metering.Create_Usage_Record("validPayload", 200)
		fmt.Println("Metering Response", result)
		fmt.Println("Metering Response code", err)
		Expect(result).To(Equal(true), "Failed to create metering record with gRPC scope")
	})

	It("Validate user can't read coupon without view scope", func() {
		code, _ := financials.GetJwtCognito(cognitoUrl, "client_credentials", cognitoClientId, clientSecret, "idc-staging-staging-internal/coupons")
		Expect(code).To(Equal(200), "Failed to create cognito JWT token.")
		fmt.Println("Global grpc: ", grpcGlobalUrl)
		result := authorization.GetCoupons(grpcGlobalUrl)
		Expect(result).To(Equal(false), "Failed to get 403 error using token without scope.")
	})
})
