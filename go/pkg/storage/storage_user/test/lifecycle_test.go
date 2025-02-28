package server

import (
	"errors"

	"github.com/golang/mock/gomock"
	sc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BucketLifecycleServiceServer", func() {
	Context("Create rule", func() {
		It("Should create rule successfully", func() {
			//set mock
			mockS3ServiceClient.EXPECT().CreateLifecycleRules(gomock.Any(), gomock.Any()).Return(&sc.CreateLifecycleRulesResponse{
				LifecycleRules: []*sc.LifecycleRule{lcRule},
			}, nil).Times(1)
			res, err := lcServer.CreateOrUpdateLifecycleRule(ctx, lcCreate)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())
		})
		It("Should fail to create rule", func() {
			By("Throwing error from sds")
			//set mock
			mockS3ServiceClient.EXPECT().CreateLifecycleRules(gomock.Any(), gomock.Any()).Return(nil, errors.New("some error")).Times(1)
			res, err := lcServer.CreateOrUpdateLifecycleRule(ctx, lcCreate)
			Expect(err).NotTo(BeNil())
			Expect(res).To(BeNil())
		})
	})
	Context("Ping", func() {
		It("Should be successful", func() {
			_, err := lcServer.PingBucketLifecyclePrivate(ctx, nil)
			Expect(err).To(BeNil())
		})

	})
})
