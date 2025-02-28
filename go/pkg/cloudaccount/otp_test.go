// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

func CreateOtpRequest(t *testing.T) *pb.OtpRequest {
	user := "premiumuser" + uuid.NewString() + "@example.com"
	memberEmail := "member" + uuid.NewString() + "@proton.me"
	caClient := pb.NewCloudAccountServiceClient(test.ClientConn())

	cloudAccountId, err := caClient.Create(context.Background(),
		&pb.CloudAccountCreate{
			Name:  user,
			Owner: user,
			Tid:   uuid.NewString(),
			Oid:   uuid.NewString(),
			Type:  pb.AccountType_ACCOUNT_TYPE_PREMIUM,
		})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}

	otpRequest := &pb.OtpRequest{
		CloudAccountId: cloudAccountId.Id,
		MemberEmail:    memberEmail,
	}

	return otpRequest
}

func TestCreateOTP(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCreateOTP")
	logger.Info("BEGIN")
	defer logger.Info("End")

	otpServiceClient := pb.NewOtpServiceClient(test.ClientConn())
	otpRequest := CreateOtpRequest(t)

	_, err := otpServiceClient.CreateOtp(ctx, otpRequest)
	if err != nil {
		t.Fatalf("failed to create otp: %v", err)
	}
}

func TestResendOTP(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestResendOTP")
	logger.Info("BEGIN")
	defer logger.Info("End")

	otpServiceClient := pb.NewOtpServiceClient(test.ClientConn())
	otpRequest := CreateOtpRequest(t)

	_, err := otpServiceClient.CreateOtp(ctx, otpRequest)
	if err != nil {
		t.Fatalf("failed to create admin OTP record : %v", err)
	}

	_, err = otpServiceClient.ResendOtp(ctx, otpRequest)
	if err != nil {
		t.Fatalf("failed to resend otp: %v", err)
	}
}

func TestVerifyOtp(t *testing.T) {
	var svc *OtpService
	ctx := context.Background()

	logger := log.FromContext(ctx).WithName("TestVerifyOTP")
	logger.Info("BEGIN")
	defer logger.Info("End")

	otpServiceClient := pb.NewOtpServiceClient(test.ClientConn())
	otpRequest := CreateOtpRequest(t)

	err := checkMembersAddLimit(ctx, otpRequest.CloudAccountId)
	if err != nil {
		t.Fatalf("failed to check the limit to add members : %v", err)
	}

	valid, err := svc.validateOtpRequest(ctx, otpRequest)
	if !valid {
		t.Fatalf("failed to Validate OTP : %v", err)
	}
	otpData, err := svc.otpCreateHandler(ctx, otpRequest)
	if err != nil {
		t.Fatalf("failed to create admin OTP record : %v", err)
	}
	otp_req := &pb.VerifyOtpRequest{
		MemberEmail:    otpData.MemberEmail,
		CloudAccountId: otpData.CloudAccountId,
		OtpCode:        otpData.OtpCode,
	}
	res, err := otpServiceClient.VerifyOtp(ctx, otp_req)

	if err != nil {
		t.Fatalf("failed to verify OTP: %v", err)
	}
	if res.Validated != true {
		t.Errorf("OTP is not validated.")
	}
	if res.OtpState != pb.OtpState_OTP_STATE_ACCEPTED {
		t.Errorf("OTP is not Accepted.")
	}

}
