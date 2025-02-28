// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package batch_service

import (
	"context"
	"errors"
	"os"
	"testing"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestNewCloudAccountClient(t *testing.T) {
	ctx := context.Background()

	t.Run("successful execution with env var", func(t *testing.T) {
		mockResolver := NewMockResolver(func(ctx context.Context, s string) (string, error) {
			return "", nil
		})

		// set environment variable
		os.Setenv("CLOUDACCOUNT_ADDR", "localhost:port")

		// run the function under test
		cloudAccountClient, err := NewCloudAccountClient(ctx, mockResolver)
		assert.NoError(t, err)
		assert.NotNil(t, cloudAccountClient)
	})

	t.Run("resolver error", func(t *testing.T) {
		mockResolver := NewMockResolver(func(ctx context.Context, s string) (string, error) {
			return "", errors.New("resolver error")
		})

		// set environment variable
		os.Setenv("CLOUDACCOUNT_ADDR", "")

		// run the function under test
		cloudAccountClient, err := NewCloudAccountClient(ctx, mockResolver)
		assert.Error(t, err)
		assert.Nil(t, cloudAccountClient)
	})
}

func TestGetCloudAccount(t *testing.T) {
	ctx := context.Background()

	t.Run("successful execution", func(t *testing.T) {
		mockAccount := v1.CloudAccount{
			Id:   "test-account-id",
			Type: v1.AccountType_ACCOUNT_TYPE_STANDARD,
		}
		mockGetById := func(ctx context.Context, in *v1.CloudAccountId, opts ...grpc.CallOption) (*v1.CloudAccount, error) {
			return &mockAccount, nil
		}
		mockCloudAccountServiceClient := NewMockCloudAccountServiceClient(nil, nil, mockGetById, nil, nil, nil, nil, nil, nil, nil, nil)
		cloudAccountClient := &CloudAccountSvcClient{
			CloudAccountClient: mockCloudAccountServiceClient,
		}

		// run the function under test
		cloudAccount, err := cloudAccountClient.GetCloudAccount(ctx, &v1.CloudAccountId{})
		assert.NoError(t, err)
		assert.Equal(t, &mockAccount, cloudAccount)
	})

	t.Run("error in cloud account client response", func(t *testing.T) {
		mockGetById := func(ctx context.Context, in *v1.CloudAccountId, opts ...grpc.CallOption) (*v1.CloudAccount, error) {
			return nil, errors.New("error in cloud account client response")
		}
		mockCloudAccountServiceClient := NewMockCloudAccountServiceClient(nil, nil, mockGetById, nil, nil, nil, nil, nil, nil, nil, nil)
		cloudAccountClient := &CloudAccountSvcClient{
			CloudAccountClient: mockCloudAccountServiceClient,
		}

		// run the function under test
		cloudAccount, err := cloudAccountClient.GetCloudAccount(ctx, &v1.CloudAccountId{})
		assert.Error(t, err)
		assert.Nil(t, cloudAccount)
	})
}

func TestGetCloudAccountType(t *testing.T) {
	ctx := context.Background()

	t.Run("successful execution", func(t *testing.T) {
		mockAccount := v1.CloudAccount{
			Id:   "test-account-id",
			Type: v1.AccountType_ACCOUNT_TYPE_STANDARD,
		}
		mockGetById := func(ctx context.Context, in *v1.CloudAccountId, opts ...grpc.CallOption) (*v1.CloudAccount, error) {
			return &mockAccount, nil
		}
		mockCloudAccountServiceClient := NewMockCloudAccountServiceClient(nil, nil, mockGetById, nil, nil, nil, nil, nil, nil, nil, nil)
		cloudAccountClient := &CloudAccountSvcClient{
			CloudAccountClient: mockCloudAccountServiceClient,
		}

		// run the function under test
		cloudAccountType, err := cloudAccountClient.GetCloudAccountType(ctx, &v1.CloudAccountId{})
		assert.NoError(t, err)
		assert.Equal(t, v1.AccountType_ACCOUNT_TYPE_STANDARD, cloudAccountType)
	})

	t.Run("error in cloud account client response", func(t *testing.T) {
		mockGetById := func(ctx context.Context, in *v1.CloudAccountId, opts ...grpc.CallOption) (*v1.CloudAccount, error) {
			return nil, errors.New("error in cloud account client response")
		}
		mockCloudAccountServiceClient := NewMockCloudAccountServiceClient(nil, nil, mockGetById, nil, nil, nil, nil, nil, nil, nil, nil)
		cloudAccountClient := &CloudAccountSvcClient{
			CloudAccountClient: mockCloudAccountServiceClient,
		}

		// run the function under test
		cloudAccountType, err := cloudAccountClient.GetCloudAccountType(ctx, &v1.CloudAccountId{})
		assert.Error(t, err)
		assert.Equal(t, v1.AccountType_ACCOUNT_TYPE_UNSPECIFIED, cloudAccountType)
	})
}

func TestGetAllCloudAccount(t *testing.T) {
	ctx := context.Background()

	t.Run("successful execution", func(t *testing.T) {
		mockAccounts := []*v1.CloudAccount{{
			Id:   "test-account-id",
			Type: v1.AccountType_ACCOUNT_TYPE_STANDARD,
		}, {
			Id:   "test-account-id-2",
			Type: v1.AccountType_ACCOUNT_TYPE_PREMIUM,
		}}
		mockCloudAccountServiceSearchClient := NewMockCloudAccountServiceSearchClient(false, mockAccounts)
		mockSearch := func(ctx context.Context, in *v1.CloudAccountFilter, opts ...grpc.CallOption) (v1.CloudAccountService_SearchClient, error) {
			return mockCloudAccountServiceSearchClient, nil
		}
		mockCloudAccountServiceClient := NewMockCloudAccountServiceClient(nil, nil, nil, nil, nil, nil, mockSearch, nil, nil, nil, nil)
		cloudAccountClient := &CloudAccountSvcClient{
			CloudAccountClient: mockCloudAccountServiceClient,
		}

		// run the function under test
		cloudAccounts, err := cloudAccountClient.GetAllCloudAccount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(cloudAccounts))
		assert.Equal(t, mockAccounts[0], cloudAccounts[0])
		assert.Equal(t, mockAccounts[1], cloudAccounts[1])
	})

	t.Run("error in cloud account client search", func(t *testing.T) {
		mockSearch := func(ctx context.Context, in *v1.CloudAccountFilter, opts ...grpc.CallOption) (v1.CloudAccountService_SearchClient, error) {
			return nil, errors.New("error in cloud account client search")
		}
		mockCloudAccountServiceClient := NewMockCloudAccountServiceClient(nil, nil, nil, nil, nil, nil, mockSearch, nil, nil, nil, nil)
		cloudAccountClient := &CloudAccountSvcClient{
			CloudAccountClient: mockCloudAccountServiceClient,
		}

		// run the function under test
		cloudAccounts, err := cloudAccountClient.GetAllCloudAccount(ctx)
		assert.Error(t, err)
		assert.Nil(t, cloudAccounts)
	})

	t.Run("failed to read from stream", func(t *testing.T) {
		mockCloudAccountServiceSearchClient := NewMockCloudAccountServiceSearchClient(true, nil)
		mockSearch := func(ctx context.Context, in *v1.CloudAccountFilter, opts ...grpc.CallOption) (v1.CloudAccountService_SearchClient, error) {
			return mockCloudAccountServiceSearchClient, nil
		}
		mockCloudAccountServiceClient := NewMockCloudAccountServiceClient(nil, nil, nil, nil, nil, nil, mockSearch, nil, nil, nil, nil)
		cloudAccountClient := &CloudAccountSvcClient{
			CloudAccountClient: mockCloudAccountServiceClient,
		}

		// run the function under test
		cloudAccounts, err := cloudAccountClient.GetAllCloudAccount(ctx)
		assert.Error(t, err)
		assert.Nil(t, cloudAccounts)
	})
}
