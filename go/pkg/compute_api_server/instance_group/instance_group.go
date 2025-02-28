// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package instance_group

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/pbconvert"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type InstanceGroupService struct {
	pb.UnimplementedInstanceGroupServiceServer
	pb.UnimplementedInstanceGroupPrivateServiceServer
	privateInstanceService pb.InstancePrivateServiceServer
	instanceService        pb.InstanceServiceServer
	pbConverter            *pbconvert.PbConverter
}

// Create a new InstanceGroupService instance
func NewInstanceGroupService(privateInstanceService pb.InstancePrivateServiceServer) (*InstanceGroupService, error) {
	return &InstanceGroupService{
		pbConverter:            pbconvert.NewPbConverter(),
		privateInstanceService: privateInstanceService,
		instanceService:        privateInstanceService.(pb.InstanceServiceServer),
	}, nil
}

// Create API implementation.
func (s *InstanceGroupService) Create(ctx context.Context, req *pb.InstanceGroupCreateRequest) (*pb.InstanceGroup, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceGroupService.Create").
		WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId, logkeys.InstanceGroupName, req.Metadata.Name).Start()
	defer span.End()
	logger.Info("Request", logkeys.Request, req)

	resp, err := func() (*pb.InstanceGroup, error) {
		// Validate input.
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
		// Validate the name of the group.
		if err := validateClusterName(req.Metadata.Name); err != nil {
			return nil, err
		}
		// Instance count that is supported between 1 and 128
		if req.Spec.InstanceCount > 128 || req.Spec.InstanceCount < 1 {
			return nil, status.Error(codes.InvalidArgument, "invalid instance count")
		}
		multipleReq, err := s.createInstancePrivateRequest(req)
		if err != nil {
			return nil, status.Error(codes.Internal, "Failed to create instance requests")
		}

		resp, err := s.privateInstanceService.CreateMultiplePrivate(ctx, multipleReq)
		if err != nil {
			return nil, err
		}

		// log the instances created.
		for _, instanceResp := range resp.Instances {
			logger.V(9).Info("Instance Status", logkeys.InstanceGroupName, req.Metadata.Name, logkeys.ResourceId, instanceResp.Metadata.ResourceId,
				logkeys.InstanceStatus, instanceResp.Status.Phase)
		}

		return &pb.InstanceGroup{
			Metadata: &pb.InstanceGroupMetadata{
				CloudAccountId: req.Metadata.CloudAccountId,
				Name:           req.Metadata.Name,
			},
			Spec: req.Spec,
			Status: &pb.InstanceGroupStatus{
				ReadyCount: 0, // the value is zero since the instances will definitely not be ready.
			},
		}, nil

	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, utils.SanitizeError(err)
}

func (s *InstanceGroupService) CreatePrivate(ctx context.Context, req *pb.InstanceGroupCreatePrivateRequest) (*pb.InstanceGroupPrivateCreateResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceGroupService.CreatePrivate").
		WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId, logkeys.InstanceGroupName, req.Metadata.Name).Start()
	defer span.End()
	logger.Info("Request", logkeys.Request, req)

	resp, err := func() (*pb.InstanceGroupPrivateCreateResponse, error) {
		// Validate input.
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
		if err := validateClusterName(req.Metadata.Name); err != nil {
			return nil, err
		}

		// Create a request for instances to be created.
		var requestList []*pb.InstanceCreatePrivateRequest
		for i := 0; i < int(req.Spec.InstanceCount); i++ {
			resourceId, err := uuid.NewRandom()
			if err != nil {
				return nil, status.Errorf(codes.Internal, "error encountered while generating resourceId: %v", err)
			}

			// Set the default instance spec
			instanceSpecPrivate := proto.Clone(req.Spec.InstanceSpecPrivate).(*pb.InstanceSpecPrivate)
			instanceSpecPrivate.InstanceGroup = req.Metadata.Name
			scLocations := ""
			if req.Spec.Placement != nil {
				// Set the preferred SC locations. Multiple locations can be provided but only one will be selected.
				if len(req.Spec.Placement.SuperComputeGroupIds) > 0 {
					scLocations = strings.Join(req.Spec.Placement.SuperComputeGroupIds, ",")
				}
			}
			instanceSpecPrivate.SuperComputeGroupId = scLocations

			instCreateReq := &pb.InstanceCreatePrivateRequest{
				Metadata: &pb.InstanceMetadataCreatePrivate{
					CloudAccountId: req.Metadata.CloudAccountId,
					Name:           req.Metadata.Name + "-" + strconv.Itoa(i),
					ResourceId:     resourceId.String(),
					Labels:         req.Metadata.Labels,
				},
				Spec: instanceSpecPrivate,
			}
			requestList = append(requestList, instCreateReq)
		}

		instanceCreateMultiplePrivateRequest := &pb.InstanceCreateMultiplePrivateRequest{
			Instances: requestList,
			DryRun:    req.DryRun,
		}

		// Create instances.
		resp, err := s.privateInstanceService.CreateMultiplePrivate(ctx, instanceCreateMultiplePrivateRequest)
		if err != nil {
			return nil, err
		}

		superComputeIds := map[string]bool{}
		clusterGroupIds := map[string]bool{}
		clusterIds := map[string]bool{}
		nodeIds := map[string]bool{}

		for _, instanceResp := range resp.Instances {
			logger.V(9).Info("Instance Status", logkeys.InstanceGroupName, req.Metadata.Name,
				logkeys.ResourceId, instanceResp.Metadata.ResourceId, logkeys.InstanceStatus, instanceResp.Status.Phase)

			superComputeIds[instanceResp.Spec.SuperComputeGroupId] = true
			clusterGroupIds[instanceResp.Spec.ClusterGroupId] = true
			clusterIds[instanceResp.Spec.ClusterId] = true
			nodeIds[instanceResp.Spec.NodeId] = true
		}

		return &pb.InstanceGroupPrivateCreateResponse{
			Metadata:  req.Metadata,
			Instances: resp.Instances,
			Placement: &pb.InstanceGroupPlacement{
				SuperComputeGroupIds: utils.ToList(superComputeIds),
				ClusterGroupIds:      utils.ToList(clusterGroupIds),
				ClusterIds:           utils.ToList(clusterIds),
				NodeIds:              utils.ToList(nodeIds),
			},
		}, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}

func (s *InstanceGroupService) Update(ctx context.Context, req *pb.InstanceGroupUpdateRequest) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceGroupService.Update").
		WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId, logkeys.InstanceGroupName, req.Metadata.Name).Start()
	defer span.End()
	log.Info("Request", logkeys.Request, req)

	err := func() error {
		if req.Metadata == nil {
			return status.Error(codes.InvalidArgument, "missing metadata")
		}
		if req.Spec == nil {
			return status.Error(codes.InvalidArgument, "missing spec")
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return err
		}
		// Validate the name of the group.
		instanceGroup := req.Metadata.Name
		if err := validateClusterName(instanceGroup); err != nil {
			return err
		}
		instances, err := s.instanceService.Search(ctx, &pb.InstanceSearchRequest{
			Metadata: &pb.InstanceMetadataSearch{
				CloudAccountId:      cloudAccountId,
				InstanceGroup:       instanceGroup,
				InstanceGroupFilter: pb.SearchFilterCriteria_ExactValue,
			},
		})
		if err != nil {
			return err
		}

		var updateInstanceErr error
		for _, instance := range instances.Items {

			if instance.Spec.InstanceGroup != instanceGroup {
				return status.Errorf(codes.InvalidArgument, "instance of a different instance group %s found", instance.Spec.InstanceGroup)
			}
			_, updateInstanceErr = s.instanceService.Update(ctx, &pb.InstanceUpdateRequest{
				Metadata: &pb.InstanceMetadataUpdate{
					CloudAccountId: cloudAccountId,
					NameOrId: &pb.InstanceMetadataUpdate_Name{
						Name: instance.Metadata.Name,
					},
				},
				Spec: req.Spec.InstanceSpec,
			})
			if updateInstanceErr != nil {
				log.Error(err, "Failed to update instance which part of an instanceGroup", logkeys.InstanceName, instance.Metadata.Name,
					logkeys.ResourceId, instance.Metadata.ResourceId)
			}
		}
		return updateInstanceErr
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Update InstanceGroup completed")
	}

	return &emptypb.Empty{}, utils.SanitizeError(err)
}

func (s *InstanceGroupService) Search(ctx context.Context, req *pb.InstanceGroupSearchRequest) (*pb.InstanceGroupSearchResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceGroupService.Search").
		WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("Request", logkeys.Request, req)

	resp, err := func() (*pb.InstanceGroupSearchResponse, error) {
		// validate input
		if req.Metadata == nil {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return nil, err
		}

		searchRequest := &pb.InstanceSearchRequest{
			Metadata: &pb.InstanceMetadataSearch{
				CloudAccountId:      req.Metadata.CloudAccountId,
				InstanceGroupFilter: pb.SearchFilterCriteria_NonEmpty,
			},
		}
		searchResponse, err := s.instanceService.Search(ctx, searchRequest)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "error encountered while getting all the instances for cloudaccountId %s: %v", req.Metadata.CloudAccountId, err)
		}

		instanceGroups := make(map[string][]*pb.Instance)
		for _, instance := range searchResponse.Items {
			if instance.Spec.InstanceGroup != "" {
				instanceGroups[instance.Spec.InstanceGroup] = append(instanceGroups[instance.Spec.InstanceGroup], instance)
			}
		}

		var items []*pb.InstanceGroup
		for instanceGroup, instances := range instanceGroups {
			var readyCount int32
			var instance *pb.Instance
			for _, instance = range instances {
				if instance.Status.Phase == pb.InstancePhase_Ready {
					readyCount++
				}
			}
			instanceCount := len(instances)
			group := &pb.InstanceGroup{
				Metadata: &pb.InstanceGroupMetadata{
					CloudAccountId: req.Metadata.CloudAccountId,
					Name:           instanceGroup,
				},
				Spec: &pb.InstanceGroupSpec{
					InstanceSpec:  instance.Spec,
					InstanceCount: int32(instanceCount),
				},
				Status: &pb.InstanceGroupStatus{
					ReadyCount: readyCount,
				},
			}
			items = append(items, group)
		}
		resp := &pb.InstanceGroupSearchResponse{
			Items: items,
		}
		return resp, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, utils.SanitizeError(err)
}

// Delete API implementation.
func (s *InstanceGroupService) Delete(ctx context.Context, req *pb.InstanceGroupDeleteRequest) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceGroupService.Delete").
		WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId, logkeys.InstanceGroupName, req.Metadata.GetName()).Start()
	defer span.End()
	log.Info("Request", logkeys.Request, req)

	err := func() error {

		// Validate input.
		if req.Metadata == nil {
			return status.Error(codes.InvalidArgument, "missing metadata")
		}

		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return err
		}

		// Validate the name of the group.
		instanceGroup := req.Metadata.GetName()
		if err := validateClusterName(instanceGroup); err != nil {
			return err
		}

		// Search for instances
		instances, err := s.instanceService.Search(ctx, &pb.InstanceSearchRequest{
			Metadata: &pb.InstanceMetadataSearch{
				CloudAccountId:      cloudAccountId,
				InstanceGroup:       instanceGroup,
				InstanceGroupFilter: pb.SearchFilterCriteria_ExactValue,
			},
		})
		if err != nil {
			return err
		}

		var deleteInstanceError error
		for _, instance := range instances.Items {
			if instance.Spec.InstanceGroup == req.Metadata.GetName() {
				_, deleteInstanceError = s.privateInstanceService.DeletePrivate(ctx, &pb.InstanceDeletePrivateRequest{
					Metadata: &pb.InstanceMetadataReference{
						CloudAccountId: cloudAccountId,
						NameOrId: &pb.InstanceMetadataReference_Name{
							Name: instance.Metadata.Name,
						},
					},
				})
				if deleteInstanceError != nil {
					log.Error(err, "Failed to delete instance", logkeys.InstanceName, instance.Metadata.Name, logkeys.ResourceId, instance.Metadata.ResourceId)
				}
			}
		}
		if deleteInstanceError != nil {
			return err
		}
		return nil

	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Deletion invoked successfully", logkeys.InstanceGroupName, req.Metadata.NameOrId)
	}
	return &emptypb.Empty{}, utils.SanitizeError(err)
}

// DeleteMember API implementation.
func (s *InstanceGroupService) DeleteMember(ctx context.Context, req *pb.InstanceGroupMemberDeleteRequest) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceGroupService.DeleteMember").
		WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId, logkeys.InstanceGroupName, req.Metadata.GetName()).Start()
	defer span.End()
	log.Info("Request", logkeys.Request, req)

	err := func() error {
		if req.Metadata == nil {
			return status.Error(codes.InvalidArgument, "missing group metadata")
		}
		cloudAccountId := req.Metadata.CloudAccountId
		if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
			return err
		}
		groupName := req.Metadata.GetName()
		if err := validateClusterName(groupName); err != nil {
			return err
		}

		// check the existing instances in the instanceGroup
		instances, err := s.instanceService.Search(ctx, &pb.InstanceSearchRequest{
			Metadata: &pb.InstanceMetadataSearch{
				CloudAccountId:      cloudAccountId,
				InstanceGroup:       groupName,
				InstanceGroupFilter: pb.SearchFilterCriteria_ExactValue,
			},
		})
		if err != nil {
			return err
		}
		instanceCount := len(instances.Items)
		if instanceCount == 0 {
			return status.Errorf(codes.NotFound, "no instances found for instanceGroup %v", groupName)
		}

		updateGroupSize := func(groupSize int) {
			for _, instance := range instances.Items {
				instanceMetadataUpdate := &pb.InstanceMetadataUpdate{}
				if req.GetInstanceResourceId() != "" {
					instanceMetadataUpdate = &pb.InstanceMetadataUpdate{
						CloudAccountId: cloudAccountId,
						NameOrId: &pb.InstanceMetadataUpdate_ResourceId{
							ResourceId: instance.Metadata.ResourceId,
						},
					}
				} else if req.GetInstanceName() != "" {
					instanceMetadataUpdate = &pb.InstanceMetadataUpdate{
						CloudAccountId: cloudAccountId,
						NameOrId: &pb.InstanceMetadataUpdate_Name{
							Name: instance.Metadata.Name,
						},
					}
				}

				_, err := s.privateInstanceService.UpdatePrivate(ctx, &pb.InstanceUpdatePrivateRequest{
					Metadata: instanceMetadataUpdate,
					Spec: &pb.InstanceSpecPrivate{
						InstanceGroupSize: int32(groupSize),
					},
				})
				if err != nil {
					log.Error(err, "Failed to update instance which is part of an instanceGroup", logkeys.InstanceName, instance.Metadata.Name,
						logkeys.ResourceId, instance.Metadata.ResourceId)
				}
			}
		}

		deletedInstanceCount := 0
		isActiveInstanceBeingDeleted := false
		instanceNameFoundInGroup := false
		instanceResourceIdFoundInGroup := false

		for _, instance := range instances.Items {
			// check that the instance to be deleted is part of the instanceGroup
			if req.GetInstanceName() != "" {
				if instance.Metadata.Name == req.GetInstanceName() {
					instanceNameFoundInGroup = true
				}
			} else if req.GetInstanceResourceId() != "" {
				if instance.Metadata.ResourceId == req.GetInstanceResourceId() {
					instanceResourceIdFoundInGroup = true
				}
			}

			// validate the group name
			if instance.Spec.InstanceGroup != groupName {
				return status.Errorf(codes.InvalidArgument, "instance %v is not part of instanceGroup %v", instance.Metadata.Name, groupName)
			}

			// track the number of instances being deleted
			if instance.Metadata.DeletionTimestamp != nil {
				deletedInstanceCount++
				continue
			}
			if instance.Metadata.Name == req.GetInstanceName() ||
				instance.Metadata.ResourceId == req.GetInstanceResourceId() {
				isActiveInstanceBeingDeleted = true
			}
		}

		// retain the last remaining instance to use a template
		hasOneActiveInstance := deletedInstanceCount == instanceCount-1
		if hasOneActiveInstance && isActiveInstanceBeingDeleted {
			return status.Error(codes.FailedPrecondition, "deleting the last remaining instance from instanceGroup is not allowed by this method.")
		}

		// sync the instance group size based on the number of instances being deleted
		currentGroupSize := instanceCount - deletedInstanceCount
		updateGroupSize(currentGroupSize)

		// prepare the instance delete request
		instanceMetadata := &pb.InstanceMetadataReference{}
		if req.GetInstanceResourceId() != "" {
			if !instanceResourceIdFoundInGroup {
				return status.Errorf(codes.NotFound, "instance with resourceId %v not found in instanceGroup %v", req.GetInstanceResourceId(), groupName)
			}
			instanceMetadata = &pb.InstanceMetadataReference{
				CloudAccountId: cloudAccountId,
				NameOrId: &pb.InstanceMetadataReference_ResourceId{
					ResourceId: req.GetInstanceResourceId(),
				},
			}
		} else if req.GetInstanceName() != "" {
			if !instanceNameFoundInGroup {
				return status.Errorf(codes.NotFound, "instance with name %v not found in instanceGroup %v", req.GetInstanceName(), groupName)
			}
			instanceMetadata = &pb.InstanceMetadataReference{
				CloudAccountId: cloudAccountId,
				NameOrId: &pb.InstanceMetadataReference_Name{
					Name: req.GetInstanceName(),
				},
			}
		} else {
			return status.Error(codes.InvalidArgument, "missing instance name or resource id")
		}

		// delete the instance
		log.Info("Deleting instance", logkeys.InstanceMetadata, instanceMetadata)
		_, deleteErr := s.privateInstanceService.DeletePrivate(ctx, &pb.InstanceDeletePrivateRequest{
			Metadata: instanceMetadata,
		})
		if deleteErr != nil {
			return err
		}

		// update the new instance group size
		if isActiveInstanceBeingDeleted {
			// increment the number of deleted instance since an active instance was deleted.
			deletedInstanceCount++
		}
		newGroupSize := instanceCount - deletedInstanceCount
		updateGroupSize(newGroupSize)

		return nil
	}()
	if err != nil {
		log.Error(err, logkeys.Error, logkeys.Request, req)
	} else {
		log.Info("Deletion invoked successfully", logkeys.InstanceGroupName, req.Metadata.GetName())
	}
	return &emptypb.Empty{}, utils.SanitizeError(err)
}

// Scale API implementation.
// The following can be updated:
// - InstanceGroupScaleRequest.Spec.instanceCount
// - InstanceGroupScaleRequest.Spec.InstanceSpec.UserData
func (s *InstanceGroupService) ScaleUp(ctx context.Context, req *pb.InstanceGroupScaleRequest) (*pb.InstanceGroupScaleResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("InstanceGroupService.ScaleUp").
		WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId, logkeys.InstanceGroupName, req.Metadata.Name).Start()
	defer span.End()
	logger.Info("Request", logkeys.Request, req)

	resp, err := func() (*pb.InstanceGroupScaleResponse, error) {
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
		// Validate the name of the group.
		groupName := req.Metadata.Name
		if err := validateClusterName(groupName); err != nil {
			return nil, err
		}

		// Instance count that is supported between 1 and 128
		desiredCount := req.Spec.InstanceCount
		if desiredCount > 128 || desiredCount < 1 {
			return nil, status.Error(codes.InvalidArgument, "invalid instance count")
		}

		instances, err := s.privateInstanceService.SearchPrivate(ctx, &pb.InstanceSearchPrivateRequest{
			Metadata: &pb.InstanceMetadataSearch{
				CloudAccountId:      cloudAccountId,
				InstanceGroup:       groupName,
				InstanceGroupFilter: pb.SearchFilterCriteria_ExactValue,
			},
		})
		if err != nil {
			return nil, err
		}
		currentCount := int32(len(instances.Items))

		// need at least one instance to scale
		if currentCount == 0 {
			return nil, status.Errorf(codes.NotFound, "require at lest one instance in instanceGroup %v to scale", groupName)
		}

		// scaling down is unsupported
		if currentCount > desiredCount {
			return nil, status.Errorf(codes.InvalidArgument, "scaling down is unsupported. currentCount=%d, desiredCount=%d", currentCount, desiredCount)
		}

		// extract information from existing instances
		createdInstances := make(map[string]bool)
		clusterGroupIds := make(map[string]bool)
		deletedCount := 0
		currentMemberNames := []string{}
		newMemberNames := []string{}
		readyMemberNames := []string{}

		for _, instance := range instances.Items {
			instanceName := instance.Metadata.GetName()
			createdInstances[instance.Metadata.GetName()] = true
			if instance.Status.Phase == pb.InstancePhase_Ready {
				readyMemberNames = append(readyMemberNames, instanceName)
			}
			if instance.Metadata.DeletionTimestamp != nil {
				deletedCount++
			} else {
				currentMemberNames = append(currentMemberNames, instanceName)
			}
			if instance.Spec.ClusterGroupId != "" {
				clusterGroupIds[instance.Spec.ClusterGroupId] = true
			}
		}

		readyCount := int32(len(readyMemberNames))

		if deletedCount == int(currentCount) {
			return nil, status.Errorf(codes.NotFound, "instanceGroup %v is being deleted", groupName)
		}

		assignedGroupIds := []string{}
		for groupId := range clusterGroupIds {
			assignedGroupIds = append(assignedGroupIds, groupId)
		}

		updateGroupsize := func() {
			for _, instance := range instances.Items {
				_, err := s.privateInstanceService.UpdatePrivate(ctx, &pb.InstanceUpdatePrivateRequest{
					Metadata: &pb.InstanceMetadataUpdate{
						CloudAccountId: cloudAccountId,
						NameOrId: &pb.InstanceMetadataUpdate_Name{
							Name: instance.Metadata.Name,
						},
					},
					Spec: &pb.InstanceSpecPrivate{
						InstanceGroupSize: desiredCount,
					},
				})
				if err != nil {
					logger.Error(err, "Failed to update instance which part of an instanceGroup", logkeys.InstanceName, instance.Metadata.Name,
						logkeys.ResourceId, instance.Metadata.ResourceId)
				}
			}
		}

		if currentCount == desiredCount {
			updateGroupsize()
			return &pb.InstanceGroupScaleResponse{
				Metadata: &pb.InstanceGroupMetadata{
					CloudAccountId: cloudAccountId,
					Name:           groupName,
				},
				Status: &pb.InstanceGroupScaleStatus{
					CurrentCount:   int32(len(currentMemberNames)),
					DesiredCount:   desiredCount,
					ReadyCount:     readyCount,
					CurrentMembers: currentMemberNames,
					NewMembers:     newMemberNames,
					ReadyMembers:   readyMemberNames,
				},
			}, nil
		}

		// use an existing instance as a template for the new instances
		instancePrivate := instances.Items[0]
		instance := &pb.Instance{}
		if err := s.pbConverter.Transcode(instancePrivate, instance); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to transcode: %v", err)
		}

		// build all instances that should be created in this group
		createReq := &pb.InstanceGroupCreateRequest{
			Metadata: &pb.InstanceGroupMetadataCreate{
				CloudAccountId: cloudAccountId,
				Name:           groupName,
			},
			Spec: &pb.InstanceGroupSpec{
				InstanceSpec:  instance.Spec,
				InstanceCount: desiredCount,
			},
		}
		multipleReq, err := s.createInstancePrivateRequest(createReq)
		if err != nil {
			return nil, status.Error(codes.Internal, "Failed to create instance requests")
		}

		// create only non-existing instances
		var instancesToBeCreated []*pb.InstanceCreatePrivateRequest
		for _, in := range multipleReq.Instances {
			if _, created := createdInstances[in.Metadata.GetName()]; !created {
				// override user data if provided
				if req.Spec.InstanceSpec != nil &&
					req.Spec.InstanceSpec.UserData != "" {
					in.Spec.UserData = req.Spec.InstanceSpec.UserData
				}
				// set default network interface
				if len(in.Spec.Interfaces) > 1 {
					in.Spec.Interfaces = in.Spec.Interfaces[:1]
				}
				// set network information
				in.Spec.NetworkMode = instancePrivate.Spec.NetworkMode
				in.Spec.SuperComputeGroupId = instancePrivate.Spec.SuperComputeGroupId
				if len(clusterGroupIds) > 1 {
					// the final group ID will be decided by the scheduler.
					in.Spec.ClusterGroupId = strings.Join(assignedGroupIds, ",")
				} else {
					in.Spec.ClusterGroupId = instancePrivate.Spec.ClusterGroupId
				}
				in.Spec.InstanceGroupSize = desiredCount
				in.Metadata.Labels = instancePrivate.Metadata.Labels
				instancesToBeCreated = append(instancesToBeCreated, in)
			}
		}

		resp, err := s.privateInstanceService.CreateMultiplePrivate(ctx, &pb.InstanceCreateMultiplePrivateRequest{
			Instances: instancesToBeCreated,
		})
		if err != nil {
			return nil, err
		}

		for _, instance := range resp.Instances {
			logger.V(9).Info("Instance Status", logkeys.InstanceGroupName, groupName, logkeys.ResourceId, instance.Metadata.ResourceId,
				logkeys.InstanceStatus, instance.Status.Phase)
			currentMemberNames = append(currentMemberNames, instance.Metadata.GetName())
			newMemberNames = append(newMemberNames, instance.Metadata.GetName())
		}
		sort.Strings(currentMemberNames)

		// update the new instanceGroup size
		updateGroupsize()

		return &pb.InstanceGroupScaleResponse{
			Metadata: &pb.InstanceGroupMetadata{
				CloudAccountId: cloudAccountId,
				Name:           groupName,
			},
			Status: &pb.InstanceGroupScaleStatus{
				CurrentCount:   int32(len(currentMemberNames)),
				DesiredCount:   desiredCount,
				ReadyCount:     readyCount,
				CurrentMembers: currentMemberNames,
				NewMembers:     newMemberNames,
				ReadyMembers:   readyMemberNames,
			},
		}, nil
	}()
	log.LogResponseOrError(logger, req, resp, err)
	return resp, utils.SanitizeError(err)
}

func (s *InstanceGroupService) createInstancePrivateRequest(req *pb.InstanceGroupCreateRequest) (*pb.
	InstanceCreateMultiplePrivateRequest, error) {

	var requestList []*pb.InstanceCreatePrivateRequest
	for i := 0; i < int(req.Spec.InstanceCount); i++ {
		// 	Every request will have its own resource id.
		resourceId, err := uuid.NewRandom()
		if err != nil {
			return nil, err
		}

		// Transcode InstanceSpec to InstanceSpecPrivate.
		// Unmatched fields will remain at their default.
		instanceSpecPrivate := &pb.InstanceSpecPrivate{}
		if err := s.pbConverter.Transcode(req.Spec.InstanceSpec, instanceSpecPrivate); err != nil {
			return nil, fmt.Errorf("unable to transcode instance spec: %w", err)
		}
		// Update InstanceGroup in the spec
		instanceSpecPrivate.InstanceGroup = req.Metadata.Name

		instCreateReq := &pb.InstanceCreatePrivateRequest{
			Metadata: &pb.InstanceMetadataCreatePrivate{
				CloudAccountId: req.Metadata.CloudAccountId,
				ResourceId:     resourceId.String(),
				Name:           req.Metadata.Name + "-" + strconv.Itoa(i),
			},
			Spec: instanceSpecPrivate,
		}
		requestList = append(requestList, instCreateReq)
	}

	return &pb.InstanceCreateMultiplePrivateRequest{
		Instances: requestList,
	}, nil
}

// Validate clusterName.
// clustername is valid when name is starting and ending with lowercase alphanumeric
// and contains lowercase alphanumeric, '-' characters and should have at most 63 characters
func validateClusterName(name string) error {
	re := regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,57}[a-z0-9])?$`)
	matches := re.FindAllString(name, -1)
	if matches == nil {
		return status.Error(codes.InvalidArgument, "invalid instance name")
	}
	return nil
}
