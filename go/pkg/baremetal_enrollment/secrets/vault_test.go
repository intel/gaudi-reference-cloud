// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package secrets

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	vault "github.com/hashicorp/vault/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	log.SetOutput(GinkgoWriter)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vault Secrets Test Suite")
}

var Any = gomock.Any()

var _ = Describe("Vault Secret Task", func() {
	var (
		ctx context.Context

		fakeApproleID     string
		fakeSecretIDFile  string
		fakeVaultAddr     string
		fakeVaultToken    string
		fakeSecretPath    string
		fakeEnvPath       string
		fakeDefaultPath   string
		fakeSSHPrivateKey string
		fakeRenewToken    bool
		fakeKVSecret      *vault.KVSecret

		mockCtrl        *gomock.Controller
		mockVaultHelper *MockVaultHelper

		vaultHelperClient *Vault

		config *vault.Config
		vac    *vault.Client
		vas    *vault.Secret
		sa     *vault.SecretAuth
	)

	BeforeEach(func() {
		ctx = context.Background()

		fakeApproleID = "tmp_role_id"
		fakeSecretIDFile = "tmp_secret_id"
		fakeVaultAddr = "http://127.0.0.1:45678/vault"
		fakeVaultToken = "111111111111111111111111111"
		fakeSecretPath = "us-dev-1/fake/path"
		fakeEnvPath = "fakeEnv"
		fakeDefaultPath = "default"
		fakeRenewToken = false
		fakeSSHPrivateKey = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAlwAAAAdzc2gtcn
NhAAAAAwEAAQAAAIEAqIUQXPInoV4R/WM4AMr992ws31HR8CvHnymvu1qZVZK4il2rTFTH
GC8CRbnF6NjVMsJTr4eBNvQqvpS3i3H5JLUkQcROyN0bikTzhmZ5fk9QxQBfkQVKg4tj5B
xn5SUXAt7PFEg65GxuxVdDcGUZXC3aVPOTm7Fdeuvz1V9+g4sAAAIYyKAjiMigI4gAAAAH
c3NoLXJzYQAAAIEAqIUQXPInoV4R/WM4AMr992ws31HR8CvHnymvu1qZVZK4il2rTFTHGC
8CRbnF6NjVMsJTr4eBNvQqvpS3i3H5JLUkQcROyN0bikTzhmZ5fk9QxQBfkQVKg4tj5Bxn
5SUXAt7PFEg65GxuxVdDcGUZXC3aVPOTm7Fdeuvz1V9+g4sAAAADAQABAAAAgHZkFVzXGx
R5HDZh8EROWCHtM5Eo0E7k0vd0t+rt+W9vBorex6t2m/DXhccqfmnZe96PO2/DyPmsjCMc
I96pkZgecKHa3lHps7+hG+TfBaLmIL4BrK60/Md/DbDEIIyqNRC05Ls/rFpRY20Af1D9ms
D9HQkJiNnvQqVdYJTBGwBxAAAAQHqYrnWq7wxpxHqwFOsP+dFf0g5VxsXBhnAK26FgUpQK
DXmNLrR1nAMMann5zfpue2IPV5CGk3789HdhdHBnISgAAABBANMgd7K/WlWjHfrhujbZSd
JqYwh69BGLfji0cWgWrKHOA9UeF9BrgRKFAahZ9Yr6Fy6cIm98eogLuK8Ozt2bjfUAAABB
AMxWUooaaJ/K1tEd6ze38oBfF7RsJRxWiwXngaPRjjZvM1CFATP3rUqaGLWX8ubHSFV4MF
J2DiJMm0iMimVJW38AAAAdc2RwQGE0YmYwMTFkMzRmMS5qZi5pbnRlbC5jb20BAgMEBQY=
-----END OPENSSH PRIVATE KEY-----`

		createFile(fakeApproleID)
		createFile(fakeSecretIDFile)

		fakeKVSecret = &vault.KVSecret{}

		Expect(os.Setenv("VAULT_APPROLE_ROLE_ID_FILE", fakeApproleID)).To(Succeed())
		Expect(os.Setenv("VAULT_APPROLE_SECRET_ID_FILE", fakeSecretIDFile)).To(Succeed())
		Expect(os.Setenv("VAULT_ADDR", fakeVaultAddr)).To(Succeed())
		Expect(os.Setenv("VAULT_TOKEN", fakeVaultToken)).To(Succeed())

		// create mock objects
		mockCtrl = gomock.NewController(GinkgoT())
		mockVaultHelper = NewMockVaultHelper(mockCtrl)

		config = vault.DefaultConfig()
		vac, _ = vault.NewClient(config)
		sa = &vault.SecretAuth{Renewable: true, LeaseDuration: int(time.Millisecond)}
		vas = &vault.Secret{Renewable: true, Auth: sa}

		vaultHelperClient = &Vault{
			client:       vac,
			secret:       vas,
			roleID:       fakeApproleID,
			secretIDFile: fakeSecretIDFile,
			vaultHelper:  mockVaultHelper,
			validate:     fakeRenewToken,
		}

	})

	AfterEach(func() {
		mockCtrl.Finish()
		deleteFile(fakeApproleID)
		deleteFile(fakeSecretIDFile)
	})

	Describe("preparing for vault secret test", func() {
		It("should initialize dependencies", func() {
			By("initializing Vault secret manager client")
			Expect(mockVaultHelper).NotTo(BeNil())
		})
	})
	Describe("get control plane secrets", func() {
		When("returns no errors as get operation is succeeded", func() {
			It("should return kv secrets and no errors", func() {
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(fakeKVSecret, nil).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				secret, err := vaultHelperClient.GetControlPlaneSecrets(ctx, fakeSecretPath)
				fmt.Printf("secret: %+v, error: %+v\n", secret, err)
				Expect(secret).NotTo(BeNil())
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("returns errors as get operation is failed", func() {
			It("should return errors", func() {
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(nil, fmt.Errorf("failed to get the Control plane secrets")).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				secret, err := vaultHelperClient.GetControlPlaneSecrets(ctx, fakeSecretPath)
				fmt.Printf("secret: %+v, error: %+v\n", secret, err)
				Expect(secret).Should(BeNil())
				Expect(err).Should(HaveOccurred())
			})
		})
	})
	Describe("get BMC vault secrets", func() {
		When("returns no errors as get operation is succeeded", func() {
			It("should return kv secrets and no errors", func() {
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(fakeKVSecret, nil).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				secret, err := vaultHelperClient.getBMCVaultSecrets(ctx, fakeSecretPath)
				fmt.Printf("secret: %+v, error: %+v\n", secret, err)
				Expect(secret).NotTo(BeNil())
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("returns errors", func() {
			It("returns errors as get operation is failed", func() {
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(nil, fmt.Errorf("failed to get the Control plane secrets")).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				secret, err := vaultHelperClient.getBMCVaultSecrets(ctx, fakeSecretPath)
				fmt.Printf("secret: %+v, error: %+v\n", secret, err)
				Expect(secret).Should(BeNil())
				Expect(err).Should(HaveOccurred())
			})
		})
	})
	Describe("get BMC credentials", func() {
		When("username and password are missing", func() {
			It("should return errors as username and password is missing", func() {
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(fakeKVSecret, nil).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				username, password, err := vaultHelperClient.GetBMCCredentials(ctx, fakeSecretPath)
				fmt.Printf("username: %+v, password: %+v,error: %+v\n", username, password, err)
				Expect(username).To(Equal(""))
				Expect(password).To(Equal(""))
				Expect(err).Should(HaveOccurred())
			})
		})
		When("username is present but password is missing", func() {
			It("should return user but errors because no password", func() {
				fakeUserData := map[string]interface{}{"username": "fakeUser"}
				fakeKVSecret.Data = fakeUserData
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(fakeKVSecret, nil).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				username, password, err := vaultHelperClient.GetBMCCredentials(ctx, fakeSecretPath)
				fmt.Printf("username: %+v, password: %+v,error: %+v\n", username, password, err)
				Expect(username).To(Equal(""))
				Expect(password).To(Equal(""))
				Expect(err).Should(HaveOccurred())
			})
		})
		When("username and password are present", func() {
			It("should return user and password", func() {
				fakeUserData := map[string]interface{}{"username": "fakeUser", "password": "fakePassword"}
				fakeKVSecret.Data = fakeUserData
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(fakeKVSecret, nil).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				username, password, err := vaultHelperClient.GetBMCCredentials(ctx, fakeSecretPath)
				fmt.Printf("username: %+v, password: %+v,error: %+v\n", username, password, err)
				Expect(username).To(Equal("fakeUser"))
				Expect(password).To(Equal("fakePassword"))
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("failed to get the secret", func() {
			It("should return error as secret is failed", func() {
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(nil, fmt.Errorf("failed to get the vault KVv2 secret")).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				username, password, err := vaultHelperClient.GetBMCCredentials(ctx, fakeSecretPath)
				fmt.Printf("username: %+v, password: %+v,error: %+v\n", username, password, err)
				Expect(username).To(Equal(""))
				Expect(password).To(Equal(""))
				Expect(err).Should(HaveOccurred())
			})
		})
	})
	Describe("get Enrollment apiservice credentials", func() {
		When("username and password are missing", func() {
			It("should return errors as username and password is missing", func() {
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(fakeKVSecret, nil).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				username, password, err := vaultHelperClient.GetEnrollBasicAuth(ctx, fakeSecretPath, fakeRenewToken)
				fmt.Printf("username: %+v, password: %+v,error: %+v\n", username, password, err)
				Expect(username).To(Equal(""))
				Expect(password).To(Equal(""))
				Expect(err).Should(HaveOccurred())
			})
		})
		When("username is present but password is missing", func() {
			It("should return user but errors because no password", func() {
				fakeUserData := map[string]interface{}{"username": "fakeUser"}
				fakeKVSecret.Data = fakeUserData
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(fakeKVSecret, nil).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				username, password, err := vaultHelperClient.GetEnrollBasicAuth(ctx, fakeSecretPath, fakeRenewToken)
				fmt.Printf("username: %+v, password: %+v,error: %+v\n", username, password, err)
				Expect(username).To(Equal(""))
				Expect(password).To(Equal(""))
				Expect(err).Should(HaveOccurred())
			})
		})
		When("username and password are present", func() {
			It("should return user and password", func() {
				fakeUserData := map[string]interface{}{"username": "fakeUser", "password": "fakePassword"}
				fakeKVSecret.Data = fakeUserData
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(fakeKVSecret, nil).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				username, password, err := vaultHelperClient.GetEnrollBasicAuth(ctx, fakeSecretPath, fakeRenewToken)
				fmt.Printf("username: %+v, password: %+v,error: %+v\n", username, password, err)
				Expect(username).To(Equal("fakeUser"))
				Expect(password).To(Equal("fakePassword"))
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("failed to get the secret", func() {
			It("should return error as secret is failed", func() {
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(nil, fmt.Errorf("failed to get the vault KVv2 secret")).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				username, password, err := vaultHelperClient.GetEnrollBasicAuth(ctx, fakeSecretPath, fakeRenewToken)
				fmt.Printf("username: %+v, password: %+v,error: %+v\n", username, password, err)
				Expect(username).To(Equal(""))
				Expect(password).To(Equal(""))
				Expect(err).Should(HaveOccurred())
			})
		})
	})
	Describe("get DDI credentials", func() {
		When("username and password are missing", func() {
			It("should return errors as username and password is missing", func() {
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(fakeKVSecret, nil).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				username, password, err := vaultHelperClient.GetDDICredentials(ctx, fakeSecretPath)
				fmt.Printf("username: %+v, password: %+v,error: %+v\n", username, password, err)
				Expect(username).To(Equal(""))
				Expect(password).To(Equal(""))
				Expect(err).Should(HaveOccurred())
			})
		})
		When("username is present but password is missing", func() {
			It("should return user but errors because no password", func() {
				fakeUserData := map[string]interface{}{"username": "fakeUser"}
				fakeKVSecret.Data = fakeUserData
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(fakeKVSecret, nil).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				username, password, err := vaultHelperClient.GetDDICredentials(ctx, fakeSecretPath)
				fmt.Printf("username: %+v, password: %+v,error: %+v\n", username, password, err)
				Expect(username).To(Equal(""))
				Expect(password).To(Equal(""))
				Expect(err).Should(HaveOccurred())
			})
		})
		When("username and password are present", func() {
			It("should return user and password", func() {
				fakeUserData := map[string]interface{}{"username": "fakeUser", "password": "fakePassword"}
				fakeKVSecret.Data = fakeUserData
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(fakeKVSecret, nil).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				username, password, err := vaultHelperClient.GetDDICredentials(ctx, fakeSecretPath)
				fmt.Printf("username: %+v, password: %+v,error: %+v\n", username, password, err)
				Expect(username).To(Equal("fakeUser"))
				Expect(password).To(Equal("fakePassword"))
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("failed to get the secret", func() {
			It("should return error as secret is failed", func() {
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(nil, fmt.Errorf("failed to get the vault KVv2 secret")).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				username, password, err := vaultHelperClient.GetDDICredentials(ctx, fakeSecretPath)
				fmt.Printf("username: %+v, password: %+v,error: %+v\n", username, password, err)
				Expect(username).To(Equal(""))
				Expect(password).To(Equal(""))
				Expect(err).Should(HaveOccurred())
			})
		})
	})
	Describe("get BIOS credentials", func() {
		When("BIOS password is missing", func() {
			It("should return errors as password is missing", func() {
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(fakeKVSecret, nil).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				password, err := vaultHelperClient.GetBMCBIOSPassword(ctx, fakeSecretPath)
				fmt.Printf("password: %+v,error: %+v\n", password, err)
				Expect(password).To(Equal(""))
				Expect(err).Should(HaveOccurred())
			})
		})
		When("BIOS password is present", func() {
			It("should return password", func() {
				fakeUserData := map[string]interface{}{"password": "fakePassword"}
				fakeKVSecret.Data = fakeUserData
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(fakeKVSecret, nil).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				password, err := vaultHelperClient.GetBMCBIOSPassword(ctx, fakeSecretPath)
				fmt.Printf("password: %+v,error: %+v\n", password, err)
				Expect(password).To(Equal("fakePassword"))
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("failed to get the secret", func() {
			It("should return error as secret is failed", func() {
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(nil, fmt.Errorf("failed to get the vault KVv2 secret")).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				password, err := vaultHelperClient.GetBMCBIOSPassword(ctx, fakeSecretPath)
				fmt.Printf("password: %+v,error: %+v\n", password, err)
				Expect(password).To(Equal(""))
				Expect(err).Should(HaveOccurred())
			})
		})
	})
	Describe("get netbox token", func() {
		When("netbox token is missing", func() {
			It("should return errors as token is missing", func() {
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(fakeKVSecret, nil).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				token, err := vaultHelperClient.GetNetBoxAPIToken(ctx, fakeSecretPath)
				fmt.Printf("password: %+v,error: %+v\n", token, err)
				Expect(token).To(Equal(""))
				Expect(err).Should(HaveOccurred())
			})
		})
		When("netbox token is present", func() {
			It("should return token", func() {
				fakeUserData := map[string]interface{}{"token": "123456"}
				fakeKVSecret.Data = fakeUserData
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(fakeKVSecret, nil).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				token, err := vaultHelperClient.GetNetBoxAPIToken(ctx, fakeSecretPath)
				fmt.Printf("token: %+v,error: %+v\n", token, err)
				Expect(token).To(Equal("123456"))
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("failed to get the secret", func() {
			It("should return error as secret is failed", func() {
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(nil, fmt.Errorf("failed to get the vault KVv2 secret")).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				token, err := vaultHelperClient.GetNetBoxAPIToken(ctx, fakeSecretPath)
				fmt.Printf("password: %+v,error: %+v\n", token, err)
				Expect(token).To(Equal(""))
				Expect(err).Should(HaveOccurred())
			})
		})
	})
	Describe("get IPA SSH key credentials", func() {
		When("IPA SSH key is missing", func() {
			It("should return errors as ssh key is missing", func() {
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(fakeKVSecret, nil).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				sshKey, err := vaultHelperClient.GetIPAImageSSHPrivateKey(ctx, fakeSecretPath)
				fmt.Printf("password: %+v,error: %+v\n", sshKey, err)
				Expect(sshKey).To(Equal(""))
				Expect(err).Should(HaveOccurred())
			})
		})
		When("IPA SSH key is present", func() {
			It("should return SSH key", func() {
				fakeUserData := map[string]interface{}{"privateKey": fakeSSHPrivateKey}
				fakeKVSecret.Data = fakeUserData
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(fakeKVSecret, nil).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				key, err := vaultHelperClient.GetIPAImageSSHPrivateKey(ctx, fakeSecretPath)
				fmt.Printf("token: %+v,error: %+v\n", key, err)
				Expect(key).To(Equal(fakeSSHPrivateKey))
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("IPA SSH key length is 0", func() {
			It("should return error as SSH key length is 0", func() {
				fakeUserData := map[string]interface{}{"privateKey": ""}
				fakeKVSecret.Data = fakeUserData
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(fakeKVSecret, nil).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				key, err := vaultHelperClient.GetIPAImageSSHPrivateKey(ctx, fakeSecretPath)
				fmt.Printf("token: %+v,error: %+v\n", key, err)
				Expect(key).To(Equal(""))
				Expect(err).Should(HaveOccurred())
			})
		})
		When("failed to get the secret", func() {
			It("should return error as secret is failed", func() {
				mockVaultHelper.EXPECT().getVaultSecrets(Any, Any, Any, Any, false).Return(nil, fmt.Errorf("failed to get the vault KVv2 secret")).Times(2)
				vaultHelperClient.vaultHelper.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				sshKey, err := vaultHelperClient.GetIPAImageSSHPrivateKey(ctx, fakeSecretPath)
				fmt.Printf("password: %+v,error: %+v\n", sshKey, err)
				Expect(sshKey).To(Equal(""))
				Expect(err).Should(HaveOccurred())
			})
		})
	})
	Describe("put control plane secrets", func() {
		When("returns no errors as put operation is succeeded", func() {
			It("should return kv secrets and no errors", func() {
				fakeKVPutData := map[string]interface{}{"username": "fakeUser", "password": "fakePassword"}
				fakeKVSecret.Data = fakeKVPutData
				mockVaultHelper.EXPECT().putVaultSecrets(Any, Any, Any, Any, Any, false).Return(fakeKVSecret, nil).Times(2)
				vaultHelperClient.vaultHelper.putVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeKVPutData, fakeRenewToken)
				secret, err := vaultHelperClient.PutControlplaneVaultSecrets(ctx, fakeSecretPath, fakeKVPutData, fakeRenewToken)
				fmt.Printf("secret: %+v, error: %+v\n", secret, err)
				Expect(secret).NotTo(BeNil())
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("returns errors as put operation is failed", func() {
			It("should return errors", func() {
				fakeKVPutData := map[string]interface{}{"username": "fakeUser", "password": "fakePassword"}
				fakeKVSecret.Data = fakeKVPutData
				mockVaultHelper.EXPECT().putVaultSecrets(Any, Any, Any, Any, Any, false).Return(nil, fmt.Errorf("failed to put the Control plane secrets")).Times(2)
				vaultHelperClient.vaultHelper.putVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeKVPutData, fakeRenewToken)
				secret, err := vaultHelperClient.PutControlplaneVaultSecrets(ctx, fakeSecretPath, fakeKVPutData, false)
				fmt.Printf("secret: %+v, error: %+v\n", secret, err)
				Expect(secret).Should(BeNil())
				Expect(err).Should(HaveOccurred())
			})
		})
	})
	Describe("put BMC secrets", func() {
		When("returns no errors as put operation is succeeded", func() {
			It("should return kv secrets and no errors", func() {
				fakeKVPutData := map[string]interface{}{"username": "fakeUser", "password": "fakePassword"}
				fakeKVSecret.Data = fakeKVPutData
				mockVaultHelper.EXPECT().putVaultSecrets(Any, Any, Any, Any, Any, false).Return(fakeKVSecret, nil).Times(2)
				vaultHelperClient.vaultHelper.putVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeKVPutData, fakeRenewToken)
				secret, err := vaultHelperClient.PutBMCSecrets(ctx, fakeSecretPath, fakeKVPutData)
				fmt.Printf("secret: %+v, error: %+v\n", secret, err)
				Expect(secret).NotTo(BeNil())
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("returns errors as put operation is failed", func() {
			It("should return errors", func() {
				fakeKVPutData := map[string]interface{}{"username": "fakeUser", "password": "fakePassword"}
				fakeKVSecret.Data = fakeKVPutData
				mockVaultHelper.EXPECT().putVaultSecrets(Any, Any, Any, Any, Any, false).Return(nil, fmt.Errorf("failed to put the Control plane secrets")).Times(2)
				vaultHelperClient.vaultHelper.putVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeKVPutData, fakeRenewToken)
				secret, err := vaultHelperClient.PutBMCSecrets(ctx, fakeSecretPath, fakeKVPutData)
				fmt.Printf("secret: %+v, error: %+v\n", secret, err)
				Expect(secret).Should(BeNil())
				Expect(err).Should(HaveOccurred())
			})
		})
	})
	Describe("delete BMC secrets", func() {
		When("returns no errors as delete operation is succeeded", func() {
			It("should return kv secrets and no errors", func() {
				mockVaultHelper.EXPECT().deleteVaultSecrets(Any, Any, Any, Any, false).Return(nil).Times(2)
				vaultHelperClient.vaultHelper.deleteVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				err := vaultHelperClient.DeleteBMCSecrets(ctx, fakeSecretPath)
				fmt.Printf("error: %+v\n", err)
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("returns errors as delete operation is failed", func() {
			It("should return errors", func() {
				mockVaultHelper.EXPECT().deleteVaultSecrets(Any, Any, Any, Any, false).Return(fmt.Errorf("failed to delete the Control plane secrets")).Times(2)
				vaultHelperClient.vaultHelper.deleteVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeRenewToken)
				err := vaultHelperClient.DeleteBMCSecrets(ctx, fakeSecretPath)
				fmt.Printf("error: %+v\n", err)
				Expect(err).Should(HaveOccurred())
			})
		})
	})
	Describe("vault get roleID test", func() {
		When("have a valid role ID", func() {
			It("should get role ID without error", func() {
				err := vaultHelperClient.getRoleID()
				fmt.Printf("error: %+v\n", err)
				Expect(err).Should(BeNil())
			})
		})
		When("have no valid role ID", func() {
			JustBeforeEach(func() {
				deleteFile(fakeApproleID)
			})
			It("should result in error as roleID is not present", func() {
				err := vaultHelperClient.getRoleID()
				fmt.Printf("error: %+v\n", err)
				Expect(err).Should(HaveOccurred())
			})
		})
	})
	Describe("vault get secretID test", func() {
		When("have a valid secretID file", func() {
			It("should get secretID file name without error", func() {
				err := vaultHelperClient.getSecretIDFile()
				fmt.Printf("error: %+v\n", err)
				Expect(err).Should(BeNil())
			})
		})
		When("missing secretID file env var", func() {
			BeforeEach(func() {
				Expect(os.Setenv("VAULT_APPROLE_SECRET_ID_FILE", "")).To(Succeed())
			})
			It("should get secretID file name without error", func() {
				err := vaultHelperClient.getSecretIDFile()
				fmt.Printf("error: %+v\n", err)
				Expect(err).Should(HaveOccurred())
			})
		})
	})
	Describe("vault test new vault client", func() {
		It("should get new vault client", func() {
			nv, err := NewVaultClient(ctx)
			fmt.Printf("client: %+v\n", nv)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(nv).NotTo(BeNil())
		})
	})
	Describe("get vault client test", func() {
		When("have a valid vault client", func() {
			It("should get vault client without error", func() {
				mockVaultHelper.EXPECT().getVaultClient(Any).Return(nil).Times(2)
				vaultHelperClient.vaultHelper.getVaultClient(ctx)
				Expect(vaultHelperClient.getClient(ctx)).To(Succeed())
			})
		})
		When("have valid vault client - w/o mock", func() {
			It("should get vault client without error", func() {
				err := vaultHelperClient.getVaultClient(ctx)
				fmt.Printf("client:%v error: %+v\n", vaultHelperClient.client, err)
				Expect(err).Should(BeNil())
			})
		})
	})
	Describe("get vault auth info test", func() {
		When("have a valid auth info response", func() {
			It("should get vault auth info without error", func() {
				mockVaultHelper.EXPECT().getVaultAuthInfo(Any).Return(nil).Times(2)
				vaultHelperClient.vaultHelper.getVaultAuthInfo(ctx)
				Expect(vaultHelperClient.getAuthInfo(ctx)).To(Succeed())
			})
		})
		When("no connectivity to the vault server to get vault authinfo - w/o mock", func() {
			It("should get vault client without error", func() {
				vaultHelperClient.client.SetClientTimeout(1 * time.Millisecond)
				vaultHelperClient.client.SetMinRetryWait(1 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetryWait(2 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetries(0)
				err := vaultHelperClient.getVaultAuthInfo(ctx)
				fmt.Printf("client:%v error: %+v\n", vaultHelperClient.client, err)
				Expect(err).Should(HaveOccurred())
			})
		})
	})
	Describe("get vault KVv2 secrets test", func() {
		When("trying to get secrets from the vault server", func() {
			BeforeEach(func() {
				Expect(os.Setenv(fakeEnvPath, "bmc")).To(Succeed())
			})
			It("should fail as vault server connection is missing - get", func() {
				vaultHelperClient.client.SetClientTimeout(1 * time.Millisecond)
				vaultHelperClient.client.SetMinRetryWait(1 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetryWait(2 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetries(0)
				secret, err := vaultHelperClient.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, false)
				fmt.Printf("secret:%v error: %+v\n", secret, err)
				Expect(secret).Should(BeNil())
				Expect(err).Should(HaveOccurred())
			})
		})
		When("secret engine to get KVv2 secret is empty - get", func() {
			BeforeEach(func() {
				Expect(os.Setenv(fakeEnvPath, "")).To(Succeed())
			})
			It("should fail as secret engine is empty", func() {
				vaultHelperClient.client.SetClientTimeout(1 * time.Millisecond)
				vaultHelperClient.client.SetMinRetryWait(1 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetryWait(2 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetries(0)
				secret, err := vaultHelperClient.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, false)
				fmt.Printf("secret:%v error: %+v\n", secret, err)
				Expect(secret).Should(BeNil())
				Expect(err).Should(HaveOccurred())
			})
		})
	})
	Describe("put vault KVv2 secrets test", func() {
		BeforeEach(func() {
			Expect(os.Setenv(fakeEnvPath, "bmc")).To(Succeed())
		})
		When("trying to put secrets to the vault server", func() {
			BeforeEach(func() {
				Expect(os.Setenv(fakeEnvPath, "bmc")).To(Succeed())
			})
			It("should fail as server connection is missing - put", func() {
				vaultHelperClient.client.SetClientTimeout(1 * time.Millisecond)
				vaultHelperClient.client.SetMinRetryWait(1 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetryWait(2 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetries(0)
				fakeKVPutData := map[string]interface{}{"username": "fakeUser", "password": "fakePassword"}
				fakeKVSecret.Data = fakeKVPutData
				secret, err := vaultHelperClient.putVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeKVPutData, false)
				fmt.Printf("secret:%v error: %+v\n", secret, err)
				Expect(secret).Should(BeNil())
				Expect(err).Should(HaveOccurred())
			})
		})
		When("secret engine to put KVv2 secret is empty", func() {
			BeforeEach(func() {
				Expect(os.Setenv(fakeEnvPath, "")).To(Succeed())
			})
			It("secret engine to get KVv2 secret is empty - put", func() {
				vaultHelperClient.client.SetClientTimeout(1 * time.Millisecond)
				vaultHelperClient.client.SetMinRetryWait(1 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetryWait(2 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetries(0)
				fakeKVPutData := map[string]interface{}{"username": "fakeUser", "password": "fakePassword"}
				fakeKVSecret.Data = fakeKVPutData
				secret, err := vaultHelperClient.putVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeKVPutData, false)
				fmt.Printf("secret:%v error: %+v\n", secret, err)
				Expect(secret).Should(BeNil())
				Expect(err).Should(HaveOccurred())
			})
		})
	})
	Describe("delete vault KVv2 secrets test", func() {
		When("trying to delete secrets from the vault server", func() {
			BeforeEach(func() {
				Expect(os.Setenv(fakeEnvPath, "bmc")).To(Succeed())
			})
			It("should fail as vault server connection is missing - delete", func() {
				vaultHelperClient.client.SetClientTimeout(1 * time.Millisecond)
				vaultHelperClient.client.SetMinRetryWait(1 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetryWait(2 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetries(0)
				err := vaultHelperClient.deleteVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, false)
				fmt.Printf("error: %+v\n", err)
				Expect(err).Should(HaveOccurred())
			})
		})
		When("secret engine to delete KVv2 secret is empty", func() {
			BeforeEach(func() {
				Expect(os.Setenv(fakeEnvPath, "")).To(Succeed())
			})
			It("should fail as secret engine is empty - delete", func() {
				vaultHelperClient.client.SetClientTimeout(1 * time.Millisecond)
				vaultHelperClient.client.SetMinRetryWait(1 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetryWait(2 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetries(0)
				err := vaultHelperClient.deleteVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, false)
				fmt.Printf("error: %+v\n", err)
				Expect(err).Should(HaveOccurred())
			})
		})
	})
	Describe("validate token renewal", func() {
		When("using a renewable token", func() {
			It("should check token renewal with a renewable token", func() {
				vaultHelperClient.client.SetClientTimeout(1 * time.Millisecond)
				vaultHelperClient.client.SetMinRetryWait(1 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetryWait(2 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetries(0)
				RenewalTimeout = 500 * time.Millisecond
				status := vaultHelperClient.renewVaultToken(ctx)
				fmt.Printf("renewal status: %+v\n", status)
				Expect(status).To(Equal(false))
			})
		})
		When("using a non renewable token", func() {
			It("should check token renewal with a renewable token", func() {
				vaultHelperClient.client.SetClientTimeout(1 * time.Millisecond)
				vaultHelperClient.client.SetMinRetryWait(1 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetryWait(2 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetries(0)
				RenewalTimeout = 500 * time.Millisecond
				status := vaultHelperClient.renewVaultToken(ctx)
				fmt.Printf("renewal status: %+v\n", status)
				Expect(status).To(Equal(false))
			})
		})
		When("renew token with get vault secrets", func() {
			It("should check token renewal - get", func() {
				vaultHelperClient.client.SetClientTimeout(1 * time.Millisecond)
				vaultHelperClient.client.SetMinRetryWait(1 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetryWait(2 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetries(0)
				RenewalTimeout = 500 * time.Millisecond
				vaultHelperClient.secret.Auth.Renewable = false
				secret, err := vaultHelperClient.getVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, true)
				fmt.Printf("secret:%+v\n error: %+v\n", secret, err)
				Expect(secret).Should(BeNil())
				Expect(err).Should(HaveOccurred())
			})
		})
		When("renew token with put vault secrets", func() {
			It("should check token renewal - put", func() {
				fakeKVPutData := map[string]interface{}{"username": "fakeUser", "password": "fakePassword"}
				fakeKVSecret.Data = fakeKVPutData
				vaultHelperClient.client.SetClientTimeout(1 * time.Millisecond)
				vaultHelperClient.client.SetMinRetryWait(1 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetryWait(2 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetries(0)
				RenewalTimeout = 500 * time.Millisecond
				vaultHelperClient.secret.Auth.Renewable = false
				secret, err := vaultHelperClient.putVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, fakeKVPutData, true)
				fmt.Printf("secret:%+v\n error: %+v\n", secret, err)
				Expect(secret).Should(BeNil())
				Expect(err).Should(HaveOccurred())
			})
		})
		When("renew token with delete vault secrets", func() {
			It("should check token renewal - delete", func() {
				vaultHelperClient.client.SetClientTimeout(1 * time.Millisecond)
				vaultHelperClient.client.SetMinRetryWait(1 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetryWait(2 * time.Millisecond)
				vaultHelperClient.client.SetMaxRetries(0)
				RenewalTimeout = 500 * time.Millisecond
				vaultHelperClient.secret.Auth.Renewable = false
				err := vaultHelperClient.deleteVaultSecrets(ctx, fakeEnvPath, fakeDefaultPath, fakeSecretPath, true)
				fmt.Printf("error: %+v\n", err)
				Expect(err).Should(HaveOccurred())
			})
		})
	})
	Describe("test validate vault client", func() {
		When("all the required parameters available without errors", func() {
			It("get vault client and auth info without error", func() {
				err := vaultHelperClient.ValidateVaultClient(ctx)
				fmt.Printf("error: %+v\n", err)
				Expect(err).Should(HaveOccurred())
			})
		})
	})
})

func createFile(fileName string) {
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = f.WriteString(fileName)
	if err != nil {
		log.Fatal(err)
	}
}

func deleteFile(fileName string) {
	if _, err := os.Stat(fileName); err == nil {
		err := os.Remove(fileName)
		if err != nil {
			log.Fatal(err)
		}
	}
}
