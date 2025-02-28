// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tasks

import (
	"context"
	"time"

	"github.com/golang/mock/gomock"
	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakedynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/dcim"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/mocks"
)

var _ = Describe("Disenrollment Task", func() {
	const testNamespace = "test-namespace"

	var (
		task   *DisenrollmentTask
		ctx    context.Context
		helper *testHelper
		bmh    *baremetalv1alpha1.BareMetalHost
		err    error

		mockCtrl      *gomock.Controller
		netbox        *mocks.MockDCIM
		vault         *mocks.MockSecretManager
		clientSet     *fake.Clientset
		dynamicClient *fakedynamic.FakeDynamicClient
	)

	BeforeEach(func() {
		ctx = context.Background()

		// create mock objects
		mockCtrl = gomock.NewController(GinkgoT())
		netbox = mocks.NewMockDCIM(mockCtrl)
		vault = mocks.NewMockSecretManager(mockCtrl)
		clientSet = fake.NewSimpleClientset()
		scheme := runtime.NewScheme()
		baremetalv1alpha1.AddToScheme(scheme)
		dynamicClient = fakedynamic.NewSimpleDynamicClient(scheme)

		// create test struct
		task = &DisenrollmentTask{
			deviceData: &DeviceData{
				Name:   "device-1",
				ID:     1,
				Rack:   "test-rack",
				Region: "us-dev-1",
			},
			netBox:        netbox,
			vault:         vault,
			clientSet:     clientSet,
			dynamicClient: dynamicClient,
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("running the disenrollment task", func() {
		BeforeEach(func() {
			By("creating Metal3 namespace")
			Expect(helper.createNamespace(ctx, clientSet, testNamespace)).Error().NotTo(HaveOccurred())

			By("creating BareMetalHost")
			Expect(helper.createBareMetalHost(ctx, dynamicClient, task.deviceData.Name, testNamespace)).Error().NotTo(HaveOccurred())
			bmh, err = helper.getBareMetalHost(ctx, dynamicClient, task.deviceData.Name, testNamespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(bmh).NotTo(BeNil())

			By("expecting the device to start with 'Disenrolling' status")
			netbox.EXPECT().UpdateDeviceCustomFields(ctx, Any, Any, &dcim.DeviceCustomFields{
				BMEnrollmentStatus:  dcim.BMDisenrolling,
				BMEnrollmentComment: "Disenrollment is in progress",
			}).Return(nil).Times(1)
		})

		When("disenrollment is successful", func() {
			BeforeEach(func() {
				By("expecting the device to end up with 'Disenrolled' status")
				netbox.EXPECT().UpdateDeviceCustomFields(ctx, Any, Any, &dcim.DeviceCustomFields{
					BMEnrollmentStatus:  dcim.BMDisenrolled,
					BMEnrollmentComment: "Disenrollment is complete",
				}).Return(nil).Times(1)
			})

			It("deletes the host", func() {
				Expect(task.Run(ctx)).To(Succeed())
				_, err := helper.getBareMetalHost(ctx, dynamicClient, task.deviceData.Name, testNamespace)
				Expect(errors.IsNotFound(err)).To(BeTrue())
			})

			It("deletes the host's secret", func() {
				Expect(task.Run(ctx)).To(Succeed())
				_, err := clientSet.CoreV1().Secrets(testNamespace).Get(ctx, task.deviceData.Name, metav1.GetOptions{})
				Expect(errors.IsNotFound(err)).To(BeTrue())
			})

			It("finds no matching host in metal3 namespaces", func() {
				By("deleting all hosts")
				Expect(helper.deleteAllBareMetalHosts(ctx, dynamicClient)).To(Succeed())
				Expect(task.Run(ctx)).To(Succeed())
			})

			When("host is already being deleted", func() {
				It("skips the deleting and only waits for the host to be deleted", func() {
					By("adding a deletion timestamp to the host")
					bmh.DeletionTimestamp = &metav1.Time{Time: time.Now()}
					Expect(helper.updateBareMetalHost(ctx, dynamicClient, bmh)).Error().NotTo(HaveOccurred())

					By("simulating k8s reconciler to clean up the host")
					go func() {
						time.Sleep(500 * time.Millisecond)
						Expect(helper.deleteBareMetalHost(ctx, task.dynamicClient, task.deviceData.Name, testNamespace)).To(Succeed())
					}()

					Eventually(task.Run(ctx)).Should(Succeed())
					_, err := helper.getBareMetalHost(ctx, dynamicClient, task.deviceData.Name, testNamespace)
					Expect(errors.IsNotFound(err)).To(BeTrue())
				})
			})
		})

		When("disenrollment is unsuccessful", func() {
			BeforeEach(func() {
				By("expecting the device to end up with 'DisenrollmentFailed' status")
				netbox.EXPECT().UpdateDeviceCustomFields(ctx, Any, Any, Any).Return(nil).Times(1)
			})

			It("finds no metal3 namespaces", func() {
				By("deleting all metal3 namespaces")
				Expect(helper.deleteAllNamespaces(ctx, clientSet)).To(Succeed())
				Expect(task.Run(ctx)).Error().To(HaveOccurred())
			})

			It("finds the host with a consumer", func() {
				By("adding a consumer reference")
				bmh.Spec.ConsumerRef = &v1.ObjectReference{}
				Expect(helper.updateBareMetalHost(ctx, dynamicClient, bmh)).Error().NotTo(HaveOccurred())
				Expect(task.Run(ctx)).Error().To(HaveOccurred())
			})
		})

		AfterEach(func() {
			By("cleaning up")
			Expect(helper.deleteAllBareMetalHosts(ctx, dynamicClient)).To(Succeed())
			Expect(helper.deleteAllSecrets(ctx, clientSet, testNamespace)).To(Succeed())
			Expect(helper.deleteAllNamespaces(ctx, clientSet)).To(Succeed())
		})
	})
})
