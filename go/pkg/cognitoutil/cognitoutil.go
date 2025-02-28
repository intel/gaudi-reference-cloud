// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cognitoutil

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type COGNITOUtil struct {
	cognitoClient *cognitoidentityprovider.CognitoIdentityProvider
}

func (cognitoUtil *COGNITOUtil) Init(ctx context.Context, region string, credentialsFile string, userPool string) error {
	logger := log.FromContext(ctx).WithName("COGNITOUtil.Init")
	logger.Info("Init BEGIN")
	defer logger.Info("Init END")
	logger.Info("cognito init", "region", region, "credentialsFile", credentialsFile)
	cognitoClient, err := cognitoUtil.CreateCognitoClient(ctx, region, credentialsFile, userPool)
	if err != nil {
		logger.Error(err, "failed to create cognito client with session")
		return err
	}
	cognitoUtil.cognitoClient = cognitoClient

	return nil
}

func (cognitoUtil *COGNITOUtil) CreateCognitoClient(ctx context.Context, region string, credentialsFile string, userPool string) (*cognitoidentityprovider.CognitoIdentityProvider, error) {
	logger := log.FromContext(ctx).WithName("COGNITOUtil.CreateCognitoClient")
	logger.Info("BEGIN")
	defer logger.Info("END")
	logger.Info("cognito init", "region", region, "credentialsFile", credentialsFile)
	var sess *session.Session
	var err error
	if credentialsFile != "" {
		options := session.Options{
			Config: aws.Config{
				Region: aws.String(region),
			},
			SharedConfigState: session.SharedConfigEnable,
		}
		options.SharedConfigFiles = []string{credentialsFile}
		sess, err = session.NewSessionWithOptions(options)
		if err != nil {
			return nil, err
		}
	} else {
		sess, err = session.NewSession(&aws.Config{
			Region: aws.String("us-west-2"),
		})
		if err != nil {
			return nil, err
		}
	}
	cognitoClient := cognitoidentityprovider.New(sess)
	return cognitoClient, nil
}

func (cognitoUtil *COGNITOUtil) CreateCredentials(ctx context.Context, appClientName string, userPoolId string, email string, customScope string, cloudAccountId string) (string, string, error) {
	logger := log.FromContext(ctx).WithName("COGNITOUtil.CreateCredentials")
	logger.Info("BEGIN")
	defer logger.Info("END")
	//Make this configurable
	scopes := []*string{
		aws.String(customScope),
	}
	// Define the OAuth flow and scopes
	oauthFlows := []*string{
		aws.String("client_credentials"),
	}
	appClient := appClientName + "-" + cloudAccountId
	input := &cognitoidentityprovider.CreateUserPoolClientInput{
		ClientName:                      aws.String(appClient),
		UserPoolId:                      aws.String(userPoolId),
		GenerateSecret:                  aws.Bool(true),
		AllowedOAuthFlows:               oauthFlows,
		AllowedOAuthScopes:              scopes,         // Define your OAuth scopes
		AllowedOAuthFlowsUserPoolClient: aws.Bool(true), // Enable OAuth2 flows
	}
	result, err := cognitoUtil.cognitoClient.CreateUserPoolClient(input)
	if err != nil || result.UserPoolClient == nil {
		awsErr := cognitoUtil.getAWSError(err)
		logger.Error(awsErr, "error creating client credentials", "cloudAccountId", cloudAccountId)
		return "", "", status.Errorf(codes.Aborted, "error creating client credentials: %v", awsErr)
	}
	return *result.UserPoolClient.ClientId, *result.UserPoolClient.ClientSecret, nil

}

func (cognitoUtil *COGNITOUtil) DeleteCredentials(ctx context.Context, clientId string, userPoolId string) (*wrapperspb.BoolValue, error) {
	logger := log.FromContext(ctx).WithName("COGNITOUtil.DeleteCredentials")
	logger.Info("BEGIN")
	defer logger.Info("END")

	logger.Info("Delete credentials for ", "clientID :", clientId)
	// Create DeleteUserPoolClientInput
	input := &cognitoidentityprovider.DeleteUserPoolClientInput{
		UserPoolId: aws.String(userPoolId),
		ClientId:   aws.String(clientId),
	}
	// Call DeleteUserPoolClient API
	_, err := cognitoUtil.cognitoClient.DeleteUserPoolClient(input)

	if err != nil {
		logger.Error(err, "error Deleting Client Credentials", "clientID", clientId)
		return nil, err
	} else {
		return &wrapperspb.BoolValue{Value: true}, err
	}
}

func (cognitoUtil *COGNITOUtil) getAWSError(err error) error {
	if awsErr, ok := err.(awserr.Error); ok {
		switch awsErr.Code() {
		case cognitoidentityprovider.ErrCodeNotAuthorizedException:
			return fmt.Errorf("not authorized: %s", awsErr.Message())
		case cognitoidentityprovider.ErrCodeResourceNotFoundException:
			return fmt.Errorf("resource not found: %s", awsErr.Message())
		case cognitoidentityprovider.ErrCodeScopeDoesNotExistException:
			return fmt.Errorf("scope does not exist: %s", awsErr.Message())
		case cognitoidentityprovider.ErrCodeLimitExceededException:
			return fmt.Errorf("user exceeds the limit: %s", awsErr.Message())
		case cognitoidentityprovider.ErrCodeTooManyRequestsException:
			return fmt.Errorf("too many requests: %s", awsErr.Message())
		default:
			return fmt.Errorf("aws error: %s - %s", awsErr.Code(), awsErr.Message())
		}
	}
	return fmt.Errorf("unexpected error: %v", err)
}
