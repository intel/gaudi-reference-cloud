// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package vnet

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/pbconvert"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/protodb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/emptypb"
)

type VNetService struct {
	pb.UnimplementedVNetServiceServer
	pb.UnimplementedVNetPrivateServiceServer
	db                       *sql.DB
	protoJsonTable           *protodb.ProtoJsonTable
	ipResourceManagerService pb.IpResourceManagerServiceServer
	// objectStorageServicePrivate may be nil
	objectStorageServicePrivate pb.ObjectStorageServicePrivateClient
	pbConverter                 *pbconvert.PbConverter
	config                      VNetServiceConfig
}

type VNetServiceConfig struct {
	// The pattern for creating SubnetConsumerId.
	// VNets that have the same SubnetConsumerId will share a subnet.
	// Sharing subnets should only occur in development environments.
	// The following values will be replaced in the string:
	// 		{CloudAccountId}
	//		{Name}            (VNet name)
	//		{ResourceId}      (VNet resource ID)
	//		{Region}
	// 		{AvailabilityZone}
	SubnetConsumerIdPattern string `koanf:"subnetConsumerIdPattern"`
}

const defaultSubnetConsumerIdPattern = "{ResourceId}.{CloudAccountId}.vnet"

func NewVNetService(db *sql.DB, ipResourceManagerService pb.IpResourceManagerServiceServer, objectStorageServicePrivate pb.ObjectStorageServicePrivateClient, config VNetServiceConfig) (*VNetService, error) {
	if db == nil {
		panic("db is required")
	}
	if config.SubnetConsumerIdPattern == "" {
		config.SubnetConsumerIdPattern = defaultSubnetConsumerIdPattern
	}
	return &VNetService{
		db:                          db,
		protoJsonTable:              NewVNetProtoJsonTable(db),
		ipResourceManagerService:    ipResourceManagerService,
		objectStorageServicePrivate: objectStorageServicePrivate,
		pbConverter:                 pbconvert.NewPbConverter(),
		config:                      config,
	}, nil
}

func NewVNetProtoJsonTable(db *sql.DB) *protodb.ProtoJsonTable {
	return &protodb.ProtoJsonTable{
		Db:                  db,
		TableName:           "vnet",
		KeyColumns:          []string{"cloud_account_id", "resource_id"},
		SecondaryKeyColumns: []string{"cloud_account_id", "name"},
		JsonDocumentColumn:  "value",
		EmptyMessage:        &pb.VNetPrivate{},
		ConflictTarget:      "cloud_account_id, name",
		GetKeyValuesFunc: func(m proto.Message) ([]any, error) {
			var cloudaccountId string
			var resourceId string
			if msg, ok := m.(*pb.VNetPrivate); ok {
				cloudaccountId = msg.Metadata.CloudAccountId
				resourceId = msg.Metadata.ResourceId
			} else if msg, ok := m.(*pb.VNetGetRequest); ok {
				cloudaccountId = msg.Metadata.CloudAccountId
				resourceId = msg.Metadata.GetResourceId()
			} else if msg, ok := m.(*pb.VNetDeleteRequest); ok {
				cloudaccountId = msg.Metadata.CloudAccountId
				resourceId = msg.Metadata.GetResourceId()
			} else {
				return nil, fmt.Errorf("unsupported type")
			}
			return []any{cloudaccountId, resourceId}, nil
		},
		GetSecondaryKeyValuesFunc: func(m proto.Message) ([]any, error) {
			var cloudaccountId string
			var name string
			if msg, ok := m.(*pb.VNetPrivate); ok {
				cloudaccountId = msg.Metadata.CloudAccountId
				name = msg.Metadata.Name
			} else if msg, ok := m.(*pb.VNetGetRequest); ok {
				cloudaccountId = msg.Metadata.CloudAccountId
				name = msg.Metadata.GetName()
			} else if msg, ok := m.(*pb.VNetDeleteRequest); ok {
				cloudaccountId = msg.Metadata.CloudAccountId
				name = msg.Metadata.GetName()
			} else {
				return nil, fmt.Errorf("unsupported type")
			}
			return []any{cloudaccountId, name}, nil
		},
		SearchFilterFunc: func(m proto.Message) (protodb.Flattened, error) {
			var cloudAccountId string
			flattened := protodb.Flattened{}
			if msg, ok := m.(*pb.VNetSearchRequest); ok {
				cloudAccountId = msg.Metadata.CloudAccountId
			} else {
				return flattened, fmt.Errorf("unsupported type")
			}
			flattened.Add("cloud_account_id", cloudAccountId)
			return flattened, nil
		},
	}
}

func (s *VNetService) Put(ctx context.Context, req *pb.VNetPutRequest) (*pb.VNet, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("VNetService.Put").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.VNet, error) {
		// Validate input
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}
		if err := cloudaccount.CheckValidId(req.Metadata.CloudAccountId); err != nil {
			return nil, err
		}
		if req.Metadata.Name == "" {
			return nil, status.Error(codes.InvalidArgument, "missing metadata.name")
		}
		if req.Spec == nil {
			return nil, status.Error(codes.InvalidArgument, "missing spec")
		}
		if req.Spec.Region == "" {
			return nil, status.Error(codes.InvalidArgument, "missing spec.region")
		}
		if req.Spec.AvailabilityZone == "" {
			return nil, status.Error(codes.InvalidArgument, "missing spec.availabilityZone")
		}
		if req.Spec.PrefixLength < 1 || req.Spec.PrefixLength > 32 {
			return nil, status.Error(codes.InvalidArgument, "spec.prefixLength should be between 1 and 32 ")
		}

		vNetP, err := s.protoJsonTable.GetBySecondaryKey(ctx, &pb.VNetGetRequest{
			Metadata: &pb.VNetGetRequest_Metadata{
				CloudAccountId: req.Metadata.CloudAccountId,
				NameOrId: &pb.VNetGetRequest_Metadata_Name{
					Name: req.Metadata.Name,
				},
			},
		})

		if err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
				logger.Info("vNet not found for this account, creating now", logkeys.VNetName, req.Metadata.Name, logkeys.CloudAccountId, req.Metadata.CloudAccountId)
			} else {
				return nil, err
			}
		}

		var resourceId uuid.UUID
		if vNetP != nil {
			logger.Info("vnet already exists for this account", logkeys.VNetName, req.Metadata.Name, logkeys.CloudAccountId, req.Metadata.CloudAccountId)
			vNetPri := vNetP.(*pb.VNetPrivate)
			resourceId, err = uuid.Parse(vNetPri.Metadata.GetResourceId())
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "invalid resourceId")
			}
		} else {
			// Calculate resourceId.
			resourceId, err = uuid.NewRandom()
			if err != nil {
				return nil, err
			}
		}
		resourceIdStr := resourceId.String()

		vNetPrivate := &pb.VNetPrivate{
			Metadata: &pb.VNetPrivate_Metadata{
				Name:           req.Metadata.Name,
				CloudAccountId: req.Metadata.CloudAccountId,
				ResourceId:     resourceIdStr,
			},
			Spec: &pb.VNetSpecPrivate{
				Region:           req.Spec.Region,
				AvailabilityZone: req.Spec.AvailabilityZone,
				PrefixLength:     req.Spec.PrefixLength,
			},
		}

		if err := s.protoJsonTable.Put(ctx, vNetPrivate); err != nil {
			return nil, err
		}
		resp := &pb.VNet{
			Metadata: &pb.VNet_Metadata{
				Name:           req.Metadata.Name,
				CloudAccountId: req.Metadata.CloudAccountId,
				ResourceId:     resourceIdStr,
			},
			Spec: req.Spec,
		}
		return resp, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, utils.SanitizeError(err)
}

func (s *VNetService) Delete(ctx context.Context, req *pb.VNetDeleteRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("VNetService.Delete").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*emptypb.Empty, error) {
		// Validate input.
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}
		if err := cloudaccount.CheckValidId(req.Metadata.CloudAccountId); err != nil {
			return nil, err
		}

		// Get VNet record from DB.
		vNetPrivate, err := s.getPrivate(ctx, req.Metadata.CloudAccountId, req.Metadata.GetResourceId(), req.Metadata.GetName())
		if err != nil {
			return nil, err
		}

		// Release the subnet. Ignore NotFound error. This will fail if any addresses are consumed.
		_, err = s.releaseSubnet(ctx, vNetPrivate)
		if err != nil && status.Code(err) != codes.NotFound {
			return nil, err
		}

		// Delete VNet record from DB.
		err = s.protoJsonTable.Delete(ctx, &pb.VNetDeleteRequest{
			Metadata: &pb.VNetDeleteRequest_Metadata{
				CloudAccountId: vNetPrivate.Metadata.CloudAccountId,
				NameOrId: &pb.VNetDeleteRequest_Metadata_ResourceId{
					ResourceId: vNetPrivate.Metadata.ResourceId,
				},
			},
		})
		if err != nil {
			return nil, err
		}

		return &emptypb.Empty{}, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, utils.SanitizeError(err)
}

func (s *VNetService) Get(ctx context.Context, req *pb.VNetGetRequest) (*pb.VNet, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("VNetService.Get").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.VNet, error) {
		// Validate input.
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}
		if err := cloudaccount.CheckValidId(req.Metadata.CloudAccountId); err != nil {
			return nil, err
		}

		vNetPrivate, err := s.getPrivate(ctx, req.Metadata.CloudAccountId, req.Metadata.GetResourceId(), req.Metadata.GetName())
		if err != nil {
			return nil, err
		}
		vNet := &pb.VNet{}
		if err := s.pbConverter.Transcode(vNetPrivate, vNet); err != nil {
			return nil, err
		}
		return vNet, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, utils.SanitizeError(err)
}

func (s *VNetService) getPrivate(ctx context.Context, cloudAccountId string, resourceId string, name string) (*pb.VNetPrivate, error) {
	// Validate input.
	if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
		return nil, err
	}

	var vNetPrivate protoreflect.ProtoMessage
	var err error
	if resourceId != "" {
		vNetPrivate, err = s.protoJsonTable.Get(ctx, &pb.VNetGetRequest{
			Metadata: &pb.VNetGetRequest_Metadata{
				CloudAccountId: cloudAccountId,
				NameOrId: &pb.VNetGetRequest_Metadata_ResourceId{
					ResourceId: resourceId,
				},
			},
		})
		if err != nil {
			return nil, err
		}

	} else if name != "" {
		vNetPrivate, err = s.protoJsonTable.GetBySecondaryKey(ctx, &pb.VNetGetRequest{
			Metadata: &pb.VNetGetRequest_Metadata{
				CloudAccountId: cloudAccountId,
				NameOrId: &pb.VNetGetRequest_Metadata_Name{
					Name: name,
				},
			},
		})
		if err != nil {
			return nil, err
		}
	} else {
		return nil, status.Error(codes.InvalidArgument, "missing both metadata.name and metadata.resourceId. Need at least one of them")
	}
	return vNetPrivate.(*pb.VNetPrivate), nil
}

func (s *VNetService) Search(ctx context.Context, req *pb.VNetSearchRequest) (*pb.VNetSearchResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("VNetService.Search").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()

	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.VNetSearchResponse, error) {
		items := make([]*pb.VNet, 0)
		handlerFunc := func(m proto.Message) error {
			vNetPrivate := m.(*pb.VNetPrivate)
			vNet := &pb.VNet{}
			if err := s.pbConverter.Transcode(vNetPrivate, vNet); err != nil {
				return err
			}
			items = append(items, vNet)
			return nil
		}
		if err := s.protoJsonTable.Search(ctx, req, handlerFunc); err != nil {
			return nil, err
		}
		resp := &pb.VNetSearchResponse{
			Items: items,
		}
		return resp, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, utils.SanitizeError(err)
}

func (s *VNetService) SearchStream(req *pb.VNetSearchRequest, svc pb.VNetService_SearchStreamServer) error {
	ctx := svc.Context()
	logger := log.FromContext(ctx).WithName("VNetService.SearchStream")
	logger.Info("BEGIN", logkeys.Request, req)
	defer logger.Info("END")
	err := func() error {
		handlerFunc := func(m proto.Message) error {
			vNetPrivate := m.(*pb.VNetPrivate)
			vNet := &pb.VNet{}
			if err := s.pbConverter.Transcode(vNetPrivate, vNet); err != nil {
				return err
			}
			return svc.Send(vNet)
		}
		if err := s.protoJsonTable.Search(ctx, req, handlerFunc); err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		logger.Error(err, logkeys.Error, logkeys.Request, req)
	}
	return utils.SanitizeError(err)
}

func (s *VNetService) ReserveSubnet(ctx context.Context, req *pb.VNetReserveSubnetRequest) (*pb.VNetPrivate, error) {
	logger := log.FromContext(ctx).WithName("VNetService.ReserveSubnet")
	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.VNetPrivate, error) {
		// Get VNet record from DB.
		vNetPrivate, err := s.getPrivate(ctx, req.VNetReference.CloudAccountId, "", req.VNetReference.Name)
		if err != nil {
			return nil, err
		}
		subnetConsumerId := s.subnetConsumerIdForVNet(vNetPrivate)

		// Set prefixLengthHint to the minimum of vNetPrivate.Spec.PrefixLength and req.MaximumPrefixLength
		prefixLengthHint := vNetPrivate.Spec.PrefixLength
		if 0 < req.MaximumPrefixLength && req.MaximumPrefixLength < vNetPrivate.Spec.PrefixLength {
			logger.Info("using prefix length from VNetReserveSubnetRequest", logkeys.MaximumPrefixLength, req.MaximumPrefixLength, logkeys.SpecPrefixLength, vNetPrivate.Spec.PrefixLength)
			prefixLengthHint = req.MaximumPrefixLength
		}

		subnet, err := s.ipResourceManagerService.ReserveSubnet(ctx, &pb.ReserveSubnetRequest{
			SubnetReference: &pb.SubnetReference{
				SubnetConsumerId: subnetConsumerId,
			},
			Spec: &pb.ReserveSubnetRequest_Spec{
				Region:           vNetPrivate.Spec.Region,
				AvailabilityZone: vNetPrivate.Spec.AvailabilityZone,
				PrefixLengthHint: prefixLengthHint,
				VlanDomain:       req.VlanDomain,
				AddressSpace:     req.AddressSpace,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("IPRM ReserveSubnet: %w", err)
		}
		vNetPrivate.Spec.Region = subnet.Region
		vNetPrivate.Spec.AvailabilityZone = subnet.AvailabilityZone
		vNetPrivate.Spec.Subnet = subnet.Subnet
		vNetPrivate.Spec.PrefixLength = subnet.PrefixLength
		vNetPrivate.Spec.Gateway = subnet.Gateway
		vNetPrivate.Spec.VlanId = subnet.VlanId
		vNetPrivate.Spec.VlanDomain = subnet.VlanDomain
		vNetPrivate.Spec.AddressSpace = subnet.AddressSpace
		if s.objectStorageServicePrivate != nil {
			_, err = s.objectStorageServicePrivate.AddBucketSubnet(ctx, vNetPrivate)
			if err != nil {
				return nil, fmt.Errorf("ObjectStorage AddBucketSubnet: %w", err)
			}
		}
		return vNetPrivate, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

func (s *VNetService) ReleaseSubnet(ctx context.Context, req *pb.VNetReleaseSubnetRequest) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("VNetService.ReleaseSubnet")
	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*emptypb.Empty, error) {
		// Calling releaseSubnet before RemoveBucketSubnet ensures there are no
		// consumers of the subnet before the bucket subnet is removed.
		vNetPrivate, err := s.getPrivate(ctx, req.VNetReference.CloudAccountId, "", req.VNetReference.Name)
		if err != nil {
			return nil, err
		}
		_, err = s.releaseSubnet(ctx, vNetPrivate)
		if status.Code(err) == codes.NotFound {
			// When the subnet is already released we still want to ensure the call to
			// RemoveBucketSubnet below will eventually complete without error.
			logger.Info("Subnet already released")
		} else if err != nil {
			return nil, fmt.Errorf("IPRM ReleaseSubnet: %w", err)
		}
		if s.objectStorageServicePrivate != nil {
			_, err := s.objectStorageServicePrivate.RemoveBucketSubnet(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("ObjectStorage RemoveBucketSubnet: %w", err)
			}
		}
		return &emptypb.Empty{}, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

func (s *VNetService) releaseSubnet(ctx context.Context, vNetPrivate *pb.VNetPrivate) (*emptypb.Empty, error) {
	subnetConsumerId := s.subnetConsumerIdForVNet(vNetPrivate)
	_, err := s.ipResourceManagerService.ReleaseSubnet(ctx, &pb.ReleaseSubnetRequest{
		SubnetReference: &pb.SubnetReference{
			SubnetConsumerId: subnetConsumerId,
		},
	})
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *VNetService) ReserveAddress(ctx context.Context, req *pb.VNetReserveAddressRequest) (*pb.VNetReserveAddressResponse, error) {
	logger := log.FromContext(ctx).WithName("VNetService.ReserveAddress")
	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*pb.VNetReserveAddressResponse, error) {
		vNetPrivate, err := s.getPrivate(ctx, req.VNetReference.CloudAccountId, "", req.VNetReference.Name)
		if err != nil {
			return nil, err
		}
		subnetConsumerId := s.subnetConsumerIdForVNet(vNetPrivate)
		address, err := s.ipResourceManagerService.ReserveAddress(ctx, &pb.ReserveAddressRequest{
			SubnetReference: &pb.SubnetReference{
				SubnetConsumerId: subnetConsumerId,
			},
			AddressReference: &pb.AddressReference{
				AddressConsumerId: req.AddressReference.AddressConsumerId,
				Address:           req.AddressReference.Address,
			},
		})
		if err != nil {
			return nil, err
		}
		resp := &pb.VNetReserveAddressResponse{
			Address: address.Address,
		}
		return resp, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

func (s *VNetService) ReleaseAddress(ctx context.Context, req *pb.VNetReleaseAddressRequest) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("VNetService.ReleaseAddress")
	logger.Info("Request", logkeys.Request, req)
	resp, err := func() (*emptypb.Empty, error) {
		if req.AddressReference.Address != "" {
			return nil, status.Error(codes.InvalidArgument, "address must be empty")
		}
		vNetPrivate, err := s.getPrivate(ctx, req.VNetReference.CloudAccountId, "", req.VNetReference.Name)
		if err != nil {
			return nil, err
		}
		subnetConsumerId := s.subnetConsumerIdForVNet(vNetPrivate)
		return s.ipResourceManagerService.ReleaseAddress(ctx, &pb.ReleaseAddressRequest{
			SubnetReference: &pb.SubnetReference{
				SubnetConsumerId: subnetConsumerId,
			},
			AddressReference: &pb.AddressReference{
				AddressConsumerId: req.AddressReference.AddressConsumerId,
			},
		})
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

func (s *VNetService) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("VNetService.Ping")
	log.Info("Ping")
	return &emptypb.Empty{}, nil
}

func (s *VNetService) PingPrivate(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("VNetService.PingPrivate")
	log.Info("Ping")
	return &emptypb.Empty{}, nil
}

func (s *VNetService) subnetConsumerIdForVNet(vNet *pb.VNetPrivate) string {
	return subnetConsumerIdForVNet(
		s.config.SubnetConsumerIdPattern,
		vNet.Metadata.CloudAccountId,
		vNet.Metadata.Name,
		vNet.Metadata.ResourceId,
		vNet.Spec.Region,
		vNet.Spec.AvailabilityZone,
	)
}

func subnetConsumerIdForVNet(
	subnetConsumerIdPattern string,
	cloudAccountId string,
	name string,
	resourceId string,
	region string,
	availabilityZone string) string {

	replacer := strings.NewReplacer(
		"{CloudAccountId}", cloudAccountId,
		"{Name}", name,
		"{ResourceId}", resourceId,
		"{Region}", region,
		"{AvailabilityZone}", availabilityZone,
	)
	return replacer.Replace(subnetConsumerIdPattern)
}
