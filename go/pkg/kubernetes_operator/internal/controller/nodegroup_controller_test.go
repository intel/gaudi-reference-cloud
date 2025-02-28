// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controller

import (
	"context"
	"fmt"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"strings"
	"testing"
)

func TestRequiredActiveNodes(t *testing.T) {
	tests := []struct {
		Description    string
		GivenNumber    int
		GivenNodes     []privatecloudv1alpha1.NodeStatus
		ExpectedResult bool
	}{
		{
			"Should return true since there are 2 active nodes",
			2,
			[]privatecloudv1alpha1.NodeStatus{
				{
					Name:  "node01",
					State: privatecloudv1alpha1.ActiveNodegroupState,
				},
				{
					Name:  "node02",
					State: privatecloudv1alpha1.ActiveNodegroupState,
				},
				{
					Name:  "node03",
					State: privatecloudv1alpha1.ActiveNodegroupState,
				},
			},
			true,
		},
		{
			"Should return false since there isn't an active node",
			1,
			[]privatecloudv1alpha1.NodeStatus{
				{
					Name:  "node01",
					State: privatecloudv1alpha1.UpdatingNodegroupState,
				},
				{
					Name:  "node02",
					State: privatecloudv1alpha1.UpdatingNodegroupState,
				},
				{
					Name:  "node03",
					State: privatecloudv1alpha1.ErrorNodegroupState,
				},
			},
			false,
		},
	}

	for _, test := range tests {
		res := requiredActiveNodes(test.GivenNumber, test.GivenNodes)
		assert.Assert(t, cmp.Equal(test.ExpectedResult, res), test.Description)
	}
}

func TestRemoveDeletedNode(t *testing.T) {
	tests := []struct {
		Description          string
		GivenDeletedNodeName string
		GivenNodes           []privatecloudv1alpha1.NodeStatus
		ExpectedNodes        []privatecloudv1alpha1.NodeStatus
	}{
		{
			"Should return nodes without node02",
			"node02",
			[]privatecloudv1alpha1.NodeStatus{
				{
					Name:  "node01",
					State: privatecloudv1alpha1.ActiveNodegroupState,
				},
				{
					Name:  "node02",
					State: privatecloudv1alpha1.DeletingNodegroupState,
				},
				{
					Name:  "node03",
					State: privatecloudv1alpha1.ActiveNodegroupState,
				},
			},
			[]privatecloudv1alpha1.NodeStatus{
				{
					Name:  "node01",
					State: privatecloudv1alpha1.ActiveNodegroupState,
				},
				{
					Name:  "node03",
					State: privatecloudv1alpha1.ActiveNodegroupState,
				},
			},
		},
	}

	for _, test := range tests {
		updatedNodes := removeDeletedNode(test.GivenNodes, test.GivenDeletedNodeName)
		assert.Assert(t, cmp.DeepEqual(test.ExpectedNodes, updatedNodes), test.Description)
	}
}

func TestGetInstanceGroupName(t *testing.T) {
	tests := []struct {
		GivenNodeName             string
		ExpectedInstanceGroupName string
	}{
		{"ng-ntat6s264e-ig-12345-0", "ng-ntat6s264e-ig-12345"},
		{"ng-ntat6s264e-ig-abcde-4", "ng-ntat6s264e-ig-abcde"},
		{"ng-ntat6s264e-ig-zzzzz-2", "ng-ntat6s264e-ig-zzzzz"},
	}

	for _, test := range tests {
		instanceGroupName := getInstanceGroupName(test.GivenNodeName)
		assert.Assert(t, cmp.Equal(test.ExpectedInstanceGroupName, instanceGroupName))
	}
}

func TestIsInstanceGroup(t *testing.T) {
	tests := []struct {
		GivenNodegroup *privatecloudv1alpha1.Nodegroup
		ExpectedResult bool
	}{
		{
			&privatecloudv1alpha1.Nodegroup{
				Spec: privatecloudv1alpha1.NodegroupSpec{
					InstanceType: "bm-icp-gaudi2-cluster-4",
				},
			},
			true,
		},
		{
			&privatecloudv1alpha1.Nodegroup{
				Spec: privatecloudv1alpha1.NodegroupSpec{
					InstanceType: "bm-icp-gaudi2",
				},
			},
			false,
		},
		{
			&privatecloudv1alpha1.Nodegroup{
				Spec: privatecloudv1alpha1.NodegroupSpec{
					InstanceType: "vm-spr-sml",
				},
			},
			false,
		},
	}

	for _, test := range tests {
		res := isInstanceGroup(test.GivenNodegroup)
		assert.Assert(t, cmp.Equal(test.ExpectedResult, res))
	}
}

type MockNodeProvider struct{}

func (p *MockNodeProvider) CreateNode(ctx context.Context, _ string, nameserver string, gateway string, registrationCmd string, bootstrapScript string, nodegroup privatecloudv1alpha1.Nodegroup) (privatecloudv1alpha1.NodeStatus, error) {
	return privatecloudv1alpha1.NodeStatus{}, nil
}

func (p *MockNodeProvider) GetNodes(ctx context.Context, selector string, cloudaccountid string) ([]privatecloudv1alpha1.NodeStatus, error) {
	return []privatecloudv1alpha1.NodeStatus{}, nil
}

func (p *MockNodeProvider) GetNode(ctx context.Context, nodeName string, cloudaccountid string) (privatecloudv1alpha1.NodeStatus, error) {
	nodes := []privatecloudv1alpha1.NodeStatus{
		{
			Name:        "ng-jqqottnc4q-b9180",
			IpAddress:   "100.80.65.2",
			InstanceIMI: "iks-vm-u22-cd-wk-1-28-7-v20240227",
			State:       privatecloudv1alpha1.ActiveNodegroupState,
		},
		{
			Name:        "ng-jqqottnc4q-b9182",
			IpAddress:   "100.80.65.4",
			InstanceIMI: "iks-vm-u22-cd-wk-1-28-7-v20240227",
			State:       privatecloudv1alpha1.ErrorNodegroupState,
		},
		{
			Name:        "ng-jqqottnc4q-b9181",
			IpAddress:   "100.80.65.5",
			InstanceIMI: "iks-vm-u22-cd-wk-1-28-7-v20240227",
			State:       privatecloudv1alpha1.ActiveNodegroupState,
		},
		{
			Name:        "ng-jqqottnc4q-b918a",
			IpAddress:   "100.80.65.6",
			InstanceIMI: "iks-vm-u22-cd-wk-1-28-7-v20240227",
			State:       privatecloudv1alpha1.ActiveNodegroupState,
		},
		{
			Name:        "ng-jqqottnc4q-b918b",
			IpAddress:   "100.80.65.7",
			InstanceIMI: "iks-vm-u22-cd-wk-1-28-7-v20240227",
			State:       privatecloudv1alpha1.ErrorNodegroupState,
		},
	}

	for _, node := range nodes {
		if node.Name == nodeName {
			return node, nil
		}
	}

	return privatecloudv1alpha1.NodeStatus{}, grpcstatus.Error(grpccodes.NotFound, "node not found")
}

func (p *MockNodeProvider) DeleteNode(ctx context.Context, nodeName string, cloudaccountid string) error {
	return nil
}

func (p *MockNodeProvider) CreateInstanceGroup(ctx context.Context, registrationCmd string, bootstrapScript string, instanceType string, instanceCount int, nodegroup privatecloudv1alpha1.Nodegroup) ([]privatecloudv1alpha1.NodeStatus, string, error) {
	return []privatecloudv1alpha1.NodeStatus{}, "", nil
}

func (p *MockNodeProvider) CreatePrivateInstanceGroup(ctx context.Context, registrationCmd string, bootstrapScript string, instanceType string, instanceCount int, nodegroup privatecloudv1alpha1.Nodegroup) ([]privatecloudv1alpha1.NodeStatus, string, error) {
	return []privatecloudv1alpha1.NodeStatus{}, "", nil
}

func (p *MockNodeProvider) DeleteInstanceGroupMember(ctx context.Context, nodeName string, cloudaccountid string, instanceGroup string) error {
	return nil
}

func (p *MockNodeProvider) ScaleUpInstanceGroup(context.Context, string, string, string, int, privatecloudv1alpha1.Nodegroup, string) ([]privatecloudv1alpha1.NodeStatus, string, error) {
	return nil, "", nil
}

func (p *MockNodeProvider) SearchInstanceGroup(ctx context.Context, cloudaccountid string, instanceGroup string) (bool, error) {
	return false, nil
}

type MockKubernetesProvider struct{}

func (p *MockKubernetesProvider) InitCluster(ctx context.Context, secret *corev1.Secret, cluster *privatecloudv1alpha1.Cluster, etcdLB string, apiserverLB string, publicApiserverLB string, konnectivityLB string, etcdLBPort int, apiserverLBPort int, publicApiserverLBPort int) error {
	return nil
}

func (p *MockKubernetesProvider) GetCluster(ctx context.Context) (*privatecloudv1alpha1.ClusterStatus, error) {
	return &privatecloudv1alpha1.ClusterStatus{}, nil
}

func (p *MockKubernetesProvider) CleanUpCluster(context.Context, string) error {
	return nil
}

func (p *MockKubernetesProvider) GetNodes(ctx context.Context, nodegroupName string) ([]privatecloudv1alpha1.NodeStatus, error) {
	return []privatecloudv1alpha1.NodeStatus{}, nil
}

func (p *MockKubernetesProvider) GetNode(ctx context.Context, nodeName string) (privatecloudv1alpha1.NodeStatus, error) {
	nodes := []privatecloudv1alpha1.NodeStatus{
		{
			Name:               "ng-jqqottnc4q-b9180",
			IpAddress:          "100.80.65.2",
			State:              privatecloudv1alpha1.ActiveNodegroupState,
			KubeletVersion:     "v1.28.7",
			KubeProxyVersion:   "v1.28.7",
			Unschedulable:      false,
			AutoRepairDisabled: true,
		},
		{
			Name:               "ng-jqqottnc4q-b9183",
			IpAddress:          "100.80.65.3",
			State:              privatecloudv1alpha1.UpdatingNodegroupState,
			KubeletVersion:     "v1.28.7",
			KubeProxyVersion:   "v1.28.7",
			Unschedulable:      false,
			AutoRepairDisabled: true,
		},
		{
			Name:               "ng-jqqottnc4q-b9182",
			IpAddress:          "100.80.65.4",
			State:              privatecloudv1alpha1.UpdatingNodegroupState,
			InstanceIMI:        "iks-vm-u22-cd-wk-1-28-7-v20240227",
			KubeletVersion:     "v1.28.7",
			KubeProxyVersion:   "v1.28.7",
			Unschedulable:      false,
			AutoRepairDisabled: false,
		},
		{
			Name:               "ng-jqqottnc4q-b9181",
			IpAddress:          "100.80.65.5",
			State:              privatecloudv1alpha1.UpdatingNodegroupState,
			KubeletVersion:     "v1.28.7",
			KubeProxyVersion:   "v1.28.7",
			Unschedulable:      false,
			AutoRepairDisabled: false,
		},
	}

	for _, node := range nodes {
		if node.Name == nodeName {
			return node, nil
		}
	}

	return privatecloudv1alpha1.NodeStatus{}, k8serrors.NewNotFound(corev1.Resource("node"), nodeName)
}

func (p *MockKubernetesProvider) DeleteNode(ctx context.Context, nodeName string) error {
	return nil
}

func (p *MockKubernetesProvider) DrainNode(ctx context.Context, nodeName string) error {
	return nil
}

func (p *MockKubernetesProvider) GetBootstrapScript(nodegroupType privatecloudv1alpha1.NodegroupType) (string, error) {
	return "", nil
}

func (p *MockKubernetesProvider) CreateBootstrapTokenSecret(ctx context.Context, secret *corev1.Secret) error {
	return nil
}

func (p *MockKubernetesProvider) ApproveKubeletServingCertificateSigningRequests(ctx context.Context, nodeNamePrefix string) error {
	return nil
}

func (p *MockKubernetesProvider) CreateNamespace(ctx context.Context, namespace *corev1.Namespace) error {
	return nil
}

func (p *MockKubernetesProvider) CreateSecret(ctx context.Context, namespace string, secret *corev1.Secret) error {
	return nil
}

func (p *MockKubernetesProvider) GetSecret(ctx context.Context, name, namespace string) (*corev1.Secret, error) {
	return nil, nil
}

func TestGetWorkerNodeStatus(t *testing.T) {
	var clusterName = "cl-rwhm3mcroi"
	var nodeproviderName = "Compute"
	var cloudaccountid = "000000000000"

	tests := []struct {
		Description   string
		GivenNodes    []privatecloudv1alpha1.NodeStatus
		ExpectedNodes []privatecloudv1alpha1.NodeStatus
	}{
		{
			Description:   "Empty list of nodes should return empty list of nodes",
			GivenNodes:    make([]privatecloudv1alpha1.NodeStatus, 0),
			ExpectedNodes: make([]privatecloudv1alpha1.NodeStatus, 0),
		},
		{
			Description: "Node not found in node provider and found in kubernetes provider with updating state should return node in kubernetes provider state",
			GivenNodes: []privatecloudv1alpha1.NodeStatus{
				{
					Name:               "ng-jqqottnc4q-b9183",
					IpAddress:          "100.80.65.3",
					State:              privatecloudv1alpha1.UpdatingNodegroupState,
					KubeletVersion:     "v1.28.7",
					KubeProxyVersion:   "v1.28.7",
					Unschedulable:      false,
					AutoRepairDisabled: true,
				},
			},
			ExpectedNodes: []privatecloudv1alpha1.NodeStatus{
				{
					Name:               "ng-jqqottnc4q-b9183",
					IpAddress:          "100.80.65.3",
					State:              privatecloudv1alpha1.UpdatingNodegroupState,
					KubeletVersion:     "v1.28.7",
					KubeProxyVersion:   "v1.28.7",
					Unschedulable:      false,
					AutoRepairDisabled: true,
				},
			},
		},
		{
			Description: "Node not found in node provider and kubernetes provider should return node in updating state",
			GivenNodes: []privatecloudv1alpha1.NodeStatus{
				{
					Name:               "ng-jqqottnc4q-b918e",
					IpAddress:          "100.80.65.10",
					InstanceIMI:        "iks-vm-u22-cd-wk-1-28-7-v20240227",
					State:              privatecloudv1alpha1.ActiveNodegroupState,
					KubeletVersion:     "v1.28.7",
					KubeProxyVersion:   "v1.28.7",
					Unschedulable:      false,
					AutoRepairDisabled: false,
				},
			},
			ExpectedNodes: []privatecloudv1alpha1.NodeStatus{
				{
					Name:               "ng-jqqottnc4q-b918e",
					IpAddress:          "100.80.65.10",
					InstanceIMI:        "iks-vm-u22-cd-wk-1-28-7-v20240227",
					State:              privatecloudv1alpha1.UpdatingNodegroupState,
					KubeletVersion:     "v1.28.7",
					KubeProxyVersion:   "v1.28.7",
					Unschedulable:      false,
					AutoRepairDisabled: false,
					Message:            "Checking node",
					Reason:             "WorkerNotReady",
				},
			},
		},
		{
			Description: "Node found in node provider with active state and not found in kubernetes provider should return node in updating state",
			GivenNodes: []privatecloudv1alpha1.NodeStatus{
				{
					Name: "ng-jqqottnc4q-b918a",
				},
			},
			ExpectedNodes: []privatecloudv1alpha1.NodeStatus{
				{
					Name:               "ng-jqqottnc4q-b918a",
					IpAddress:          "100.80.65.6",
					InstanceIMI:        "iks-vm-u22-cd-wk-1-28-7-v20240227",
					State:              privatecloudv1alpha1.UpdatingNodegroupState,
					Unschedulable:      false,
					AutoRepairDisabled: false,
					Message:            "Checking node",
					Reason:             "WorkerNotReady",
				},
			},
		},
		{
			Description: "Node found in node provider with error state and not found in kubernetes provider should return node in updating state",
			GivenNodes: []privatecloudv1alpha1.NodeStatus{
				{
					Name: "ng-jqqottnc4q-b918b",
				},
			},
			ExpectedNodes: []privatecloudv1alpha1.NodeStatus{
				{
					Name:               "ng-jqqottnc4q-b918b",
					IpAddress:          "100.80.65.7",
					InstanceIMI:        "iks-vm-u22-cd-wk-1-28-7-v20240227",
					State:              privatecloudv1alpha1.UpdatingNodegroupState,
					Unschedulable:      false,
					AutoRepairDisabled: false,
					Message:            "Checking node",
					Reason:             "WorkerNotReady",
				},
			},
		},
		{
			Description: "Node found in node provider with active state and found in kubernetes provider with updating state should return node in updating state",
			GivenNodes: []privatecloudv1alpha1.NodeStatus{
				{
					Name: "ng-jqqottnc4q-b9181",
				},
			},
			ExpectedNodes: []privatecloudv1alpha1.NodeStatus{
				{
					Name:               "ng-jqqottnc4q-b9181",
					IpAddress:          "100.80.65.5",
					InstanceIMI:        "iks-vm-u22-cd-wk-1-28-7-v20240227",
					State:              privatecloudv1alpha1.UpdatingNodegroupState,
					Unschedulable:      false,
					AutoRepairDisabled: false,
					KubeletVersion:     "v1.28.7",
					KubeProxyVersion:   "v1.28.7",
				},
			},
		},
		{
			Description: "Node found in node provider with error state and found in kubernetes provider with updating state should return node in updating state",
			GivenNodes: []privatecloudv1alpha1.NodeStatus{
				{
					Name: "ng-jqqottnc4q-b9182",
				},
			},
			ExpectedNodes: []privatecloudv1alpha1.NodeStatus{
				{
					Name:               "ng-jqqottnc4q-b9182",
					IpAddress:          "100.80.65.4",
					State:              privatecloudv1alpha1.UpdatingNodegroupState,
					InstanceIMI:        "iks-vm-u22-cd-wk-1-28-7-v20240227",
					KubeletVersion:     "v1.28.7",
					KubeProxyVersion:   "v1.28.7",
					Unschedulable:      false,
					AutoRepairDisabled: false,
				},
			},
		},
		{
			Description: "Node found in node provider with active state and found in kubernetes provider with active state should return node in active state",
			GivenNodes: []privatecloudv1alpha1.NodeStatus{
				{
					Name: "ng-jqqottnc4q-b9180",
				},
			},
			ExpectedNodes: []privatecloudv1alpha1.NodeStatus{
				{
					Name:               "ng-jqqottnc4q-b9180",
					IpAddress:          "100.80.65.2",
					InstanceIMI:        "iks-vm-u22-cd-wk-1-28-7-v20240227",
					State:              privatecloudv1alpha1.ActiveNodegroupState,
					Unschedulable:      false,
					AutoRepairDisabled: true,
					KubeletVersion:     "v1.28.7",
					KubeProxyVersion:   "v1.28.7",
				},
			},
		},
	}

	ctx := context.Background()
	for _, test := range tests {
		nodes, err := getWorkerNodeStatus(ctx, clusterName, nodeproviderName, test.GivenNodes, &MockNodeProvider{}, &MockKubernetesProvider{}, cloudaccountid)
		assert.Assert(t, cmp.Equal(err, nil), test.Description)
		assert.Assert(t, cmp.DeepEqual(test.ExpectedNodes, nodes), test.Description)
	}
}

func Test_workerNodeAdditionalLabels(t *testing.T) {
	tests := []struct {
		name         string
		instanceType *pb.InstanceType
		machineImage *pb.MachineImage
		want         string
	}{
		{
			name: "Happy path",
			instanceType: &pb.InstanceType{
				Metadata: &pb.InstanceType_Metadata{
					Name: "test-instance",
				},
				Spec: &pb.InstanceTypeSpec{
					Name: "test",
					Cpu: &pb.CpuSpec{
						Cores:     1,
						Id:        "id",
						ModelName: "Intel",
						Sockets:   2,
						Threads:   3,
					},
					Memory: &pb.MemorySpec{
						Size: "128m",
					},
					Gpu: &pb.GpuSpec{
						ModelName: "Test GPU",
						Count:     2,
					},
					HbmMode: "flat",
				},
			},
			machineImage: &pb.MachineImage{
				Spec: &pb.MachineImageSpec{
					Components: []*pb.MachineImageComponent{
						{
							Name:    "Intel Gaudi SW",
							Type:    "Firmware kit",
							Version: "1.18.0-524",
						},
						{
							Name:    "Intel Gaudi SW",
							Type:    "Firmware kit",
							Version: "1.16.0-526",
						},
						{
							Name:    "Intel Gaudi SW",
							Type:    "Software kit",
							Version: "1.16.0",
						},
					},
				},
			},
			want: "cloud.intel.com/hbm-mode=flat,cloud.intel.com/host-cpu-cores=1,cloud.intel.com/host-cpu-model-name=Intel,cloud.intel.com/host-cpu-sockets=2,cloud.intel.com/host-gpu-model=TestGPU,cloud.intel.com/host-memory-size=128m,cloud.intel.com/intel-gaudi-sw-sk-1.16.0=true,cloud.intel.com/instance-type=test-instance,cloud.intel.com/host-cpu-id=id,cloud.intel.com/host-cpu-threads=3,cloud.intel.com/host-gpu-count=2,cloud.intel.com/intel-gaudi-sw-fk-1.18.0-524=true,cloud.intel.com/intel-gaudi-sw-fk-1.16.0-526=true",
		},
		{
			name: "Empty instance Spec returns empty labels string",
			instanceType: &pb.InstanceType{
				Spec: &pb.InstanceTypeSpec{},
			},
			want: "",
		},
		{
			name:         "Empty instance type returns empty labels string",
			instanceType: nil,
			want:         "",
		},
		{
			name: "skip too long label name",
			instanceType: &pb.InstanceType{
				Spec: &pb.InstanceTypeSpec{
					Name: "test",
					Cpu: &pb.CpuSpec{
						ModelName: strings.Repeat("a", 64), // this will be skipped
					},
				},
			},
			machineImage: &pb.MachineImage{
				Spec: &pb.MachineImageSpec{},
			},
			want: "",
		},
		{
			name: "special characters should be removed from label value",
			instanceType: &pb.InstanceType{
				Spec: &pb.InstanceTypeSpec{
					Name: "test",
					Cpu: &pb.CpuSpec{
						ModelName: "@Intel",
					},
				},
			},
			machineImage: &pb.MachineImage{
				Spec: &pb.MachineImageSpec{
					Components: []*pb.MachineImageComponent{},
				},
			},
			want: "cloud.intel.com/host-cpu-model-name=Intel",
		},
		{
			name: "no firmware kit version defined, still returns another labels",
			instanceType: &pb.InstanceType{
				Spec: &pb.InstanceTypeSpec{
					Name: "test",
					Cpu: &pb.CpuSpec{
						ModelName: "Intel",
					},
				},
			},
			machineImage: &pb.MachineImage{},
			want:         "cloud.intel.com/host-cpu-model-name=Intel",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := workerNodeAdditionalLabels(context.Background(), tt.instanceType, tt.machineImage)

			// If the expected value is empty, we can just compare the strings
			if tt.want == "" {
				assert.Equal(t, got, tt.want)
				return
			}

			// The order of the labels is not guaranteed, so we need to compare the maps
			gotMap := make(map[string]string)
			for _, pair := range strings.Split(got, ",") {
				kv := strings.Split(pair, "=")
				gotMap[kv[0]] = kv[1]
			}
			wantMap := make(map[string]string)
			for _, pair := range strings.Split(tt.want, ",") {
				kv := strings.Split(pair, "=")
				wantMap[kv[0]] = kv[1]
			}

			assert.Equal(t, fmt.Sprint(gotMap), fmt.Sprint(wantMap))
		})
	}
}

func TestDetectUserSelectedNodeForDeletion_SelectedNodeExists(t *testing.T) {
	nodeReconciler := NodegroupReconciler{
		InstanceServiceClient: NewMockInstanceServiceClient(t),
	}
	// test selected node group exists in
	ng := &privatecloudv1alpha1.Nodegroup{
		Status: privatecloudv1alpha1.NodegroupStatus{
			Nodes: []privatecloudv1alpha1.NodeStatus{
				{
					Name: "1-node-exists-in-status", // index 0
				},
				{
					Name: "2-node-exists-in-status", // index 1
				},
			},
		},
	}
	assert.Equal(t, 1,
		nodeReconciler.syncDeletedInstancesToNodegroupStatus(
			context.Background(),
			"123",
			-1,
			&ng.Status),
	)
}

func TestDetectUserSelectedNodeForDeletion_SelectedNodeNotExists(t *testing.T) {
	nodeReconciler := NodegroupReconciler{
		InstanceServiceClient: NewMockInstanceServiceClient(t),
	}
	// test selected node group exists in
	ng := &privatecloudv1alpha1.Nodegroup{
		Status: privatecloudv1alpha1.NodegroupStatus{
			Nodes: []privatecloudv1alpha1.NodeStatus{
				{
					Name: "foo", // index 0
				},
			},
		},
	}
	assert.Equal(t, -1,
		nodeReconciler.syncDeletedInstancesToNodegroupStatus(
			context.Background(),
			"123",
			-1,
			&ng.Status),
	)
}

func TestDetectUserSelectedNodeForDeletion_SelectedNodeNilMetadata(t *testing.T) {
	nodeReconciler := NodegroupReconciler{
		InstanceServiceClient: NewMockInstanceServiceClient(t),
	}
	// test selected node group exists in
	ng := &privatecloudv1alpha1.Nodegroup{
		Status: privatecloudv1alpha1.NodegroupStatus{
			Nodes: []privatecloudv1alpha1.NodeStatus{
				{
					Name: "nil-metadata", // index 0
				},
			},
		},
	}
	assert.Equal(t, -1,
		nodeReconciler.syncDeletedInstancesToNodegroupStatus(
			context.Background(),
			"123",
			-1,
			&ng.Status),
	)
}

func TestDetectUserSelectedNodeForDeletion_SelectedNodeNotFound(t *testing.T) {
	nodeReconciler := NodegroupReconciler{
		InstanceServiceClient: NewMockInstanceServiceClient(t),
	}
	// test selected node group exists in
	ng := &privatecloudv1alpha1.Nodegroup{
		Status: privatecloudv1alpha1.NodegroupStatus{
			Nodes: []privatecloudv1alpha1.NodeStatus{
				{
					Name: "bar-node", // index 0
				},
				{
					Name: "foo-node", // index 1
				},
				{
					Name: "not-exists-node", // index 2
				},
			},
		},
	}
	assert.Equal(t, 2,
		nodeReconciler.syncDeletedInstancesToNodegroupStatus(
			context.Background(),
			"123",
			-1,
			&ng.Status),
	)
}

func TestDetectUserSelectedNodeForDeletion_OldestNodeIndex(t *testing.T) {
	nodeReconciler := NodegroupReconciler{
		InstanceServiceClient: NewMockInstanceServiceClient(t),
	}
	// test selected node group exists in
	ng := &privatecloudv1alpha1.Nodegroup{
		Status: privatecloudv1alpha1.NodegroupStatus{
			Nodes: []privatecloudv1alpha1.NodeStatus{
				{
					Name: "bar-node", // index 0
				},
				{
					Name: "foo-node", // index 1
				},
				{
					Name: "not-exists-node", // index 2
				},
			},
		},
	}
	assert.Equal(t, 2,
		nodeReconciler.syncDeletedInstancesToNodegroupStatus(
			context.Background(),
			"123",
			1,
			&ng.Status),
	)
}
