// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package basic

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	computeapiserverconfig "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/openapi"
	computeapiserver "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/server"
	fleetadminapiserverconfig "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/fleet_admin/api_server/config"
	fleetadminapiserver "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/fleet_admin/api_server/server"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpc_rest_gateway"
	instancecontrollertest "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/test"
	instanceoperatorutil "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/util"
	vminstancecontroller "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/vm/controllers"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_replicator/replicator"
	vminstanceschedulerserver "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/server"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	idcclientset "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/cloudintelclient/clientset/versioned"
	idcinformerfactory "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/cloudintelclient/informers/externalversions"
	loadbalancer "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/pkg/constants"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	privatecloudsshproxy "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/ssh_proxy_operator/controllers/private.cloud"
	toolsk8s "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/k8s/test"
	test_tools "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/ssh"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/stoppable"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	kubevirtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

type BasicTestEnvOptions struct {
}

type BasicTestEnv struct {
	// context is cancelled when the test environment should stop.
	ctx                            context.Context
	cancel                         context.CancelFunc
	managerStoppable               *stoppable.Stoppable
	namespace                      string
	instanceResourceId             string
	authorizedKeysFilePath         string
	computeApiServerGrpcListenPort uint16
	computeGrpcService             *computeapiserver.GrpcService
	fleetAdminGrpcService          *fleetadminapiserver.GrpcService
	fleetAdminServiceClient        pb.FleetAdminServiceClient
	qmsClient                      pb.QuotaManagementPrivateServiceClient
	restService                    *grpc_rest_gateway.RestService
	harvesterTestEnvs              []*envtest.Environment
	harvesterRestConfigList        []*restclient.Config
	vmClustersKubeConfigDir        string
	publicKeyOp                    string
	privateKeyOp                   string
	computeNodePoolId              string
}

// Create a test environment for integration testing.
//   - Start Kubernetes envtest.Environment (Kubernetes API Server, etcd)
//   - Compute Database (Postgres)
//   - Compute API Server (GRPC)
//   - Compute API Gateway (GRPC-REST gateway)
//   - Fleet Admin Database (Postgres)
//   - Fleet Admin API Server (GRPC)
//   - VM Instance Scheduler
//   - Instance Replicator (Compute API Server to K8s Instance)
//   - VM Instance Operator (K8s Instance to Kubevirt VirtualMachine)
func NewBasicTestEnv(opts BasicTestEnvOptions) *BasicTestEnv {
	// Create a context that will last throughout the lifetime of the test environment.
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	testEnv := &BasicTestEnv{
		ctx:    ctx,
		cancel: cancel,
	}
	return testEnv
}

func (e *BasicTestEnv) Start() {
	ctx := e.ctx
	log := log.FromContext(ctx).WithName("BasicTestEnv.Start")
	log.Info("BEGIN")
	defer log.Info("END")

	var err error
	e.privateKeyOp, e.publicKeyOp, err = test_tools.CreateSshRsaKeyPair(4096, "idcuser@example.com")
	Expect(err).NotTo(HaveOccurred())

	clearDatabase(ctx)

	By("Creating Kubernetes test environments to simulate Harvester clusters")
	e.harvesterTestEnvs, e.harvesterRestConfigList, e.vmClustersKubeConfigDir = toolsk8s.CreateTestEnvs(1, crdDirectoryPaths)

	By("Creating Kubernetes client with testenv rest.Config")
	clientset, err := kubernetes.NewForConfig(e.harvesterRestConfigList[0])
	Expect(err).NotTo(HaveOccurred())

	By("Creating dynamic clientsets")
	dynamicClientset, err := dynamic.NewForConfig(e.harvesterRestConfigList[0])
	Expect(err).NotTo(HaveOccurred())

	unstructuredHarvesterSettingObj, gvr := toolsk8s.CreateUnstructuredHarvesterSettingObject()
	_, err = dynamicClientset.Resource(gvr).Create(ctx, unstructuredHarvesterSettingObj, metav1.CreateOptions{})
	Expect(err).NotTo(HaveOccurred())

	By("Creating Node")
	nodeName := "node0"
	topologyPartition := "us-dev-1a-p0"
	e.computeNodePoolId = "pool0"
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
			Labels: map[string]string{
				"kubernetes.io/hostname":                                              nodeName,
				instanceoperatorutil.TopologySpreadTopologyKey:                        topologyPartition,
				instanceoperatorutil.LabelKeyForInstanceType("vm-spr-sml"):            "true",
				instanceoperatorutil.LabelKeyForComputeNodePools(e.computeNodePoolId): "true",
			},
		},
		Status: corev1.NodeStatus{
			Allocatable: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourcePods:   resource.MustParse("110"),
				corev1.ResourceCPU:    resource.MustParse("100"),
				corev1.ResourceMemory: resource.MustParse("100Gi"),
			},
		},
	}
	_, err = clientset.CoreV1().Nodes().Create(ctx, node, metav1.CreateOptions{})
	Expect(err).NotTo(HaveOccurred())

	By("Removing node.kubernetes.io/not-ready taint added by node admission controller")
	node.Spec.Taints = node.Spec.Taints[:0]
	_, err = clientset.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	Expect(err).NotTo(HaveOccurred())

	By("Set manager options to enable instance operator to filter by labels")
	managerOptions := ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
	}
	vmInstanceOperatorConfig := instancecontrollertest.NewTestVmInstanceOperatorConfig("../../instance_operator/testdata", scheme)
	err = instanceoperatorutil.SetManagerOptions(ctx, &managerOptions, &vmInstanceOperatorConfig.InstanceOperator)
	Expect(err).Should(Succeed())

	By("Creating manager")
	k8sManager, err := ctrl.NewManager(k8sRestConfig, managerOptions)
	Expect(err).ToNot(HaveOccurred())

	By("Starting VM Instance Scheduling Server")
	instanceSchedulingServerListener, err := net.Listen("tcp", "localhost:")
	Expect(err).Should(Succeed())
	instanceSchedulingServerListenPort = uint16(instanceSchedulingServerListener.Addr().(*net.TCPAddr).Port)
	vmInstanceSchedulerConfig := &privatecloudv1alpha1.VmInstanceSchedulerConfig{
		ListenPort:              instanceSchedulingServerListenPort,
		VmClustersKubeConfigDir: e.vmClustersKubeConfigDir,
		OvercommitConfig: privatecloudv1alpha1.OvercommitConfig{
			CPU:     100,
			Memory:  100,
			Storage: 100,
		},
	}
	computeApiServerGrpcServerListener, err := net.Listen("tcp", "localhost:")
	Expect(err).Should(Succeed())
	e.computeApiServerGrpcListenPort = uint16(computeApiServerGrpcServerListener.Addr().(*net.TCPAddr).Port)
	computeApiServerGrpcListenPort = e.computeApiServerGrpcListenPort

	computeApiServerAddress := fmt.Sprintf("localhost:%d", computeApiServerGrpcListenPort)
	clientConn, err := grpc.Dial(computeApiServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).Should(Succeed())
	instanceTypeClient := pb.NewInstanceTypeServiceClient(clientConn)
	_, err = vminstanceschedulerserver.NewSchedulingServer(ctx, vmInstanceSchedulerConfig, k8sManager, instanceSchedulingServerListener, instanceTypeClient)
	Expect(err).Should(Succeed())

	By("Creating VM Instance Scheduling Service client")
	vmInstanceSchedulingServerAddress := fmt.Sprintf("localhost:%d", instanceSchedulingServerListenPort)
	vmInstanceSchedulingServerClientConn, err := grpc.Dial(vmInstanceSchedulingServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	vmInstanceSchedulingService := pb.NewInstanceSchedulingServiceClient(vmInstanceSchedulingServerClientConn)
	Expect(err).Should(Succeed())

	By("Starting Fleet Admin Server")
	fleetAdminServerGrpcServerListener, err := net.Listen("tcp", "localhost:")
	Expect(err).Should(Succeed())
	fleetAdminServerGrpcListenPort := uint16(fleetAdminServerGrpcServerListener.Addr().(*net.TCPAddr).Port)
	fleetAdminConfig := &fleetadminapiserverconfig.Config{
		ComputeNodePoolForUnknownCloudAccount: e.computeNodePoolId,
	}
	e.fleetAdminGrpcService, err = fleetadminapiserver.New(ctx, fleetAdminConfig, fleetAdminManagedDb, fleetAdminServerGrpcServerListener)
	Expect(err).Should(Succeed())
	Expect(e.fleetAdminGrpcService.Start(ctx)).Should(Succeed())

	By("Creating Fleet Admin Service client")
	fleetAdminServerAddress := fmt.Sprintf("localhost:%d", fleetAdminServerGrpcListenPort)
	fleetAdminServerClientConn, err := grpc.Dial(fleetAdminServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	e.fleetAdminServiceClient = pb.NewFleetAdminServiceClient(fleetAdminServerClientConn)
	Expect(err).Should(Succeed())

	cloudAccountService := NewMockCloudAccountServiceClient()

	cloudAccountAppClientService := NewMockCloudAccountAppClientServiceClient()

	objectStorageServicePrivate := NewMockObjectStorageServicePrivateClient()

	By("Starting Compute API Server (GRPC)")
	computeApiServerGrpcServerListener, err = net.Listen("tcp", "localhost:")
	Expect(err).Should(Succeed())
	e.computeApiServerGrpcListenPort = uint16(computeApiServerGrpcServerListener.Addr().(*net.TCPAddr).Port)
	computeApiServerGrpcListenPort = e.computeApiServerGrpcListenPort
	cloudaccounts := make(map[string]computeapiserverconfig.LaunchQuota)
	instanceQuota := map[string]int{
		"vm-spr-sml":        5,
		"vm-spr-med":        0,
		"vm-spr-lrg":        0,
		"bm-spr":            1,
		"bm-spr-pvc-1100-4": 1,
		"bm-icp-gaudi2":     0,
	}

	cloudaccounts["STANDARD"] = computeapiserverconfig.LaunchQuota{
		InstanceQuota: instanceQuota,
	}

	computeConfig := &computeapiserverconfig.Config{
		ListenPort:                     computeApiServerGrpcListenPort,
		PurgeInstanceInterval:          time.Duration(5 * time.Minute),
		PurgeInstanceAge:               time.Duration(5 * time.Minute),
		GetDeactivateInstancesInterval: time.Duration(5 * time.Minute),
		DbMaxIdleConnectionCount:       2,
		CloudAccountQuota: computeapiserverconfig.CloudAccountQuota{
			CloudAccounts: cloudaccounts,
		},
		FeatureFlags: computeapiserverconfig.FeatureFlags{
			EnableComputeNodePoolsForScheduling: true,
		},
	}
	e.computeGrpcService, err = computeapiserver.New(
		ctx,
		computeConfig,
		computeManagedDb,
		vmInstanceSchedulingService,
		billingDeactivateInstancesService,
		cloudAccountService,
		cloudAccountAppClientService,
		objectStorageServicePrivate,
		e.fleetAdminServiceClient,
		e.qmsClient,
		computeApiServerGrpcServerListener,
	)
	Expect(err).Should(Succeed())
	Expect(e.computeGrpcService.Start(ctx)).Should(Succeed())

	By("Starting Compute API Gateway (GRPC-REST)")
	computeApiServerRestListener, err := net.Listen("tcp", "localhost:")
	Expect(err).Should(Succeed())
	computeApiServerRestListenPort = uint16(computeApiServerRestListener.Addr().(*net.TCPAddr).Port)
	e.restService, err = grpc_rest_gateway.New(ctx, &grpc_rest_gateway.Config{
		TargetAddr: fmt.Sprintf("localhost:%d", computeApiServerGrpcListenPort),
	}, computeApiServerRestListener)
	Expect(err).Should(Succeed())
	e.restService.AddService(pb.RegisterVNetServiceHandler)
	e.restService.AddService(pb.RegisterSshPublicKeyServiceHandler)
	e.restService.AddService(pb.RegisterInstanceServiceHandler)
	Expect(e.restService.Start(ctx)).Should(Succeed())

	By("Creating OpenAPI client for Compute API Gateway")
	clientConfig := openapi.NewConfiguration()
	clientConfig.Scheme = "http"
	clientConfig.Host = fmt.Sprintf("localhost:%d", computeApiServerRestListenPort)
	openApiClient = openapi.NewAPIClient(clientConfig)

	By("Pinging Compute API Gateway until it comes up")
	Eventually(func() error {
		_, _, err = openApiClient.SshPublicKeyServiceApi.SshPublicKeyServicePing(ctx).Execute()
		return err
	}, time.Millisecond*5000, time.Millisecond*500).Should(Succeed())

	By("Creating Instance Service client")
	computeApiServerAddress = fmt.Sprintf("localhost:%d", computeApiServerGrpcListenPort)
	computeApiServerClientConn, err := grpc.Dial(computeApiServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).Should(Succeed())
	instanceClient := pb.NewInstancePrivateServiceClient(computeApiServerClientConn)

	By("Creating IP Resource Manager Client")
	ipResourceManagerClient := pb.NewIpResourceManagerServiceClient(computeApiServerClientConn)

	By("Creating subnet")
	_, err = ipResourceManagerClient.PutSubnet(ctx, &pb.CreateSubnetRequest{
		Region:           "us-dev-1",
		AvailabilityZone: "us-dev-1a",
		Subnet:           "172.16.0.0",
		PrefixLength:     24,
		Gateway:          "172.16.0.1",
		VlanId:           1,
		Address:          []string{"172.16.0.3", "172.16.0.4", "172.16.0.5", "172.16.0.6"},
	})
	Expect(err).Should(Succeed())

	By("Creating VNet Private Client")
	vNetPrivateClient := pb.NewVNetPrivateServiceClient(computeApiServerClientConn)

	By("Creating VNet Client")
	vNetClient := pb.NewVNetServiceClient(computeApiServerClientConn)

	By("Creating instance replicator")
	_, err = replicator.NewReplicator(ctx, k8sManager, instanceClient)
	Expect(err).Should(Succeed())

	By("Creating instance operator")
	_, err = vminstancecontroller.NewVmInstanceReconciler(ctx, k8sManager, vNetPrivateClient, vNetClient, vmInstanceOperatorConfig)
	Expect(err).Should(Succeed())

	By("Creating SshProxy operator")
	sshKeyAuthorizedFileTmpDir, err = os.MkdirTemp("", "")
	Expect(err).ToNot(HaveOccurred())
	e.authorizedKeysFilePath = fmt.Sprintf("%s/.ssh/authorized_keys", sshKeyAuthorizedFileTmpDir)
	log.Info("authorizedKeysFilePath", "authorizedKeysFilePath", e.authorizedKeysFilePath)
	scpTarget := "scp://guest@127.0.0.1:22/home/guest/.ssh/authorized_keys"
	// Passing a fake scpTarget List so as to avoid running SCP/SSH code for transferring the file to proxy server
	authorizedKeysScpTargets := []string{scpTarget}

	sshProxyConfig := privatecloudsshproxy.SshProxyTunnelConfig{
		AuthorizedKeysFilePath:   e.authorizedKeysFilePath,
		ProxyUser:                "guest",
		ProxyAddress:             "ssh.us-dev-1.cloud.intel.com",
		ProxyPort:                22,
		AuthorizedKeysScpTargets: authorizedKeysScpTargets,
		PublicKey:                e.publicKeyOp,
		PrivateKey:               e.privateKeyOp,
	}
	kubeClientSet, err := kubernetes.NewForConfig(k8sRestConfig)
	Expect(err).NotTo(HaveOccurred())
	Expect(kubeClientSet).NotTo(BeNil())

	idcClientSet, err := idcclientset.NewForConfig(k8sRestConfig)
	Expect(err).NotTo(HaveOccurred())
	Expect(kubeClientSet).NotTo(BeNil())

	informerFactory := idcinformerfactory.NewSharedInformerFactory(idcClientSet, 10*time.Minute)

	sshProxyController, err := privatecloudsshproxy.
		NewSshProxyController(ctx, kubeClientSet, idcClientSet, informerFactory.Private().V1alpha1().SshProxyTunnels(), sshProxyConfig)
	Expect(err).NotTo(HaveOccurred())

	sshProxyController.MockScpTargetsMutex.Lock()
	sshProxyController.MockScpTargets[scpTarget] = nil
	sshProxyController.MockScpTargetsMutex.Unlock()

	informerFactory.Start(ctx.Done())

	err = k8sManager.Add(manager.RunnableFunc(func(context.Context) error {
		return sshProxyController.Run(ctx, ctx.Done())
	}))
	Expect(err).NotTo(HaveOccurred())

	By("Starting manager")
	e.managerStoppable = stoppable.New(k8sManager.Start)
	e.managerStoppable.Start(ctx)
}

func (e *BasicTestEnv) Stop() {
	ctx := e.ctx
	log := log.FromContext(ctx).WithName("BasicTestEnv.Stop")
	log.Info("BEGIN")
	defer log.Info("END")

	if e.instanceResourceId != "" {
		By("Cleanup instances")
		instanceLookupKey := types.NamespacedName{Name: e.instanceResourceId, Namespace: e.namespace}
		instanceRef := &privatecloudv1alpha1.Instance{}
		_, _, _ = openApiClient.InstanceServiceApi.InstanceServiceDelete(ctx, e.namespace, e.instanceResourceId).Execute()
		Eventually(func(g Gomega) {
			g.Expect(apierrors.IsNotFound(k8sClient.Get(ctx, instanceLookupKey, instanceRef))).Should(BeTrue())
		}, "10s").Should(Succeed())

		By("Instance deletion should cleanup sshProxy tunnel")
		tunnelLookupKey := types.NamespacedName{Name: e.instanceResourceId, Namespace: e.namespace}
		tunnelRef := &privatecloudv1alpha1.SshProxyTunnel{}
		Eventually(func(g Gomega) {
			g.Expect(apierrors.IsNotFound(k8sClient.Get(ctx, tunnelLookupKey, tunnelRef))).Should(BeTrue())
		}, "10s").Should(Succeed())
	}

	By("Cancelling BasicTestEnv context")
	e.cancel()

	if e.managerStoppable != nil {
		By("Stopping manager")
		Expect(e.managerStoppable.Stop(ctx)).Should(Succeed())
		By("Manager stopped")
	}
	if e.restService != nil {
		By("Stopping Compute API Gateway (GRPC-REST)")
		Expect(e.restService.Stop(ctx)).Should(Succeed())
	}
	if e.computeGrpcService != nil {
		By("Stopping Compute API Server (GRPC)")
		Expect(e.computeGrpcService.Stop(ctx)).Should(Succeed())
	}
	if e.fleetAdminGrpcService != nil {
		By("Stopping Fleet Admin API Server (GRPC)")
		Expect(e.fleetAdminGrpcService.Stop(ctx)).Should(Succeed())
	}

	toolsk8s.StopTestEnvs(e.harvesterTestEnvs)

	if sshKeyAuthorizedFileTmpDir != "" {
		By("Deleting authorized file directories")
		Expect(os.RemoveAll(sshKeyAuthorizedFileTmpDir)).Should(Succeed())
	}
}

func getInstanceGrpcClient(e *BasicTestEnv) pb.InstancePrivateServiceClient {
	clientConn, err := grpc.Dial(fmt.Sprintf("localhost:%d", e.computeApiServerGrpcListenPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).NotTo(HaveOccurred())
	return pb.NewInstancePrivateServiceClient(clientConn)
}

func NewPutVNetRequest(name string) *openapi.VNetServicePutRequest {
	region := "us-dev-1"
	availabilityZone := "us-dev-1a"
	prefixLength := int32(24)
	return &openapi.VNetServicePutRequest{
		Metadata: &openapi.VNetServicePutRequestMetadata{
			Name: &name,
		},
		Spec: &openapi.ProtoVNetSpec{
			Region:           &region,
			AvailabilityZone: &availabilityZone,
			PrefixLength:     &prefixLength,
		},
	}
}

func NewCreateSshPublicKeyRequest(name string) *openapi.SshPublicKeyServiceCreateRequest {
	_, sshPublicKey, err := test_tools.CreateSshRsaKeyPair(4096, name)
	Expect(err).Should(Succeed())
	return &openapi.SshPublicKeyServiceCreateRequest{
		Metadata: &openapi.SshPublicKeyServiceCreateRequestMetadata{
			Name: &name,
		},
		Spec: &openapi.ProtoSshPublicKeySpec{
			SshPublicKey: &sshPublicKey,
		},
	}
}

func NewCreateInstanceRequest(sshPublicKeyNames []string, instanceType string, machineImage string, vNetName string) *openapi.InstanceServiceCreateRequest {
	availabilityZone := "us-dev-1a"
	runStrategyStr := "RerunOnFailure"
	runStrategy, err := openapi.NewProtoRunStrategyFromValue(runStrategyStr)
	Expect(err).Should(Succeed())
	interfaces := []openapi.ProtoNetworkInterface{{VNet: &vNetName}}
	return &openapi.InstanceServiceCreateRequest{
		Metadata: &openapi.InstanceServiceCreateRequestMetadata{
			Labels: &map[string]string{
				"iks-cluster-name": "my-iks-cluster-1",
				"iks-role":         "master",
			},
		},
		Spec: &openapi.ProtoInstanceSpec{
			AvailabilityZone:  &availabilityZone,
			InstanceType:      &instanceType,
			MachineImage:      &machineImage,
			RunStrategy:       runStrategy,
			SshPublicKeyNames: sshPublicKeyNames,
			Interfaces:        interfaces,
		},
	}
}

func CreateInstanceType(ctx context.Context) string {
	By("Creating instance type client")
	computeApiServerAddress := fmt.Sprintf("localhost:%d", computeApiServerGrpcListenPort)
	clientConn, err := grpc.Dial(computeApiServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).Should(Succeed())
	instanceTypeClient := pb.NewInstanceTypeServiceClient(clientConn)

	By("Creating an InstanceType")
	instanceTypeName := "vm-spr-sml"
	instanceType := &pb.InstanceType{
		Metadata: &pb.InstanceType_Metadata{
			Name: instanceTypeName,
		},
		Spec: &pb.InstanceTypeSpec{
			Name:             instanceTypeName,
			InstanceCategory: pb.InstanceCategory_VirtualMachine,
			Cpu: &pb.CpuSpec{
				Cores:     8,
				Sockets:   1,
				Threads:   1,
				Id:        "0x806F2",
				ModelName: "Intel® Xeon 4th Gen® Scalable processor formerly known as Sapphire Rapids",
			},
			Description: "Intel® Xeon 4th Gen® Scalable processor formerly known as Sapphire Rapids",
			Disks: []*pb.DiskSpec{
				{Size: "20Gi"},
			},
			DisplayName: "Small VM",
			Memory: &pb.MemorySpec{
				DimmCount: 1,
				Speed:     3200,
				DimmSize:  "16Gi",
				Size:      "16Gi",
			},
		},
	}
	_, err = instanceTypeClient.Put(ctx, instanceType)
	Expect(err).Should(Succeed())
	return instanceTypeName
}

func CreateMachineImage(ctx context.Context) string {
	By("Creating machine image client")
	computeApiServerAddress := fmt.Sprintf("localhost:%d", computeApiServerGrpcListenPort)
	clientConn, err := grpc.Dial(computeApiServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).Should(Succeed())
	machineImageClient := pb.NewMachineImageServiceClient(clientConn)

	By("Creating MachineImage")
	// virtualMachine image name must be less than 32 characters
	machineImageName := uuid.NewString()[:30]
	machineImage := &pb.MachineImage{
		Metadata: &pb.MachineImage_Metadata{
			Name: machineImageName,
		},
		Spec: &pb.MachineImageSpec{
			DisplayName: "Ubuntu 22.04 LTS (Jammy Jellyfish) v20230128",
			Description: "Ubuntu Server is a version of the Ubuntu operating system designed and engineered as a backbone for the internet.",
			Icon:        "https://www.intel.com/content/dam/www/central-libraries/us/en/images/3rd-gen-xeon-scalable-processor-badge-left-rwd.png.rendition.intel.web.368.207.png",
			InstanceCategories: []pb.InstanceCategory{
				0,
			},
			InstanceTypes: []string{
				"vm-spr-sml",
				"vm-spr-lrg",
			},
			Md5Sum:    "764efa883dda1e11db47671c4a3bbd9e",
			Sha256Sum: "98ea6e4f216f2fb4b69fff9b3a44842c38686ca685f3f55dc48c5d3fb1107be4",
			Labels: map[string]string{
				"architecture": "X86_64",
				"family":       "ubuntu-2204-lts",
			},
			ImageCategories: []string{
				"AI",
			},
			Components: []*pb.MachineImageComponent{
				{
					Name:        "Ubuntu 22.04 LTS",
					Type:        "OS",
					Version:     "22.04",
					Description: "Ubuntu Server is a version of the Ubuntu operating system designed and engineered as a backbone for the internet.",
					InfoUrl:     "https://discourse.ubuntu.com/t/jammy-jellyfish-release-notes/24668?_ga=2.61253994.716223186.1673128475-452578204.1673128475",
					ImageUrl:    "https://www.intel.com/content/dam/www/central-libraries/us/en/images/3rd-gen-xeon-scalable-processor-badge-left-rwd.png.rendition.intel.web.368.207.png",
				},
			},
		},
	}
	_, err = machineImageClient.Put(ctx, machineImage)
	Expect(err).Should(Succeed())
	return machineImageName
}

func CreateVNet(ctx context.Context, cloudAccountId string) string {
	By("Creating VNet")
	vNetName := "us-dev-1a-default"
	createVNetReq := NewPutVNetRequest(vNetName)
	_, _, err := openApiClient.VNetServiceApi.VNetServicePut(ctx, cloudAccountId).Body(*createVNetReq).Execute()
	Expect(err).Should(Succeed())
	return vNetName
}

func CreateSshPublicKey(ctx context.Context, cloudAccountId string) string {
	By("Creating SSH Public Key")
	sshPublicKeyName := "key1-" + uuid.NewString()
	createKeyReq := NewCreateSshPublicKeyRequest(sshPublicKeyName)
	_, _, err := openApiClient.SshPublicKeyServiceApi.SshPublicKeyServiceCreate(ctx, cloudAccountId).Body(*createKeyReq).Execute()
	Expect(err).Should(Succeed())
	return sshPublicKeyName
}

var _ = Describe("Instance happy path using OpenAPI client", Serial, func() {
	ctx := context.Background()
	log := log.FromContext(ctx)
	_ = log

	It("When an instance is created in the Compute API Server, the Kubevirt VirtualMachine should get created along with the Ssh proxy tunnel with user sshKey and VM ip address", func() {
		e := NewBasicTestEnv(BasicTestEnvOptions{})
		defer e.Stop()
		e.Start()

		cloudAccountId := cloudaccount.MustNewId()
		e.namespace = cloudAccountId
		instanceType := CreateInstanceType(ctx)
		machineImage := CreateMachineImage(ctx)
		vNetName := CreateVNet(ctx, cloudAccountId)

		By("SshPublicKeyServiceCreate")
		sshPublicKeyName := "key1-" + uuid.NewString()
		createKeyReq := NewCreateSshPublicKeyRequest(sshPublicKeyName)
		createKeyResp, _, err := openApiClient.SshPublicKeyServiceApi.SshPublicKeyServiceCreate(ctx, cloudAccountId).Body(*createKeyReq).Execute()
		Expect(err).Should(Succeed())
		resourceId := *createKeyResp.Metadata.ResourceId
		getKeyResp, _, err := openApiClient.SshPublicKeyServiceApi.SshPublicKeyServiceGet(ctx, cloudAccountId, resourceId).Execute()
		Expect(err).Should(Succeed())
		pubKey1 := *getKeyResp.Spec.SshPublicKey
		cleanPublicKey, err := privatecloudsshproxy.CleanPublicKey(ctx, pubKey1)
		Expect(err).Should(Succeed())

		By("InstanceServiceCreate")
		createInstanceReq := NewCreateInstanceRequest([]string{sshPublicKeyName}, instanceType, machineImage, vNetName)
		createResp, _, err := openApiClient.InstanceServiceApi.InstanceServiceCreate(ctx, cloudAccountId).Body(*createInstanceReq).Execute()
		Expect(err).Should(Succeed())
		e.instanceResourceId = *createResp.Metadata.ResourceId

		By("Waiting for instance to be created in K8s by the Instance Replicator")
		instanceLookupKey := types.NamespacedName{Namespace: e.namespace, Name: e.instanceResourceId}
		instanceRef := &privatecloudv1alpha1.Instance{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
		}, "10s").Should(Succeed())

		By("Waiting for Kubevirt VirtualMachine to be created in K8s by the Instance Operator")
		harvesterVMlookupKey := types.NamespacedName{Namespace: e.namespace, Name: e.instanceResourceId}
		harvesterVmRef := &kubevirtv1.VirtualMachine{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, harvesterVMlookupKey, harvesterVmRef)).Should(Succeed())
		}, "10s").Should(Succeed())

		By("Kubevirt VirtualMachine should have an affinity rule for instance type")
		affinityTerms := harvesterVmRef.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
		Expect(len(affinityTerms)).Should(BeNumerically(">", 0))
		for _, term := range affinityTerms {
			instanceTypeMatchExpression := corev1.NodeSelectorRequirement{
				Key:      instanceoperatorutil.LabelKeyForInstanceType(instanceType),
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{"true"},
			}
			Expect(term.MatchExpressions).Should(ContainElements(instanceTypeMatchExpression))
		}

		By("Waiting for SshProxyTunnel to be created in K8s by the Instance Operator")
		tunnelLookupKey := types.NamespacedName{Namespace: e.namespace, Name: e.instanceResourceId}
		tunnelRef := &privatecloudv1alpha1.SshProxyTunnel{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, tunnelLookupKey, tunnelRef)).Should(Succeed())
		}, "10s").Should(Succeed())

		By("Waiting for SshProxyTunnel spec Interfaces to be in sync with Instance status Interfaces")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, tunnelLookupKey, tunnelRef)).Should(Succeed())
			g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
			g.Expect(len(instanceRef.Status.Interfaces)).Should(Equal(1))
			g.Expect(len(instanceRef.Status.Interfaces[0].Addresses)).Should(Equal(1))
			g.Expect(tunnelRef.Spec.TargetAddresses).
				Should(Equal([]string{instanceRef.Status.Interfaces[0].Addresses[0]}))
		}, "10s").Should(Succeed())

		expectedKeyData := e.publicKeyOp +
			"permitopen=\"" + instanceRef.Status.Interfaces[0].Addresses[0] + ":22" + "\"," +
			privatecloudsshproxy.SshAuthorizedKeysOptions + " " + cleanPublicKey + "\n"
		Eventually(func() (string, error) {
			str, err := ReadAuthorizedKeysFile(e.authorizedKeysFilePath)
			return str, err
		}, "10s").Should(Equal(expectedKeyData))
	})

	It("When an instance with topology spread constraints is created in the Compute API Server, the Kubevirt VirtualMachine should have corresponding affinity rules", func() {
		e := NewBasicTestEnv(BasicTestEnvOptions{})
		defer e.Stop()
		e.Start()

		cloudAccountId := cloudaccount.MustNewId()
		e.namespace = cloudAccountId
		instanceType := CreateInstanceType(ctx)
		machineImage := CreateMachineImage(ctx)
		vNetName := CreateVNet(ctx, cloudAccountId)
		sshPublicKeyName := CreateSshPublicKey(ctx, cloudAccountId)

		By("InstanceServiceCreate with TopologySpreadConstraints")
		createInstanceReq := NewCreateInstanceRequest([]string{sshPublicKeyName}, instanceType, machineImage, vNetName)
		Expect(len(*createInstanceReq.Metadata.Labels)).Should(BeNumerically(">", 0))
		createInstanceReq.Spec.TopologySpreadConstraints = []openapi.ProtoTopologySpreadConstraints{
			{
				LabelSelector: &openapi.ProtoLabelSelector{
					MatchLabels: createInstanceReq.Metadata.Labels,
				},
			},
		}
		createResp, _, err := openApiClient.InstanceServiceApi.InstanceServiceCreate(ctx, cloudAccountId).Body(*createInstanceReq).Execute()
		Expect(err).Should(Succeed())
		e.instanceResourceId = *createResp.Metadata.ResourceId

		By("Waiting for Kubevirt VirtualMachine to be created in K8s by the Instance Operator")
		harvesterVMlookupKey := types.NamespacedName{Namespace: e.namespace, Name: e.instanceResourceId}
		harvesterVmRef := &kubevirtv1.VirtualMachine{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, harvesterVMlookupKey, harvesterVmRef)).Should(Succeed())
		}, "10s").Should(Succeed())

		By("Kubevirt VirtualMachine labels should match user-provided Instance labels")
		for k, v := range *createInstanceReq.Metadata.Labels {
			Expect(harvesterVmRef.Labels[k]).Should(Equal(v))
		}

		By("Kubevirt VirtualMachine should have an affinity rule for instance type")
		affinityTerms := harvesterVmRef.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
		Expect(len(affinityTerms)).Should(BeNumerically(">", 0))
		for _, term := range affinityTerms {
			instanceTypeMatchExpression := corev1.NodeSelectorRequirement{
				Key:      instanceoperatorutil.LabelKeyForInstanceType(instanceType),
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{"true"},
			}
			Expect(term.MatchExpressions).Should(ContainElements(instanceTypeMatchExpression))
		}

		By("Kubevirt VirtualMachine should have an affinity rule for partition")
		partitionMatchExpression := corev1.NodeSelectorRequirement{
			Key:      instanceoperatorutil.TopologySpreadTopologyKey,
			Operator: corev1.NodeSelectorOpIn,
			Values:   []string{"us-dev-1a-p0"},
		}
		Expect(harvesterVmRef.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions).Should(ContainElement(partitionMatchExpression))
	})

	It("When an instance is requested to be deleted from the Compute API Server, the instance, Kubevirt VirtualMachine, and SshProxyTunnel should get deleted", func() {
		e := NewBasicTestEnv(BasicTestEnvOptions{})
		defer e.Stop()
		e.Start()

		cloudAccountId := cloudaccount.MustNewId()
		e.namespace = cloudAccountId
		instanceType := CreateInstanceType(ctx)
		machineImage := CreateMachineImage(ctx)
		vNetName := CreateVNet(ctx, cloudAccountId)
		sshPublicKeyName := CreateSshPublicKey(ctx, cloudAccountId)

		By("InstanceServiceCreate")
		createInstanceReq := NewCreateInstanceRequest([]string{sshPublicKeyName}, instanceType, machineImage, vNetName)
		createResp, _, err := openApiClient.InstanceServiceApi.InstanceServiceCreate(ctx, cloudAccountId).Body(*createInstanceReq).Execute()
		Expect(err).Should(Succeed())
		Expect(*createResp.Metadata.ResourceId).ShouldNot(BeEmpty())
		e.instanceResourceId = *createResp.Metadata.ResourceId

		By("Waiting for instance to be created in K8s")
		instanceLookupKey := types.NamespacedName{Namespace: e.namespace, Name: e.instanceResourceId}
		instanceRef := &privatecloudv1alpha1.Instance{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
		}, "10s").Should(Succeed())

		By("Waiting for Kubevirt VirtualMachine to be created in K8s by the Instance Operator")
		harvesterVMlookupKey := types.NamespacedName{Namespace: e.namespace, Name: e.instanceResourceId}
		harvesterVmRef := &kubevirtv1.VirtualMachine{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, harvesterVMlookupKey, harvesterVmRef)).Should(Succeed())
		}, "10s").Should(Succeed())

		By("Waiting for SshProxyTunnel to be created in K8s by the Instance Operator")
		tunnelLookupKey := types.NamespacedName{Namespace: e.namespace, Name: e.instanceResourceId}
		tunnelRef := &privatecloudv1alpha1.SshProxyTunnel{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, tunnelLookupKey, tunnelRef)).Should(Succeed())
			g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
			g.Expect(len(instanceRef.Status.Interfaces)).Should(Equal(1))
			g.Expect(len(instanceRef.Status.Interfaces[0].Addresses)).Should(Equal(1))
			g.Expect(tunnelRef.Spec.TargetAddresses).
				Should(Equal([]string{instanceRef.Status.Interfaces[0].Addresses[0]}))
		}, "10s").Should(Succeed())

		By("InstanceServiceDelete")
		_, _, err = openApiClient.InstanceServiceApi.InstanceServiceDelete(ctx, cloudAccountId, e.instanceResourceId).Execute()
		Expect(err).Should(Succeed())

		By("Waiting for instance to be deleted from K8s")
		Eventually(func(g Gomega) {
			g.Expect(apierrors.IsNotFound(k8sClient.Get(ctx, instanceLookupKey, instanceRef))).Should(BeTrue())
		}, "10s").Should(Succeed())

		By("Waiting for Kubevirt VirtualMachine to be deleted from K8s")
		// TODO: This should happen before instance is deleted from K8s. Should not need to wait.
		Eventually(func(g Gomega) {
			g.Expect(apierrors.IsNotFound(k8sClient.Get(ctx, harvesterVMlookupKey, harvesterVmRef))).Should(BeTrue())
		}, "10s").Should(Succeed())

		By("Waiting for SshProxyTunnel to be deleted from K8s")
		Eventually(func(g Gomega) {
			g.Expect(apierrors.IsNotFound(k8sClient.Get(ctx, tunnelLookupKey, tunnelRef))).Should(BeTrue())
		}, "10s").Should(Succeed())

		By("Waiting for InstanceServiceGet to return NotFound")
		Eventually(func(g Gomega) {
			_, httpResponse, err := openApiClient.InstanceServiceApi.InstanceServiceGet(ctx, cloudAccountId, e.instanceResourceId).Execute()
			g.Expect(err).ShouldNot(Succeed())
			g.Expect(httpResponse.StatusCode).Should(Equal(http.StatusNotFound))
		}, "10s").Should(Succeed())
	})

	It("When an instance is requested to be deleted from the Compute API Server, deletion should be blocked while the LoadBalancer finalizer exists", func() {
		e := NewBasicTestEnv(BasicTestEnvOptions{})
		defer e.Stop()
		e.Start()

		cloudAccountId := cloudaccount.MustNewId()
		e.namespace = cloudAccountId
		instanceType := CreateInstanceType(ctx)
		machineImage := CreateMachineImage(ctx)
		vNetName := CreateVNet(ctx, cloudAccountId)
		sshPublicKeyName := CreateSshPublicKey(ctx, cloudAccountId)

		By("InstanceServiceCreate")
		createInstanceReq := NewCreateInstanceRequest([]string{sshPublicKeyName}, instanceType, machineImage, vNetName)
		createResp, _, err := openApiClient.InstanceServiceApi.InstanceServiceCreate(ctx, cloudAccountId).Body(*createInstanceReq).Execute()
		Expect(err).Should(Succeed())
		Expect(*createResp.Metadata.ResourceId).ShouldNot(BeEmpty())
		e.instanceResourceId = *createResp.Metadata.ResourceId

		By("Waiting for instance to be created in K8s")
		instanceLookupKey := types.NamespacedName{Namespace: e.namespace, Name: e.instanceResourceId}
		instanceRef := &privatecloudv1alpha1.Instance{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
		}, "10s").Should(Succeed())

		By("Waiting for Kubevirt VirtualMachine to be created in K8s by the Instance Operator")
		harvesterVMlookupKey := types.NamespacedName{Namespace: e.namespace, Name: e.instanceResourceId}
		harvesterVmRef := &kubevirtv1.VirtualMachine{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, harvesterVMlookupKey, harvesterVmRef)).Should(Succeed())
		}, "10s").Should(Succeed())

		By("Waiting for SshProxyTunnel to be created in K8s by the Instance Operator")
		tunnelLookupKey := types.NamespacedName{Namespace: e.namespace, Name: e.instanceResourceId}
		tunnelRef := &privatecloudv1alpha1.SshProxyTunnel{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, tunnelLookupKey, tunnelRef)).Should(Succeed())
			g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
			g.Expect(len(instanceRef.Status.Interfaces)).Should(Equal(1))
			g.Expect(len(instanceRef.Status.Interfaces[0].Addresses)).Should(Equal(1))
			g.Expect(tunnelRef.Spec.TargetAddresses).
				Should(Equal([]string{instanceRef.Status.Interfaces[0].Addresses[0]}))
		}, "10s").Should(Succeed())

		By("Add LoadBalancer finalizer to the Instance")
		Eventually(func(g Gomega) {
			Expect(retry.RetryOnConflict(retry.DefaultRetry, func() error {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				g.Expect(controllerutil.AddFinalizer(instanceRef, loadbalancer.LoadbalancerFinalizer)).Should((Equal(true)))
				if err := k8sClient.Update(ctx, instanceRef); err != nil {
					return fmt.Errorf("PersistStatusUpdate: %w", err)
				}
				return nil
			})).Should(Succeed())
		}, "10s").Should(Succeed())

		By("Instance should have the finalizer attached")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
		}, "10s").Should(Succeed())

		By("InstanceServiceDelete")
		_, _, err = openApiClient.InstanceServiceApi.InstanceServiceDelete(ctx, cloudAccountId, e.instanceResourceId).Execute()
		Expect(err).Should(Succeed())

		By("Instance should not be removed, but waiting for finalizer")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
			g.Expect(controllerutil.ContainsFinalizer(instanceRef, loadbalancer.LoadbalancerFinalizer)).Should(Equal(true))
		}, "10s").Should(Succeed())

		By("Remove LoadBalancer finalizer from the Instance")
		Eventually(func(g Gomega) {
			Expect(retry.RetryOnConflict(retry.DefaultRetry, func() error {
				g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
				g.Expect(controllerutil.RemoveFinalizer(instanceRef, loadbalancer.LoadbalancerFinalizer)).Should((Equal(true)))
				if err := k8sClient.Update(ctx, instanceRef); err != nil {
					return fmt.Errorf("PersistStatusUpdate: %w", err)
				}
				return nil
			})).Should(Succeed())
		}, "10s").Should(Succeed())

		By("Waiting for instance to be deleted from K8s")
		Eventually(func(g Gomega) {
			g.Expect(apierrors.IsNotFound(k8sClient.Get(ctx, instanceLookupKey, instanceRef))).Should(BeTrue())
		}, "10s").Should(Succeed())

		By("Waiting for Kubevirt VirtualMachine to be deleted from K8s")
		// TODO: This should happen before instance is deleted from K8s. Should not need to wait.
		Eventually(func(g Gomega) {
			g.Expect(apierrors.IsNotFound(k8sClient.Get(ctx, harvesterVMlookupKey, harvesterVmRef))).Should(BeTrue())
		}, "10s").Should(Succeed())

		By("Waiting for SshProxyTunnel to be deleted from K8s")
		Eventually(func(g Gomega) {
			g.Expect(apierrors.IsNotFound(k8sClient.Get(ctx, tunnelLookupKey, tunnelRef))).Should(BeTrue())
		}, "10s").Should(Succeed())

		By("Waiting for InstanceServiceGet to return NotFound")
		Eventually(func(g Gomega) {
			_, httpResponse, err := openApiClient.InstanceServiceApi.InstanceServiceGet(ctx, cloudAccountId, e.instanceResourceId).Execute()
			g.Expect(err).ShouldNot(Succeed())
			g.Expect(httpResponse.StatusCode).Should(Equal(http.StatusNotFound))
		}, "10s").Should(Succeed())
	})

	It("When the instance RunStrategy is updated in the Compute API Server, it should not get updated in K8s if instance is Provisioning", func() {
		e := NewBasicTestEnv(BasicTestEnvOptions{})
		defer e.Stop()
		e.Start()

		cloudAccountId := cloudaccount.MustNewId()
		e.namespace = cloudAccountId
		instanceType := CreateInstanceType(ctx)
		machineImage := CreateMachineImage(ctx)
		vNetName := CreateVNet(ctx, cloudAccountId)
		sshPublicKeyName := CreateSshPublicKey(ctx, cloudAccountId)
		runStrategyHaltedStr := "Halted"
		runStrategyHalted, err := openapi.NewProtoRunStrategyFromValue(runStrategyHaltedStr)
		Expect(err).Should(Succeed())
		statusPhaseStr := "Provisioning"
		runStrategyRerunOnFailureStr := "RerunOnFailure"

		By("InstanceServiceCreate")
		createInstanceReq := NewCreateInstanceRequest([]string{sshPublicKeyName}, instanceType, machineImage, vNetName)
		createResp, _, err := openApiClient.InstanceServiceApi.InstanceServiceCreate(ctx, cloudAccountId).Body(*createInstanceReq).Execute()
		Expect(err).Should(Succeed())
		Expect(*createResp.Metadata.ResourceId).ShouldNot(BeEmpty())
		e.instanceResourceId = *createResp.Metadata.ResourceId

		By("Waiting for instance to be created in K8s")
		instanceLookupKey := types.NamespacedName{Namespace: e.namespace, Name: e.instanceResourceId}
		instanceRef := &privatecloudv1alpha1.Instance{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
		}, "10s").Should(Succeed())

		By("Ensure that instance is in Provisioning state")
		Expect(string(*createResp.Status.Phase)).Should(Equal(statusPhaseStr))
		By("InstanceServiceUpdate RunStrategy should not succeed for Halted RunStrategy because instance is Provisioning")
		_, _, err = openApiClient.InstanceServiceApi.InstanceServiceUpdate(ctx, cloudAccountId, e.instanceResourceId).Body(
			openapi.InstanceServiceUpdateRequest{
				Spec: &openapi.ProtoInstanceSpec{
					RunStrategy:       runStrategyHalted,
					SshPublicKeyNames: createInstanceReq.Spec.SshPublicKeyNames,
				},
			}).Execute()
		Expect(err).ShouldNot(Succeed())

		By("InstanceServiceGet should return same RunStrategy")
		getResp, _, err := openApiClient.InstanceServiceApi.InstanceServiceGet(ctx, cloudAccountId, e.instanceResourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(string(*getResp.Spec.RunStrategy)).Should(Equal(runStrategyRerunOnFailureStr))

		By("Waiting for instance to be updated in K8s")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
			g.Expect(string(instanceRef.Spec.RunStrategy)).Should(Equal(runStrategyRerunOnFailureStr))
		}, "10s").Should(Succeed())
	})

	It("When the instance RunStrategy is updated in the Compute API Server, it should get updated in K8s if instance is Ready", func() {
		e := NewBasicTestEnv(BasicTestEnvOptions{})
		defer e.Stop()
		e.Start()

		cloudAccountId := cloudaccount.MustNewId()
		e.namespace = cloudAccountId
		instanceType := CreateInstanceType(ctx)
		machineImage := CreateMachineImage(ctx)
		vNetName := CreateVNet(ctx, cloudAccountId)
		sshPublicKeyName := CreateSshPublicKey(ctx, cloudAccountId)
		runStrategyHaltedStr := "Halted"
		runStrategyHalted, err := openapi.NewProtoRunStrategyFromValue(runStrategyHaltedStr)
		Expect(err).Should(Succeed())
		statusPhaseProvisioningStr := "Provisioning"
		statusPhaseReadyStr := "Ready"
		statusMessage := "Instance is running and has completed running startup scripts"

		By("InstanceServiceCreate")
		createInstanceReq := NewCreateInstanceRequest([]string{sshPublicKeyName}, instanceType, machineImage, vNetName)
		createResp, _, err := openApiClient.InstanceServiceApi.InstanceServiceCreate(ctx, cloudAccountId).Body(*createInstanceReq).Execute()
		Expect(err).Should(Succeed())
		Expect(*createResp.Metadata.ResourceId).ShouldNot(BeEmpty())
		e.instanceResourceId = *createResp.Metadata.ResourceId

		By("Waiting for instance to be created in K8s")
		instanceLookupKey := types.NamespacedName{Namespace: e.namespace, Name: e.instanceResourceId}
		instanceRef := &privatecloudv1alpha1.Instance{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
		}, "10s").Should(Succeed())

		By("Ensure that instance is Provisoning")
		Expect(string(*createResp.Status.Phase)).Should(Equal(statusPhaseProvisioningStr))

		By("InstanceUpdateStatusRequest should update the instance to Ready Phase")
		grpcClient := getInstanceGrpcClient(e)
		_, err = grpcClient.UpdateStatus(ctx, &pb.InstanceUpdateStatusRequest{
			Metadata: &pb.InstanceIdReference{
				CloudAccountId: cloudAccountId,
				ResourceId:     e.instanceResourceId,
			},
			Status: &pb.InstanceStatusPrivate{
				Phase:   pb.InstancePhase(pb.InstancePhase_value[statusPhaseReadyStr]),
				Message: statusMessage,
			},
		})
		Expect(err).Should(Succeed())

		By("InstanceServiceGet should return update Phase")
		getResp, _, err := openApiClient.InstanceServiceApi.InstanceServiceGet(ctx, cloudAccountId, e.instanceResourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(string(*getResp.Status.Phase)).Should(Equal(statusPhaseReadyStr))

		By("InstanceServiceUpdate RunStrategy should succeed when instance is in Ready state")
		_, _, err = openApiClient.InstanceServiceApi.InstanceServiceUpdate(ctx, cloudAccountId, e.instanceResourceId).Body(
			openapi.InstanceServiceUpdateRequest{
				Spec: &openapi.ProtoInstanceSpec{
					RunStrategy:       runStrategyHalted,
					SshPublicKeyNames: createInstanceReq.Spec.SshPublicKeyNames,
				},
			}).Execute()
		Expect(err).Should(Succeed())

		By("InstanceServiceGet should return same RunStrategy")
		getResp, _, err = openApiClient.InstanceServiceApi.InstanceServiceGet(ctx, cloudAccountId, e.instanceResourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(string(*getResp.Spec.RunStrategy)).Should(Equal(runStrategyHaltedStr))

		By("Waiting for instance to be updated in K8s")
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
			g.Expect(string(instanceRef.Spec.RunStrategy)).Should(Equal(runStrategyHaltedStr))
		}, "10s").Should(Succeed())
	})

	It("When the instance SshPublicKeys is updated in the Compute API Server, it should get updated in K8s", func() {
		e := NewBasicTestEnv(BasicTestEnvOptions{})
		defer e.Stop()
		e.Start()

		cloudAccountId := cloudaccount.MustNewId()
		e.namespace = cloudAccountId
		instanceType := CreateInstanceType(ctx)
		machineImage := CreateMachineImage(ctx)
		vNetName := CreateVNet(ctx, cloudAccountId)
		sshPublicKeyName1 := "key1-" + uuid.NewString()
		sshPublicKeyName2 := "key2-" + uuid.NewString()

		By("SshPublicKeyServiceCreate")
		createKeyReq1 := NewCreateSshPublicKeyRequest(sshPublicKeyName1)
		_, _, err := openApiClient.SshPublicKeyServiceApi.SshPublicKeyServiceCreate(ctx, cloudAccountId).Body(*createKeyReq1).Execute()
		Expect(err).Should(Succeed())
		createKeyReq2 := NewCreateSshPublicKeyRequest(sshPublicKeyName2)
		_, _, err = openApiClient.SshPublicKeyServiceApi.SshPublicKeyServiceCreate(ctx, cloudAccountId).Body(*createKeyReq2).Execute()
		Expect(err).Should(Succeed())

		By("InstanceServiceCreate with key1")
		createInstanceReq := NewCreateInstanceRequest([]string{sshPublicKeyName1}, instanceType, machineImage, vNetName)
		createResp, _, err := openApiClient.InstanceServiceApi.InstanceServiceCreate(ctx, cloudAccountId).Body(*createInstanceReq).Execute()
		Expect(err).Should(Succeed())
		Expect(*createResp.Metadata.ResourceId).ShouldNot(BeEmpty())
		e.instanceResourceId = *createResp.Metadata.ResourceId

		By("Waiting for instance to be created in K8s")
		instanceLookupKey := types.NamespacedName{Namespace: e.namespace, Name: e.instanceResourceId}
		instanceRef := &privatecloudv1alpha1.Instance{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
		}, "30s").Should(Succeed())

		By("Waiting for SshProxyTunnel to be created in K8s by the Instance Operator")
		tunnelLookupKey := types.NamespacedName{Namespace: e.namespace, Name: e.instanceResourceId}
		tunnelRef := &privatecloudv1alpha1.SshProxyTunnel{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, tunnelLookupKey, tunnelRef)).Should(Succeed())
			g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
			g.Expect(len(instanceRef.Status.Interfaces)).Should(Equal(1))
			g.Expect(len(instanceRef.Status.Interfaces[0].Addresses)).Should(Equal(1))
			g.Expect(tunnelRef.Spec.TargetAddresses).
				Should(Equal([]string{instanceRef.Status.Interfaces[0].Addresses[0]}))
		}, "10s").Should(Succeed())

		By("InstanceServiceUpdate SshPublicKeyNames with key1,key2")
		_, _, err = openApiClient.InstanceServiceApi.InstanceServiceUpdate(ctx, cloudAccountId, e.instanceResourceId).Body(
			openapi.InstanceServiceUpdateRequest{
				Spec: &openapi.ProtoInstanceSpec{
					RunStrategy:       createInstanceReq.Spec.RunStrategy,
					SshPublicKeyNames: []string{sshPublicKeyName1, sshPublicKeyName2},
				},
			}).Execute()
		Expect(err).Should(Succeed())

		By("InstanceServiceGet should return updated SshPublicKeyNames")
		getResp, _, err := openApiClient.InstanceServiceApi.InstanceServiceGet(ctx, cloudAccountId, e.instanceResourceId).Execute()
		Expect(err).Should(Succeed())
		Expect(getResp.Spec.SshPublicKeyNames).Should(Equal([]string{sshPublicKeyName1, sshPublicKeyName2}))

		By("Waiting for instance to be updated in K8s")
		expectedSshPublicKeys := []privatecloudv1alpha1.SshPublicKeySpec{
			{
				SshPublicKey: *createKeyReq1.Spec.SshPublicKey,
			},
			{
				SshPublicKey: *createKeyReq2.Spec.SshPublicKey,
			},
		}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, instanceLookupKey, instanceRef)).Should(Succeed())
			g.Expect(instanceRef.Spec.SshPublicKeySpecs).Should(Equal(expectedSshPublicKeys))
		}, "30s").Should(Succeed())

		By("Waiting for SshProxyTunnel to be updated in K8s with key1,key2")
		cleanPublicKey1, err := privatecloudsshproxy.CleanPublicKey(ctx, *createKeyReq1.Spec.SshPublicKey)
		Expect(err).Should(Succeed())
		cleanPublicKey2, err := privatecloudsshproxy.CleanPublicKey(ctx, *createKeyReq2.Spec.SshPublicKey)
		Expect(err).Should(Succeed())
		expectedKeyDataOp := e.publicKeyOp
		expectedKeyData1 := "permitopen=\"" + instanceRef.Status.Interfaces[0].Addresses[0] + ":22" + "\"," +
			privatecloudsshproxy.SshAuthorizedKeysOptions + " " + cleanPublicKey1 + "\n"
		expectedKeyData2 := "permitopen=\"" + instanceRef.Status.Interfaces[0].Addresses[0] + ":22" + "\"," +
			privatecloudsshproxy.SshAuthorizedKeysOptions + " " + cleanPublicKey2 + "\n"

		Eventually(func() bool {
			str, _ := ReadAuthorizedKeysFile(e.authorizedKeysFilePath)
			return strings.Contains(str, expectedKeyDataOp) &&
				strings.Contains(str, expectedKeyData1) &&
				strings.Contains(str, expectedKeyData2)
		}, "30s").Should(BeTrue())
	})
})

func ReadAuthorizedKeysFile(authorizedKeysPath string) (string, error) {
	buf, err := os.ReadFile(authorizedKeysPath)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func NewMockCloudAccountServiceClient() pb.CloudAccountServiceClient {
	mockController := gomock.NewController(GinkgoT())
	cloudAccountClient := pb.NewMockCloudAccountServiceClient(mockController)

	cloudAccount := &pb.CloudAccount{
		Type: pb.AccountType_ACCOUNT_TYPE_STANDARD,
	}

	cloudAccountClient.EXPECT().GetById(gomock.Any(), gomock.Any()).Return(cloudAccount, nil).AnyTimes()
	return cloudAccountClient
}

func NewMockCloudAccountAppClientServiceClient() pb.CloudAccountAppClientServiceClient {
	mockController := gomock.NewController(GinkgoT())
	cloudAccountAppClientServiceClient := pb.NewMockCloudAccountAppClientServiceClient(mockController)

	cloudAccount := &pb.CloudAccount{
		Type: pb.AccountType_ACCOUNT_TYPE_STANDARD,
	}

	cloudAccountAppClientServiceClient.EXPECT().GetAppClientCloudAccount(gomock.Any(), gomock.Any()).Return(cloudAccount, nil).AnyTimes()
	return cloudAccountAppClientServiceClient
}

func NewMockObjectStorageServicePrivateClient() pb.ObjectStorageServicePrivateClient {
	mockController := gomock.NewController(GinkgoT())
	objectStorageServicePrivateClient := pb.NewMockObjectStorageServicePrivateClient(mockController)

	objectStorageServicePrivateClient.EXPECT().AddBucketSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	objectStorageServicePrivateClient.EXPECT().RemoveBucketSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	return objectStorageServicePrivateClient
}
