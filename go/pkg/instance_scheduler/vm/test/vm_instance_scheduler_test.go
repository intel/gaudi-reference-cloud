// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"fmt"
	"net"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/uuid"
	instanceoperatorutil "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/util"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/scheduler"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/server"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	toolsk8s "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/k8s/test"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

type SchedulerOptions struct {
	Clusters         []ClusterOptions
	OvercommitConfig *cloudv1alpha1.OvercommitConfig
}

type ClusterOptions struct {
	Nodes []*NodeOptions
}

type NodeOptions struct {
	Pods              []*corev1.Pod
	Instances         []*pb.InstancePrivate
	AllocatableCpu    string
	AllocatableMemory string
	Partition         string
	// Node labels such as pool.cloud.intel.com/pool_id: true
	NodeLabels map[string]string
}

type SchedulerTestEnv struct {
	ManagerStoppable       *stoppable.Stoppable
	SchedulerServiceClient pb.InstanceSchedulingServiceClient
	SchedulingServer       *server.SchedulingServer
	SchedulerOpts          SchedulerOptions
	Ctx                    context.Context
	Cancel                 context.CancelFunc
	HarvesterTestEnvs      []*envtest.Environment
	Clientsets             []*kubernetes.Clientset
	DynamicClientSets      []dynamic.Interface
	OvercommitConfig       cloudv1alpha1.OvercommitConfig
}

// Create a test environment for testing the scheduler.
//   - Start Kubernetes envtest.Environment (Kubernetes API Server, etcd) per Harvester cluster to simulate.
//     This supports any number of clusters.
//   - Create Nodes and initial Pods (simulated).
//   - Start controller manager.
//   - Start SchedulingServer.
func NewSchedulerTestEnv(opts SchedulerOptions) *SchedulerTestEnv {
	// Create a context that will last throughout the lifetime of the test environment.
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	overcommitConfig := cloudv1alpha1.OvercommitConfig{
		CPU:     100,
		Memory:  100,
		Storage: 100,
	}
	if opts.OvercommitConfig != nil {
		overcommitConfig = *opts.OvercommitConfig
	}
	testEnv := &SchedulerTestEnv{
		SchedulerOpts:    opts,
		Ctx:              ctx,
		Cancel:           cancel,
		OvercommitConfig: overcommitConfig,
	}
	return testEnv
}

func (e *SchedulerTestEnv) Start() {
	ctx := e.Ctx
	log := log.FromContext(ctx).WithName("SchedulerTestEnv.Start")
	log.Info("BEGIN")
	defer log.Info("END")

	By("Creating Kubernetes test environments that simulate Harvester clusters")
	harvesterTestEnvs, harvesterRestConfigList, vmClustersKubeConfigDir := toolsk8s.CreateTestEnvs(len(e.SchedulerOpts.Clusters), crdDirectoryPaths)
	e.HarvesterTestEnvs = harvesterTestEnvs

	By("Creating clientsets")
	for _, restConfig := range harvesterRestConfigList {
		clientset, err := kubernetes.NewForConfig(restConfig)
		Expect(err).NotTo(HaveOccurred())
		e.Clientsets = append(e.Clientsets, clientset)
	}

	By("Creating dynamic clientsets")
	for _, restConfig := range harvesterRestConfigList {
		dynamicClientset, err := dynamic.NewForConfig(restConfig)
		Expect(err).NotTo(HaveOccurred())
		e.DynamicClientSets = append(e.DynamicClientSets, dynamicClientset)

		unstructuredHarvesterSettingObj, gvr := toolsk8s.CreateUnstructuredHarvesterSettingObject()
		_, err = dynamicClientset.Resource(gvr).Create(ctx, unstructuredHarvesterSettingObj, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())
	}

	By("Creating manager")
	k8sManager, err := ctrl.NewManager(k8sRestConfig, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
	})
	Expect(err).ToNot(HaveOccurred())
	e.ManagerStoppable = stoppable.New(k8sManager.Start)

	By("Creating scheduling server")
	grpcServerListener, err := net.Listen("tcp", "localhost:")
	Expect(err).Should(Succeed())
	grpcListenPort := uint16(grpcServerListener.Addr().(*net.TCPAddr).Port)
	cfg := &cloudv1alpha1.VmInstanceSchedulerConfig{
		ListenPort:              grpcListenPort,
		VmClustersKubeConfigDir: vmClustersKubeConfigDir,
		EnableBMaaSLocal:        false,
		OvercommitConfig:        e.OvercommitConfig,
	}
	mockController := gomock.NewController(GinkgoT())
	instanceTypeClient := pb.NewMockInstanceTypeServiceClient(mockController)
	e.SchedulingServer, err = server.NewSchedulingServer(ctx, cfg, k8sManager, grpcServerListener, instanceTypeClient)
	Expect(err).Should(Succeed())

	By("Creating Nodes and Pods")
	for clusterIndex, clusterOpts := range e.SchedulerOpts.Clusters {
		clientset := e.Clientsets[clusterIndex]
		for nodeIndex, nodeOpts := range clusterOpts.Nodes {
			nodeName := fmt.Sprintf("node%d", nodeIndex)
			partition := nodeOpts.Partition
			if partition == "" {
				partition = "us-dev-1a-p0"
			}
			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: nodeName,
					Labels: map[string]string{
						"kubernetes.io/hostname":                             nodeName,
						instanceoperatorutil.TopologySpreadTopologyKey:       partition,
						instanceoperatorutil.LabelKeyForInstanceType("tiny"): "true",
					},
				},
				Status: corev1.NodeStatus{
					Allocatable: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourcePods:   resource.MustParse("100"),
						corev1.ResourceCPU:    resource.MustParse(nodeOpts.AllocatableCpu),
						corev1.ResourceMemory: resource.MustParse(nodeOpts.AllocatableMemory),
					},
				},
			}
			// adding pools labels
			for key, value := range nodeOpts.NodeLabels {
				node.ObjectMeta.Labels[key] = value
			}
			By(fmt.Sprintf("Creating node %v in cluster %v, partition %v", node.Name, clusterIndex, partition))
			node, err := clientset.CoreV1().Nodes().Create(ctx, node, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
			log.Info("Created node", "node", node)

			By("Removing node.kubernetes.io/not-ready taint added by node admission controller")
			node.Spec.Taints = node.Spec.Taints[:0]
			node, err = clientset.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
			Expect(err).NotTo(HaveOccurred())
			log.Info("Updated node", "node", node)

			By("Converting initial instances to pods")
			pods := append([]*corev1.Pod{}, nodeOpts.Pods...)
			for _, instance := range nodeOpts.Instances {
				pod, err := e.SchedulingServer.SchedulingService.Sched.InstanceToPod(ctx, instance, &instanceoperatorutil.InstanceNetworkInfo{})
				Expect(err).NotTo(HaveOccurred())
				pods = append(pods, pod)
			}

			By("Creating initial pods")
			for _, pod := range pods {
				By("Creating namespace for initial pod")
				_, err = clientset.CoreV1().Namespaces().Create(ctx, NewNamespace(pod.Namespace), metav1.CreateOptions{})
				Expect(err).Should(Succeed())

				By("Creating initial pod")
				pod.Spec.NodeName = node.Name
				_, err = clientset.CoreV1().Pods(pod.Namespace).Create(ctx, pod, metav1.CreateOptions{})
				Expect(err).NotTo(HaveOccurred())
			}
		}
		nodeList, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		log.Info("Nodes created", "nodeList", nodeList)
	}

	By("Starting manager")
	e.ManagerStoppable.Start(ctx)

	By("Creating GRPC client")
	clientConn, err := grpc.Dial(fmt.Sprintf("localhost:%d", grpcListenPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).NotTo(HaveOccurred())
	e.SchedulerServiceClient = pb.NewInstanceSchedulingServiceClient(clientConn)

	By("Waiting for service to become ready")
	Eventually(func(g Gomega) {
		_, err := e.SchedulerServiceClient.Ready(ctx, &emptypb.Empty{})
		g.Expect(err).Should(Succeed())
	}, "10s").Should(Succeed())
	By("Service is ready")
}

func (e *SchedulerTestEnv) Stop() {
	ctx := context.Background()
	log := log.FromContext(ctx).WithName("SchedulerTestEnv.Start")
	log.Info("BEGIN")
	defer log.Info("END")

	By("Cancelling context")
	e.Cancel()

	if e.ManagerStoppable != nil {
		By("Stopping manager")
		Expect(e.ManagerStoppable.Stop(ctx)).Should(Succeed())
	}

	By("Stopping Kubernetes test environments")
	toolsk8s.StopTestEnvs(e.HarvesterTestEnvs)
}

func (e *SchedulerTestEnv) ScheduleOneInstance(ctx context.Context, req *pb.ScheduleRequest) (*pb.ScheduleInstanceResult, error) {
	resp, err := e.SchedulerServiceClient.Schedule(ctx, req)
	if err != nil {
		return nil, err
	}
	Expect(len(resp.InstanceResults)).Should(Equal(1))
	return resp.InstanceResults[0], err
}

func NewScheduleRequest(name string, instanceType string, cpu int32, memory string) *pb.ScheduleRequest {
	req := &pb.ScheduleRequest{
		Instances: []*pb.InstancePrivate{
			NewInstance(name, instanceType, cpu, memory),
		},
	}
	return req
}

func NewScheduleResourcesRequest(numInstances int, namePrefix string, instanceType string, cpu int32, memory string) *pb.ScheduleRequest {
	req := &pb.ScheduleRequest{}
	for i := 0; i < numInstances; i++ {
		name := fmt.Sprintf("%s%d", namePrefix, i)
		req.Instances = append(req.Instances, NewInstance(name, instanceType, cpu, memory))
	}
	return req
}

func NewInstance(name string, instanceType string, cpu int32, memory string) *pb.InstancePrivate {
	resourceId := uuid.NewString()
	return &pb.InstancePrivate{
		Metadata: &pb.InstanceMetadataPrivate{
			CloudAccountId: uuid.NewString(),
			Name:           fmt.Sprintf("%s-instance-%s", name, resourceId),
			ResourceId:     resourceId,
			Labels: map[string]string{
				"iks-cluster-name": "my-iks-cluster-1",
				"iks-role":         "master",
			},
		},
		Spec: &pb.InstanceSpecPrivate{
			AvailabilityZone: "us-dev-1a",
			InstanceType:     instanceType,
			InstanceTypeSpec: &pb.InstanceTypeSpec{
				Name: instanceType,
				Cpu: &pb.CpuSpec{
					Cores:   cpu,
					Sockets: 1,
					Threads: 1,
				},
				Memory: &pb.MemorySpec{
					Size: memory,
				},
				Disks: []*pb.DiskSpec{{
					Size: "10Gi",
				}},
			},
		},
	}
}

func SetTopologySpreadConstraint(instance *pb.InstancePrivate) {
	instance.Spec.TopologySpreadConstraints = []*pb.TopologySpreadConstraints{
		{
			LabelSelector: &pb.LabelSelector{
				MatchLabels: instance.Metadata.Labels,
			},
		},
	}
}

func NewExistingPod(name string, cpu string, memory string) *corev1.Pod {
	uid := uuid.NewString()
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", "name", uid),
			Namespace: uuid.NewString(),
			UID:       types.UID(uid),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Resources: corev1.ResourceRequirements{
					Requests: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceCPU:    resource.MustParse(cpu),
						corev1.ResourceMemory: resource.MustParse(memory),
					},
				},
				Image: "image0",
				Name:  "container0",
			}},
		},
	}
}

func InstanceToExpectedPod(sched *scheduler.Scheduler, instance *pb.InstancePrivate, nodeName string) *corev1.Pod {
	pod, err := sched.InstanceToPod(context.Background(), instance, &instanceoperatorutil.InstanceNetworkInfo{})
	Expect(err).Should(Succeed())
	pod.Spec.NodeName = nodeName
	return pod
}

func NewNamespace(namespace string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
}

var _ = Describe("VM Instance Scheduler", func() {
	ctx := context.Background()
	log := log.FromContext(ctx).WithName("VM Instance Scheduler")
	_ = log

	It("Schedule instance should succeed (1 cluster, 1 node, 0 pods)", func() {
		schedulerTestEnv := NewSchedulerTestEnv(SchedulerOptions{
			Clusters: []ClusterOptions{{
				Nodes: []*NodeOptions{{
					AllocatableCpu:    "100",
					AllocatableMemory: "100Gi",
				}},
			}},
		})
		defer schedulerTestEnv.Stop()
		schedulerTestEnv.Start()
		req := NewScheduleRequest("i1", "tiny", 4, "8Gi")
		resp, err := schedulerTestEnv.ScheduleOneInstance(ctx, req)
		Expect(err).Should(Succeed())
		log.Info("resp", "resp", resp)
		Expect(resp.ClusterId).Should(Equal("cluster0"))
		Expect(resp.NodeId).Should(Equal("node0"))
	})

	It("Schedule instance with insufficient memory should fail (1 cluster, 1 node, 0 pods)", func() {
		schedulerTestEnv := NewSchedulerTestEnv(SchedulerOptions{
			Clusters: []ClusterOptions{{
				Nodes: []*NodeOptions{{
					AllocatableCpu:    "100",
					AllocatableMemory: "1Gi",
				}},
			}},
		})
		defer schedulerTestEnv.Stop()
		schedulerTestEnv.Start()
		req := NewScheduleRequest("i2", "tiny", 4, "8Gi")
		_, err := schedulerTestEnv.ScheduleOneInstance(ctx, req)
		Expect(err).ShouldNot(Succeed())
		Expect(err.Error()).Should(MatchRegexp(".*" + server.InsufficientResourcesError))
	})

	It("Schedule instance with insufficient cpu should fail (1 cluster, 1 node, 0 pods)", func() {
		schedulerTestEnv := NewSchedulerTestEnv(SchedulerOptions{
			Clusters: []ClusterOptions{{
				Nodes: []*NodeOptions{{
					AllocatableCpu:    "4",
					AllocatableMemory: "100Gi",
				}},
			}},
		})
		defer schedulerTestEnv.Stop()
		schedulerTestEnv.Start()
		req := NewScheduleRequest("i2", "tiny", 16, "8Gi")
		_, err := schedulerTestEnv.ScheduleOneInstance(ctx, req)
		Expect(err).ShouldNot(Succeed())
		Expect(err.Error()).Should(MatchRegexp(".*" + server.InsufficientResourcesError))
	})

	It("Schedule instance with no nodes that support instance type should fail (1 cluster, 1 node, 0 pods)", func() {
		schedulerTestEnv := NewSchedulerTestEnv(SchedulerOptions{
			Clusters: []ClusterOptions{{
				Nodes: []*NodeOptions{{
					AllocatableCpu:    "100",
					AllocatableMemory: "100Gi",
				}},
			}},
		})
		defer schedulerTestEnv.Stop()
		schedulerTestEnv.Start()
		req := NewScheduleRequest("i2", "no-nodes-instance-type", 4, "8Gi")
		_, err := schedulerTestEnv.ScheduleOneInstance(ctx, req)
		Expect(err).ShouldNot(Succeed())
		Expect(err.Error()).Should(MatchRegexp(".*" + server.InsufficientResourcesError))
	})

	It("Schedule instance should choose least allocated node (1 cluster, 2 nodes, 1 pod)", func() {
		schedulerTestEnv := NewSchedulerTestEnv(SchedulerOptions{
			Clusters: []ClusterOptions{{
				Nodes: []*NodeOptions{
					{
						AllocatableCpu:    "100",
						AllocatableMemory: "100Gi",
						Pods: []*corev1.Pod{
							NewExistingPod("pod0", "50", "50Gi"), // 50% allocated
						},
					},
					{
						AllocatableCpu:    "100",
						AllocatableMemory: "100Gi",
					},
				},
			}},
		})
		log.Info("schedulerTestEnv", "clusters", schedulerTestEnv.SchedulerOpts.Clusters)
		defer schedulerTestEnv.Stop()
		schedulerTestEnv.Start()
		req := NewScheduleRequest("i3", "tiny", 4, "8Gi")
		log.Info("NewScheduleRequest", "req", req)
		resp, err := schedulerTestEnv.ScheduleOneInstance(ctx, req)
		Expect(err).Should(Succeed())
		log.Info("resp", "resp", resp)
		Expect(resp.ClusterId).Should(Equal("cluster0"))
		Expect(resp.NodeId).Should(Equal("node1"))
	})

	It("Schedule instance should choose least allocated node (2 clusters, 1 node/cluster, 1 pod)", func() {
		schedulerTestEnv := NewSchedulerTestEnv(SchedulerOptions{
			Clusters: []ClusterOptions{
				{
					Nodes: []*NodeOptions{
						{
							AllocatableCpu:    "100",
							AllocatableMemory: "100Gi",
							Pods: []*corev1.Pod{
								NewExistingPod("pod0", "50", "50Gi"), // 50% allocated
							},
						},
					},
				},
				{
					Nodes: []*NodeOptions{
						{
							AllocatableCpu:    "100",
							AllocatableMemory: "100Gi",
						},
					},
				},
			},
		})
		log.Info("schedulerTestEnv", "clusters", schedulerTestEnv.SchedulerOpts.Clusters)
		defer schedulerTestEnv.Stop()
		schedulerTestEnv.Start()
		req := NewScheduleRequest("i3", "tiny", 4, "8Gi")
		log.Info("NewScheduleRequest", "req", req)
		resp, err := schedulerTestEnv.ScheduleOneInstance(ctx, req)
		Expect(err).Should(Succeed())
		log.Info("resp", "resp", resp)
		Expect(resp.ClusterId).Should(Equal("cluster1"))
		Expect(resp.NodeId).Should(Equal("node0"))
	})

	It("Schedule instance should fail with insufficient memory due to assumed pod (1 cluster, 1 node)", func() {
		schedulerTestEnv := NewSchedulerTestEnv(SchedulerOptions{
			Clusters: []ClusterOptions{{
				Nodes: []*NodeOptions{{
					AllocatableCpu:    "10",
					AllocatableMemory: "10Gi",
				}},
			}},
		})
		defer schedulerTestEnv.Stop()
		schedulerTestEnv.Start()
		By("Scheduling 1st instance should succeed")
		req := NewScheduleRequest("i1", "tiny", 4, "8Gi")
		resp, err := schedulerTestEnv.ScheduleOneInstance(ctx, req)
		Expect(err).Should(Succeed())
		log.Info("resp", "resp", resp)
		Expect(resp.ClusterId).Should(Equal("cluster0"))
		Expect(resp.NodeId).Should(Equal("node0"))

		By("Scheduling 2nd instance should fail")
		req = NewScheduleRequest("i2", "tiny", 4, "8Gi")
		resp, err = schedulerTestEnv.ScheduleOneInstance(ctx, req)
		log.Info("resp", "resp", resp)
		Expect(err).ShouldNot(Succeed())
		Expect(err.Error()).Should(MatchRegexp(".*" + server.InsufficientResourcesError))
	})

	It("Scheduler cache should be consistent when instance runs on a node different from what was scheduled. (1 cluster, 2 nodes, 1 pod)", func() {
		schedulerTestEnv := NewSchedulerTestEnv(SchedulerOptions{
			Clusters: []ClusterOptions{{
				Nodes: []*NodeOptions{
					{
						AllocatableCpu:    "100",
						AllocatableMemory: "100Gi",
						Pods: []*corev1.Pod{
							NewExistingPod("pod0", "50", "50Gi"), // 50% allocated
						},
					},
					{
						AllocatableCpu:    "100",
						AllocatableMemory: "100Gi",
					},
				},
			}},
		})
		defer schedulerTestEnv.Stop()
		schedulerTestEnv.Start()
		sched := schedulerTestEnv.SchedulingServer.SchedulingService.Sched
		By("Scheduling instance should succeed")
		req := NewScheduleRequest("i1", "tiny", 4, "8Gi")
		log.Info("NewScheduleRequest", "req", req)
		resp, err := schedulerTestEnv.ScheduleOneInstance(ctx, req)
		Expect(err).Should(Succeed())
		log.Info("resp", "resp", resp)
		Expect(resp.ClusterId).Should(Equal("cluster0"))
		Expect(resp.NodeId).Should(Equal("node1"))

		By("Creating namespace for instance (simulate Instance Replicator + Instance Operator)")
		pod := InstanceToExpectedPod(sched, req.Instances[0], resp.NodeId)
		_, err = schedulerTestEnv.Clientsets[0].CoreV1().Namespaces().Create(ctx, NewNamespace(pod.Namespace), metav1.CreateOptions{})
		Expect(err).Should(Succeed())

		By("Checking that scheduler assumes pod")
		dump := sched.Cache.Dump()
		Expect(dump.AssumedPods.Has(req.Instances[0].Metadata.ResourceId)).Should(BeTrue())

		By("Creating pod for instance in non-recommended node (simulate Instance Replicator + Instance Operator + Kubevirt)")
		pod.Spec.NodeName = "node0"
		pod, err = schedulerTestEnv.Clientsets[0].CoreV1().Pods(pod.Namespace).Create(ctx, pod, metav1.CreateOptions{})
		Expect(err).Should(Succeed())
		log.Info("Created pod", "pod", pod)

		By("Waiting for scheduler to find new pod and no longer assume it")
		// The following will be logged:
		// I0211 04:03:47.462919      27 cache.go:467] "Pod was added to a different node than it was assumed" pod="fc7c4c61-01f8-43c2-911d-ffa5f7935217/5a125003-571d-4dc9-9843-1767c6feeadf" assumedNode="cluster0/node0" currentNode="cluster0/node1"
		Eventually(func(g Gomega) {
			dump := sched.Cache.Dump()
			g.Expect(dump.AssumedPods.Has(req.Instances[0].Metadata.ResourceId)).Should(BeFalse())
		}, "5s")
	})

	It("Deleted instance should make resources allocatable (1 cluster, 1 node)", func() {
		schedulerTestEnv := NewSchedulerTestEnv(SchedulerOptions{
			Clusters: []ClusterOptions{{
				Nodes: []*NodeOptions{{
					AllocatableCpu:    "10",
					AllocatableMemory: "10Gi",
				}},
			}},
		})
		defer schedulerTestEnv.Stop()
		schedulerTestEnv.Start()
		sched := schedulerTestEnv.SchedulingServer.SchedulingService.Sched
		By("Scheduling 1st instance should succeed")
		req := NewScheduleRequest("i1", "tiny", 4, "8Gi")
		resp, err := schedulerTestEnv.ScheduleOneInstance(ctx, req)
		Expect(err).Should(Succeed())
		log.Info("resp", "resp", resp)
		Expect(resp.ClusterId).Should(Equal("cluster0"))
		Expect(resp.NodeId).Should(Equal("node0"))

		By("Creating namespace for 1st instance (simulate Instance Replicator + Instance Operator)")
		pod := InstanceToExpectedPod(sched, req.Instances[0], resp.NodeId)
		_, err = schedulerTestEnv.Clientsets[0].CoreV1().Namespaces().Create(ctx, NewNamespace(pod.Namespace), metav1.CreateOptions{})
		Expect(err).Should(Succeed())

		By("Checking that scheduler assumes pod")
		dump := sched.Cache.Dump()
		Expect(dump.AssumedPods.Has(req.Instances[0].Metadata.ResourceId)).Should(BeTrue())

		By("Scheduling 2nd instance should fail due to assumed pod")
		req2 := NewScheduleRequest("i2", "tiny", 4, "8Gi")
		_, err = schedulerTestEnv.ScheduleOneInstance(ctx, req2)
		Expect(err).ShouldNot(Succeed())
		Expect(err.Error()).Should(MatchRegexp(".*" + server.InsufficientResourcesError))

		By("Creating pod for 1st instance (simulate Instance Replicator + Instance Operator + Kubevirt)")
		pod, err = schedulerTestEnv.Clientsets[0].CoreV1().Pods(pod.Namespace).Create(ctx, pod, metav1.CreateOptions{})
		Expect(err).Should(Succeed())
		log.Info("Created pod", "pod", pod)

		By("Waiting for scheduler to find new pod and no longer assume it")
		Eventually(func(g Gomega) {
			dump := sched.Cache.Dump()
			g.Expect(dump.AssumedPods.Has(req.Instances[0].Metadata.ResourceId)).Should(BeFalse())
		}, "5s")

		By("Scheduling 3rd instance should fail due to non-assumed pod")
		req3 := NewScheduleRequest("i3", "tiny", 4, "8Gi")
		_, err = schedulerTestEnv.ScheduleOneInstance(ctx, req3)
		Expect(err).ShouldNot(Succeed())
		Expect(err.Error()).Should(MatchRegexp(".*" + server.InsufficientResourcesError))

		By("Deleting pod (simulate Instance Replicator + Instance Operator + Kubevirt)")
		err = schedulerTestEnv.Clientsets[0].CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
		Expect(err).Should(Succeed())

		By("Waiting for scheduling of 4th pod to succeed after resources are freed up")
		Eventually(func(g Gomega) {
			req4 := NewScheduleRequest("i4", "tiny", 4, "8Gi")
			resp4, err := schedulerTestEnv.ScheduleOneInstance(ctx, req4)
			Expect(err).Should(Succeed())
			log.Info("resp", "resp", resp4)
			Expect(resp4.ClusterId).Should(Equal("cluster0"))
			Expect(resp4.NodeId).Should(Equal("node0"))
		}, "5s")
	})

	It("2 instances should be scheduled on the same node because the node remains least allocated (1 cluster, 2 nodes)", func() {
		schedulerTestEnv := NewSchedulerTestEnv(SchedulerOptions{
			Clusters: []ClusterOptions{{
				// cluster0
				Nodes: []*NodeOptions{
					// node0
					{
						AllocatableCpu:    "100",
						AllocatableMemory: "100Gi",
						Pods: []*corev1.Pod{
							NewExistingPod("pod0", "50", "50Gi"), // 50% allocated
						},
					},
					// node1
					{
						AllocatableCpu:    "100",
						AllocatableMemory: "100Gi",
					},
				},
			}},
		})
		defer schedulerTestEnv.Stop()
		schedulerTestEnv.Start()

		By("Scheduling 1st instance should succeed, on node1")
		req1 := NewScheduleRequest("i1", "tiny", 4, "8Gi")
		resp1, err := schedulerTestEnv.ScheduleOneInstance(ctx, req1)
		Expect(err).Should(Succeed())
		log.Info("resp1", "resp1", resp1)
		Expect(resp1.ClusterId).Should(Equal("cluster0"))
		Expect(resp1.NodeId).Should(Equal("node1"))

		By("Scheduling 2nd instance should succeed, on node1")
		req2 := NewScheduleRequest("i1", "tiny", 4, "8Gi")
		resp2, err := schedulerTestEnv.ScheduleOneInstance(ctx, req2)
		Expect(err).Should(Succeed())
		log.Info("resp2", "resp2", resp2)
		Expect(resp2.ClusterId).Should(Equal("cluster0"))
		Expect(resp2.NodeId).Should(Equal("node1"))
	})

	It("When there is only 1 node with an existing instance with a topology spread constraint, scheduling a new instance with a topology spread constraint should succeed (1 cluster, 1 nodes)", func() {
		instance1 := NewInstance("i1", "tiny", 4, "8Gi")
		SetTopologySpreadConstraint(instance1)
		schedulerTestEnv := NewSchedulerTestEnv(SchedulerOptions{
			Clusters: []ClusterOptions{{
				Nodes: []*NodeOptions{
					{
						AllocatableCpu:    "100",
						AllocatableMemory: "100Gi",
						Instances: []*pb.InstancePrivate{
							instance1,
						},
					},
				},
			}},
		})
		defer schedulerTestEnv.Stop()
		schedulerTestEnv.Start()

		req2 := NewScheduleRequest("i1", "tiny", 4, "8Gi")
		SetTopologySpreadConstraint(req2.Instances[0])
		req2.Instances[0].Metadata.CloudAccountId = instance1.Metadata.CloudAccountId
		_, err := schedulerTestEnv.ScheduleOneInstance(ctx, req2)
		Expect(err).Should(Succeed())
	})

	It("When there are 2 nodes and 1 existing instance with a topology spread constraint, a new instance with a topology spread constraint and a different CloudAccount should not be affected by the topology spread constraint (1 cluster, 2 nodes)", func() {
		instance1 := NewInstance("i1", "tiny", 4, "8Gi")
		SetTopologySpreadConstraint(instance1)
		schedulerTestEnv := NewSchedulerTestEnv(SchedulerOptions{
			Clusters: []ClusterOptions{{
				// cluster0
				Nodes: []*NodeOptions{
					// node0
					{
						AllocatableCpu:    "100",
						AllocatableMemory: "100Gi",
						Pods: []*corev1.Pod{
							NewExistingPod("pod0", "50", "50Gi"), // 50% allocated
						},
						Partition: "us-dev-1a-p0",
					},
					// node1
					{
						AllocatableCpu:    "100",
						AllocatableMemory: "100Gi",
						Instances: []*pb.InstancePrivate{
							instance1,
						},
						Partition: "us-dev-1a-p1",
					},
				},
			}},
		})
		defer schedulerTestEnv.Stop()
		schedulerTestEnv.Start()

		By("Scheduling 1st instance should succeed, on node1 (least allocated)")
		req1 := NewScheduleRequest("i1", "tiny", 4, "8Gi")
		SetTopologySpreadConstraint(req1.Instances[0])
		resp1, err := schedulerTestEnv.ScheduleOneInstance(ctx, req1)
		Expect(err).Should(Succeed())
		log.Info("resp1", "resp1", resp1)
		Expect(resp1.ClusterId).Should(Equal("cluster0"))
		Expect(resp1.NodeId).Should(Equal("node1"))
	})

	It("When there are 2 nodes and 1 existing instance with a topology spread constraint, a new instance with a topology spread constraint should be scheduled on the other node (1 cluster, 2 nodes)", func() {
		instance1 := NewInstance("i1", "tiny", 4, "8Gi")
		SetTopologySpreadConstraint(instance1)
		schedulerTestEnv := NewSchedulerTestEnv(SchedulerOptions{
			Clusters: []ClusterOptions{{
				// cluster0
				Nodes: []*NodeOptions{
					// node0
					{
						AllocatableCpu:    "100",
						AllocatableMemory: "100Gi",
						Pods: []*corev1.Pod{
							NewExistingPod("pod0", "50", "50Gi"), // 50% allocated
						},
						Partition: "us-dev-1a-p0",
					},
					// node1
					{
						AllocatableCpu:    "100",
						AllocatableMemory: "100Gi",
						Instances: []*pb.InstancePrivate{
							instance1,
						},
						Partition: "us-dev-1a-p1",
					},
				},
			}},
		})
		defer schedulerTestEnv.Stop()
		schedulerTestEnv.Start()

		By("Scheduling 1st instance should succeed, on node0")
		req1 := NewScheduleRequest("i1", "tiny", 4, "8Gi")
		SetTopologySpreadConstraint(req1.Instances[0])
		req1.Instances[0].Metadata.CloudAccountId = instance1.Metadata.CloudAccountId
		resp1, err := schedulerTestEnv.ScheduleOneInstance(ctx, req1)
		Expect(err).Should(Succeed())
		log.Info("resp1", "resp1", resp1)
		Expect(resp1.ClusterId).Should(Equal("cluster0"))
		Expect(resp1.NodeId).Should(Equal("node0"))
	})

	It("When there are 2 nodes, 2 instances with topology spread constraints should be scheduled on different nodes (1 cluster, 2 nodes, 1 instance assumed)", func() {
		schedulerTestEnv := NewSchedulerTestEnv(SchedulerOptions{
			Clusters: []ClusterOptions{{
				// cluster0
				Nodes: []*NodeOptions{
					// node0
					{
						AllocatableCpu:    "100",
						AllocatableMemory: "100Gi",
						Pods: []*corev1.Pod{
							NewExistingPod("pod0", "50", "50Gi"), // 50% allocated
						},
						Partition: "us-dev-1a-p0",
					},
					// node1
					{
						AllocatableCpu:    "100",
						AllocatableMemory: "100Gi",
						Partition:         "us-dev-1a-p1",
					},
				},
			}},
		})
		defer schedulerTestEnv.Stop()
		schedulerTestEnv.Start()

		By("Scheduling 1st instance should succeed, on node1")
		req1 := NewScheduleRequest("i1", "tiny", 4, "8Gi")
		SetTopologySpreadConstraint(req1.Instances[0])
		resp1, err := schedulerTestEnv.ScheduleOneInstance(ctx, req1)
		Expect(err).Should(Succeed())
		log.Info("resp1", "resp1", resp1)
		Expect(resp1.ClusterId).Should(Equal("cluster0"))
		Expect(resp1.NodeId).Should(Equal("node1"))

		By("Scheduling 2nd instance should succeed, on node0")
		req2 := NewScheduleRequest("i1", "tiny", 4, "8Gi")
		SetTopologySpreadConstraint(req2.Instances[0])
		req2.Instances[0].Metadata.CloudAccountId = req1.Instances[0].Metadata.CloudAccountId
		resp2, err := schedulerTestEnv.ScheduleOneInstance(ctx, req2)
		Expect(err).Should(Succeed())
		log.Info("resp2", "resp2", resp2)
		Expect(resp2.ClusterId).Should(Equal("cluster0"))
		Expect(resp2.NodeId).Should(Equal("node0"))
	})

	It("Schedule 2 instances atomically should succeed (1 cluster, 1 node)", func() {
		schedulerTestEnv := NewSchedulerTestEnv(SchedulerOptions{
			Clusters: []ClusterOptions{{
				Nodes: []*NodeOptions{{
					AllocatableCpu:    "100",
					AllocatableMemory: "100Gi",
				}},
			}},
		})
		defer schedulerTestEnv.Stop()
		schedulerTestEnv.Start()
		req := NewScheduleResourcesRequest(2, "i1", "tiny", 4, "8Gi")
		resp, err := schedulerTestEnv.SchedulerServiceClient.Schedule(ctx, req)
		Expect(err).Should(Succeed())
		log.Info("resp", "resp", resp)
		Expect(len(resp.InstanceResults)).Should(Equal(len(req.Instances)))
		Expect(resp.InstanceResults[0].ClusterId).Should(Equal("cluster0"))
		Expect(resp.InstanceResults[1].ClusterId).Should(Equal("cluster0"))
		Expect(resp.InstanceResults[0].NodeId).Should(Equal("node0"))
		Expect(resp.InstanceResults[1].NodeId).Should(Equal("node0"))
	})

	It("Schedule 3 instances atomically with insufficent cpus should fail and free up resources so that 2 instances can be scheduled (1 cluster, 1 node)", func() {
		schedulerTestEnv := NewSchedulerTestEnv(SchedulerOptions{
			Clusters: []ClusterOptions{{
				Nodes: []*NodeOptions{{
					AllocatableCpu:    "8", // enough for 2 instances but not 3
					AllocatableMemory: "100Gi",
				}},
			}},
		})
		defer schedulerTestEnv.Stop()
		schedulerTestEnv.Start()

		By("Schedule 3 instances atomically with insufficent cpus should fail")
		req := NewScheduleResourcesRequest(3, "i1", "tiny", 4, "8Gi")
		resp, err := schedulerTestEnv.SchedulerServiceClient.Schedule(ctx, req)
		Expect(err).ShouldNot(Succeed())
		log.Info("resp", "resp", resp)

		By("Ensure that resources were freed by scheduling 2 instances")
		req2 := NewScheduleResourcesRequest(2, "i1", "tiny", 4, "8Gi")
		resp2, err := schedulerTestEnv.SchedulerServiceClient.Schedule(ctx, req2)
		Expect(err).Should(Succeed())
		log.Info("resp", "resp2", resp2)
	})

	It("When Harvester overcommit config settings are set (cpu:200 memory:100 and storage:200) scheduler should patch the new settings", func() {
		updatedOverCommitConfig := cloudv1alpha1.OvercommitConfig{
			CPU:     200,
			Memory:  100,
			Storage: 200,
		}
		gvr := schema.GroupVersionResource{
			Group:    "harvesterhci.io",
			Version:  "v1beta1",
			Resource: "settings",
		}
		schedulerTestEnv := NewSchedulerTestEnv(SchedulerOptions{
			Clusters: []ClusterOptions{{
				Nodes: []*NodeOptions{{
					AllocatableCpu:    "100",
					AllocatableMemory: "100Gi",
				}},
			}},
			OvercommitConfig: &updatedOverCommitConfig,
		})
		defer schedulerTestEnv.Stop()
		schedulerTestEnv.Start()

		for _, dynamicClientSet := range schedulerTestEnv.DynamicClientSets {
			obj, err := dynamicClientSet.Resource(gvr).Get(ctx, "overcommit-config", metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(obj.UnstructuredContent()["value"].(string)).Should(Equal(fmt.Sprintf(`{"cpu":%d,"memory":%d,"storage":%d}`, 200, 100, 200)))
		}
	})

	It("Instance scheduler should consider node pool labels", func() {
		schedulerTestEnv := NewSchedulerTestEnv(SchedulerOptions{
			Clusters: []ClusterOptions{{
				// cluster0
				Nodes: []*NodeOptions{
					// node0
					{
						AllocatableCpu:    "100",
						AllocatableMemory: "100Gi",
						Partition:         "us-dev-1a-p0",
					},
					// node1
					{
						AllocatableCpu:    "100",
						AllocatableMemory: "100Gi",
						Pods: []*corev1.Pod{
							NewExistingPod("pod0", "50", "50Gi"), // 50% allocated
						},
						Partition: "us-dev-1a-p1",
						NodeLabels: map[string]string{
							"pool.cloud.intel.com/general": "true",
						},
					},
				},
			}},
		})
		defer schedulerTestEnv.Stop()
		schedulerTestEnv.Start()
		req := NewScheduleRequest("i1", "tiny", 4, "8Gi")
		req.Instances[0].Spec.ComputeNodePools = []string{"general"}
		resp, err := schedulerTestEnv.ScheduleOneInstance(ctx, req)
		Expect(err).Should(Succeed())
		log.Info("resp", "resp", resp)
		Expect(resp.ClusterId).Should(Equal("cluster0"))
		Expect(resp.NodeId).Should(Equal("node1"))
		Expect(resp.ComputeNodePools[0]).Should(Equal("general"))
	})

	It("The final node pool labels should be the intersection of the compute node pools specified in the request and those recommended node", func() {
		schedulerOptions := SchedulerOptions{
			Clusters: []ClusterOptions{{
				// cluster0
				Nodes: []*NodeOptions{
					// node0
					{
						AllocatableCpu:    "100",
						AllocatableMemory: "100Gi",
						Partition:         "us-dev-1a-p0",
						NodeLabels: map[string]string{
							"pool.cloud.intel.com/general": "true",
							"pool.cloud.intel.com/habana":  "true",
						},
					},
					// node1
					{
						AllocatableCpu:    "100",
						AllocatableMemory: "100Gi",
						Partition:         "us-dev-1a-p1",
						NodeLabels: map[string]string{
							"pool.cloud.intel.com/habana": "true",
						},
					},
				},
			}},
		}
		schedulerTestEnv := NewSchedulerTestEnv(schedulerOptions)
		defer schedulerTestEnv.Stop()
		schedulerTestEnv.Start()
		req := NewScheduleRequest("i1", "tiny", 4, "8Gi")
		req.Instances[0].Spec.ComputeNodePools = []string{"general", "test"}
		resp, err := schedulerTestEnv.ScheduleOneInstance(ctx, req)
		Expect(err).Should(Succeed())
		log.Info("resp", "resp", resp)
		Expect(resp.ClusterId).Should(Equal("cluster0"))
		Expect(resp.NodeId).Should(Equal("node0"))
		Expect(len(resp.ComputeNodePools)).Should(Equal(1))
		Expect(resp.ComputeNodePools[0]).Should(Equal("general"), "Final node pool labels should be intersection of compute node pools in request and those in recommended node")
	})

	It("Instance scheduler should fail when no node with maching label exists", func() {
		schedulerOptions := SchedulerOptions{
			Clusters: []ClusterOptions{{
				// cluster0
				Nodes: []*NodeOptions{
					// node0
					{
						AllocatableCpu:    "100",
						AllocatableMemory: "100Gi",
						Partition:         "us-dev-1a-p0",
						NodeLabels: map[string]string{
							"pool.cloud.intel.com/general": "true",
							"pool.cloud.intel.com/habana":  "true",
						},
					},
				},
			}},
		}
		schedulerTestEnv := NewSchedulerTestEnv(schedulerOptions)
		defer schedulerTestEnv.Stop()
		schedulerTestEnv.Start()
		req := NewScheduleRequest("i1", "tiny", 4, "8Gi")
		req.Instances[0].Spec.ComputeNodePools = []string{"test"}
		_, err := schedulerTestEnv.ScheduleOneInstance(ctx, req)
		Expect(err).ShouldNot(Succeed())
		Expect(err.Error()).Should(MatchRegexp(".*" + server.InsufficientResourcesError))
	})

	It("Instance scheduler ignore compute node pools when an instance has an empty ComputeNodePools list", func() {
		schedulerOptions := SchedulerOptions{
			Clusters: []ClusterOptions{{
				// cluster0
				Nodes: []*NodeOptions{
					// node0
					{
						AllocatableCpu:    "100",
						AllocatableMemory: "100Gi",
						Partition:         "us-dev-1a-p0",
						NodeLabels: map[string]string{
							"pool.cloud.intel.com/general": "true",
							"pool.cloud.intel.com/habana":  "true",
						},
					},
				},
			}},
		}
		schedulerTestEnv := NewSchedulerTestEnv(schedulerOptions)
		defer schedulerTestEnv.Stop()
		schedulerTestEnv.Start()
		req := NewScheduleRequest("i1", "tiny", 4, "8Gi")
		resp, err := schedulerTestEnv.ScheduleOneInstance(ctx, req)
		Expect(err).Should(Succeed())
		log.Info("resp", "resp", resp)
		Expect(resp.ClusterId).Should(Equal("cluster0"))
		Expect(resp.NodeId).Should(Equal("node0"))
	})
})
