// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package helper

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"encoding/json"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	computeopenapi "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/openapi"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tlsutil"
	toolsssh "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/ssh"
	testcommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/test/common"
	restyclient "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/test/compute/restyclient"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/test/kindtestenv"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ComputeTestHelper struct {
	ComputeOpenApiClient            *computeopenapi.APIClient
	ComputeGrpcClientConn           *grpc.ClientConn
	ComputeSchedulerGrpcClientConn  *grpc.ClientConn
	GlobalGrpcClientConn            *grpc.ClientConn
	RestyClient                     *restyclient.RestyClient
	InstanceGroupServiceClient      pb.InstanceGroupServiceClient
	InstancePrivateServiceClient    pb.InstancePrivateServiceClient
	CloudAccountServiceClient       pb.CloudAccountServiceClient
	FilesystemPrivateServiceClient  pb.FilesystemPrivateServiceClient
	CloudAccountRestUrl             string
	InstanceSchedulingServiceClient pb.InstanceSchedulingServiceClient
	FleetAdminServiceClient         pb.FleetAdminServiceClient
	// List of instances created, which will be deleted by the Cleanup() function.
	createdInstances   []qualifiedResourceId
	createdFilesystems []qualifiedResourceId
}

type qualifiedResourceId struct {
	CloudAccountId string
	ResourceId     string
}

func NewComputeTestHelper(
	computeOpenApiClient *computeopenapi.APIClient,
	computeGrpcClientConn *grpc.ClientConn,
	computeSchedulerGrpcClientConn *grpc.ClientConn,
	globalGrpcClientConn *grpc.ClientConn,
	cloudAccountRestUrl string,
) *ComputeTestHelper {
	return &ComputeTestHelper{
		ComputeOpenApiClient:            computeOpenApiClient,
		ComputeGrpcClientConn:           computeGrpcClientConn,
		ComputeSchedulerGrpcClientConn:  computeSchedulerGrpcClientConn,
		GlobalGrpcClientConn:            globalGrpcClientConn,
		InstancePrivateServiceClient:    pb.NewInstancePrivateServiceClient(computeGrpcClientConn),
		CloudAccountServiceClient:       pb.NewCloudAccountServiceClient(globalGrpcClientConn),
		FilesystemPrivateServiceClient:  pb.NewFilesystemPrivateServiceClient(computeGrpcClientConn),
		CloudAccountRestUrl:             cloudAccountRestUrl,
		InstanceSchedulingServiceClient: pb.NewInstanceSchedulingServiceClient(computeSchedulerGrpcClientConn),
		FleetAdminServiceClient:         pb.NewFleetAdminServiceClient(computeGrpcClientConn),
	}
}

func NewComputeTestHelperFromKindTestEnv(ctx context.Context, kindTestEnv *kindtestenv.KindTestEnv) *ComputeTestHelper {
	computeOpenApiClient := NewComputeOpenApiClientFromKindTestEnv(ctx, kindTestEnv)
	computeGrpcClientConn := NewComputeGrpcClientConnFromKindTestEnv(ctx, kindTestEnv)
	computeAZGrpcClientConn := NewComputeSchedulerGrpcClientConnFromKindTestEnv(ctx, kindTestEnv)
	globalGrpcClientConn := NewGlobalGrpcClientConnFromKindTestEnv(ctx, kindTestEnv)
	cloudAccountRestUrl := fmt.Sprintf("https://dev.api.cloud.intel.com.kind.local:%d", kindTestEnv.IngressHttpsPort())
	return NewComputeTestHelper(
		computeOpenApiClient,
		computeGrpcClientConn,
		computeAZGrpcClientConn,
		globalGrpcClientConn,
		cloudAccountRestUrl,
	)
}

func NewComputeOpenApiClientFromKindTestEnv(ctx context.Context, kindTestEnv *kindtestenv.KindTestEnv) *computeopenapi.APIClient {
	clientConfig := computeopenapi.NewConfiguration()
	clientConfig.Scheme = "https"
	clientConfig.Host = fmt.Sprintf("dev.compute.us-dev-1.api.cloud.intel.com.kind.local:%d", kindTestEnv.IngressHttpsPort())
	clientConfig.HTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	return computeopenapi.NewAPIClient(clientConfig)
}

func NewComputeGrpcClientConnFromKindTestEnv(ctx context.Context, kindTestEnv *kindtestenv.KindTestEnv) *grpc.ClientConn {
	computeApiServerGrpcAddress := fmt.Sprintf("dev.compute.us-dev-1.grpcapi.cloud.intel.com.kind.local:%d", kindTestEnv.IngressHttpsPort())
	pkiDir := filepath.Join(kindTestEnv.SecretsDir, "pki", "client1")
	certFile := filepath.Join(pkiDir, "cert.pem")
	keyFile := filepath.Join(pkiDir, "cert.key")
	caCertFile := filepath.Join(pkiDir, "ca.pem")
	tlsProvider := tlsutil.NewTestTlsProvider(certFile, keyFile, caCertFile)
	tlsConfig, err := tlsProvider.ClientTlsConfig(ctx)
	Expect(err).Should(Succeed())
	transportCredentials := credentials.NewTLS(tlsConfig)
	computeGrpcClientConn, err := grpc.Dial(computeApiServerGrpcAddress, grpc.WithTransportCredentials(transportCredentials))
	Expect(err).Should(Succeed())
	return computeGrpcClientConn
}

func NewGlobalGrpcClientConnFromKindTestEnv(ctx context.Context, kindTestEnv *kindtestenv.KindTestEnv) *grpc.ClientConn {
	globalApiServerGrpcAddress := fmt.Sprintf("dev.grpcapi.cloud.intel.com.kind.local:%d", kindTestEnv.IngressHttpsPort())
	pkiDir := filepath.Join(kindTestEnv.SecretsDir, "pki", "client1")
	certFile := filepath.Join(pkiDir, "cert.pem")
	keyFile := filepath.Join(pkiDir, "cert.key")
	caCertFile := filepath.Join(pkiDir, "ca.pem")
	tlsProvider := tlsutil.NewTestTlsProvider(certFile, keyFile, caCertFile)
	tlsConfig, err := tlsProvider.ClientTlsConfig(ctx)
	Expect(err).Should(Succeed())
	transportCredentials := credentials.NewTLS(tlsConfig)
	globalGrpcClientConn, err := grpc.Dial(globalApiServerGrpcAddress, grpc.WithTransportCredentials(transportCredentials))
	Expect(err).Should(Succeed())
	return globalGrpcClientConn
}

func NewComputeSchedulerGrpcClientConnFromKindTestEnv(ctx context.Context, kindTestEnv *kindtestenv.KindTestEnv) *grpc.ClientConn {
	computeAZApiServerGrpcAddress := fmt.Sprintf("dev.compute.us-dev-1a.grpcapi.cloud.intel.com.kind.local:%d", kindTestEnv.IngressHttpsPort())
	pkiDir := filepath.Join(kindTestEnv.SecretsDir, "pki", "client1")
	certFile := filepath.Join(pkiDir, "cert.pem")
	keyFile := filepath.Join(pkiDir, "cert.key")
	caCertFile := filepath.Join(pkiDir, "ca.pem")
	tlsProvider := tlsutil.NewTestTlsProvider(certFile, keyFile, caCertFile)
	tlsConfig, err := tlsProvider.ClientTlsConfig(ctx)
	Expect(err).Should(Succeed())
	transportCredentials := credentials.NewTLS(tlsConfig)
	computeAZGrpcClientConn, err := grpc.Dial(computeAZApiServerGrpcAddress, grpc.WithTransportCredentials(transportCredentials))
	Expect(err).Should(Succeed())
	return computeAZGrpcClientConn
}

func (h *ComputeTestHelper) PingComputeApiServer(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("ComputeTestHelper.PingComputeApiServer")
	_, _, err := h.ComputeOpenApiClient.SshPublicKeyServiceApi.SshPublicKeyServicePing(ctx).Execute()
	if err != nil {
		log.Error(err, "Service is not ready")
		return err
	}
	log.Info("Service is ready")
	return nil
}

func (h *ComputeTestHelper) PingCloudAccountServer(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("ComputeTestHelper.PingCloudAccountServer")
	pingUrl := fmt.Sprintf("%s/v1/cloudaccount/ping", h.CloudAccountRestUrl)
	log.Info("Pinging service", "pingUrl", pingUrl)
	response, err := h.RestyClient.Request(ctx, "GET", pingUrl, nil)
	if err != nil {
		log.Error(err, "Service is not ready")
		return err
	}
	if response.StatusCode() != 200 {
		err := fmt.Errorf("ping returned status code %d", response.StatusCode())
		log.Error(err, "Service is not ready")
		return err
	}
	log.Info("Service is ready")
	return nil
}

func (h *ComputeTestHelper) PingInstanceScheduler(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("ComputeTestHelper.PingInstanceScheduler")
	_, err := h.InstanceSchedulingServiceClient.Ready(ctx, &emptypb.Empty{})
	if err != nil {
		log.Error(err, "Service is not ready")
		return err
	}
	log.Info("Service is ready")
	return nil
}

func (h *ComputeTestHelper) PingFleetAdminService(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("ComputeTestHelper.PingFleetAdminService")
	_, err := h.FleetAdminServiceClient.Ping(ctx, &emptypb.Empty{})
	if err != nil {
		log.Error(err, "Service is not ready")
		return err
	}
	log.Info("Service is ready")
	return nil
}

func (h *ComputeTestHelper) CreateVNet(ctx context.Context, cloudAccountId string, vNetName string, availabilityZone string) (string, *computeopenapi.ProtoVNet) {
	By("Creating VNet")
	createVNetReq := h.NewPutVNetRequest(vNetName, availabilityZone)
	vnetResp, _, err := h.ComputeOpenApiClient.VNetServiceApi.VNetServicePut(ctx, cloudAccountId).Body(*createVNetReq).Execute()
	Expect(err).Should(Succeed())
	return vNetName, vnetResp
}

func (h *ComputeTestHelper) NewPutVNetRequest(name string, availabilityZone string) *computeopenapi.VNetServicePutRequest {
	region := "us-dev-1"
	prefixLength := int32(22)
	return &computeopenapi.VNetServicePutRequest{
		Metadata: &computeopenapi.VNetServicePutRequestMetadata{
			Name: &name,
		},
		Spec: &computeopenapi.ProtoVNetSpec{
			Region:           &region,
			AvailabilityZone: &availabilityZone,
			PrefixLength:     &prefixLength,
		},
	}
}

func (h *ComputeTestHelper) CreateSshPublicKey(ctx context.Context, cloudAccountId string) (name string, privateKey string, publicKey string, CreateSshPublicKeyResp *computeopenapi.ProtoSshPublicKey) {
	By("Creating SSH Key Pair")
	name = "key-" + uuid.NewString()
	privateKey, publicKey, err := toolsssh.CreateSshRsaKeyPair(4096, name)
	Expect(err).Should(Succeed())
	createKeyReq := h.NewCreateSshPublicKeyRequest(name, publicKey)
	CreateSshPublicKeyResp, _, err = h.ComputeOpenApiClient.SshPublicKeyServiceApi.SshPublicKeyServiceCreate(ctx, cloudAccountId).Body(*createKeyReq).Execute()
	Expect(err).Should(Succeed())
	return
}

func (h *ComputeTestHelper) NewCreateSshPublicKeyRequest(name string, publicKey string) *computeopenapi.SshPublicKeyServiceCreateRequest {
	return &computeopenapi.SshPublicKeyServiceCreateRequest{
		Metadata: &computeopenapi.SshPublicKeyServiceCreateRequestMetadata{
			Name: &name,
		},
		Spec: &computeopenapi.ProtoSshPublicKeySpec{
			SshPublicKey: &publicKey,
		},
	}
}

func (h *ComputeTestHelper) CreateInstance(ctx context.Context, cloudAccountId string, createInstanceReq *computeopenapi.InstanceServiceCreateRequest) (*computeopenapi.ProtoInstance, error) {
	By("Creating Instance")
	createInstanceResp, _, err := h.ComputeOpenApiClient.InstanceServiceApi.InstanceServiceCreate(ctx, cloudAccountId).Body(*createInstanceReq).Execute()
	if err != nil {
		return nil, err
	}
	h.createdInstances = append(h.createdInstances, qualifiedResourceId{
		CloudAccountId: cloudAccountId,
		ResourceId:     *createInstanceResp.Metadata.ResourceId,
	})
	return createInstanceResp, nil
}

func (h *ComputeTestHelper) CreateInstanceGroup(ctx context.Context, cloudAccountId string, createInstanceGroup *computeopenapi.InstanceGroupServiceCreateRequest) (*computeopenapi.ProtoInstanceGroup, error) {
	log := log.FromContext(ctx).WithName("ComputeTestHelper.CreateInstanceGroup")
	createInstanceGroupResp, _, err := h.ComputeOpenApiClient.InstanceGroupServiceApi.InstanceGroupServiceCreate(ctx, cloudAccountId).Body(*createInstanceGroup).Execute()
	if err != nil {
		return nil, err
	}
	log.Info("Create InstanceGroup body", "createInstanceGroup: ", createInstanceGroup)

	return createInstanceGroupResp, nil
}

func (h *ComputeTestHelper) NewCreateInstanceGroupRequest(sshPublicKeyNames []string, instanceType string, machineImage string, vNetName string,
	availabilityZone string, instanceGroupName string, instanceCount int32) *computeopenapi.InstanceGroupServiceCreateRequest {

	runStrategyStr := "RerunOnFailure"
	runStrategy, err := computeopenapi.NewProtoRunStrategyFromValue(runStrategyStr)
	Expect(err).Should(Succeed())
	interfaces := []computeopenapi.ProtoNetworkInterface{{VNet: &vNetName}}
	return &computeopenapi.InstanceGroupServiceCreateRequest{
		Metadata: &computeopenapi.InstanceGroupServiceCreateRequestMetadata{
			Name: &instanceGroupName,
		},
		Spec: &computeopenapi.ProtoInstanceGroupSpec{
			InstanceSpec: &computeopenapi.ProtoInstanceSpec{
				AvailabilityZone:  &availabilityZone,
				InstanceType:      &instanceType,
				MachineImage:      &machineImage,
				RunStrategy:       runStrategy,
				SshPublicKeyNames: sshPublicKeyNames,
				Interfaces:        interfaces,
			},
			InstanceCount: &instanceCount,
		},
	}
}

func (h *ComputeTestHelper) NewCreateInstanceRequest(sshPublicKeyNames []string, instanceType string, machineImage string, vNetName string, availabilityZone string) *computeopenapi.InstanceServiceCreateRequest {
	runStrategyStr := "RerunOnFailure"
	runStrategy, err := computeopenapi.NewProtoRunStrategyFromValue(runStrategyStr)
	Expect(err).Should(Succeed())
	interfaces := []computeopenapi.ProtoNetworkInterface{{VNet: &vNetName}}
	return &computeopenapi.InstanceServiceCreateRequest{
		Metadata: &computeopenapi.InstanceServiceCreateRequestMetadata{
			Labels: &map[string]string{
				"iks-cluster-name": "my-iks-cluster-1",
				"iks-role":         "master",
			},
		},
		Spec: &computeopenapi.ProtoInstanceSpec{
			AvailabilityZone:  &availabilityZone,
			InstanceType:      &instanceType,
			MachineImage:      &machineImage,
			RunStrategy:       runStrategy,
			SshPublicKeyNames: sshPublicKeyNames,
			Interfaces:        interfaces,
		},
	}
}

func (h *ComputeTestHelper) CreateInstancePrivateGrpc(ctx context.Context, createInstanceReq *pb.InstanceCreatePrivateRequest) (*pb.InstancePrivate, error) {
	By("Creating Instance")
	log := log.FromContext(ctx).WithName("ComputeTestHelper.CreateInstancePrivateGrpc")
	instance, err := h.InstancePrivateServiceClient.CreatePrivate(ctx, createInstanceReq)
	if err != nil {
		return nil, err
	}
	h.createdInstances = append(h.createdInstances, qualifiedResourceId{
		CloudAccountId: instance.Metadata.CloudAccountId,
		ResourceId:     instance.Metadata.ResourceId,
	})
	log.Info("Created instance", "createdInstances", h.createdInstances)
	return instance, nil
}

func (h *ComputeTestHelper) NewCreateInstancePrivateRequestGrpc(cloudAccountId string, sshPublicKeyNames []string, instanceType string, machineImage string,
	vNetName string, availabilityZone string) *pb.InstanceCreatePrivateRequest {
	interfaces := []*pb.NetworkInterfacePrivate{{VNet: vNetName}}
	req := &pb.InstanceCreatePrivateRequest{
		Metadata: &pb.InstanceMetadataCreatePrivate{
			CloudAccountId: cloudAccountId,
			SkipQuotaCheck: true,
		},
		Spec: &pb.InstanceSpecPrivate{
			AvailabilityZone:  availabilityZone,
			InstanceType:      instanceType,
			MachineImage:      machineImage,
			SshPublicKeyNames: sshPublicKeyNames,
			Interfaces:        interfaces,
		},
	}
	return req
}

func (h *ComputeTestHelper) CreateFilesystemPrivateGrpc(ctx context.Context, createFilesystemReq *pb.FilesystemCreateRequestPrivate) (*pb.FilesystemPrivate, error) {
	By("Creating volume")
	log := log.FromContext(ctx).WithName("ComputeTestHelper.CreateFilesystemPrivateGrpc")
	filesystem, err := h.FilesystemPrivateServiceClient.CreatePrivate(ctx, createFilesystemReq)
	if err != nil {
		return nil, err
	}
	h.createdFilesystems = append(h.createdInstances, qualifiedResourceId{
		CloudAccountId: filesystem.Metadata.CloudAccountId,
		ResourceId:     filesystem.Metadata.ResourceId,
	})
	log.Info("Created instance", "createdFileSystem", h.createdFilesystems)
	return filesystem, nil
}

func (h *ComputeTestHelper) NewCreateFilesystemPrivateRequestGrpc(cloudAccountId string) *pb.FilesystemCreateRequestPrivate {
	name := "storage-automation-" + h.GetRandomStringWithLimit(6)
	availabilityZone := "us-dev-1a"
	request := &pb.FilesystemCapacity{Storage: "5GB"}
	storageClass := pb.FilesystemStorageClass_AIOptimized
	req := &pb.FilesystemCreateRequestPrivate{
		Metadata: &pb.FilesystemMetadataPrivate{
			Name:             name,
			CloudAccountId:   cloudAccountId,
			SkipQuotaCheck:   true,
			SkipProductCheck: true,
		},
		Spec: &pb.FilesystemSpecPrivate{
			AvailabilityZone: availabilityZone,
			StorageClass:     storageClass,
			Request:          request,
		},
	}
	fmt.Println(req)
	return req
}

func (h *ComputeTestHelper) DeleteInstance(ctx context.Context, cloudAccountId string, instanceResourceId string) error {
	log := log.FromContext(ctx).WithName("ComputeTestHelper.DeleteInstance").WithValues("cloudAccountId", cloudAccountId, "instanceResourceId", instanceResourceId)
	By("Deleting Instance")
	useOpenApi := false
	if useOpenApi {
		// TODO: Below fails because openapi client does not send Content-Type header.
		results, _, err := h.ComputeOpenApiClient.InstanceServiceApi.InstanceServiceDelete(ctx, cloudAccountId, instanceResourceId).Execute()
		log.Info("Result", "results", results, "err", err)
		return err
	} else {
		config := h.ComputeOpenApiClient.GetConfig()
		urlPrefix := config.Scheme + "://" + config.Host
		cmd := exec.CommandContext(ctx,
			"curl",
			"-vk",
			"-H", "Accept: application/json",
			"-H", "Content-type: application/json",
			"-X", "DELETE",
			urlPrefix+"/v1/cloudaccounts/"+cloudAccountId+"/instances/id/"+instanceResourceId)
		return testcommon.RunCmd(ctx, cmd)
	}
}

func (h *ComputeTestHelper) DeleteInstanceViaResty(ctx context.Context, instance_endpoint string, cloudAccountId string, instanceResourceId string) (response *resty.Response) {
	log := log.FromContext(ctx).WithName("ComputeTestHelper.DeleteInstance").WithValues("cloudAccountId", cloudAccountId, "instanceResourceId", instanceResourceId)
	log.Info("Deletion of instance :", "Instance resource id", instanceResourceId)
	instance_delete_ep := instance_endpoint + "/id/" + instanceResourceId
	response, _ = h.RestyClient.Request(ctx, "DELETE", instance_delete_ep, nil)
	return
}

func (h *ComputeTestHelper) DeleteInstanceGroupViaResty(ctx context.Context, instancegroup_endpoint string, cloudAccountId string, instanceGroupName string) (response *resty.Response) {
	log := log.FromContext(ctx).WithName("ComputeTestHelper.DeleteInstance").WithValues("cloudAccountId", cloudAccountId, "instanceGroupName", instanceGroupName)
	log.Info("Deletion of instance group :", "Instance group name", instanceGroupName)
	instance_delete_ep := instancegroup_endpoint + "/name/" + instanceGroupName
	response, _ = h.RestyClient.Request(ctx, "DELETE", instance_delete_ep, nil)
	return
}

func (h *ComputeTestHelper) GetInstance(ctx context.Context, cloudAccountId string, instanceResourceId string) (*computeopenapi.ProtoInstance, error) {
	instance, _, err := h.ComputeOpenApiClient.InstanceServiceApi.InstanceServiceGet(ctx, cloudAccountId, instanceResourceId).Execute()
	return instance, err
}

func (h *ComputeTestHelper) GetAllInstanceGroups(ctx context.Context, cloudAccountId string) (*computeopenapi.ProtoInstanceGroupSearchResponse, error) {
	instanceGroup, _, err := h.ComputeOpenApiClient.InstanceGroupServiceApi.InstanceGroupServiceSearch(ctx, cloudAccountId).Execute()
	return instanceGroup, err
}

func (h *ComputeTestHelper) SearchInstanceGroup(ctx context.Context, cloudAccountId string, searchInstanceGroupBody *computeopenapi.InstanceServiceSearch2Request) (*computeopenapi.ProtoInstanceSearchResponse, error) {
	log := log.FromContext(ctx).WithName("ComputeTestHelper.SearchInstanceGroup")
	instanceGroupSearch, _, err := h.ComputeOpenApiClient.InstanceServiceApi.InstanceServiceSearch2(ctx, cloudAccountId).Body(*searchInstanceGroupBody).Execute()
	if err != nil {
		return nil, err
	}
	log.Info("Search InstanceGroup body", "searchInstanceGroupBody: ", searchInstanceGroupBody)

	return instanceGroupSearch, err
}

func (h *ComputeTestHelper) NewSearchInstanceGroupRequest(instanceGroupName string) *computeopenapi.InstanceServiceSearch2Request {
	request := &computeopenapi.InstanceServiceSearch2Request{
		Metadata: &computeopenapi.InstanceServiceSearch2RequestMetadata{
			InstanceGroup: &instanceGroupName,
		},
	}
	return request
}

// Returns nil if instance is not found. Otherwise returns an error.
func (h *ComputeTestHelper) CheckInstanceNotFound(ctx context.Context, cloudAccountId string, instanceResourceId string) error {
	log := log.FromContext(ctx).WithName("ComputeTestHelper.CheckInstanceNotFound").WithValues("cloudAccountId", cloudAccountId, "instanceResourceId", instanceResourceId)
	instance, httpResponse, err := h.ComputeOpenApiClient.InstanceServiceApi.InstanceServiceGet(ctx, cloudAccountId, instanceResourceId).Execute()
	log.Info("InstanceServiceGet", "instance", instance, "httpResponse", httpResponse, "err", err)
	if httpResponse.StatusCode == http.StatusNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	return fmt.Errorf("instance found: %v", instance)
}

// Returns nil if instance group is not found. Otherwise returns an error.
func (h *ComputeTestHelper) CheckInstanceGroupNotFound(ctx context.Context, cloudAccountId string) error {
	log := log.FromContext(ctx).WithName("ComputeTestHelper.CheckInstanceGroupNotFound").WithValues("cloudAccountId", cloudAccountId)
	instanceGroup, httpResponse, err := h.ComputeOpenApiClient.InstanceGroupServiceApi.InstanceGroupServiceSearch(ctx, cloudAccountId).Execute()
	log.Info("GetAllInstanceGroups", "instanceGroup", instanceGroup, "httpResponse", httpResponse, "err", err)
	jsonBytes, _ := json.Marshal(instanceGroup)
	jsonString := string(jsonBytes)
	if httpResponse.StatusCode == http.StatusOK {
		if jsonString == `{"items":[]}` {
			return nil
		}
	}
	if err != nil {
		return err
	}
	return fmt.Errorf("instance found: %v", instanceGroup)
}

func (h *ComputeTestHelper) CloudAccountCreation(ctx context.Context, username string, account_type string) string {
	tid := h.GetRandomStringWithLimit(12)
	oid := h.GetRandomStringWithLimit(12)
	name := "compute-cloudaccount-" + h.GetRandomStringWithLimit(6)
	cloudaccount_payload := fmt.Sprintf(`{"name":"%s","owner":"%s","tid":"%s","oid":"%s","type":"%s"}`, name, username, tid, oid, account_type)
	url := fmt.Sprintf("%s/v1/cloudaccounts", h.CloudAccountRestUrl)
	response, _ := h.RestyClient.Request(ctx, "POST", url, []byte(cloudaccount_payload))
	Expect(response.StatusCode()).To(Equal(200))
	cloudAccount := gjson.Get(response.String(), "id").String()
	return cloudAccount
}

func (h *ComputeTestHelper) GetUnixTime(timestamp string) int64 {
	layout := time.RFC3339
	t, _ := time.Parse(layout, timestamp)
	return t.UnixMilli()
}

func (h *ComputeTestHelper) ValidateTimeStamp(i, min, max int64) bool {
	if (i >= min) && (i <= max) {
		return true
	} else {
		return false
	}
}

func (h *ComputeTestHelper) GetRandomStringWithLimit(n int) string {
	var charset = []rune("abcdefghijklmnopqrstuvwxyz0987654321")
	rand.Seed(time.Now().UnixNano())
	str := make([]rune, n)
	for i := range str {
		str[i] = charset[rand.Intn(len(charset))]
	}
	return string(str)
}

func (h *ComputeTestHelper) LoadSuiteLevelTestData(ctx context.Context, kindTestEnv *kindtestenv.KindTestEnv) (instance_ep string, instance_group_ep string,
	ssh_ep string, vnet_ep string, instance_type_ep string, machine_image_ep string, cloudAccount string) {
	compute_url := "https://dev.compute.us-dev-1.api.cloud.intel.com.kind.local"

	port_number := strconv.Itoa(kindTestEnv.IngressHttpsPort())
	username := h.GetRandomStringWithLimit(10) + "@intel.com"
	cloudAccount = h.CloudAccountCreation(ctx, username, "ACCOUNT_TYPE_INTEL")

	// populate the endpoints after cloud account creation
	instance_ep = compute_url + ":" + port_number + "/v1/cloudaccounts/" + cloudAccount + "/" + "instances"
	instance_group_ep = compute_url + ":" + port_number + "/v1/cloudaccounts/" + cloudAccount + "/" + "instancegroups"
	ssh_ep = compute_url + ":" + port_number + "/v1/cloudaccounts/" + cloudAccount + "/" + "sshpublickeys"
	vnet_ep = compute_url + ":" + port_number + "/v1/cloudaccounts/" + cloudAccount + "/vnets"
	instance_type_ep = compute_url + ":" + port_number + "/v1/instancetypes"
	machine_image_ep = compute_url + ":" + port_number + "/v1/machineimages"
	return
}

func (h *ComputeTestHelper) LoadStorageSuiteTestData(ctx context.Context, kindTestEnv *kindtestenv.KindTestEnv) (filesystem_ep string, cloudAccount string) {
	compute_url := "https://dev.compute.us-dev-1.api.cloud.intel.com.kind.local"

	port_number := strconv.Itoa(kindTestEnv.IngressHttpsPort())
	username := h.GetRandomStringWithLimit(10) + "@intel.com"
	cloudAccount = h.CloudAccountCreation(ctx, username, "ACCOUNT_TYPE_INTEL")

	// populate the endpoints after cloud account creation
	filesystem_ep = compute_url + ":" + port_number + "/v1/cloudaccounts/" + cloudAccount + "/" + "filesystems"
	return
}

// Attempt to delete all instances created by this type.
func (h *ComputeTestHelper) Cleanup(ctx context.Context) error {
	By("Cleanup Instances")
	log := log.FromContext(ctx).WithName("ComputeTestHelper.Cleanup")
	log.Info("BEGIN", "createdInstances", h.createdInstances)
	defer log.Info("END")

	// Gomega fail handler that only logs errors.
	g := gomega.NewGomega(func(message string, callerSkip ...int) { log.Info(message) })

	// Attempt to delete instances.
	log.Info("Instances to be deleted :", "Instance list", h.createdInstances)
	for _, createdInstance := range h.createdInstances {
		if err := h.DeleteInstance(ctx, createdInstance.CloudAccountId, createdInstance.ResourceId); err != nil {
			log.Info("Unable to delete instance: %w", err)
			// continue trying to clean up.
		}
	}

	// Wait for all instances to be not found.
	g.Eventually(func(g Gomega) {
		for _, createdInstance := range h.createdInstances {
			g.Expect(h.CheckInstanceNotFound(ctx, createdInstance.CloudAccountId, createdInstance.ResourceId)).Should(Succeed())
		}
	}, "2m", "1s").Should(Succeed())
	return nil
}

func ReadFileAsString(ctx context.Context, filename string) (string, error) {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(fileBytes), err
}

func ScanHostPubKeyfile(ctx context.Context, filePathParent string) error {
	AbsfilePath := filepath.Join(filePathParent, "host_public_key")

	ssh_proxy_cmd := exec.CommandContext(ctx, "/bin/sh", "-c", fmt.Sprintf(`ssh-keyscan -t rsa ${HOST_IP} | awk '{print $2, $3}'> %s`, AbsfilePath))
	if err := testcommon.RunCmd(ctx, ssh_proxy_cmd); err != nil {
		return err
	}
	return nil
}
