// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storage_scheduler/pkg/server"
	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/test/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	mockCtrl *gomock.Controller
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VM Instance Scheduler Suite")
}

// Constructor returns mock StorageSchedulerServiceServer
func NewMockStorageSchedulerServer() (*server.StorageSchedulerServiceServer, error) {
	sClient := &sc.StorageControllerClient{}
	sClient.Init(context.Background(), "localhost", false)

	mockCtrl = gomock.NewController(GinkgoT())
	mockClient = mocks.NewMockClusterServiceClient(mockCtrl)
	mockStatefulClient = mocks.NewMockStatefulClientServiceClient(mockCtrl)
	mockNsClient = mocks.NewMockNamespaceServiceClient(mockCtrl)
	mockFsClient = mocks.NewMockFilesystemServiceClient(mockCtrl)
	mockKmsClient = NewMockStorageKMSPrivateServiceClient()

	sClient.ClusterSvcClient = mockClient
	sClient.NamespaceSvcClient = mockNsClient
	sClient.WekaFilesystemSvcClient = mockFsClient
	sClient.StatefulSvcClient = mockStatefulClient
	client, err := server.NewStorageSchedulerService(sClient, mockKmsClient, false)
	if err != nil {
		return nil, errors.New("error intializing scheduler struct")
	}
	return client, nil
}
