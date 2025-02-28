// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cognitoutil

import (
	"context"
	"os"

	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

func TestCreateCredentials(t *testing.T) {
	log.SetDefaultLogger()
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCreateCredentials")
	logger.Info("TestCreateCredentials BEGIN")
	defer logger.Info("TestCreateCredentials END")
	awsCognitoRegion := os.Getenv("AWS_COGNITO_REGION")
	awsCredentialsFile := os.Getenv("AWS_CREDENTIAL_FILE")
	awsUserPoolName := os.Getenv("AWS_COGNITO_USER_POOL_NAME")
	awsUserPoolId := os.Getenv("AWS_COGNITO_USER_POOL_ID")
	logger.Info("AWS COGNITO", "awsCognitoRegion", awsCognitoRegion, "awsUserPoolName", awsUserPoolName, "awsUserPool", awsUserPoolId)
	if len(awsUserPoolName) == 0 || len(awsUserPoolId) == 0 || len(awsCognitoRegion) == 0 {
		t.Skip("aws envionment variable is not set")
	}
	cognitoUtil := COGNITOUtil{}
	if err := cognitoUtil.Init(ctx, awsCognitoRegion, awsCredentialsFile, awsUserPoolId); err != nil {
		logger.Error(err, "couldn't init cognitoUtil client")
	}
	clientId, clientSecret, err := cognitoUtil.CreateCredentials(ctx, "test.appClientName", awsUserPoolId, "userEmail@test.com", awsUserPoolName, "123456789101")
	if err != nil {
		logger.Error(err, "Error creating user pool client")
		t.Error("expected no error when creating credentials")

	} else {
		logger.Info("expected credentials", "clientId", clientId, "clientSecret", clientSecret)
	}
}

func TestCreateCredentialsFailed(t *testing.T) {
	log.SetDefaultLogger()
	ctx := context.Background()
	logger := log.FromContext(ctx).WithName("TestCreateCredentialsFailed")
	logger.Info("TestCreateCredentialsFailed BEGIN")
	defer logger.Info("TestCreateCredentialsFailed END")
	awsCognitoRegion := os.Getenv("AWS_COGNITO_REGION")
	awsCredentialsFile := os.Getenv("AWS_CREDENTIAL_FILE")
	awsUserPoolName := os.Getenv("AWS_COGNITO_USER_POOL_NAME")
	awsUserPool := os.Getenv("AWS_COGNITO_USER_POOL_ID")
	logger.Info("AWS COGNITO", "awsCognitoRegion", awsCognitoRegion, "awsCredentialsFile", awsCredentialsFile, "awsUserPool", awsUserPool)
	if len(awsUserPoolName) == 0 || len(awsUserPool) == 0 || len(awsCognitoRegion) == 0 {
		t.Skip("aws envionment variable is not set")
	}
	cognitoUtil := COGNITOUtil{}
	if err := cognitoUtil.Init(ctx, awsCognitoRegion, awsCredentialsFile, awsUserPool); err != nil {
		logger.Error(err, "couldn't init cognitoUtil client")
	}
	_, _, err := cognitoUtil.CreateCredentials(ctx, "test", "", "userEmail@test.com", awsUserPoolName, "")
	if err == nil {
		t.Error("expected error when creating credentials with empty awsUserPool, but got none")
	} else {
		logger.Info("expected error occurred", "error", err)
	}
}
