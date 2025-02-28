// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package processor

import (
	"context"
	"fmt"
	"testing"

	firewallv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/api/v1alpha1"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	loadbalancerv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/pkg/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	cloudaccountid1  = "123456789123"
	cloudaccountid2  = "222222222222"
	numFlakeAttempts = 2
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = Describe("TestProcessor_PersistFirewallRuleStatusUpdate", FlakeAttempts(numFlakeAttempts), func() {
	var (
		ctx       context.Context
		k8sclient client.WithWatch
		processor *Processor
	)

	BeforeEach(func() {
		ctx = context.Background()
		var err error
		k8sclient, err = getFakeClient()
		Expect(err).NotTo(HaveOccurred())
		processor = NewProcessor(k8sclient, nil, nil, "us-dev-1", "us-dev-1a")
	})

	It("should update firewall rule status", func() {
		fwRule := &firewallv1alpha1.FirewallRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "firewallrule1",
				Namespace: cloudaccountid1,
			},
			Spec: firewallv1alpha1.FirewallRuleSpec{},
			Status: firewallv1alpha1.FirewallRuleStatus{
				Conditions: []metav1.Condition{
					{
						Type:   "Ready",
						Status: metav1.ConditionFalse,
					},
				},
			},
		}

		// Note: Status sub-resource is special with the fake client. It needs
		// to be added when the client is generated, otherwise a "not-found" error
		// will always be returned.
		ctx = context.Background()
		var err error
		k8sclient, err = getFakeClient(fwRule)
		Expect(err).NotTo(HaveOccurred())
		processor = NewProcessor(k8sclient, nil, nil, "us-dev-1", "us-dev-1a")

		fwRule.Status.Conditions[0].Status = metav1.ConditionTrue
		err = processor.PersistFirewallRuleStatusUpdate(ctx, fwRule, types.NamespacedName{Name: fwRule.Name, Namespace: fwRule.Namespace})
		Expect(err).NotTo(HaveOccurred())

		updatedFWRule := &firewallv1alpha1.FirewallRule{}
		Expect(k8sclient.Get(ctx, client.ObjectKey{Name: fwRule.Name, Namespace: fwRule.Namespace}, updatedFWRule)).NotTo(HaveOccurred())
		Expect(updatedFWRule.Status.Conditions[0].Status).To(Equal(metav1.ConditionTrue))
	})

	It("should handle firewall rule not found", func() {
		fwRule := &firewallv1alpha1.FirewallRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "non-existent-firewallrule",
				Namespace: cloudaccountid1,
			},
			Spec: firewallv1alpha1.FirewallRuleSpec{},
			Status: firewallv1alpha1.FirewallRuleStatus{
				Conditions: []metav1.Condition{
					{
						Type:   "Ready",
						Status: metav1.ConditionFalse,
					},
				},
			},
		}

		err := processor.PersistFirewallRuleStatusUpdate(ctx, fwRule, types.NamespacedName{Name: fwRule.Name, Namespace: fwRule.Namespace})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Operation cannot be fulfilled on v1alpha1.private.cloud.intel.com \"non-existent-firewallrule\""))
	})

	It("should not update firewall rule status if unchanged", func() {
		fwRule := &firewallv1alpha1.FirewallRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "firewallrule1",
				Namespace: cloudaccountid1,
			},
			Spec: firewallv1alpha1.FirewallRuleSpec{},
			Status: firewallv1alpha1.FirewallRuleStatus{
				Conditions: []metav1.Condition{
					{
						Type:   "Ready",
						Status: metav1.ConditionTrue,
					},
				},
			},
		}

		// Note: Status sub-resource is special with the fake client. It needs
		// to be added when the client is generated, otherwise a "not-found" error
		// will always be returned.
		ctx = context.Background()
		var err error
		k8sclient, err = getFakeClient(fwRule)
		Expect(err).NotTo(HaveOccurred())
		processor = NewProcessor(k8sclient, nil, nil, "us-dev-1", "us-dev-1a")

		err = processor.PersistFirewallRuleStatusUpdate(ctx, fwRule, types.NamespacedName{Name: fwRule.Name, Namespace: fwRule.Namespace})
		Expect(err).NotTo(HaveOccurred())

		updatedFWRule := &firewallv1alpha1.FirewallRule{}
		Expect(k8sclient.Get(ctx, client.ObjectKey{Name: fwRule.Name, Namespace: fwRule.Namespace}, updatedFWRule)).NotTo(HaveOccurred())
		Expect(updatedFWRule.Status.Conditions[0].Status).To(Equal(metav1.ConditionTrue))
	})
})

var _ = Describe("TestProcessor_PersistFirewallRuleFinalizer", FlakeAttempts(numFlakeAttempts), func() {
	var (
		ctx       context.Context
		k8sclient client.WithWatch
		processor *Processor
	)

	BeforeEach(func() {
		ctx = context.Background()
		var err error
		k8sclient, err = getFakeClient()
		Expect(err).NotTo(HaveOccurred())
		processor = NewProcessor(k8sclient, nil, nil, "us-dev-1", "us-dev-1a")
	})

	It("should add finalizer to firewall rule", func() {
		fwRule := &firewallv1alpha1.FirewallRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "firewallrule1",
				Namespace: cloudaccountid1,
			},
			Spec: firewallv1alpha1.FirewallRuleSpec{},
		}
		Expect(k8sclient.Create(ctx, fwRule)).NotTo(HaveOccurred())

		err := processor.PersistFirewallRuleFinalizer(ctx, add, *fwRule)
		Expect(err).NotTo(HaveOccurred())

		updatedFWRule := &firewallv1alpha1.FirewallRule{}
		Expect(k8sclient.Get(ctx, client.ObjectKey{Name: fwRule.Name, Namespace: fwRule.Namespace}, updatedFWRule)).NotTo(HaveOccurred())
		Expect(controllerutil.ContainsFinalizer(updatedFWRule, constants.LoadbalancerFinalizer)).To(BeTrue())
	})

	It("should remove finalizer from firewall rule", func() {
		fwRule := &firewallv1alpha1.FirewallRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "firewallrule1",
				Namespace: cloudaccountid1,
			},
			Spec: firewallv1alpha1.FirewallRuleSpec{},
		}
		controllerutil.AddFinalizer(fwRule, constants.LoadbalancerFinalizer)
		Expect(k8sclient.Create(ctx, fwRule)).NotTo(HaveOccurred())

		err := processor.PersistFirewallRuleFinalizer(ctx, remove, *fwRule)
		Expect(err).NotTo(HaveOccurred())

		updatedFWRule := &firewallv1alpha1.FirewallRule{}
		Expect(k8sclient.Get(ctx, client.ObjectKey{Name: fwRule.Name, Namespace: fwRule.Namespace}, updatedFWRule)).NotTo(HaveOccurred())
		Expect(controllerutil.ContainsFinalizer(updatedFWRule, constants.LoadbalancerFinalizer)).To(BeFalse())
	})

	It("should handle firewall rule not found", func() {
		fwRule := &firewallv1alpha1.FirewallRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "non-existent-firewallrule",
				Namespace: cloudaccountid1,
			},
			Spec: firewallv1alpha1.FirewallRuleSpec{},
		}

		err := processor.PersistFirewallRuleFinalizer(ctx, add, *fwRule)
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("TestProcessor_PersistLoadbalancerFinalizer", FlakeAttempts(numFlakeAttempts), func() {
	var (
		ctx       context.Context
		k8sclient client.WithWatch
		processor *Processor
	)

	BeforeEach(func() {
		ctx = context.Background()
		var err error
		k8sclient, err = getFakeClient()
		Expect(err).NotTo(HaveOccurred())
		processor = NewProcessor(k8sclient, nil, nil, "us-dev-1", "us-dev-1a")
	})

	It("should add finalizer to loadbalancer", func() {
		loadbalancer := &loadbalancerv1alpha1.Loadbalancer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "loadbalancer1",
				Namespace: cloudaccountid1,
			},
			Spec: loadbalancerv1alpha1.LoadbalancerSpec{},
		}
		Expect(k8sclient.Create(ctx, loadbalancer)).NotTo(HaveOccurred())

		err := processor.PersistLoadbalancerFinalizer(ctx, add, *loadbalancer)
		Expect(err).NotTo(HaveOccurred())

		updatedLoadbalancer := &loadbalancerv1alpha1.Loadbalancer{}
		Expect(k8sclient.Get(ctx, client.ObjectKey{Name: loadbalancer.Name, Namespace: loadbalancer.Namespace}, updatedLoadbalancer)).NotTo(HaveOccurred())
		Expect(controllerutil.ContainsFinalizer(updatedLoadbalancer, constants.LoadbalancerFinalizer)).To(BeTrue())
	})

	It("should remove finalizer from loadbalancer", func() {
		loadbalancer := &loadbalancerv1alpha1.Loadbalancer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "loadbalancer1",
				Namespace: cloudaccountid1,
			},
			Spec: loadbalancerv1alpha1.LoadbalancerSpec{},
		}
		controllerutil.AddFinalizer(loadbalancer, constants.LoadbalancerFinalizer)
		Expect(k8sclient.Create(ctx, loadbalancer)).NotTo(HaveOccurred())

		err := processor.PersistLoadbalancerFinalizer(ctx, remove, *loadbalancer)
		Expect(err).NotTo(HaveOccurred())

		updatedLoadbalancer := &loadbalancerv1alpha1.Loadbalancer{}
		Expect(k8sclient.Get(ctx, client.ObjectKey{Name: loadbalancer.Name, Namespace: loadbalancer.Namespace}, updatedLoadbalancer)).NotTo(HaveOccurred())
		Expect(controllerutil.ContainsFinalizer(updatedLoadbalancer, constants.LoadbalancerFinalizer)).To(BeFalse())
	})

	It("should handle loadbalancer not found", func() {
		loadbalancer := &loadbalancerv1alpha1.Loadbalancer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "non-existent-loadbalancer",
				Namespace: cloudaccountid1,
			},
			Spec: loadbalancerv1alpha1.LoadbalancerSpec{},
		}

		err := processor.PersistLoadbalancerFinalizer(ctx, add, *loadbalancer)
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("TestProcessor_PersistInstanceFinalizer", FlakeAttempts(numFlakeAttempts), func() {
	var (
		ctx       context.Context
		k8sclient client.WithWatch
		processor *Processor
	)

	BeforeEach(func() {
		ctx = context.Background()
		var err error
		k8sclient, err = getFakeClient()
		Expect(err).NotTo(HaveOccurred())
		processor = NewProcessor(k8sclient, nil, nil, "us-dev-1", "us-dev-1a")
	})

	It("should add finalizer to instance", func() {
		instance := &cloudv1alpha1.Instance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "instance1",
				Namespace: cloudaccountid1,
			},
			Spec: cloudv1alpha1.InstanceSpec{
				Labels: map[string]string{
					"lb":     "true",
					"expose": "externallb",
				},
			},
			Status: instanceReadyStatus,
		}
		Expect(k8sclient.Create(ctx, instance)).NotTo(HaveOccurred())

		err := processor.PersistInstanceFinalizer(ctx, add, *instance)
		Expect(err).NotTo(HaveOccurred())

		updatedInstance := &cloudv1alpha1.Instance{}
		Expect(k8sclient.Get(ctx, client.ObjectKey{Name: instance.Name, Namespace: instance.Namespace}, updatedInstance)).NotTo(HaveOccurred())
		Expect(controllerutil.ContainsFinalizer(updatedInstance, constants.LoadbalancerFinalizer)).To(BeTrue())
	})

	It("should remove finalizer from instance", func() {
		instance := &cloudv1alpha1.Instance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "instance1",
				Namespace: cloudaccountid1,
			},
			Spec: cloudv1alpha1.InstanceSpec{
				Labels: map[string]string{
					"lb":     "true",
					"expose": "externallb",
				},
			},
			Status: instanceReadyStatus,
		}
		controllerutil.AddFinalizer(instance, constants.LoadbalancerFinalizer)
		Expect(k8sclient.Create(ctx, instance)).NotTo(HaveOccurred())

		err := processor.PersistInstanceFinalizer(ctx, remove, *instance)
		Expect(err).NotTo(HaveOccurred())

		updatedInstance := &cloudv1alpha1.Instance{}
		Expect(k8sclient.Get(ctx, client.ObjectKey{Name: instance.Name, Namespace: instance.Namespace}, updatedInstance)).NotTo(HaveOccurred())
		Expect(controllerutil.ContainsFinalizer(updatedInstance, constants.LoadbalancerFinalizer)).To(BeFalse())
	})

	It("should handle instance not found", func() {
		instance := &cloudv1alpha1.Instance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "non-existent-instance",
				Namespace: cloudaccountid1,
			},
			Spec: cloudv1alpha1.InstanceSpec{
				Labels: map[string]string{
					"lb":     "true",
					"expose": "externallb",
				},
			},
			Status: instanceReadyStatus,
		}

		err := processor.PersistInstanceFinalizer(ctx, add, *instance)
		Expect(err).NotTo(HaveOccurred())
	})

})

var (
	instanceReadyStatus = cloudv1alpha1.InstanceStatus{
		Phase: cloudv1alpha1.PhaseReady,
	}

	instance1 = &cloudv1alpha1.Instance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "instance1",
			Namespace: cloudaccountid1,
		},
		Spec: cloudv1alpha1.InstanceSpec{
			Labels: map[string]string{
				"lb":     "true",
				"expose": "externallb",
			},
		},
		Status: instanceReadyStatus,
	}

	instance2 = &cloudv1alpha1.Instance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "instance2",
			Namespace: cloudaccountid1,
		},
		Spec: cloudv1alpha1.InstanceSpec{
			Labels: map[string]string{
				"lb": "true",
			},
		},
		Status: instanceReadyStatus,
	}

	instance3 = &cloudv1alpha1.Instance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "instance3",
			Namespace: cloudaccountid1,
		},
		Spec:   cloudv1alpha1.InstanceSpec{},
		Status: instanceReadyStatus,
	}

	// Same as instance2 but in a different cloud account
	instance4 = &cloudv1alpha1.Instance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "instance2",
			Namespace: cloudaccountid2,
		},
		Spec: cloudv1alpha1.InstanceSpec{
			Labels: map[string]string{
				"lb": "true",
			},
		},
		Status: instanceReadyStatus,
	}

	// Same as instance3 but in a different cloud account
	instance5 = &cloudv1alpha1.Instance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "instance3",
			Namespace: cloudaccountid2,
		},
		Spec:   cloudv1alpha1.InstanceSpec{},
		Status: instanceReadyStatus,
	}
)

var _ = Describe("TestLoadbalancerCache_OnLoadbalancerChange", FlakeAttempts(numFlakeAttempts), func() {

	seedInstances := []*cloudv1alpha1.Instance{
		instance1, instance2, instance3, instance4, instance5,
	}

	tests := map[string]struct {
		loadbalancer  *loadbalancerv1alpha1.Loadbalancer
		seedInstances []*cloudv1alpha1.Instance
		want          []*cloudv1alpha1.Instance
		wantErr       bool
	}{
		"static pool - no instances configured": {
			loadbalancer: &loadbalancerv1alpha1.Loadbalancer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "lb1",
					Namespace: cloudaccountid1,
				},
				Spec: loadbalancerv1alpha1.LoadbalancerSpec{},
			},
			want:    nil,
			wantErr: true,
		},
		"static pool - invalid instance": {
			loadbalancer: &loadbalancerv1alpha1.Loadbalancer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "lb1",
					Namespace: cloudaccountid1,
				},
				Spec: loadbalancerv1alpha1.LoadbalancerSpec{
					Listeners: []loadbalancerv1alpha1.LoadbalancerListener{{
						Pool: loadbalancerv1alpha1.VPool{
							Members: []loadbalancerv1alpha1.VMember{{
								InstanceResourceId: "invalid",
							}},
							InstanceSelectors: nil,
						},
					}},
				},
			},
			want:    []*cloudv1alpha1.Instance{},
			wantErr: false,
		},
		"static pool - single instance": {
			seedInstances: seedInstances,
			loadbalancer: &loadbalancerv1alpha1.Loadbalancer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "lb1",
					Namespace: cloudaccountid1,
				},
				Spec: loadbalancerv1alpha1.LoadbalancerSpec{
					Listeners: []loadbalancerv1alpha1.LoadbalancerListener{{
						Pool: loadbalancerv1alpha1.VPool{
							Members: []loadbalancerv1alpha1.VMember{{
								InstanceResourceId: instance1.Name,
							}},
							InstanceSelectors: nil,
						},
					}},
				},
			},
			want:    []*cloudv1alpha1.Instance{instance1},
			wantErr: false,
		},
		"static pool - multiple instances": {
			seedInstances: seedInstances,
			loadbalancer: &loadbalancerv1alpha1.Loadbalancer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "lb1",
					Namespace: cloudaccountid1,
				},
				Spec: loadbalancerv1alpha1.LoadbalancerSpec{
					Listeners: []loadbalancerv1alpha1.LoadbalancerListener{{
						Pool: loadbalancerv1alpha1.VPool{
							Members: []loadbalancerv1alpha1.VMember{{
								InstanceResourceId: instance1.Name,
							}, {
								InstanceResourceId: instance3.Name,
							}},
							InstanceSelectors: nil,
						},
					}},
				},
			},
			want:    []*cloudv1alpha1.Instance{instance1, instance3},
			wantErr: false,
		},
		"instance selector - single instance": {
			seedInstances: seedInstances,
			loadbalancer: &loadbalancerv1alpha1.Loadbalancer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "lb1",
					Namespace: cloudaccountid1,
				},
				Spec: loadbalancerv1alpha1.LoadbalancerSpec{
					Listeners: []loadbalancerv1alpha1.LoadbalancerListener{{
						Pool: loadbalancerv1alpha1.VPool{
							InstanceSelectors: map[string]string{"expose": "externallb"},
						},
					}},
				},
			},
			want:    []*cloudv1alpha1.Instance{instance1},
			wantErr: false,
		},
		"instance selector - multiple instances": {
			seedInstances: seedInstances,
			loadbalancer: &loadbalancerv1alpha1.Loadbalancer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "lb1",
					Namespace: cloudaccountid1,
				},
				Spec: loadbalancerv1alpha1.LoadbalancerSpec{
					Listeners: []loadbalancerv1alpha1.LoadbalancerListener{{
						Pool: loadbalancerv1alpha1.VPool{
							InstanceSelectors: map[string]string{"lb": "true"},
						},
					}},
				},
			},
			want:    []*cloudv1alpha1.Instance{instance1, instance2},
			wantErr: false,
		},
		"instance selector - no matches": {
			seedInstances: seedInstances,
			loadbalancer: &loadbalancerv1alpha1.Loadbalancer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "lb1",
					Namespace: cloudaccountid1,
				},
				Spec: loadbalancerv1alpha1.LoadbalancerSpec{
					Listeners: []loadbalancerv1alpha1.LoadbalancerListener{{
						Pool: loadbalancerv1alpha1.VPool{
							InstanceSelectors: map[string]string{"invalid": "selector"},
						},
					}},
				},
			},
			want:    []*cloudv1alpha1.Instance{},
			wantErr: false,
		},
	}
	for name, tt := range tests {

		It(name, func() {
			scheme := runtime.NewScheme()
			Expect(cloudv1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())

			ctx := context.Background()

			k8sclient, _ := getFakeClient()
			for _, instance := range tt.seedInstances {
				instance.ResourceVersion = ""
				Expect(k8sclient.Create(ctx, instance)).NotTo(HaveOccurred())
			}

			c := NewProcessor(k8sclient, nil, nil, "us-dev-1", "us-dev-1a")

			for _, l := range tt.loadbalancer.Spec.Listeners {
				got, err := c.GetLoadbalancerInstances(ctx, tt.loadbalancer.Namespace, l)

				if (err != nil) != tt.wantErr {
					Fail(fmt.Sprintf("UpsertLoadbalancer() error = %v, wantErr %v", err, tt.wantErr))
					return
				}
				Expect(len(tt.want)).Should(Equal(len(got)))
			}
		})
	}
})

var _ = Describe("TestLoadbalancerCache_GetLoadbalancers", FlakeAttempts(numFlakeAttempts), func() {

	loadbalancer0 := &loadbalancerv1alpha1.Loadbalancer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "loadbalancer1",
			Namespace: cloudaccountid2,
		},
		Spec: loadbalancerv1alpha1.LoadbalancerSpec{
			Listeners: []loadbalancerv1alpha1.LoadbalancerListener{{
				Pool: loadbalancerv1alpha1.VPool{
					Members: []loadbalancerv1alpha1.VMember{{
						InstanceResourceId: instance1.Name,
					}},
				},
			}},
		},
	}

	loadbalancer1 := &loadbalancerv1alpha1.Loadbalancer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "loadbalancer1",
			Namespace: cloudaccountid1,
		},
		Spec: loadbalancerv1alpha1.LoadbalancerSpec{
			Listeners: []loadbalancerv1alpha1.LoadbalancerListener{{
				Pool: loadbalancerv1alpha1.VPool{
					Members: []loadbalancerv1alpha1.VMember{{
						InstanceResourceId: instance1.Name,
					}},
				},
			}},
		},
	}

	loadbalancer2 := &loadbalancerv1alpha1.Loadbalancer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "loadbalancer2",
			Namespace: cloudaccountid1,
		},
		Spec: loadbalancerv1alpha1.LoadbalancerSpec{
			Listeners: []loadbalancerv1alpha1.LoadbalancerListener{{
				Pool: loadbalancerv1alpha1.VPool{
					Members: []loadbalancerv1alpha1.VMember{{
						InstanceResourceId: instance2.Name,
					}},
				},
			}},
		},
	}

	loadbalancer3 := &loadbalancerv1alpha1.Loadbalancer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "loadbalancer3",
			Namespace: cloudaccountid1,
		},
		Spec: loadbalancerv1alpha1.LoadbalancerSpec{
			Listeners: []loadbalancerv1alpha1.LoadbalancerListener{{
				Pool: loadbalancerv1alpha1.VPool{
					Members: []loadbalancerv1alpha1.VMember{{
						InstanceResourceId: instance2.Name,
					}},
				},
			}},
		},
	}

	loadbalancer4 := &loadbalancerv1alpha1.Loadbalancer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "loadbalancer4",
			Namespace: cloudaccountid1,
		},
		Spec: loadbalancerv1alpha1.LoadbalancerSpec{
			Listeners: []loadbalancerv1alpha1.LoadbalancerListener{{
				Pool: loadbalancerv1alpha1.VPool{
					InstanceSelectors: map[string]string{
						"lb": "true",
					},
				},
			}},
		},
	}

	loadbalancer5 := &loadbalancerv1alpha1.Loadbalancer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "loadbalancer5",
			Namespace: cloudaccountid1,
		},
		Spec: loadbalancerv1alpha1.LoadbalancerSpec{
			Listeners: []loadbalancerv1alpha1.LoadbalancerListener{{
				Pool: loadbalancerv1alpha1.VPool{
					InstanceSelectors: map[string]string{
						"lb": "true",
					},
				},
			}},
		},
	}

	tests := map[string]struct {
		instance          *cloudv1alpha1.Instance
		seedLoadbalancers []*loadbalancerv1alpha1.Loadbalancer
		want              []*loadbalancerv1alpha1.Loadbalancer
		wantErr           bool
	}{
		"static pool - no instances configured": {
			seedLoadbalancers: []*loadbalancerv1alpha1.Loadbalancer{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "loadbalancer1",
					Namespace: cloudaccountid1,
				},
				Spec: loadbalancerv1alpha1.LoadbalancerSpec{
					Listeners: []loadbalancerv1alpha1.LoadbalancerListener{{
						Pool: loadbalancerv1alpha1.VPool{},
					}},
				},
			}},
			instance: &cloudv1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "instance1",
					Namespace: cloudaccountid1,
				},
				Spec: cloudv1alpha1.InstanceSpec{},
			},
			want:    nil,
			wantErr: false,
		},
		"static pool - one lb matching": {
			seedLoadbalancers: []*loadbalancerv1alpha1.Loadbalancer{
				loadbalancer1,
			},
			instance: &cloudv1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "instance1",
					Namespace: cloudaccountid1,
				},
				Spec: cloudv1alpha1.InstanceSpec{},
			},
			want:    []*loadbalancerv1alpha1.Loadbalancer{loadbalancer1},
			wantErr: false,
		},
		"static pool - two lb matching": {
			seedLoadbalancers: []*loadbalancerv1alpha1.Loadbalancer{
				loadbalancer2, loadbalancer3,
			},
			instance: &cloudv1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "instance2",
					Namespace: cloudaccountid1,
				},
				Spec: cloudv1alpha1.InstanceSpec{},
			},
			want:    []*loadbalancerv1alpha1.Loadbalancer{loadbalancer2, loadbalancer3},
			wantErr: false,
		},
		"instance selectors - one lb matching": {
			seedLoadbalancers: []*loadbalancerv1alpha1.Loadbalancer{
				loadbalancer4,
			},
			instance: &cloudv1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "instance1",
					Namespace: cloudaccountid1,
				},
				Spec: cloudv1alpha1.InstanceSpec{
					Labels: map[string]string{
						"lb": "true",
					},
				},
			},
			want:    []*loadbalancerv1alpha1.Loadbalancer{loadbalancer4},
			wantErr: false,
		},
		"instance selectors - two lb matching - multiple accounts": {
			seedLoadbalancers: []*loadbalancerv1alpha1.Loadbalancer{
				loadbalancer0, loadbalancer4, loadbalancer5,
			},
			instance: &cloudv1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "instance1",
					Namespace: cloudaccountid1,
				},
				Spec: cloudv1alpha1.InstanceSpec{
					Labels: map[string]string{
						"lb": "true",
					},
				},
			},
			want:    []*loadbalancerv1alpha1.Loadbalancer{loadbalancer4, loadbalancer5},
			wantErr: false,
		},
	}
	for name, tt := range tests {

		It(name, func() {
			ctx := context.Background()
			scheme := runtime.NewScheme()
			Expect(cloudv1alpha1.AddToScheme(scheme)).NotTo(HaveOccurred())

			k8sclient, _ := getFakeClient()
			for _, instance := range tt.seedLoadbalancers {
				instance.ResourceVersion = ""
				Expect(k8sclient.Create(ctx, instance)).NotTo(HaveOccurred())
			}

			c := NewProcessor(k8sclient, nil, nil, "us-dev-1", "us-dev-1a")
			got, err := c.GetLoadbalancers(ctx, tt.instance)

			if (err != nil) != tt.wantErr {
				Fail(fmt.Sprintf("GetLoadbalancers() error = %v, wantErr %v", err, tt.wantErr))
				return
			}

			Expect(len(tt.want)).Should(Equal(len(got)))
		})
	}
})

func getFakeClient(initObjs ...client.Object) (client.WithWatch, error) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := cloudv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := loadbalancerv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := firewallv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	return fake.NewClientBuilder().
		WithScheme(scheme).WithObjects(initObjs...).
		WithStatusSubresource(initObjs...).
		Build(), nil
}
