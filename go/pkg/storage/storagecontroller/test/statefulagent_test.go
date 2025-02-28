// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"

	"github.com/golang/mock/gomock"
	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	stcnt_api "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api/weka"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/test/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Statefulagent", func() {
	var (
		client             *sc.StorageControllerClient
		ctrl               *gomock.Controller
		mockS3Client       *mocks.MockS3ServiceClient
		mockStatefulClient *mocks.MockStatefulClientServiceClient
		ctx                context.Context
		clusterUUID        string
	)
	BeforeEach(func() {
		// Initialize StorageControllerClient
		ctrl = gomock.NewController(GinkgoT())
		mockStatefulClient = mocks.NewMockStatefulClientServiceClient(ctrl)
		mockS3Client = mocks.NewMockS3ServiceClient(ctrl)
		client = &sc.StorageControllerClient{
			S3ServiceClient:   mockS3Client,
			StatefulSvcClient: mockStatefulClient,
		}
		// Set up the test input data (payload) before each test
		ctx = context.Background()
		clusterUUID = "66efeaca-e493-4a39-b683-15978aac90d5"

	}) //BeforEach
	AfterEach(func() {
		ctrl.Finish()
	}) //AfterEach

	Context("Register Client", func() {
		It("Should register client succesfully", func() {
			// set mock expectation
			mockStatefulClient.EXPECT().CreateStatefulClient(gomock.Any(), gomock.Any()).Return(&stcnt_api.CreateStatefulClientResponse{
				StatefulClient: &stcnt_api.StatefulClient{
					Id: &stcnt_api.StatefulClientIdentifier{
						ClusterId: &api.ClusterIdentifier{
							Uuid: clusterUUID,
						},
					},
					Name: "sClient",
				},
			}, nil).Times(1)
			//make request with valid payload
			res, err := client.RegisterClient(ctx, &sc.CreateClientRequest{
				ClusterId: clusterUUID,
				Name:      "sClient",
				IPAddr:    "0.0.0.0",
			})
			//Should not have error
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
		})
	}) //Context
	Context("DeRegister Client", func() {
		It("Should de-register client successfully", func() {
			//set mock expectation
			mockStatefulClient.EXPECT().DeleteStatefulClient(gomock.Any(), gomock.Any()).Return(&stcnt_api.DeleteStatefulClientResponse{}, nil).Times(1)
			err := client.DeRegisterClient(ctx, clusterUUID, clusterUUID)
			Expect(err).To(BeNil())
		})
	})
	Context("Get Client", func() {
		It("Should get client successfully", func() {
			stateful := &stcnt_api.StatefulClient{
				Id: &stcnt_api.StatefulClientIdentifier{
					ClusterId: &api.ClusterIdentifier{
						Uuid: clusterUUID,
					},
				},
				Name: "sClient",
				Status: &stcnt_api.StatefulClient_PredefinedStatus{
					PredefinedStatus: stcnt_api.StatefulClient_STATUS_DEGRADED_UNSPECIFIED,
				},
			}
			//set mock expectation
			mockStatefulClient.EXPECT().GetStatefulClient(gomock.Any(), gomock.Any()).Return(&stcnt_api.GetStatefulClientResponse{
				StatefulClient: stateful,
			}, nil).Times(1)
			res, err := client.GetClient(ctx, clusterUUID, clusterUUID)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
			//Status Down
			stateful.Status = &stcnt_api.StatefulClient_PredefinedStatus{
				PredefinedStatus: stcnt_api.StatefulClient_STATUS_DOWN,
			}
			mockStatefulClient.EXPECT().GetStatefulClient(gomock.Any(), gomock.Any()).Return(&stcnt_api.GetStatefulClientResponse{
				StatefulClient: stateful,
			}, nil).Times(1)
			res, err = client.GetClient(ctx, clusterUUID, clusterUUID)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
			//Status Up
			stateful.Status = &stcnt_api.StatefulClient_PredefinedStatus{
				PredefinedStatus: stcnt_api.StatefulClient_STATUS_UP,
			}
			mockStatefulClient.EXPECT().GetStatefulClient(gomock.Any(), gomock.Any()).Return(&stcnt_api.GetStatefulClientResponse{
				StatefulClient: stateful,
			}, nil).Times(1)
			res, err = client.GetClient(ctx, clusterUUID, clusterUUID)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
			//Status Unkown
			stateful.Status = &stcnt_api.StatefulClient_CustomStatus{
				CustomStatus: "unkown",
			}
			mockStatefulClient.EXPECT().GetStatefulClient(gomock.Any(), gomock.Any()).Return(&stcnt_api.GetStatefulClientResponse{
				StatefulClient: stateful,
			}, nil).Times(1)
			res, err = client.GetClient(ctx, clusterUUID, clusterUUID)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
		})
	})
	Context("List Client", func() {
		It("Should list client successfully", func() {
			stateful := &stcnt_api.StatefulClient{
				Id: &stcnt_api.StatefulClientIdentifier{
					ClusterId: &api.ClusterIdentifier{
						Uuid: clusterUUID,
					},
				},
				Name: "sClient",
				Status: &stcnt_api.StatefulClient_PredefinedStatus{
					PredefinedStatus: stcnt_api.StatefulClient_STATUS_DEGRADED_UNSPECIFIED,
				},
			}
			names := []string{"sClient"}
			//set mock expectation
			mockStatefulClient.EXPECT().ListStatefulClients(gomock.Any(), gomock.Any()).Return(&stcnt_api.ListStatefulClientsResponse{
				StatefulClients: []*stcnt_api.StatefulClient{stateful},
			}, nil).Times(1)
			res, err := client.ListClients(ctx, clusterUUID, names)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
		})
	})
}) //Describe
