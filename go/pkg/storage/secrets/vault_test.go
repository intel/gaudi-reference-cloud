// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package secrets

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	vault "github.com/hashicorp/vault/api"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	v     *Vault
	vMock *MockVaultHelper
)

func TestYourFunction(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "YourFunction Suite")
}

var _ = BeforeSuite(func() {
	mockCtrl := gomock.NewController(GinkgoT())
	vMock = NewMockVaultHelper(mockCtrl)
	Expect(vMock).NotTo(BeNil())
	v = NewVaultClient(context.Background())
	Expect(v).NotTo(BeNil())
	v.vaultHelper = vMock

})
var _ = Describe("Vault functions", func() {
	defer GinkgoRecover()

	Context("validate vault client functions", func() {
		It("should return error if no roleid file is present", func() {
			ctx := context.Background()
			err := v.getRoleID()
			Expect(err).NotTo(BeNil())

			err = v.getSecretIDFile()
			Expect(err).To(BeNil())

			err = v.getVaultClient(ctx)
			Expect(err).NotTo(BeNil())

			err = v.getVaultAuthInfo(ctx)
			Expect(err).NotTo(BeNil())

			err3 := v.ValidateVaultClient(ctx)
			Expect(err3).NotTo(BeNil())

			vMock.EXPECT().getVaultAuthInfo(gomock.Any()).Return(nil)
			err = v.getAuthInfo(ctx)
			Expect(err).NotTo(BeNil())

			vMock.EXPECT().getVaultClient(gomock.Any()).Return(nil)
			err = v.getClient(ctx)
			Expect(err).NotTo(BeNil())
		})

	})
	Context("getStorageCredentials", func() {
		It("Should succeed", func() {
			myMap := make(map[string]interface{})
			// Adding key-value pairs to the map
			myMap["username"] = "John"
			myMap["password"] = "123"
			myMap["userId"] = "2"
			myMap["namespaceId"] = "11"

			resp := &vault.KVSecret{
				Data: myMap,
			}
			//Set up mock behavior
			mockCtrl := gomock.NewController(GinkgoT())
			defer mockCtrl.Finish()
			vMock := NewMockVaultHelper(mockCtrl)
			Expect(vMock).NotTo(BeNil())
			v := NewVaultClient(context.Background())
			Expect(v).NotTo(BeNil())
			v.vaultHelper = vMock

			vMock.EXPECT().getVaultSecrets(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(resp, nil).Times(1)
			res, err := v.getStorageVaultSecrets(context.Background(), "path/to/secret", true)
			Expect(err).To(BeNil())
			Expect(res).NotTo(BeNil())

			vMock.EXPECT().getVaultSecrets(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(resp, nil).Times(1)
			user, pass, _, _, err2 := v.GetStorageCredentials(context.Background(), "path/to/secret", true)
			Expect(err2).To(BeNil())
			Expect(user).NotTo(BeNil())
			Expect(pass).NotTo(BeNil())
		}) //It

		It("Should fail", func() {
			By("Setting up mocks")
			//Set up mock behavior
			mockCtrl := gomock.NewController(GinkgoT())
			defer mockCtrl.Finish()
			vMock := NewMockVaultHelper(mockCtrl)
			Expect(vMock).NotTo(BeNil())
			v := NewVaultClient(context.Background())
			Expect(v).NotTo(BeNil())
			v.vaultHelper = vMock

			By("Having inner vault api call return error")
			vMock.EXPECT().getVaultSecrets(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
			user, pass, _, _, err2 := v.GetStorageCredentials(context.Background(), "path/to/secret", true)
			Expect(err2).NotTo(BeNil())
			Expect(user).NotTo(BeNil())
			Expect(pass).NotTo(BeNil())
		})
	}) //Context

	Context("PutStorageSecrets", func() {
		It("Should Succeed", func() {
			kv := make(map[string]interface{})
			// Adding key-value pairs to the map
			kv["username"] = "John"
			kv["password"] = "123"
			kv["userId"] = "2"
			kv["namespaceId"] = "11"
			res := &vault.KVSecret{
				Data: kv,
			}
			//Set up mock behavior
			mockCtrl := gomock.NewController(GinkgoT())
			defer mockCtrl.Finish()
			vMock := NewMockVaultHelper(mockCtrl)
			Expect(vMock).NotTo(BeNil())
			v := NewVaultClient(context.Background())
			Expect(v).NotTo(BeNil())
			v.vaultHelper = vMock

			By("Having vault API calls succeed")
			vMock.EXPECT().putVaultSecrets(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(res, nil)
			secret, err := v.PutStorageSecrets(context.Background(), "path/to/secret", kv, true)
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			vMock.EXPECT().putVaultSecrets(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(res, nil)
			ctrlPlaneSecret, err2 := v.PutControlplaneVaultSecrets(context.Background(), "path/to/secret", kv, true)
			Expect(err2).To(BeNil())
			Expect(ctrlPlaneSecret).NotTo(BeNil())

		})
		It("Should Fail", func() {
			kv := make(map[string]interface{})
			// Adding key-value pairs to the map
			kv["username"] = "John"
			kv["password"] = "123"
			kv["userId"] = "2"
			kv["namespaceId"] = "11"

			By("Setting up mocks")
			//Set up mock behavior
			mockCtrl := gomock.NewController(GinkgoT())
			defer mockCtrl.Finish()
			vMock := NewMockVaultHelper(mockCtrl)
			Expect(vMock).NotTo(BeNil())
			v := NewVaultClient(context.Background())
			Expect(v).NotTo(BeNil())
			v.vaultHelper = vMock
			//v.secret.Auth.Renewable = false
			secret := &vault.Secret{
				Auth: &vault.SecretAuth{
					Renewable: false,
				},
			}
			v.secret = secret

			By("Having vault API calls return error")
			vMock.EXPECT().putVaultSecrets(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
			_, err := v.PutStorageSecrets(context.Background(), "path/to/secret", kv, true)
			Expect(err).NotTo(BeNil())

			vMock.EXPECT().putVaultSecrets(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))
			_, err2 := v.PutControlplaneVaultSecrets(context.Background(), "path/to/secret", kv, true)
			Expect(err2).NotTo(BeNil())

			//vMock.EXPECT().getVaultAuthInfo(gomock.Any()).Return(nil)
			_, err3 := v.putVaultSecrets(context.Background(), "", "", "path/to/Secret", kv, true)
			Expect(err3).NotTo(BeNil())
		})
	})
	Context("getEnrollBasicAuth", func() {
		It("Should succeed", func() {
			myMap := make(map[string]interface{})
			// Adding key-value pairs to the map
			myMap["username"] = "John"
			myMap["password"] = "123"
			myMap["userId"] = "2"
			myMap["namespaceId"] = "11"

			resp := &vault.KVSecret{
				Data: myMap,
			}
			By("Setting up mocks")
			//Set up mock behavior
			mockCtrl := gomock.NewController(GinkgoT())
			defer mockCtrl.Finish()
			vMock := NewMockVaultHelper(mockCtrl)
			Expect(vMock).NotTo(BeNil())
			v := NewVaultClient(context.Background())
			Expect(v).NotTo(BeNil())
			v.vaultHelper = vMock

			By("Having vault API calls succeed")
			vMock.EXPECT().getVaultSecrets(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(resp, nil).Times(1)
			secret, err := v.GetControlPlaneSecrets(context.Background(), "path/to/secret", true)
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			vMock.EXPECT().getVaultSecrets(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(resp, nil).Times(1)
			user, pass, err2 := v.GetEnrollBasicAuth(context.Background(), "path/to/secret", true)
			Expect(err2).To(BeNil())
			Expect(user).NotTo(BeNil())
			Expect(pass).NotTo(BeNil())

		})
	})

	Context("Utils", func() {
		It("should succeed", func() {
			ctx := context.Background()
			By("Setting up mocks")
			//Set up mock behavior
			mockCtrl := gomock.NewController(GinkgoT())
			defer mockCtrl.Finish()
			vMock := NewMockVaultHelper(mockCtrl)
			Expect(vMock).NotTo(BeNil())
			v := NewVaultClient(context.Background())
			Expect(v).NotTo(BeNil())
			v.vaultHelper = vMock

			vMock.EXPECT().getVaultAuthInfo(ctx).Return(nil)
			err := v.getAuthInfo(ctx)
			Expect(err).To(BeNil())

			vMock.EXPECT().getVaultClient(ctx).Return(nil)
			err = v.getClient(ctx)
			Expect(err).To(BeNil())
		})
	})

})
