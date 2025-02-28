//go:remove_metal3_pod || system
// +build remove_metal3_pod system

package system_tests_test

import (
	"log"
	"fmt"
	"time"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"goFramework/framework/library/bmaas/kube"
)

var namespace = "metal3-1"
var pod_name string
var serviceaccount_name string

var _ = Describe("Remove Metal3 Pod", Ordered, func() {

	It("Verify Metal3 namespace exists", func() {
		_, err := kube.GetNamespace(namespace)
		Expect(err).Error().ShouldNot(HaveOccurred())
		
		log.Printf("Namespace: %s", namespace)
	})

	It("Get one pod from Metal3 namespace", func() {
		response, err := kube.GetPodList(namespace)
		Expect(err).Error().ShouldNot(HaveOccurred())
		Expect(response.Items).NotTo(BeNil())
		Expect(len(response.Items) > 0).To(BeTrue())
		Expect(response.Items[0]).NotTo(BeNil())
		
		pod_name = response.Items[0].Name
		serviceaccount_name = response.Items[0].Spec.ServiceAccountName
		log.Printf("Service account: %s", serviceaccount_name)
	})

	It("Delete " + pod_name +" pod", func() {
		err := kube.DeletePod(pod_name, namespace)
		Expect(err).Error().ShouldNot(HaveOccurred())
		log.Printf("Deleted pod: %s", pod_name)
		time.Sleep(5 * time.Second)
	})

	It("Validate" + pod_name +" pod has been recreated", func() {
		response, err := kube.GetPodListByFieldSelector(namespace, fmt.Sprintf("spec.serviceAccountName=%s", serviceaccount_name))
		
		Expect(err).Error().ShouldNot(HaveOccurred())
		Expect(response.Items).NotTo(BeNil())
		Expect(len(response.Items) > 0).To(BeTrue())
		Expect(response.Items[0]).NotTo(BeNil())

		pod_name = response.Items[0].Name
		pod_phase := string(response.Items[0].Status.Phase)
		pod_age := uint(time.Now().Sub(response.Items[0].Status.StartTime.Time).Seconds())

		log.Printf("Pod %s has been recreated %ss ago and it is in %s phase", pod_name, fmt.Sprint(pod_age), pod_phase)

		Expect(pod_age <= 60).To(BeTrue())
		Expect(pod_phase).To(Equal("Running"))
	})
})
