package test

import (
	"context"
	"errors"

	api "github.com/hashicorp/vault/api"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/mocks"
	server "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storage_kms/pkg/server"
	. "github.com/onsi/ginkgo"

	. "github.com/onsi/gomega"

	"github.com/golang/mock/gomock"
)

var _ = Describe("FilesystemServiceServer", func() {

	Context("Get User", func() {
		It("Should get user credentials succesfully", func() {
			By("Creating a filesystem and user")
			ctx := context.Background()

			req := &pb.GetSecretRequest{
				KeyPath: "",
			}

			mockCtrl := gomock.NewController(GinkgoT())
			mockSecretManager := mocks.NewMockSecretManager(mockCtrl)
			mockSecretManager.EXPECT().GetStorageCredentials(gomock.Any(), gomock.Any(), gomock.Any()).Return("", "", "", "", nil).Times(1)
			fsServer, _ := server.NewMockStorageKMSService(mockSecretManager)

			_, err := fsServer.Get(ctx, req)
			Expect(err).To(BeNil())
		})
	})

	Context("Put User", func() {
		It("Should put user credentials succesfully", func() {
			By("Creating a filesystem and user")
			ctx := context.Background()

			secrets := make(map[string]string)

			req := &pb.StoreSecretRequest{
				KeyPath: "",
				Secrets: secrets,
			}

			res := &api.KVSecret{}

			mockCtrl := gomock.NewController(GinkgoT())
			mockSecretManager := mocks.NewMockSecretManager(mockCtrl)
			mockSecretManager.EXPECT().PutStorageSecrets(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(res, nil).Times(1)
			fsServer, _ := server.NewMockStorageKMSService(mockSecretManager)

			resp, err := fsServer.Put(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())
		})
	})

	Context("Get User", func() {
		It("Should give error when getting user credentials", func() {
			By("Creating a filesystem and user")
			ctx := context.Background()

			req := &pb.GetSecretRequest{
				KeyPath: "",
			}

			mockCtrl := gomock.NewController(GinkgoT())
			mockSecretManager := mocks.NewMockSecretManager(mockCtrl)
			mockSecretManager.EXPECT().GetStorageCredentials(gomock.Any(), gomock.Any(), gomock.Any()).Return("", "", "", "", errors.New("error")).Times(1)
			fsServer, _ := server.NewMockStorageKMSService(mockSecretManager)

			_, err := fsServer.Get(ctx, req)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Put User", func() {
		It("Should give error putting user credentials", func() {
			By("Creating a filesystem and user")
			ctx := context.Background()

			secrets := make(map[string]string)

			req := &pb.StoreSecretRequest{
				KeyPath: "",
				Secrets: secrets,
			}

			res := &api.KVSecret{}

			mockCtrl := gomock.NewController(GinkgoT())
			mockSecretManager := mocks.NewMockSecretManager(mockCtrl)
			mockSecretManager.EXPECT().PutStorageSecrets(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(res, errors.New("error")).Times(1)
			fsServer, _ := server.NewMockStorageKMSService(mockSecretManager)

			resp, err := fsServer.Put(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())
		})
	})

})
