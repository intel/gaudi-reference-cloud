// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"errors"

	gomock "github.com/golang/mock/gomock"
	mock "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storage_scheduler/pkg/server"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	stcnt_api "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api/weka"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/test/mocks"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	schedulerSvc       *server.StorageSchedulerServiceServer
	mockStatefulClient *mocks.MockStatefulClientServiceClient
)
var _ = Describe("FilesystemServiceServer", func() {
	BeforeEach(func() {
		// Initialize calling object
		var err error
		schedulerSvc, err = NewMockStorageSchedulerServer()
		Expect(err).To(BeNil())
	})

	Context("RegisterAgent", func() {
		It("Should RegisterAgent successfully", func() {
			By("Providing valid input and recieving no errors from RegisterAgent internal function calls")
			mockStatefulClient.EXPECT().CreateStatefulClient(gomock.Any(), gomock.Any()).Return(&stcnt_api.CreateStatefulClientResponse{
				StatefulClient: &stcnt_api.StatefulClient{
					Id: &stcnt_api.StatefulClientIdentifier{
						ClusterId: &api.ClusterIdentifier{
							Uuid: "1bf067b5-2f69-4f7f-8468-a5a3b9e44f4c",
						},
					},
					Name: "sClient",
				},
			}, nil).Times(1)

			req := &mock.RegisterAgentRequest{
				ClusterId: "918b5026-d516-48c8-bfd3-5998547265b2",
				Name:      "sClient",
				IpAddr:    "127.0.0.1",
			}
			out, err := schedulerSvc.RegisterAgent(context.Background(), req)
			Expect(err).To(BeNil())
			Expect(out).NotTo(BeNil())

			By("Providing valid input and recieving errors from RegisterAgent internal function calls")
			mockStatefulClient.EXPECT().CreateStatefulClient(gomock.Any(), gomock.Any()).Return(nil, errors.New("error creating stateful client")).Times(1)

			_, err = schedulerSvc.RegisterAgent(context.Background(), req)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("DeRegisterAgent", func() {
		It("Should de register agent successfully", func() {
			By("Providing valid input and recieving no errors from de register agent internal function calls")
			mockStatefulClient.EXPECT().DeleteStatefulClient(gomock.Any(), gomock.Any()).Return(&stcnt_api.DeleteStatefulClientResponse{}, nil).Times(1)

			req := &mock.DeRegisterAgentRequest{
				ClusterId: "918b5026-d516-48c8-bfd3-5998547265b2",
				ClientId:  "testIKS",
			}
			out, err := schedulerSvc.DeRegisterAgent(context.Background(), req)
			Expect(err).To(BeNil())
			Expect(out).NotTo(BeNil())

			By("Providing valid input and recieving errors from de register agent internal function calls")
			mockStatefulClient.EXPECT().DeleteStatefulClient(gomock.Any(), gomock.Any()).Return(nil, errors.New("error deleting stateful client")).Times(1)
			_, err = schedulerSvc.DeRegisterAgent(context.Background(), req)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("GetRegisteredAgent", func() {
		It("Should get registered agent successfully", func() {
			By("Providing valid input and recieving no errors from get registered agent internal function calls")
			stateful := &stcnt_api.StatefulClient{
				Id: &stcnt_api.StatefulClientIdentifier{
					ClusterId: &api.ClusterIdentifier{
						Uuid: "1bf067b5-2f69-4f7f-8468-a5a3b9e44f4c",
					},
				},
				Name: "sClient",
				Status: &stcnt_api.StatefulClient_PredefinedStatus{
					PredefinedStatus: stcnt_api.StatefulClient_STATUS_DEGRADED_UNSPECIFIED,
				},
			}
			mockStatefulClient.EXPECT().GetStatefulClient(gomock.Any(), gomock.Any()).Return(&stcnt_api.GetStatefulClientResponse{
				StatefulClient: stateful,
			}, nil).Times(1)

			req := &mock.GetRegisterAgentRequest{
				ClusterId: "918b5026-d516-48c8-bfd3-5998547265b2",
				ClientId:  "testIKS",
			}
			out, err := schedulerSvc.GetRegisteredAgent(context.Background(), req)
			Expect(err).To(BeNil())
			Expect(out).NotTo(BeNil())

			By("Providing valid input and recieving errors from get registered agent internal function calls")

			mockStatefulClient.EXPECT().GetStatefulClient(gomock.Any(), gomock.Any()).Return(nil, errors.New("error getting stateful client")).Times(1)
			_, err = schedulerSvc.GetRegisteredAgent(context.Background(), req)
			Expect(err).NotTo(BeNil())

		})
	})

	Context("ListRegisteredAgents", func() {
		It("Should list registered agent successfully", func() {
			By("Providing valid input and recieving no errors from list registered agent internal function calls")
			stateful := &stcnt_api.StatefulClient{
				Id: &stcnt_api.StatefulClientIdentifier{
					ClusterId: &api.ClusterIdentifier{
						Uuid: "1bf067b5-2f69-4f7f-8468-a5a3b9e44f4c",
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

			req := &mock.ListRegisteredAgentRequest{
				ClusterId: "918b5026-d516-48c8-bfd3-5998547265b2",
				Names:     names,
			}

			// Create a mock instance of WekaStatefulAgentPrivateService_ListRegisteredAgentsServer
			ctrl := gomock.NewController(GinkgoT())
			defer ctrl.Finish()
			mockStream := mock.NewMockWekaStatefulAgentPrivateService_ListRegisteredAgentsServer(ctrl)
			mockStream.EXPECT().Context().Return(context.Background()).AnyTimes()
			mockStream.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()
			err := schedulerSvc.ListRegisteredAgents(req, mockStream)
			Expect(err).To(BeNil())

			By("Providing valid input and recieving errors from list registered agent internal function calls")
			mockStatefulClient.EXPECT().ListStatefulClients(gomock.Any(), gomock.Any()).Return(nil, errors.New("error listing stateful client")).Times(1)
			err = schedulerSvc.ListRegisteredAgents(req, mockStream)
			Expect(err).NotTo(BeNil())

		})
	})
})
