// Copyright (C) 2023 Intel Corporation
package instance

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	bmenroll "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/tasks"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/pbconvert"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/protodb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

const (
	maxUserdataSizeBytes                  = 256 * 1024
	bgpClusterInterfaceName               = "bgp0"
	acceleratorClusterInterfaceName       = "gpu0"
	tenantDefaultInterfaceName            = "eth0"
	acceleratorClusterNetworkPrefixLength = 22
	defaultUserData                       = "OMITTED"
	ResourceExhaustedMessage              = "we are currently experiencing high demand for this instance type. Please try again later."
)

type InstanceService struct {
	pb.UnimplementedInstanceServiceServer
	pb.UnimplementedInstancePrivateServiceServer
	db                          *sql.DB
	sshPublicKeyService         pb.SshPublicKeyServiceServer
	instanceTypeService         pb.InstanceTypeServiceServer
	machineImageService         pb.MachineImageServiceServer
	vNetService                 pb.VNetServiceServer
	sqlTransformer              *InstanceSqlTransformer
	pbConverter                 *pbconvert.PbConverter
	cfg                         config.Config
	vmInstanceSchedulingService pb.InstanceSchedulingServiceClient
	cloudAccountServiceClient   pb.CloudAccountServiceClient
	// fleetAdminServiceClient may be nil
	fleetAdminServiceClient pb.FleetAdminServiceClient
	qmsClient               pb.QuotaManagementPrivateServiceClient
}

func NewInstanceService(
	db *sql.DB,
	sshPublicKeyService pb.SshPublicKeyServiceServer,
	instanceTypeService pb.InstanceTypeServiceServer,
	machineImageService pb.MachineImageServiceServer,
	vNetService pb.VNetServiceServer,
	vmInstanceSchedulingService pb.InstanceSchedulingServiceClient,
	config config.Config,
	cloudAccountServiceClient pb.CloudAccountServiceClient,
	fleetAdminServiceClient pb.FleetAdminServiceClient,
	qmsClient pb.QuotaManagementPrivateServiceClient,

) (*InstanceService, error) {
	if db == nil {
		return nil, fmt.Errorf("db is required")
	}
	return &InstanceService{
		db:                          db,
		sshPublicKeyService:         sshPublicKeyService,
		instanceTypeService:         instanceTypeService,
		machineImageService:         machineImageService,
		vNetService:                 vNetService,
		vmInstanceSchedulingService: vmInstanceSchedulingService,
		sqlTransformer:              NewInstanceSqlTransformer(),
		pbConverter:                 pbconvert.NewPbConverter(),
		cfg:                         config,
		cloudAccountServiceClient:   cloudAccountServiceClient,
		fleetAdminServiceClient:     fleetAdminServiceClient,
		qmsClient:                   qmsClient,
	}, nil
}

// Validate instanceName.
// instanceName is valid when name is starting and ending with lowercase alphanumeric
// and contains lowercase alphanumeric, '-' characters and should have at most 63 characters
func validateInstanceName(name string) error {
	re := regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
	matches := re.FindAllString(name, -1)
	if matches == nil {
		return status.Error(codes.InvalidArgument, "invalid instance name")
	}
	return nil
}

func validateTopologySpreadConstraints(constraints []*pb.TopologySpreadConstraints) error {
	if len(constraints) == 0 {
		return nil
	} else if len(constraints) > 1 {
		return status.Error(codes.InvalidArgument, "a maximum of one TopologySpreadConstraint is allowed")
	}
	for _, constraint := range constraints {
		if err := utils.ValidateLabels(constraint.LabelSelector.MatchLabels); err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}
	}
	return nil
}

// This function verifies that the combined total of existing and requested instances does not surpass the quota limit for the specified instance type
//
// Parameters:
// - ctx: A context.Context instance for carrying deadlines, cancellation signals, and other request-scoped values.
// - cloudAccountId: A string representing the unique identifier of the user.
// - inputInstance: A pointer to an InstancePrivate object representing the details of the instance being requested.
// - requestedInstanceCount: An integer representing the number of new instances being requested of inputInstance type.
func (s *InstanceService) validateQuotaLimits(ctx context.Context, cloudAccountId string, instanceType string, requestedInstanceCount int) error {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceService.validateQuotaLimits").WithValues(logkeys.CloudAccountId, cloudAccountId).Start()
	defer span.End()
	log.Info("validateQuotaLimits", logkeys.CloudAccountId, cloudAccountId, logkeys.InstanceType, instanceType, logkeys.RequestedInstanceCount, requestedInstanceCount)
	if requestedInstanceCount == 0 {
		return nil
	}
	instances, err := s.Search(ctx, &pb.InstanceSearchRequest{
		Metadata: &pb.InstanceMetadataSearch{
			CloudAccountId:      cloudAccountId,
			InstanceGroupFilter: pb.SearchFilterCriteria_Any,
		},
	})
	if err != nil {
		log.Error(err, "error in search call")
		return status.Error(codes.Internal, "error in quota processing")
	}

	currentInstanceCount := 0
	for _, instance := range instances.Items {
		if instance.Spec.InstanceType == instanceType {
			currentInstanceCount++
		}
	}
	log.Info("validateQuotaLimits", logkeys.InstanceType, instanceType, logkeys.CurrentInstanceCount, currentInstanceCount)

	allowedQuota := 0
	if s.qmsClient != nil {
		log.Info("Fetching quota limit from QMS service", "ResourceType", instanceType)
		req := &pb.ServiceQuotaResourceRequestPrivate{
			ServiceName:    "compute",
			ResourceType:   instanceType,
			CloudAccountId: cloudAccountId,
		}
		resp, err := s.qmsClient.GetResourceQuotaPrivate(ctx, req)
		if err != nil {
			log.Error(err, "error in quota processing")
			return status.Errorf(codes.Internal, "error in quota processing")
		}
		if resp == nil {
			log.Info("Received an empty response from the QMS service")
			return status.Errorf(codes.Internal, "error in quota processing")
		}

		log.Info("Received ResourceQuota Response", "ServiceQuotasPrivate", resp)

		var quotaResources []*pb.ServiceQuotaResource
		if resp.CustomQuota != nil && resp.CustomQuota.ServiceResources != nil && len(resp.CustomQuota.ServiceResources) > 0 {
			quotaResources = resp.CustomQuota.ServiceResources
		} else if resp.DefaultQuota != nil && resp.DefaultQuota.ServiceResources != nil && len(resp.DefaultQuota.ServiceResources) > 0 {
			quotaResources = resp.DefaultQuota.ServiceResources
		}

		if len(quotaResources) > 0 && quotaResources[0] != nil && quotaResources[0].QuotaConfig != nil {
			allowedQuota = int(quotaResources[0].QuotaConfig.Limits)
		} else {
			log.Info("Received an empty QuotaConfig limit from the QMS service")
		}
	} else {
		log.Info("Fetching quota limit using legacy approach")
		cAccount, err := s.cloudAccountServiceClient.GetById(ctx, &pb.CloudAccountId{Id: cloudAccountId})
		if err != nil {
			log.Error(err, "Error retrieving cloud account details")
			return status.Error(codes.Internal, "error in quota processing")
		}

		accountType := GetAccountType(cAccount.GetType())
		log.Info("Cloud account details", logkeys.CloudAccount, cAccount, logkeys.CloudAccountType, accountType)

		allowedQuota = s.cfg.CloudAccountQuota.CloudAccounts[accountType].InstanceQuota[instanceType]
	}

	log.Info("Quota validation", logkeys.AllowedQuota, allowedQuota, logkeys.CurrentInstanceCount, currentInstanceCount, logkeys.RequestedInstanceCount, requestedInstanceCount)

	if allowedQuota < (currentInstanceCount + requestedInstanceCount) {
		return status.Errorf(codes.OutOfRange, "Your account has reached the maximum allowed limit for the %s instance you requested", instanceType)
	}

	return nil

}

func GetAccountType(typ pb.AccountType) string {
	switch typ {
	case pb.AccountType_ACCOUNT_TYPE_STANDARD:
		return "STANDARD"
	case pb.AccountType_ACCOUNT_TYPE_PREMIUM:
		return "PREMIUM"
	case pb.AccountType_ACCOUNT_TYPE_ENTERPRISE:
		return "ENTERPRISE"
	case pb.AccountType_ACCOUNT_TYPE_ENTERPRISE_PENDING:
		return "ENTERPRISE_PENDING"
	case pb.AccountType_ACCOUNT_TYPE_INTEL:
		return "INTEL"
	default:
		return "STANDARD"
	}
}

// Launch a new baremetal or virtual machine instance.
// Public API.
func (s *InstanceService) Create(ctx context.Context, req *pb.InstanceCreateRequest) (*pb.Instance, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceService.Create").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.Instance, error) {
		// Validate input.
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}
		if req.Spec == nil {
			return nil, status.Error(codes.InvalidArgument, "missing spec")
		}
		if len(req.Spec.UserData) > maxUserdataSizeBytes {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("cloud-init userdata exceeds maximum size of %d bytes", maxUserdataSizeBytes))
		}

		cloudAccountId := req.Metadata.CloudAccountId

		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		// Transcode InstanceSpec to InstanceSpecPrivate.
		// Unmatched fields will remain at their default.
		instanceSpecPrivate := &pb.InstanceSpecPrivate{}
		if err := s.pbConverter.Transcode(req.Spec, instanceSpecPrivate); err != nil {
			return nil, status.Errorf(codes.Internal, "unable to transcode instance spec: %v", err)
		}
		ts := timestamppb.Now()

		instance := &pb.InstancePrivate{
			Metadata: &pb.InstanceMetadataPrivate{
				CloudAccountId:    cloudAccountId,
				Name:              req.Metadata.Name,
				Labels:            req.Metadata.Labels,
				CreationTimestamp: ts,
			},
			Spec: instanceSpecPrivate,
			Status: &pb.InstanceStatusPrivate{
				Phase:    pb.InstancePhase_Provisioning,
				Message:  "Instance reconciliation has not started",
				SshProxy: &pb.SshProxyTunnelStatus{},
			},
		}

		if err := s.create(ctx, []*pb.InstancePrivate{instance}, false); err != nil {
			return nil, err
		}

		// Query database and return response.
		return s.get(ctx, cloudAccountId, "resource_id", instance.Metadata.ResourceId)
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, utils.SanitizeError(err)
}

// Launch multiple instances.
// This is a private GRPC API intended to be called by trusted IDC services such as Intel Kubernetes Services and Instance Groups.
func (s *InstanceService) CreateMultiplePrivate(ctx context.Context, req *pb.InstanceCreateMultiplePrivateRequest) (*pb.InstanceCreateMultiplePrivateResponse, error) {
	var cloudAccountId0 string
	if len(req.Instances) > 0 {
		cloudAccountId0 = req.Instances[0].Metadata.CloudAccountId
	}
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceService.CreateMultiplePrivate").WithValues(logkeys.CloudAccountId, cloudAccountId0).Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)
	instances := []*pb.InstancePrivate{}
	resp, err := func() (*pb.InstanceCreateMultiplePrivateResponse, error) {
		if len(req.Instances) == 0 {
			return nil, status.Error(codes.InvalidArgument, "at least one instance is required")
		}
		for _, reqInstance := range req.Instances {
			// Validate input.
			if reqInstance.Metadata == nil {
				return nil, status.Error(codes.InvalidArgument, "missing metadata")
			}
			if reqInstance.Spec == nil {
				return nil, status.Error(codes.InvalidArgument, "missing spec")
			}
			if err := cloudaccount.CheckValidId(reqInstance.Metadata.CloudAccountId); err != nil {
				return nil, err
			}
			ts := timestamppb.Now()

			instance := &pb.InstancePrivate{
				Metadata: &pb.InstanceMetadataPrivate{
					CloudAccountId:    reqInstance.Metadata.CloudAccountId,
					Name:              reqInstance.Metadata.Name,
					ResourceId:        reqInstance.Metadata.ResourceId,
					Labels:            reqInstance.Metadata.Labels,
					SkipQuotaCheck:    reqInstance.Metadata.SkipQuotaCheck,
					CreationTimestamp: ts,
				},
				Spec: reqInstance.Spec,
				Status: &pb.InstanceStatusPrivate{
					Phase:    pb.InstancePhase_Provisioning,
					Message:  "Instance reconciliation has not started",
					SshProxy: &pb.SshProxyTunnelStatus{},
				},
			}
			instances = append(instances, instance)
		}

		if err := s.create(ctx, instances, req.DryRun); err != nil {
			return nil, err
		}

		// Query database and return response.
		resp := &pb.InstanceCreateMultiplePrivateResponse{}
		for _, instance := range instances {
			var InstanceResp *pb.InstancePrivate
			var err error
			if !req.DryRun {
				InstanceResp, err = s.getPrivate(ctx, instance.Metadata.CloudAccountId, "resource_id", instance.Metadata.ResourceId)
				if err != nil {
					return nil, err
				}
			} else {
				InstanceResp = instance
			}
			if InstanceResp.Spec.UserData != "" {
				InstanceResp.Spec.UserData = defaultUserData
			}
			resp.Instances = append(resp.Instances, InstanceResp)
		}
		return resp, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, utils.SanitizeError(err)
}

// Launch a new baremetal or virtual machine instance.
// Similar to InstanceService.Create with the following differences:
//   - Caller can provide serviceType.
//   - Caller can provide resourceId.
//   - Caller can provide a private instanceType (not implemented).
//   - Caller can provide a custom instanceTypeSpec (not implemented).
//   - Caller can provide private network interface fields (not implemented).
//
// Private API.
func (s *InstanceService) CreatePrivate(ctx context.Context, req *pb.InstanceCreatePrivateRequest) (*pb.InstancePrivate, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceService.CreatePrivate").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
		logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.InstancePrivate, error) {
		// Validate input.
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}
		if req.Spec == nil {
			return nil, status.Error(codes.InvalidArgument, "missing spec")
		}
		if len(req.Spec.UserData) > maxUserdataSizeBytes {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("cloud-init userdata exceeds maximum size of %d bytes", maxUserdataSizeBytes))
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		instance := &pb.InstancePrivate{
			Metadata: &pb.InstanceMetadataPrivate{
				CloudAccountId:    cloudAccountId,
				Name:              req.Metadata.Name,
				ResourceId:        req.Metadata.ResourceId,
				Labels:            req.Metadata.Labels,
				SkipQuotaCheck:    req.Metadata.SkipQuotaCheck,
				CreationTimestamp: timestamppb.Now(),
			},
			Spec: req.Spec,
		}
		if err := s.create(ctx, []*pb.InstancePrivate{instance}, false); err != nil {
			return nil, err
		}

		// Query database and return response.
		return s.getPrivate(ctx, cloudAccountId, "resource_id", instance.Metadata.ResourceId)
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

func (s *InstanceService) checkQuota(ctx context.Context, cloudAccountId string, instances []*pb.InstancePrivate) error {
	//This contains a list of instances (with no SkipQuotaCheck) requested for an InstanceType.
	requestedInstanceCountMap := getInstanceCountMap(instances)
	//Check if cloudAccount can create instances for all the requested instanceTypes.
	for instType, requestedCount := range requestedInstanceCountMap {
		if err := s.validateQuotaLimits(ctx, cloudAccountId, instType, requestedCount); err != nil {
			// return quota error and fail fast.
			return err
		}
	}
	return nil
}

// Validates and sets defaults in the provided InstancePrivate object and stores it in the database.
func (s *InstanceService) create(ctx context.Context, instances []*pb.InstancePrivate, dryRun bool) error {
	instance0 := instances[0]
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceService.create").WithValues(logkeys.CloudAccountId, instance0.Metadata.CloudAccountId,
		logkeys.ResourceId, instance0.Metadata.GetResourceId()).Start()
	defer span.End()

	// Check quota before creation.
	if err := s.checkQuota(ctx, instance0.Metadata.CloudAccountId, instances); err != nil {
		return err
	}
	// TODO: Perform an additional quota validation following the instance creation but before
	// completing the transaction to address concurrent creation scenarios.

	userData := instance0.Spec.UserData
	instanceGroupSize := len(instances)
	for instanceIndex, instance := range instances {
		log.Info("Request", logkeys.Instance, instance, logkeys.InstanceGroupSize, instanceGroupSize)
		// Validate input.
		if instance.Metadata == nil {
			return status.Error(codes.InvalidArgument, "missing metadata")
		}
		if instance.Spec == nil {
			return status.Error(codes.InvalidArgument, "missing spec")
		}

		// Calculate resourceId if not provided.
		if instance.Metadata.ResourceId == "" {
			resourceId, err := uuid.NewRandom()
			if err != nil {
				return status.Errorf(codes.Internal, "error encountered while generating resourceId: %v", err)
			}
			instance.Metadata.ResourceId = resourceId.String()
		}

		// Validate resourceId.
		if _, err := uuid.Parse(instance.Metadata.ResourceId); err != nil {
			return status.Error(codes.InvalidArgument, "invalid resourceId")
		}

		// Calculate name if not provided.
		if instance.Metadata.Name == "" {
			instance.Metadata.Name = instance.Metadata.ResourceId
		}
		name := instance.Metadata.Name

		if err := validateInstanceName(name); err != nil {
			return err
		}

		if err := utils.ValidateLabels(instance.Metadata.Labels); err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}

		if err := validateTopologySpreadConstraints(instance.Spec.TopologySpreadConstraints); err != nil {
			return err
		}

		// Ensure that this instance name is not already used in the same request.
		for otherInstanceIndex, otherInstance := range instances {
			if instanceIndex != otherInstanceIndex && name == otherInstance.Metadata.Name {
				return status.Errorf(codes.AlreadyExists, "request contains more than one instance with name %s", name)
			}
		}

		// Ensure that an instance with the same name does not already exist in the database.
		// This is also enforced by the database but we want to validate it before scheduling reserves the resources.
		_, err := s.get(ctx, instance.Metadata.CloudAccountId, "name", name)
		if err == nil {
			return status.Errorf(codes.AlreadyExists, "instance with name %s already exists", name)
		}
		if status.Code(err) != codes.NotFound {
			return err
		}

		// Region must be empty or set to the region used by this Compute API Server.
		if instance.Spec.Region != "" && instance.Spec.Region != s.cfg.Region {
			return status.Errorf(codes.InvalidArgument, "region must be empty or %s", s.cfg.Region)
		}
		instance.Spec.Region = s.cfg.Region

		// Add InstanceTypeSpec.
		instanceType, err := s.instanceTypeService.Get(ctx, &pb.InstanceTypeGetRequest{
			Metadata: &pb.InstanceTypeGetRequest_Metadata{
				Name: instance.Spec.InstanceType,
			},
		})
		if err != nil {
			return status.Errorf(codes.InvalidArgument, "unable to get instance type %q", instance.Spec.InstanceType)
		}
		instance.Spec.InstanceTypeSpec = instanceType.Spec

		// Add MachineImageSpec.
		machineImage, err := s.machineImageService.Get(ctx, &pb.MachineImageGetRequest{
			Metadata: &pb.MachineImageGetRequest_Metadata{
				Name: instance.Spec.MachineImage,
			},
		})
		if err != nil {
			return status.Errorf(codes.InvalidArgument, "unable to get machine image %q", instance.Spec.MachineImage)
		}
		instance.Spec.MachineImageSpec = machineImage.Spec
		s.addMachineImageChecksum(machineImage, instance) // add machineImage checksum
		// Ensure that machine image is compatible with instance type.
		if err := ensureMachineImageCompatibility(machineImage, instanceType); err != nil {
			return err
		}
		if s.cfg.FeatureFlags.EnableMultipleFirmwareSupport {
			s.addMachineImageComponentVersions(machineImage, instance)
		}

		// Add NetworkInterfacePrivate details from VNet.
		if len(instance.Spec.Interfaces) != 1 {
			return status.Error(codes.InvalidArgument, "instance must have exactly 1 network interface")
		}

		intf0 := instance.Spec.Interfaces[0]
		// set the tenant interface name to eth0 for the BM instances as it is required by the SDN controller to set the VLAN
		if instance.Spec.InstanceTypeSpec.InstanceCategory == pb.InstanceCategory_BareMetalHost {
			if intf0.Name != tenantDefaultInterfaceName {
				intf0.Name = tenantDefaultInterfaceName
			}
		}
		// Validate VNet.
		vNet, err := s.vNetService.Get(ctx, &pb.VNetGetRequest{
			Metadata: &pb.VNetGetRequest_Metadata{
				CloudAccountId: instance.Metadata.CloudAccountId,
				NameOrId: &pb.VNetGetRequest_Metadata_Name{
					Name: intf0.VNet,
				},
			},
		})
		if err != nil {
			return status.Errorf(codes.InvalidArgument, "unable to get VNet with name %q", intf0.VNet)
		}
		if vNet.Spec.AvailabilityZone != instance.Spec.AvailabilityZone {
			return status.Errorf(codes.InvalidArgument, "VNet %q is in a different availability zone", intf0.VNet)
		}

		intf0.DnsName = fmt.Sprintf("%s.%s.%s.%s", name, instance.Metadata.CloudAccountId, instance.Spec.Region,
			s.cfg.DomainSuffix)
		// Setup nameservers
		intf0.Nameservers = s.cfg.Nameservers

		if s.cfg.StorageInterface.Enabled &&
			slices.Contains(s.cfg.StorageInterface.EnabledInstanceTypes, instanceType.Metadata.Name) {
			// check if node is single and single with separate storage interface is allowed
			if instance.Spec.InstanceGroup != "" || s.cfg.StorageInterface.EnabledOnSingles {
				// Prefix Length for storage vnet is same as the vNet Prefix length
				if err := s.addStorageVnet(ctx, instance, vNet.Spec.PrefixLength); err != nil {
					log.Error(err, "Failed to add storage vNet to the instance")
					return err
				}
			}
		}

		// Add SSH public keys.
		if err := s.populateSshPublicKeys(ctx, instance.Metadata.CloudAccountId, instance); err != nil {
			return err
		}

		// Number of instances in the instance group.
		if instance.Spec.InstanceGroupSize == 0 {
			instance.Spec.InstanceGroupSize = int32(instanceGroupSize)
		}

		// Set list of allowed compute node pools from Fleet Admin Service.
		if err := s.setAllowedComputeNodePoolsForScheduling(ctx, instance); err != nil {
			return err
		}

		// skip sending user data to the scheduler
		if instance.Spec.UserData != "" {
			instance.Spec.UserData = defaultUserData
		}
	}

	// Do not schedule instances if the Cluster id and node id is already specified, the instance operator will provision it.
	if instance0.Spec.ClusterId == "" && instance0.Spec.NodeId == "" {
		log.V(9).Info("Scheduling instances, starting with", logkeys.InstanceName, instance0.Metadata.Name, logkeys.InstanceSpec, instance0.Spec)
		// Schedule instances atomically.
		if err := s.scheduleResources(ctx, instances, dryRun); err != nil {
			log.Error(err, "scheduling failure")
			return status.Error(codes.ResourceExhausted, ResourceExhaustedMessage)
		}
	} else {
		log.Info("Skipping scheduling since ClusterId and NodeId is specified", logkeys.InstanceName, instance0.Metadata.Name)
	}

	for _, instance := range instances {
		// add back the user data
		instance.Spec.UserData = userData

		if s.useAcceleratorClustervNet(instance) {
			err := s.createAcceleratorClustervNet(ctx, instance)
			if err != nil {
				return err
			}
		}
	}

	// for a dry run, skip instance creation
	if dryRun {
		return nil
	}

	// Insert into database in a single transaction.
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return status.Errorf(codes.Unknown, "BeginTx: %v", err)
	}
	defer tx.Rollback()

	for _, instance := range instances {
		name := instance.Metadata.Name

		// Flatten instance into columns.
		flattened, err := s.sqlTransformer.Flatten(ctx, instance)
		if err != nil {
			return err
		}

		// Insert into database.
		query := fmt.Sprintf(`insert into instance (resource_id, cloud_account_id, name, %s) values ($1, $2, $3, %s)`,
			flattened.GetColumnsString(), flattened.GetInsertValuesString(4))
		args := append([]any{instance.Metadata.ResourceId, instance.Metadata.CloudAccountId, name}, flattened.Values...)
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			pgErr := &pgconn.PgError{}
			if errors.As(err, &pgErr) && pgErr.Code == common.KErrUniqueViolation {
				return status.Error(codes.AlreadyExists, "insert: instance "+name+" already exists")
			}
			return status.Errorf(codes.Unknown, "insert: %v", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return status.Errorf(codes.Unknown, "commit: %v", err)
	}
	return nil
}

// Public API.
func (s *InstanceService) Get(ctx context.Context, req *pb.InstanceGetRequest) (*pb.Instance, error) {

	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceService.Get").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
		logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.Instance, error) {
		// Validate input.
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}
		argName, arg, err := common.ResourceUniqueColumnAndValue(req.Metadata.GetResourceId(), req.Metadata.GetName())
		if err != nil {
			return nil, err
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		return s.get(ctx, cloudAccountId, argName, arg)
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, utils.SanitizeError(err)
}

func (s *InstanceService) getInstancePrivateRows(ctx context.Context, req *pb.InstanceSearchPrivateRequest) (*sql.Rows, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceService.getInstancePrivateRows").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()

	// Validate input.
	if req.Metadata == nil {
		return nil, status.Error(codes.InvalidArgument, "missing metadata")
	}

	cloudAccountId := req.Metadata.CloudAccountId
	if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
		return nil, err
	}

	if err := utils.ValidateLabels(req.Metadata.Labels); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	flattenedObject := protodb.Flattened{
		Columns: []string{"cloud_account_id", "deleted_timestamp"},
		Values:  []any{cloudAccountId, common.TimestampInfinityStr},
	}

	labels := req.Metadata.Labels
	for key, value := range labels {
		column := fmt.Sprintf("value->'metadata'->'labels'->>'%s'", key)
		flattenedObject.Add(column, value)
	}

	var instanceGroupFieldCheck string
	var whereString string
	instanceGroupFilter := req.Metadata.InstanceGroupFilter
	// The OR condition is used for filtering by instanceGroup if it is non-empty without explicitly setting instanceGroupFilter=ExactValue.
	if instanceGroupFilter == pb.SearchFilterCriteria_ExactValue || (instanceGroupFilter == pb.SearchFilterCriteria_Default && req.Metadata.InstanceGroup != "") {
		instanceGroup := req.Metadata.InstanceGroup
		if instanceGroup == "" {
			return nil, status.Error(codes.InvalidArgument, "instance group cannot be empty when instanceGroupFilter is set to exact value")
		}
		column := "value->'spec'->>'instanceGroup'"
		flattenedObject.Add(column, instanceGroup)
	} else if instanceGroupFilter == pb.SearchFilterCriteria_NonEmpty {
		// condition to return records with a non-empty value in the instanceGroup field
		instanceGroupFieldCheck = " and value->'spec'->'instanceGroup' is not null and value->'spec'->>'instanceGroup' <> ''"
	} else if instanceGroupFilter == pb.SearchFilterCriteria_Any {
		log.Info("return all instances")
	} else {
		// condition to return records that are not in any instance group and for already running instances
		instanceGroupFieldCheck = " and (value->'spec'->'instanceGroup' is null or value->'spec'->>'instanceGroup' = '')"
	}

	if instanceGroupFieldCheck != "" {
		whereString = flattenedObject.GetWhereString(1) + instanceGroupFieldCheck
	} else {
		whereString = flattenedObject.GetWhereString(1)
	}
	query := fmt.Sprintf(`
			select %s
			from   instance
			where  %s
			order by name
		`, s.sqlTransformer.ColumnsForFromRow(), whereString)

	args := flattenedObject.Values
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return rows, nil
}

// Private API.
func (s *InstanceService) SearchPrivate(ctx context.Context, req *pb.InstanceSearchPrivateRequest) (*pb.InstanceSearchPrivateResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceService.SearchPrivate").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)

	resp, err := func() (*pb.InstanceSearchPrivateResponse, error) {
		rows, err := s.getInstancePrivateRows(ctx, req)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var items []*pb.InstancePrivate
		for rows.Next() {
			item, err := s.rowToInstancePrivate(ctx, rows)
			if err != nil {
				return nil, err
			}
			logger.Info("User data before omitting", logkeys.UserData, item.Spec.UserData)
			if item.Spec.UserData != "" {
				item.Spec.UserData = defaultUserData
			}
			items = append(items, item)
		}
		resp := &pb.InstanceSearchPrivateResponse{
			Items: items,
		}
		return resp, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

// Public API.
func (s *InstanceService) Search(ctx context.Context, req *pb.InstanceSearchRequest) (*pb.InstanceSearchResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceService.Search").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.InstanceSearchResponse, error) {
		rows, err := s.getInstancePrivateRows(ctx, &pb.InstanceSearchPrivateRequest{
			Metadata: &pb.InstanceMetadataSearch{
				CloudAccountId:      req.Metadata.CloudAccountId,
				Labels:              req.Metadata.Labels,
				InstanceGroup:       req.Metadata.InstanceGroup,
				InstanceGroupFilter: req.Metadata.InstanceGroupFilter,
			},
		})
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var items []*pb.Instance
		for rows.Next() {
			item, err := s.rowToInstance(ctx, rows)
			if err != nil {
				return nil, err
			}
			logger.Info("User data before omitting", logkeys.UserData, item.Spec.UserData)
			if item.Spec.UserData != "" {
				item.Spec.UserData = defaultUserData
			}
			items = append(items, item)
		}
		resp := &pb.InstanceSearchResponse{
			Items: items,
		}
		return resp, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, utils.SanitizeError(err)
}

// Private API.
func (s *InstanceService) UpdatePrivate(ctx context.Context, req *pb.InstanceUpdatePrivateRequest) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceService.UpdatePrivate").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
		logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()

	log.Info("Request", logkeys.Request, req)
	resp, err := func() (*emptypb.Empty, error) {
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}
		if req.Spec == nil {
			return nil, status.Error(codes.InvalidArgument, "missing spec")
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		updateFunc := func(instance *pb.InstancePrivate) error {
			instance.Spec.InstanceGroupSize = req.Spec.InstanceGroupSize
			return nil
		}
		if err := s.update(ctx, cloudAccountId, req.Metadata.GetResourceId(), req.Metadata.GetName(), req.Metadata.ResourceVersion, updateFunc); err != nil {
			return nil, err
		}
		return &emptypb.Empty{}, nil
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Response", logkeys.Response, resp)
	}
	return resp, err
}

// Allow update of:
//   - Spec.SshPublicKeyNames
//   - Spec.RunStrategy
//
// Public API.
func (s *InstanceService) Update(ctx context.Context, req *pb.InstanceUpdateRequest) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceService.Update").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
		logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()

	log.Info("Request", logkeys.Request, req)
	resp, err := func() (*emptypb.Empty, error) {
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}
		if req.Spec == nil {
			return nil, status.Error(codes.InvalidArgument, "missing spec")
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		updateFunc := func(instance *pb.InstancePrivate) error {
			if instance.Spec.RunStrategy != req.Spec.RunStrategy {
				if req.Spec.RunStrategy == pb.RunStrategy_Halted {
					// if system is in 'Starting' or 'Provisioning', do not accept 'Halted' request
					if instance.Status.Phase == pb.InstancePhase_Starting || instance.Status.Phase == pb.InstancePhase_Provisioning {
						log.Error(errors.New(instance.Status.Phase.String()), "instance is not Ready.")
						return status.Errorf(codes.FailedPrecondition, "unable to update the Instance RunStrategy to %s during %s", req.Spec.RunStrategy, instance.Status.Phase)
					}
				}

				// if system is in 'Stopping', do not accept 'Always' or 'RerunOnFailure' request
				if instance.Status.Phase == pb.InstancePhase_Stopping {
					if req.Spec.RunStrategy == pb.RunStrategy_Always || req.Spec.RunStrategy == pb.RunStrategy_RerunOnFailure {
						log.Error(errors.New(instance.Status.Phase.String()), "instance has not completed Stopping.")
						return status.Errorf(codes.FailedPrecondition, "unable to update the instance Run Strategy to %s during %s", req.Spec.RunStrategy, instance.Status.Phase)
					}
				}
			}

			instance.Spec.RunStrategy = req.Spec.RunStrategy
			instance.Spec.SshPublicKeyNames = req.Spec.SshPublicKeyNames
			instance.Spec.SshPublicKeySpecs = nil

			if err := s.populateSshPublicKeys(ctx, cloudAccountId, instance); err != nil {
				return err
			}
			return nil
		}
		if err := s.update(ctx, cloudAccountId, req.Metadata.GetResourceId(), req.Metadata.GetName(), req.Metadata.ResourceVersion, updateFunc); err != nil {
			return nil, err
		}
		return &emptypb.Empty{}, nil
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Response", logkeys.Response, resp)
	}
	return resp, utils.SanitizeError(err)
}

// Allow update of:
//   - Status
//
// Private API.
func (s *InstanceService) UpdateStatus(ctx context.Context, req *pb.InstanceUpdateStatusRequest) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceService.UpdateStatus").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
		logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()

	log.Info("Request", logkeys.Request, req)
	resp, err := func() (*emptypb.Empty, error) {
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}
		if req.Status == nil {
			return nil, status.Error(codes.InvalidArgument, "missing status")
		}
		updateFunc := func(instance *pb.InstancePrivate) error {
			instance.Status = req.Status
			return nil
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		if err := s.update(ctx, cloudAccountId, req.Metadata.ResourceId, "", req.Metadata.ResourceVersion, updateFunc); err != nil {
			return nil, err
		}
		return &emptypb.Empty{}, nil
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Response", logkeys.Response, resp)
	}
	return resp, err
}

// Delete indicates a public API user's request to delete an instance. It sets the deletionTimestamp of the record's "value" field.
// Public API.
func (s *InstanceService) Delete(ctx context.Context, req *pb.InstanceDeleteRequest) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceService.Delete").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
		logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()
	log.Info("Request", logkeys.Request, req)

	resp, err := s.deleteInstance(ctx, req.Metadata, false)
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Response", logkeys.Response, resp)
	}
	return resp, utils.SanitizeError(err)
}

func (s *InstanceService) DeletePrivate(ctx context.Context, req *pb.InstanceDeletePrivateRequest) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceService.DeletePrivate").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
		logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()
	log.Info("Request", logkeys.Request, req)

	resp, err := s.deleteInstance(ctx, req.Metadata, true)
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Response", logkeys.Response, resp)
	}
	return resp, err
}

func (s *InstanceService) deleteInstance(ctx context.Context, metadata *pb.InstanceMetadataReference, allowInstanceGroupDeletion bool) (*emptypb.Empty, error) {
	if metadata == nil {
		return nil, status.Error(codes.InvalidArgument, "missing metadata")
	}

	argName, arg, err := common.ResourceUniqueColumnAndValue(metadata.GetResourceId(), metadata.GetName())
	if err != nil {
		return nil, err
	}

	cloudAccountId := metadata.CloudAccountId
	if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
		return nil, err
	}

	instance, err := s.get(ctx, cloudAccountId, argName, arg)
	if err != nil {
		return nil, err
	}

	if !allowInstanceGroupDeletion && instance.Spec.InstanceGroup != "" {
		return nil, status.Error(codes.InvalidArgument, "the instance cannot be deleted because it is a member of an instance group")
	}

	if err := s.DeleteInstanceInternal(ctx, cloudAccountId, metadata.GetResourceId(), metadata.GetName(), metadata.ResourceVersion); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *InstanceService) DeleteInstanceInternal(ctx context.Context, cloudAccountId string, resourceId string, name string, resourceVersion string) error {
	updateFunc := func(instance *pb.InstancePrivate) error {
		if instance.Metadata.DeletionTimestamp == nil {
			instance.Metadata.DeletionTimestamp = timestamppb.Now()
		}
		return nil
	}

	if err := s.update(ctx, cloudAccountId, resourceId, name, resourceVersion, updateFunc); err != nil {
		return err
	}
	return nil
}

func (s *InstanceService) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("InstanceService.Ping")
	log.Info("Ping")
	return &emptypb.Empty{}, nil
}

func (s *InstanceService) PingPrivate(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("InstanceService.PingPrivate")
	log.Info("PingPrivate")
	return &emptypb.Empty{}, nil
}

// RemoveFinalizer marks a record as hidden from users and controllers. It sets the record's deleted_timestamp.
// Private API.
func (s *InstanceService) RemoveFinalizer(ctx context.Context, req *pb.InstanceRemoveFinalizerRequest) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceService.RemoveFinalizer").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId,
		logkeys.ResourceId, req.Metadata.GetResourceId()).Start()
	defer span.End()

	log.Info("Request", logkeys.Request, req)
	resp, err := func() (*emptypb.Empty, error) {
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}
		query := `
			update instance
			set    deleted_timestamp = current_timestamp,
			       resource_version = nextval('instance_resource_version_seq')
			where  cloud_account_id = $1
			  and  resource_id = $2
			  and  deleted_timestamp = $3
		`
		if _, err := s.db.ExecContext(ctx, query, req.Metadata.CloudAccountId, req.Metadata.ResourceId, common.TimestampInfinityStr); err != nil {
			return nil, err
		}
		return &emptypb.Empty{}, nil
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Response", logkeys.Response, resp)
	}
	return resp, err
}

// Caller must close the returned sql.Rows.
func (s *InstanceService) selectInstance(ctx context.Context, cloudAccountId string, argName string, arg interface{}) (*sql.Rows, error) {
	query := fmt.Sprintf(`
		select %s
		from   instance
		where  cloud_account_id = $1
		  and  %s = $2
		  and  deleted_timestamp = $3
	`, s.sqlTransformer.ColumnsForFromRow(), argName)
	rows, err := s.db.QueryContext(ctx, query, cloudAccountId, arg, common.TimestampInfinityStr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "selectInstance: %v", err)
	}
	if !rows.Next() {
		defer rows.Close()
		return nil, status.Error(codes.NotFound, "resource not found")
	}
	return rows, nil
}

func (s *InstanceService) get(ctx context.Context, cloudAccountId string, argName string, arg interface{}) (*pb.Instance, error) {
	rows, err := s.selectInstance(ctx, cloudAccountId, argName, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.rowToInstance(ctx, rows)
}

func (s *InstanceService) getPrivate(ctx context.Context, cloudAccountId string, argName string, arg interface{}) (*pb.InstancePrivate, error) {
	rows, err := s.selectInstance(ctx, cloudAccountId, argName, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.sqlTransformer.FromRow(ctx, rows)
}

// For any elements missing in SshPublicKeySpecs, retrieve the SSH public key from the SSH public key service.
func (s *InstanceService) populateSshPublicKeys(ctx context.Context, cloudAccountId string, instance *pb.InstancePrivate) error {
	instance.Spec.SshPublicKeySpecs = make([]*pb.SshPublicKeySpec, len(instance.Spec.SshPublicKeyNames))
	for i := range instance.Spec.SshPublicKeyNames {
		if instance.Spec.SshPublicKeySpecs[i] == nil {
			sshPublicKey, err := s.getSshPublicKey(ctx, cloudAccountId, instance.Spec.SshPublicKeyNames[i])
			if err != nil {
				return err
			}
			instance.Spec.SshPublicKeySpecs[i] = &pb.SshPublicKeySpec{
				SshPublicKey: sshPublicKey,
			}
		}
	}
	return nil
}

func (s *InstanceService) getSshPublicKey(ctx context.Context, cloudAccountId string, name string) (string, error) {
	sshPublicKey, err := s.sshPublicKeyService.Get(ctx, &pb.SshPublicKeyGetRequest{
		Metadata: &pb.ResourceMetadataReference{
			CloudAccountId: cloudAccountId,
			NameOrId: &pb.ResourceMetadataReference_Name{
				Name: name,
			},
		},
	})
	if err != nil {
		return "", err
	}
	return sshPublicKey.Spec.SshPublicKey, nil
}

func (s *InstanceService) addStorageVnet(ctx context.Context, instance *pb.InstancePrivate, prefixLength int32) error {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceService.addStorageVnet").Start()
	defer span.End()

	storageVnetName := utils.GenerateStorageVnetName(instance.Spec.AvailabilityZone)
	// create a vNet if it does not exist for the AZ (this API is idempotent)
	storageVnet, err := s.createvNet(ctx, storageVnetName, instance, prefixLength)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to create storage vnet due to %s", err.Error())
	}
	log.Info("storage vNet info", logkeys.VNet, storageVnet)
	// append storage vnet to the interfaces
	instance.Spec.Interfaces = append(instance.Spec.Interfaces, &pb.NetworkInterfacePrivate{Name: "storage0", VNet: storageVnet.Metadata.Name,
		DnsName: fmt.Sprintf("%s.%s.storage0.%s.%s", instance.Metadata.Name, instance.Metadata.CloudAccountId, instance.Spec.Region, s.cfg.DomainSuffix)})
	log.Info("updated instance interfaces for storage", logkeys.Interfaces, instance.Spec.Interfaces)
	return nil
}

func (s *InstanceService) createAcceleratorClustervNet(ctx context.Context, instance *pb.InstancePrivate) error {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceService.createAcceleratorClustervNet").WithValues(logkeys.ResourceId, instance.Metadata.GetResourceId(),
		logkeys.InstanceName, instance.Metadata.Name).Start()
	defer span.End()

	acceleratorInterfaceName := ""
	acceleratorvNetName := ""

	if instance.Spec.NetworkMode == bmenroll.NetworkModeXBX {
		if instance.Spec.InstanceGroup == "" {
			return status.Error(codes.InvalidArgument, "instance group must be set to use BGP network")
		}
		acceleratorvNetName = utils.GenerateBGPClusterVNetName(instance.Spec.SuperComputeGroupId, instance.Spec.InstanceGroup)
		acceleratorInterfaceName = bgpClusterInterfaceName
	} else {
		acceleratorvNetName = utils.GenerateAcceleratorVNetName(instance.Metadata.Name, instance.Spec.InstanceGroup, instance.Spec.ClusterGroupId)
		acceleratorInterfaceName = acceleratorClusterInterfaceName
	}

	// Create vNet for the instance group.
	// As workaround for https://internal-placeholder.com/browse/TWC4727-823, only create vnet for first instance.
	var acceleratorvNet *pb.VNet
	var err error

	acceleratorvNet, err = s.vNetService.Get(ctx, &pb.VNetGetRequest{
		Metadata: &pb.VNetGetRequest_Metadata{
			CloudAccountId: instance.Metadata.CloudAccountId,
			NameOrId: &pb.VNetGetRequest_Metadata_Name{
				Name: acceleratorvNetName,
			},
		},
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			log.Info("vNet does not exist, create one", logkeys.VNetName, acceleratorvNetName)
			acceleratorvNet, err = s.createvNet(ctx, acceleratorvNetName, instance, acceleratorClusterNetworkPrefixLength)
			if err != nil {
				return status.Error(codes.InvalidArgument, err.Error())
			}

		} else { // failed to get vNet
			return status.Error(codes.InvalidArgument, err.Error())
		}
	}

	log.Info("accelerator cluster vNet info", logkeys.VNet, acceleratorvNet)
	// append accelerator cluster vNet to the interfaces
	domainSuffix := s.cfg.DomainSuffix
	instance.Spec.Interfaces = append(instance.Spec.Interfaces, &pb.NetworkInterfacePrivate{Name: acceleratorInterfaceName, VNet: acceleratorvNet.Metadata.Name,
		DnsName: fmt.Sprintf("accelerator.%s.%s.%s.%s", instance.Metadata.Name, instance.Metadata.CloudAccountId, instance.Spec.Region, domainSuffix)})
	log.Info("updated interfaces", logkeys.Interfaces, instance.Spec.Interfaces)
	return nil
}

func (s *InstanceService) createvNet(ctx context.Context, name string, instance *pb.InstancePrivate, prefixLength int32) (*pb.VNet, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceService.createvNet").WithValues(logkeys.CloudAccountId, instance.Metadata.CloudAccountId,
		logkeys.ResourceId, instance.Metadata.GetResourceId()).Start()
	defer span.End()

	log.Info("creating vNet", logkeys.VNetName, name)
	var vNet *pb.VNet
	vNet, err := s.vNetService.Put(ctx, &pb.VNetPutRequest{Metadata: &pb.VNetPutRequest_Metadata{CloudAccountId: instance.Metadata.CloudAccountId, Name: name},
		Spec: &pb.VNetSpec{Region: s.cfg.Region, AvailabilityZone: instance.Spec.AvailabilityZone, PrefixLength: prefixLength}})

	if err != nil {
		return nil, fmt.Errorf("failed to create vNet %s", name)
	}
	log.Info("created vNet", logkeys.VNetName, name)
	return vNet, nil
}

// Set instance.Spec.ComputeNodePools to the intersection of requested pools and allowed pools.
func (s *InstanceService) setAllowedComputeNodePoolsForScheduling(ctx context.Context, instance *pb.InstancePrivate) error {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceService.setAllowedComputeNodePoolsForScheduling").WithValues(logkeys.CloudAccountId, instance.Metadata.CloudAccountId,
		logkeys.ResourceId, instance.Metadata.GetResourceId()).Start()
	defer span.End()

	if !s.cfg.FeatureFlags.EnableComputeNodePoolsForScheduling || s.fleetAdminServiceClient == nil {
		return nil
	}

	req := &pb.SearchComputeNodePoolsForInstanceSchedulingRequest{
		CloudAccountId: instance.Metadata.CloudAccountId,
	}
	log.Info("Getting allowed compute node pools for instance scheduling from Fleet Admin Service", logkeys.Request, req)
	resp, err := s.fleetAdminServiceClient.SearchComputeNodePoolsForInstanceScheduling(ctx, req)
	if err != nil {
		return status.Errorf(codes.Internal, "error in getting allowed compute node pools from FleetAdmin Service: %v", err)
	}
	log.Info("Allowed compute node pools", logkeys.Request, req, logkeys.Response, resp)

	if len(resp.ComputeNodePools) == 0 {
		return status.Errorf(codes.PermissionDenied, "your Cloud Account is not authorized to create instances in any compute node pool in this region")
	}

	var allowedPools []string
	for _, pool := range resp.ComputeNodePools {
		allowedPools = append(allowedPools, pool.PoolId)
	}

	if len(instance.Spec.ComputeNodePools) == 0 {
		// Schedule on any allowed pool.
		instance.Spec.ComputeNodePools = allowedPools
		log.Info("Compute node pools set to all allowed pools", logkeys.ComputeNodePools, instance.Spec.ComputeNodePools)
	} else {
		// User requested to filter the set of pools to schedule on.
		instance.Spec.ComputeNodePools = utils.Intersect(allowedPools, instance.Spec.ComputeNodePools)
		log.Info("Compute node pools filtered by user", logkeys.ComputeNodePools, instance.Spec.ComputeNodePools)
	}

	return nil
}

// Schedule resources (1 or more instances).
// This sets each instance's ClusterId and NodeId.
func (s *InstanceService) scheduleResources(ctx context.Context, instances []*pb.InstancePrivate, dryRun bool) error {
	instance0 := instances[0]
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceService.scheduleResources").WithValues(logkeys.CloudAccountId, instance0.Metadata.CloudAccountId,
		logkeys.ResourceId, instance0.Metadata.GetResourceId()).Start()
	defer span.End()

	instanceSchedulingServiceClient := s.vmInstanceSchedulingService

	req := &pb.ScheduleRequest{
		Instances: instances,
		DryRun:    dryRun,
	}
	log.Info("Scheduling resources", logkeys.Request, req)
	isRetryable := func(err error) bool {
		return true
	}
	var resp *pb.ScheduleResponse
	backoff := wait.Backoff{
		Steps:    9,
		Duration: 10 * time.Millisecond,
		Factor:   2.0,
		Jitter:   0.1,
	}
	err := retry.OnError(backoff, isRetryable, func() error {
		var err error
		resp, err = instanceSchedulingServiceClient.Schedule(ctx, req)
		return err
	})
	if err != nil {
		return fmt.Errorf("InstanceService.scheduleResources: instanceSchedulingServiceClient.Schedule: %w", err)
	}
	log.Info("Scheduled resources", logkeys.Request, req, logkeys.Response, resp)
	if len(resp.InstanceResults) != len(instances) {
		return fmt.Errorf("InstanceService.scheduleResources: scheduling service did not return the expected number of results")
	}
	for i, result := range resp.InstanceResults {
		if result.ClusterId == "" || result.NodeId == "" {
			return fmt.Errorf("InstanceService.scheduleResources: scheduling service did not return ClusterId or NodeId")
		}
		if instances[i].Spec.InstanceTypeSpec.InstanceCategory == pb.InstanceCategory_BareMetalHost &&
			instances[i].Spec.InstanceGroup != "" &&
			result.GroupId == "" {
			return fmt.Errorf("InstanceService.scheduleResources: scheduling service did not return GroupId")
		}
		instances[i].Spec.ClusterId = result.ClusterId
		instances[i].Spec.NodeId = result.NodeId
		instances[i].Spec.Partition = result.Partition
		instances[i].Spec.ClusterGroupId = result.GroupId
		instances[i].Spec.ComputeNodePools = result.ComputeNodePools
		instances[i].Spec.NetworkMode = result.NetworkMode
		instances[i].Spec.SuperComputeGroupId = result.SuperComputeGroupId
	}
	return nil
}

// Update an instance record using the user-provided updateFunc to update the instance.
// This uses optimistic concurrency control to ensure that the record has not been updated between the select and update.
// Additionally, if the caller provides a resource version, optimistic concurrency control can be extended to
// previous get or search calls.
func (s *InstanceService) update(
	ctx context.Context,
	cloudAccountId string,
	resourceId string,
	name string,
	resourceVersion string,
	updateFunc func(*pb.InstancePrivate) error) error {
	argName, arg, err := common.ResourceUniqueColumnAndValue(resourceId, name)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}
	query := fmt.Sprintf(`
		select %s
		from   instance
		where  cloud_account_id = $1
			and  %s = $2
			and  deleted_timestamp = $3
	`, s.sqlTransformer.ColumnsForFromRow(), argName)

	// Retry on conflict if caller did not provide resourceVersion.
	isRetryable := func(err error) bool {
		return resourceVersion == "" && status.Code(err) == codes.FailedPrecondition
	}

	err = retry.OnError(retry.DefaultRetry, isRetryable, func() error {
		rows, err := s.db.QueryContext(ctx, query, cloudAccountId, arg, common.TimestampInfinityStr)
		if err != nil {
			return err
		}
		defer rows.Close()
		if !rows.Next() {
			return status.Error(codes.NotFound, "resource not found")
		}
		instance, err := s.sqlTransformer.FromRow(ctx, rows)
		if err != nil {
			return err
		}
		metadata := instance.Metadata
		// If resource version was provided, ensure that stored version matches.
		if resourceVersion != "" && resourceVersion != metadata.ResourceVersion {
			return status.Error(codes.FailedPrecondition, "stored resource version does not match requested resource version")
		}

		// Update Instance object.
		if err := updateFunc(instance); err != nil {
			return err
		}

		// Flatten instance into columns.
		flattened, err := s.sqlTransformer.Flatten(ctx, instance)
		if err != nil {
			return err
		}
		args := append([]any{metadata.CloudAccountId, metadata.ResourceId, metadata.ResourceVersion}, flattened.Values...)

		// Update database.
		updateQuery := fmt.Sprintf(`
		update instance
		set    resource_version = nextval('instance_resource_version_seq'),
			   %s
		where  cloud_account_id = $1
		and    resource_id = $2
		and    resource_version = $3
		`, flattened.GetUpdateSetString(4))
		sqlResult, err := s.db.ExecContext(ctx, updateQuery, args...)
		if err != nil {
			return err
		}
		rowsAffected, err := sqlResult.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected < 1 {
			return status.Error(codes.FailedPrecondition, "no records updated; possible update conflict")
		}

		return nil
	})
	if err != nil {
		st, _ := status.FromError(err)
		return status.Error(st.Code(), "update: "+err.Error())
	}
	return err
}

// Read a database row into a public Instance. Used for public APIs.
func (s *InstanceService) rowToInstance(ctx context.Context, rows *sql.Rows) (*pb.Instance, error) {
	log := log.FromContext(ctx).WithName("InstanceService.rowToInstance")
	instancePrivate, err := s.sqlTransformer.FromRow(ctx, rows)
	if err != nil {
		return nil, err
	}
	instance := &pb.Instance{}
	if err := s.pbConverter.Transcode(instancePrivate, instance); err != nil {
		return nil, err
	}
	log.V(9).Info("Read from database", logkeys.Instance, instance)
	return instance, nil
}

// Read a database row into a Private Instance. Used for private APIs.
func (s *InstanceService) rowToInstancePrivate(ctx context.Context, rows *sql.Rows) (*pb.InstancePrivate, error) {
	log := log.FromContext(ctx).WithName("InstanceService.rowToInstancePrivate")
	instance, err := s.sqlTransformer.FromRow(ctx, rows)
	if err != nil {
		return nil, fmt.Errorf("rowToInstancePrivate: %w", err)
	}
	log.V(9).Info("Read from database", logkeys.Instance, instance)
	return instance, nil
}

func ensureMachineImageCompatibility(machineImage *pb.MachineImage, instanceType *pb.InstanceType) error {
	if len(machineImage.Spec.InstanceCategories) > 0 {
		found := false
		for _, c := range machineImage.Spec.InstanceCategories {
			if instanceType.Spec.InstanceCategory == c {
				found = true
				break
			}
		}
		if !found {
			return status.Errorf(codes.InvalidArgument, "machine image %s is incompatible with category of instance type %s", machineImage.Metadata.Name, instanceType.Metadata.Name)
		}
	}
	if len(machineImage.Spec.InstanceTypes) > 0 {
		found := false
		for _, t := range machineImage.Spec.InstanceTypes {
			if instanceType.Spec.Name == t {
				found = true
				break
			}
		}
		if !found {
			return status.Errorf(codes.InvalidArgument, "machine image %s is incompatible with instance type %s", machineImage.Metadata.Name, instanceType.Metadata.Name)
		}
	}
	return nil
}

// Calculates and returns a map of instance types to their respective counts.
func getInstanceCountMap(instances []*pb.InstancePrivate) map[string]int {
	instanceCounts := make(map[string]int)
	for _, instance := range instances {
		if !instance.Metadata.SkipQuotaCheck { // do not add instance count if skip quota check is true
			instanceCounts[instance.Spec.InstanceType]++
		}
	}
	return instanceCounts
}

func (s *InstanceService) useAcceleratorClustervNet(instance *pb.InstancePrivate) bool {
	if instance.Spec.ClusterGroupId == "" {
		return false
	}
	instanceType := instance.Spec.InstanceType
	networkMode := instance.Spec.NetworkMode
	if networkMode == bmenroll.NetworkModeIgnore {
		// network mode set to ignore hence skipping accelarator Cluster vnet creation.
		return false
	}
	if s.cfg.AcceleratorInterface.EnableStaticBGP && networkMode == bmenroll.NetworkModeXBX {
		return false
	}
	return s.cfg.AcceleratorInterface.IsEnabledForInstanceType(instanceType)
}

func getMachineImageComponentVersions(machineImage *pb.MachineImage, componentType string) []string {
	versions := []string{}
	for _, component := range machineImage.Spec.Components {
		if component.Type == componentType {
			versions = append(versions, component.Version)
		}
	}
	return versions
}

func (s *InstanceService) addMachineImageComponentVersions(machineImage *pb.MachineImage, instance *pb.InstancePrivate) {
	if instance.Metadata.Labels == nil {
		instance.Metadata.Labels = make(map[string]string)
	}
	if s.cfg.AcceleratorInterface.IsEnabledForInstanceType(instance.Spec.InstanceType) {
		// set the firmware versions to ensure compatibility with the machine image.
		if versions := getMachineImageComponentVersions(machineImage, "Firmware kit"); len(versions) > 0 {
			instance.Metadata.Labels[bmenroll.FWVersionLabel] = strings.Join(versions, "_")
		}
	}
}

func (s *InstanceService) addMachineImageChecksum(machineImage *pb.MachineImage, instance *pb.InstancePrivate) {
	instance.Spec.MachineImageSpec.Md5Sum = machineImage.Spec.Md5Sum
	instance.Spec.MachineImageSpec.Sha256Sum = machineImage.Spec.Sha256Sum
	instance.Spec.MachineImageSpec.Sha512Sum = machineImage.Spec.Sha512Sum
}
