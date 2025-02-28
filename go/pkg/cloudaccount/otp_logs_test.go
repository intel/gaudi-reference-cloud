// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount

import (
	"context"
	"testing"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/stretchr/testify/assert"
)

func TestCheckThresholdReached_zeroAttempt(t *testing.T) {
	ctx := context.Background()

	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("TestCheckThresholdReached_zeroAttempt").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("End")

	cloudAccountId := "123456789"
	validate_attempt_reached, _, err := failedOtpLogs.CheckThresholdReached(ctx, db, pb.OtpType_INVITATION_VALIDATE, cloudAccountId)
	if err != nil {
		t.Fatalf("error in checking attempt number: %v", err)
	}
	assert.Equal(t, validate_attempt_reached, false)
}

func TestCheckThresholdReached_FourAttempt(t *testing.T) {
	ctx := context.Background()

	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("TestCheckThresholdReached_FourAttempt").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("End")

	cloudAccountId := "123456789"
	for i := 1; i <= 4; i++ {
		err := failedOtpLogs.WriteAttempt(ctx, db, pb.OtpType_INVITATION_VALIDATE, cloudAccountId)
		if err != nil {
			t.Fatalf("Failed to write invalid attempt on otp logs: %v", err)
		}
	}

	validate_attempt_reached, _, err := failedOtpLogs.CheckThresholdReached(ctx, db, pb.OtpType_INVITATION_VALIDATE, cloudAccountId)
	if err != nil {
		t.Fatalf("error in checking attempt number: %v", err)
	}
	assert.Equal(t, validate_attempt_reached, true)
}
