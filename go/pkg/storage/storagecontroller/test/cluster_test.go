// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"
	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/test/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cluster", func() {
	var (
		client           *sc.StorageControllerClient
		ctrl             *gomock.Controller
		mockClient       *mocks.MockClusterServiceClient
		mockListResponse *api.ListClustersResponse
		mockGetResponse  *api.GetClusterResponse
		clusterInfo      *sc.ClusterInfo
		ctx              context.Context
		clusterUUID      string
	)
	BeforeEach(func() {
		// Initialize StorageControllerClient
		ctrl = gomock.NewController(GinkgoT())
		mockClient = mocks.NewMockClusterServiceClient(ctrl)
		client = &sc.StorageControllerClient{
			ClusterSvcClient: mockClient,
		}
		// Set up the test input data (payload) before each test
		ctx = context.Background()
		clusterUUID = "66efeaca-e493-4a39-b683-15978aac90d5"
		cluster := api.Cluster{
			Id: &api.ClusterIdentifier{
				Uuid: clusterUUID,
			},
			Name:     "Backend1",
			Location: "Location1",
			Type:     api.Cluster_TYPE_WEKA,
			Health: &api.Cluster_Health{
				Status: api.Cluster_Health_STATUS_HEALTHY,
			},
			Capacity: &api.Cluster_Capacity{
				Storage: &api.Cluster_Capacity_Storage{
					TotalBytes:     2000000000,
					AvailableBytes: 1000000000,
				},
				Namespaces: &api.Cluster_Capacity_Namespaces{
					TotalCount:     256,
					AvailableCount: 250,
				},
			},
		}

		mockListResponse = &api.ListClustersResponse{
			Clusters: []*api.Cluster{
				&cluster,
			},
		}
		mockGetResponse = &api.GetClusterResponse{
			Cluster: &cluster,
		}

		clusterInfo = &sc.ClusterInfo{
			Name:               "Backend1",
			Addr:               "",
			UUID:               clusterUUID,
			DCName:             "Location1",
			DCLocation:         "Location1",
			AvailableCapacity:  1000000000,
			TotalCapacity:      2000000000,
			AvailableNamespace: 250,
			TotalNamespace:     256,
			Status:             "Online",
			Category:           "",
			S3Endpoint:         "",
			WekaBackend:        "",
			Type:               "Weka",
		}
		Expect(clusterInfo).ToNot(BeNil())
	}) //BeforEach
	AfterEach(func() {
		ctrl.Finish()
	}) //AfterEach
	Describe("ListClusters", func() {
		Context("when the clusters exist", func() {
			It("should get list of clusters with UUIDs", func() {
				// Set up expectations
				mockClient.EXPECT().ListClusters(gomock.Any(), gomock.Any()).Return(mockListResponse, nil).Times(1)

				clusters, err := client.GetClusters(ctx)
				Expect(err).To(BeNil())
				Expect(clusters).ToNot(BeNil())
				Expect(clusters).Should(ContainElement(*clusterInfo))
			})

		})
		Context("when there is an error in ListClusters", func() {
			It("should return an error", func() {
				// Set up expectations
				mockClient.EXPECT().ListClusters(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error")).Times(1)

				clusters, err := client.GetClusters(ctx)
				Expect(err).ToNot(BeNil())
				Expect(clusters).To(BeNil())
			})
		})

	}) //Describe
	Describe("GetCluster", func() {
		Context("when the cluster status is obtained", func() {
			It("should get get of cluster with UUIDs", func() {
				// Set up expectations
				mockClient.EXPECT().GetCluster(gomock.Any(), gomock.Any()).Return(mockGetResponse, nil).Times(1)

				cluster, err := client.GetClusterStatus(ctx, clusterUUID)
				Expect(err).To(BeNil())
				Expect(cluster).ToNot(BeNil())
				Expect(cluster).To(BeIdenticalTo(*clusterInfo))
			})
		})

		Context("when there is an error in GetCluster", func() {
			It("should return an error", func() {
				// Set up expectations
				mockClient.EXPECT().GetCluster(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error")).Times(1)

				cluster, err := client.GetClusterStatus(ctx, clusterUUID)
				Expect(err).ToNot(BeNil())
				Expect(cluster).To(BeZero())
			})
		})
	})
}) //Cluster
