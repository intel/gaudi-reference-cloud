// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package convert

import (
	"time"

	"github.com/google/go-cmp/cmp"
	lbv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func baselineLoadBalancer() (*pb.LoadBalancerPrivate, *lbv1alpha1.Loadbalancer) {
	creationTimestamp := time.Unix(1600000000, 0).UTC()
	deletionTimestamp := time.Unix(1700000000, 0).UTC()
	k8sDeletionTimestamp := metav1.NewTime(deletionTimestamp)
	pbLoadBalancer := &pb.LoadBalancerPrivate{
		Metadata: &pb.LoadBalancerMetadataPrivate{
			CloudAccountId:  "CloudAccountId1",
			Name:            "Name1",
			ResourceId:      "ResourceId1",
			ResourceVersion: "ResourceVersion1",
			Labels: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			CreationTimestamp: timestamppb.New(creationTimestamp),
			DeletionTimestamp: timestamppb.New(deletionTimestamp),
		},
		Spec: &pb.LoadBalancerSpecPrivate{
			Listeners: []*pb.LoadBalancerListener{{
				Port: 9090,
				Pool: &pb.LoadBalancerPool{
					Port: 8080,
					InstanceSelectors: map[string]string{
						"foo": "bar",
					},
					Monitor: pb.LoadBalancerMonitorType_http,
				},
			}},
			Security: &pb.LoadBalancerSecurity{
				Sourceips: []string{"1.2.3.4"},
			},
		},
		Status: &pb.LoadBalancerStatusPrivate{
			State: "Pending",
			Vip:   "2.2.2.2",
			Conditions: &pb.LoadBalancerConditionsStatus{
				FirewallRuleCreated: false,
				Listeners: []*pb.LoadBalancerConditionsListenerStatus{{
					Port:          0,
					PoolCreated:   true,
					VipCreated:    true,
					VipPoolLinked: false,
				}},
			},
			Listeners: []*pb.LoadBalancerListenerStatus{{
				Name:        "my-lb-1",
				VipID:       1234,
				Message:     "My update message",
				PoolMembers: []*pb.LoadBalancerPoolStatusMember{},
				PoolID:      112,
			}},
		},
	}
	k8sLoadBalancer := &lbv1alpha1.Loadbalancer{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Loadbalancer",
			APIVersion: "private.cloud.intel.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "ResourceId1",
			Namespace:         "CloudAccountId1",
			ResourceVersion:   "ResourceVersion1",
			CreationTimestamp: metav1.NewTime(creationTimestamp),
			DeletionTimestamp: &k8sDeletionTimestamp,
			Labels: map[string]string{
				"availabiltyZoneId": "us-dev-1a",
				"cloud-account-id":  "CloudAccountId1",
				"regionId":          "us-dev-1",
			},
		},
		Spec: lbv1alpha1.LoadbalancerSpec{
			Labels: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			Listeners: []lbv1alpha1.LoadbalancerListener{{
				VIP: lbv1alpha1.VServer{
					Port:       9090,
					IPProtocol: "tcp",

					IPType: string(lbv1alpha1.IPType_PUBLIC),
				},
				Pool: lbv1alpha1.VPool{
					Port:    8080,
					Monitor: string(lbv1alpha1.MonitorType_HTTP),
					InstanceSelectors: map[string]string{
						"foo": "bar",
					},
					Members: []lbv1alpha1.VMember{},
				},
				Owner: "",
			}},
			Security: lbv1alpha1.LoadbalancerSecurity{
				Sourceips: []string{"1.2.3.4"},
			},
		},
		Status: lbv1alpha1.LoadbalancerStatus{
			State: "Pending",
			Vip:   "2.2.2.2",
			Conditions: lbv1alpha1.ConditionsStatus{
				FirewallRuleCreated: false,
				Listeners: []lbv1alpha1.ConditionsListenerStatus{{
					Port:          0,
					PoolCreated:   true,
					VIPCreated:    true,
					VIPPoolLinked: false,
				}},
			},
			Listeners: []lbv1alpha1.ListenerStatus{{
				Name:        "my-lb-1",
				VipID:       1234,
				Message:     "My update message",
				PoolMembers: []lbv1alpha1.PoolStatusMember{},
				PoolID:      112,
			}},
		},
	}
	return pbLoadBalancer, k8sLoadBalancer
}

var _ = Describe("LB PbToK8s", func() {
	It("Baseline should succeed", func() {
		converter, err := NewLoadBalancerConverter("us-dev-1", "us-dev-1a")
		Expect(err).Should(Succeed())

		pbLoadBalancer, k8sLoadBalancerExpected := baselineLoadBalancer()
		k8sLoadBalancerActual, err := converter.PbToK8s(pbLoadBalancer)
		Expect(err).Should(Succeed())
		diff := cmp.Diff(k8sLoadBalancerActual, k8sLoadBalancerExpected)
		GinkgoWriter.Println(diff)
		Expect(diff).Should(Equal(""))
	})

	It("Nil DeletionTimestamp should succeed", func() {
		converter, err := NewLoadBalancerConverter("us-dev-1", "us-dev-1a")
		Expect(err).Should(Succeed())
		pbLoadBalancer, k8sLoadBalancerExpected := baselineLoadBalancer()
		pbLoadBalancer.Metadata.DeletionTimestamp = nil
		k8sLoadBalancerExpected.ObjectMeta.DeletionTimestamp = nil
		k8sLoadBalancerActual, err := converter.PbToK8s(pbLoadBalancer)
		Expect(err).Should(Succeed())
		diff := cmp.Diff(k8sLoadBalancerActual, k8sLoadBalancerExpected)
		GinkgoWriter.Println(diff)
		Expect(diff).Should(Equal(""))
	})
})

var _ = Describe("LB K8sToPb", func() {
	It("Baseline should succeed", func() {
		converter, err := NewLoadBalancerConverter("us-dev-1", "us-dev-1a")
		Expect(err).Should(Succeed())
		pbLoadBalancerExpected, k8sLoadBalancer := baselineLoadBalancer()
		// Delete fields from expected value that do not get converted.
		pbLoadBalancerExpected.Metadata.Name = ""
		pbLoadBalancerActual, err := converter.K8sToPb(k8sLoadBalancer)
		Expect(err).Should(Succeed())
		diff := cmp.Diff(pbLoadBalancerActual, pbLoadBalancerExpected, protocmp.Transform())
		GinkgoWriter.Println(diff)
		Expect(diff).Should(Equal(""))
	})

	It("Nil DeletionTimestamp should succeed", func() {
		converter, err := NewLoadBalancerConverter("us-dev-1", "us-dev-1a")
		Expect(err).Should(Succeed())
		pbLoadBalancerExpected, k8sLoadBalancer := baselineLoadBalancer()
		pbLoadBalancerExpected.Metadata.DeletionTimestamp = nil
		k8sLoadBalancer.ObjectMeta.DeletionTimestamp = nil
		// Delete fields from expected value that do not get converted.
		pbLoadBalancerExpected.Metadata.Name = ""
		pbIstanceActual, err := converter.K8sToPb(k8sLoadBalancer)
		Expect(err).Should(Succeed())
		diff := cmp.Diff(pbIstanceActual, pbLoadBalancerExpected, protocmp.Transform())
		GinkgoWriter.Println(diff)
		Expect(diff).Should(Equal(""))
	})
})
