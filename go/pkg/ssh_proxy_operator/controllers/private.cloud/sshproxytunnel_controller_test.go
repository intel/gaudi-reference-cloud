// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package privatecloud

import (
	"context"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
)

var _ = Describe("SshProxyTunnel controller", func() {

	const (
		tunnelNamespace = "default"
		tunnelName1     = "tenant-vm-1"
		tunnelName2     = "tenant-vm-2"
		tunnelName3     = "tenant-vm-3"
		tunnelName4     = "tenant-vm-4"
		tunnelName5     = "tenant-vm-5"
		pubKey1         = "ssh-rsa A/pk1/AAAAB3NzaC1yc2EAAAADAQABAAABgQDlZi5rhqqS...MwF3OE="
		pubKey2         = "ssh-rsa Z/pk2/AAAAB3NzaC1yc2EZZZZZZZZZZZZBgQDlZi5rhqqS...MwF3OE="
		pubKey3         = "ssh-rsa K/pk3/AAAAB3NzaC1yc2EKKKKKKKKKKKKBgQDlZi5rhqqS...MwF3OE="
		pubKeyOp        = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC5GVA77uCixTzSZSfSX3dBlD9Dkg9ypbvzLefB/kxWK9BX idcuser@example.com"
		timeout         = time.Second * 10
		duration        = time.Second * 10
		interval        = time.Millisecond * 250
	)

	Context("When the 1st SshProxyTunnel is created", func() {
		It("authorized_keys should have exactly have 3 lines", func() {
			ctx := context.Background()
			tunnel := &cloudv1alpha1.SshProxyTunnel{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "private.cloud.intel.com/v1alpha1",
					Kind:       "SshProxyTunnel",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      tunnelName1,
					Namespace: tunnelNamespace,
				},
				Spec: cloudv1alpha1.SshProxyTunnelSpec{
					TargetAddresses: []string{"1.1.1.1", "2.2.2.2"},
					TargetPorts:     []int{22, 443},
					SshPublicKeys:   []string{pubKey1 + " user1@example.com", pubKey2 + " user2@example.com"},
				},
			}
			Expect(k8sClient.Create(ctx, tunnel)).Should(Succeed())

			tunnelLookupKey := types.NamespacedName{Name: tunnelName1, Namespace: tunnelNamespace}
			createdTunnel := &cloudv1alpha1.SshProxyTunnel{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, tunnelLookupKey, createdTunnel)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			Expect(createdTunnel.Spec.TargetAddresses).Should(Equal([]string{"1.1.1.1", "2.2.2.2"}))

			Eventually(func() (string, error) {
				return ReadAuthorizedKeysFile()
			}, timeout, interval).Should(Equal(
				pubKeyOp + "\n" +
					"permitopen=\"1.1.1.1:22\",permitopen=\"1.1.1.1:443\",permitopen=\"2.2.2.2:22\",permitopen=\"2.2.2.2:443\"," +
					SshAuthorizedKeysOptions + " " + pubKey1 + "\n" +
					"permitopen=\"1.1.1.1:22\",permitopen=\"1.1.1.1:443\",permitopen=\"2.2.2.2:22\",permitopen=\"2.2.2.2:443\"," +
					SshAuthorizedKeysOptions + " " + pubKey2 + "\n"))
		})
	})

	Context("When the 2nd SshProxyTunnel is created with a different key", func() {
		It("authorized_keys should have exactly 4 lines", func() {
			ctx := context.Background()
			tunnel := &cloudv1alpha1.SshProxyTunnel{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "private.cloud.intel.com/v1alpha1",
					Kind:       "SshProxyTunnel",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      tunnelName2,
					Namespace: tunnelNamespace,
				},
				Spec: cloudv1alpha1.SshProxyTunnelSpec{
					TargetAddresses: []string{"1.1.1.1", "2.2.2.2"},
					TargetPorts:     []int{22, 443},
					SshPublicKeys:   []string{pubKey3 + " user3@example.com"},
				},
			}
			Expect(k8sClient.Create(ctx, tunnel)).Should(Succeed())
			// Note that items in the file are sorted deterministically so this test should be reliable.
			Eventually(func() (string, error) {
				return ReadAuthorizedKeysFile()
			}, timeout, interval).Should(Equal(
				pubKeyOp + "\n" +
					"permitopen=\"1.1.1.1:22\",permitopen=\"1.1.1.1:443\",permitopen=\"2.2.2.2:22\",permitopen=\"2.2.2.2:443\"," +
					SshAuthorizedKeysOptions + " " + pubKey1 + "\n" +
					"permitopen=\"1.1.1.1:22\",permitopen=\"1.1.1.1:443\",permitopen=\"2.2.2.2:22\",permitopen=\"2.2.2.2:443\"," +
					SshAuthorizedKeysOptions + " " + pubKey3 + "\n" +
					"permitopen=\"1.1.1.1:22\",permitopen=\"1.1.1.1:443\",permitopen=\"2.2.2.2:22\",permitopen=\"2.2.2.2:443\"," +
					SshAuthorizedKeysOptions + " " + pubKey2 + "\n"))
		})
	})

	Context("When the 3rd SshProxyTunnel is created, with a duplicate key", func() {
		It("authorized_keys should have exactly 4 lines, 1 with 3 IPs and 1 with operator public key", func() {
			ctx := context.Background()
			tunnel := &cloudv1alpha1.SshProxyTunnel{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "private.cloud.intel.com/v1alpha1",
					Kind:       "SshProxyTunnel",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      tunnelName3,
					Namespace: tunnelNamespace,
				},
				Spec: cloudv1alpha1.SshProxyTunnelSpec{
					TargetAddresses: []string{"3.3.3.3"},
					TargetPorts:     []int{22, 443},
					SshPublicKeys:   []string{pubKey2 + " user2@example.com"},
				},
			}
			Expect(k8sClient.Create(ctx, tunnel)).Should(Succeed())
			Eventually(func() (string, error) {
				return ReadAuthorizedKeysFile()
			}, timeout, interval).Should(Equal(
				pubKeyOp + "\n" +
					"permitopen=\"1.1.1.1:22\",permitopen=\"1.1.1.1:443\",permitopen=\"2.2.2.2:22\",permitopen=\"2.2.2.2:443\"," +
					SshAuthorizedKeysOptions + " " + pubKey1 + "\n" +
					"permitopen=\"1.1.1.1:22\",permitopen=\"1.1.1.1:443\",permitopen=\"2.2.2.2:22\",permitopen=\"2.2.2.2:443\"," +
					SshAuthorizedKeysOptions + " " + pubKey3 + "\n" +
					"permitopen=\"1.1.1.1:22\",permitopen=\"1.1.1.1:443\",permitopen=\"2.2.2.2:22\",permitopen=\"2.2.2.2:443\",permitopen=\"3.3.3.3:22\",permitopen=\"3.3.3.3:443\"," +
					SshAuthorizedKeysOptions + " " + pubKey2 + "\n"))
		})
	})

	Context("When the 1st and 3rd SshProxyTunnels are deleted", func() {
		It("authorized_keys should have exactly 1 line with 2 IPs and 1 line for operator public key", func() {
			ctx := context.Background()
			tunnelLookupKey := types.NamespacedName{Name: tunnelName1, Namespace: tunnelNamespace}
			tunnel := &cloudv1alpha1.SshProxyTunnel{}
			Expect(k8sClient.Get(ctx, tunnelLookupKey, tunnel)).Should(BeNil())
			Expect(k8sClient.Delete(ctx, tunnel)).Should(BeNil())
			tunnelLookupKey = types.NamespacedName{Name: tunnelName3, Namespace: tunnelNamespace}
			tunnel = &cloudv1alpha1.SshProxyTunnel{}
			Expect(k8sClient.Get(ctx, tunnelLookupKey, tunnel)).Should(BeNil())
			Expect(k8sClient.Delete(ctx, tunnel)).Should(BeNil())
			Eventually(func() (string, error) {
				return ReadAuthorizedKeysFile()
			}, timeout, interval).Should(Equal(
				pubKeyOp + "\n" +
					"permitopen=\"1.1.1.1:22\",permitopen=\"1.1.1.1:443\",permitopen=\"2.2.2.2:22\",permitopen=\"2.2.2.2:443\"," +
					SshAuthorizedKeysOptions + " " + pubKey3 + "\n"))
		})
	})

	Context("When the last SshProxyTunnel is deleted", func() {
		It("authorized_keys should have only operator public key", func() {
			ctx := context.Background()
			tunnelLookupKey := types.NamespacedName{Name: tunnelName2, Namespace: tunnelNamespace}
			tunnel := &cloudv1alpha1.SshProxyTunnel{}
			Expect(k8sClient.Get(ctx, tunnelLookupKey, tunnel)).Should(BeNil())
			Expect(k8sClient.Delete(ctx, tunnel)).Should(BeNil())
			Eventually(func() (string, error) {
				return ReadAuthorizedKeysFile()
			}, timeout, interval).Should(Equal(pubKeyOp + "\n"))
		})
	})

	Context("Mocking failure recovery while transferring authorized_keys file to any SCP target", func() {
		It("SshProxyTunnel.status should get updated with the proxy user, address, port", func() {
			ctx := context.Background()
			tunnel := &cloudv1alpha1.SshProxyTunnel{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "private.cloud.intel.com/v1alpha1",
					Kind:       "SshProxyTunnel",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      tunnelName3,
					Namespace: tunnelNamespace,
				},
				Spec: cloudv1alpha1.SshProxyTunnelSpec{
					TargetAddresses: []string{"3.3.3.3"},
					TargetPorts:     []int{22, 443},
					SshPublicKeys:   []string{pubKey2 + " user2@example.com"},
				},
			}
			sshProxyController.MockScpTargetsMutex.Lock()
			sshProxyController.MockScpTargets[scpTarget] = fmt.Errorf("injecting fault")
			sshProxyController.MockScpTargetsMutex.Unlock()
			Expect(k8sClient.Create(ctx, tunnel)).Should(Succeed())

			// mocking failure
			Consistently(func(g Gomega) {
				tunnelLookupKey := types.NamespacedName{Name: tunnelName3, Namespace: tunnelNamespace}
				retrievedtunnel := &cloudv1alpha1.SshProxyTunnel{}
				g.Expect(k8sClient.Get(ctx, tunnelLookupKey, retrievedtunnel)).Error().NotTo(HaveOccurred())
				g.Expect(retrievedtunnel.Status.ProxyUser).Should(BeNil())
				g.Expect(retrievedtunnel.Status.ProxyAddress).Should(BeNil())
				g.Expect(retrievedtunnel.Status.ProxyPort).Should(BeNil())
			}, time.Second*5, time.Millisecond*500)

			// mocking recovery
			sshProxyController.MockScpTargetsMutex.Lock()
			sshProxyController.MockScpTargets[scpTarget] = nil
			sshProxyController.MockScpTargetsMutex.Unlock()
			Eventually(func(g Gomega) {
				tunnelLookupKey := types.NamespacedName{Name: tunnelName3, Namespace: tunnelNamespace}
				retrievedtunnel := &cloudv1alpha1.SshProxyTunnel{}
				g.Expect(k8sClient.Get(ctx, tunnelLookupKey, retrievedtunnel)).Error().NotTo(HaveOccurred())
				g.Expect(retrievedtunnel.Status.ProxyUser).Should(Equal(proxyUser))
				g.Expect(retrievedtunnel.Status.ProxyAddress).Should(Equal(proxyAddress))
				g.Expect(retrievedtunnel.Status.ProxyPort).Should(Equal(proxyPort))

			}, timeout, interval).Should(Succeed())
		})
	})

	Context("Mocking failure recovery while transferring authorized_keys file to multiple SCP targets", func() {
		ctx := context.Background()
		tunnel := &cloudv1alpha1.SshProxyTunnel{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "private.cloud.intel.com/v1alpha1",
				Kind:       "SshProxyTunnel",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      tunnelName4,
				Namespace: tunnelNamespace,
			},
			Spec: cloudv1alpha1.SshProxyTunnelSpec{
				TargetAddresses: []string{"3.3.3.3"},
				TargetPorts:     []int{22, 443},
				SshPublicKeys:   []string{pubKey2 + " user2@example.com"},
			},
		}
		When("Scp failed for all the SCP targets", func() {
			It("SshProxyTunnel.status should consistently have empty proxy user, address, port", func() {
				sshProxyController.MockScpTargetsMutex.Lock()
				sshProxyController.MockScpTargets[scpTarget] = fmt.Errorf("injecting fault for target 1")
				sshProxyController.MockScpTargets[scpTarget2] = fmt.Errorf("injecting fault for target 2")
				sshProxyController.MockScpTargetsMutex.Unlock()
				Expect(k8sClient.Create(ctx, tunnel)).Should(Succeed())

				Consistently(func(g Gomega) {
					tunnelLookupKey := types.NamespacedName{Name: tunnelName4, Namespace: tunnelNamespace}
					retrievedtunnel := &cloudv1alpha1.SshProxyTunnel{}
					g.Expect(k8sClient.Get(ctx, tunnelLookupKey, retrievedtunnel)).Error().NotTo(HaveOccurred())
					g.Expect(retrievedtunnel.Status.ProxyUser).Should(BeNil())
					g.Expect(retrievedtunnel.Status.ProxyAddress).Should(BeNil())
					g.Expect(retrievedtunnel.Status.ProxyPort).Should(BeNil())
				}, time.Second*5, time.Millisecond*500)
			})
		})

		When("Scp is succesful for all the SCP targets", func() {
			It("SshProxyTunnel.status should get updated with the proxy user, address, port", func() {
				sshProxyController.MockScpTargetsMutex.Lock()
				sshProxyController.MockScpTargets[scpTarget] = nil
				sshProxyController.MockScpTargets[scpTarget2] = nil
				sshProxyController.MockScpTargetsMutex.Unlock()
				Eventually(func(g Gomega) {
					tunnelLookupKey := types.NamespacedName{Name: tunnelName4, Namespace: tunnelNamespace}
					retrievedtunnel := &cloudv1alpha1.SshProxyTunnel{}
					g.Expect(k8sClient.Get(ctx, tunnelLookupKey, retrievedtunnel)).Error().NotTo(HaveOccurred())
					g.Expect(retrievedtunnel.Status.ProxyUser).Should(Equal(proxyUser))
					g.Expect(retrievedtunnel.Status.ProxyAddress).Should(Equal(proxyAddress))
					g.Expect(retrievedtunnel.Status.ProxyPort).Should(Equal(proxyPort))
				}, timeout, interval).Should(Succeed())
			})
		})

		When("Scp is successful for 1 of 2 SCP targets", func() {
			It("SshProxyTunnel.status should get updated with the proxy user, address, port", func() {
				sshProxyController.MockScpTargetsMutex.Lock()
				sshProxyController.MockScpTargets[scpTarget] = nil
				sshProxyController.MockScpTargets[scpTarget2] = fmt.Errorf("injecting fault")
				sshProxyController.MockScpTargetsMutex.Unlock()

				tunnel1 := &cloudv1alpha1.SshProxyTunnel{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "private.cloud.intel.com/v1alpha1",
						Kind:       "SshProxyTunnel",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      tunnelName5,
						Namespace: tunnelNamespace,
					},
					Spec: cloudv1alpha1.SshProxyTunnelSpec{
						TargetAddresses: []string{"3.3.3.3"},
						TargetPorts:     []int{22, 443},
						SshPublicKeys:   []string{pubKey2 + " user2@example.com"},
					},
				}
				Expect(k8sClient.Create(ctx, tunnel1)).Should(Succeed())

				Eventually(func(g Gomega) {
					tunnelLookupKey := types.NamespacedName{Name: tunnelName5, Namespace: tunnelNamespace}
					retrievedtunnel := &cloudv1alpha1.SshProxyTunnel{}
					g.Expect(k8sClient.Get(ctx, tunnelLookupKey, retrievedtunnel)).Error().NotTo(HaveOccurred())
					g.Expect(retrievedtunnel.Status.ProxyUser).Should(Equal(proxyUser))
					g.Expect(retrievedtunnel.Status.ProxyAddress).Should(Equal(proxyAddress))
					g.Expect(retrievedtunnel.Status.ProxyPort).Should(Equal(proxyPort))
				}, timeout, interval).Should(Succeed())
			})
		})
	})
})

func ReadAuthorizedKeysFile() (string, error) {
	buf, err := os.ReadFile(authorizedKeysFilePath)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}
