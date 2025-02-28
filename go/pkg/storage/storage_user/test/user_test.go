// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	clusterUUID string
)

var _ = Describe("FilesystemServiceServer", func() {
	Context("CreateOrUpdate", func() {
		It("should update a filesystem successfully", func() {
			ctx := context.Background()
			nsUserPwd, err := utils.GenerateRandomPassword()
			Expect(err).To(BeNil())
			req := &pb.FilesystemUserCreateOrUpdateRequest{
				ClusterUUID:        clusterUUID,
				NamespaceName:      "namespace1",
				NamespaceCredsPath: "path1",
				UserName:           "abc1",
				NewUserPassword:    nsUserPwd,
			}
			Expect(req).NotTo(BeNil())
			nsMockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
				Namespaces: []*api.Namespace{nsResponse},
			}, nil).Times(2)
			mockUserClient.EXPECT().ListUsers(gomock.Any(), gomock.Any()).Return(&api.ListUsersResponse{
				Users: []*api.User{userResponse},
			}, nil).Times(2)
			mockUserClient.EXPECT().UpdateUserPassword(gomock.Any(), gomock.Any()).Return(&api.UpdateUserPasswordResponse{}, nil).Times(1)
			resp, err := fsUser.CreateOrUpdate(ctx, req)
			Expect(resp).NotTo(BeNil())
			Expect(err).To(BeNil())
		})
		It("should create a filesystem user successfully", func() {
			ctx := context.Background()
			nsUserPwd2, err := utils.GenerateRandomPassword()
			Expect(err).To(BeNil())
			req2 := &pb.FilesystemUserCreateOrUpdateRequest{
				ClusterUUID:        clusterUUID,
				NamespaceName:      "namespace1",
				NamespaceCredsPath: "path1",
				UserName:           "abc1",
				NewUserPassword:    nsUserPwd2,
			}
			Expect(req2).NotTo(BeNil())
			nsMockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
				Namespaces: []*api.Namespace{nsResponse},
			}, nil).Times(2)
			mockUserClient.EXPECT().ListUsers(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).AnyTimes()
			mockUserClient.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
			resp, err := fsUser.CreateOrUpdate(ctx, req2)
			Expect(resp).NotTo(BeNil())
			Expect(err).To(BeNil())
		})
		It("should give error while creating new user", func() {
			ctx := context.Background()
			nsUserPwd3, err := utils.GenerateRandomPassword()
			Expect(err).To(BeNil())
			req3 := &pb.FilesystemUserCreateOrUpdateRequest{
				ClusterUUID:        clusterUUID,
				NamespaceName:      "namespace1",
				NamespaceCredsPath: "path1",
				UserName:           "abc2",
				NewUserPassword:    nsUserPwd3,
			}
			Expect(req3).NotTo(BeNil())
			nsMockClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any(), gomock.Any()).Return(&api.ListNamespacesResponse{
				Namespaces: []*api.Namespace{nsResponse},
			}, nil).Times(2)
			mockUserClient.EXPECT().ListUsers(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).AnyTimes()
			mockUserClient.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
			resp, err := fsUser.CreateOrUpdate(ctx, req3)
			Expect(resp).To(BeNil())
			Expect(err).ToNot(BeNil())
		})
	})
	Context("Ping", func() {
		It("Should be successful", func() {
			_, err := fsUser.PingFileUserPrivate(ctx, nil)
			Expect(err).To(BeNil())

		})
	})
})
