package server

import (
	"context"
	"fmt"
	"io"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	stcnt_api "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
)

func (svc *UserService) HandlePrincipalSecurityGroupUpdate(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("UserService.HandlePrincipalSecurityGroupUpdate")
	streamRes, err := svc.bucketAPIClient.GetBucketSubnetEvent(ctx, &pb.SubnetEventRequest{})
	if err != nil {
		return fmt.Errorf("error reading subnet events from bucket service")
	}
	events := []*pb.BucketSubnetUpdateEvent{}
	done := make(chan bool)

	//Read events from stream
	go func() {
		for {
			resp, err := streamRes.Recv()
			if err == io.EOF {
				done <- true //close(done)
				return
			}
			if err != nil {
				log.Error(err, "error reading from agent stream")
				done <- true //close(done)
				return
			}
			events = append(events, resp)
		}
	}()
	<-done

	//Process events if not empty
	if len(events) > 0 {
		log.Info("events collected successfully", logkeys.TotalEvents, len(events))
		for _, event := range events {
			log.Info("handling event for", logkeys.CloudAccountId, event.Vnet.Metadata.CloudAccountId, logkeys.EventType, event.EventType)
			for _, p := range event.Principals {
				params := storagecontroller.ObjectUserUpdateRequest{
					ClusterId:   p.ClusterId,
					PrincipalId: p.PrincipalId,
					Policies:    getUpdatedPolicies(p.Spec, event.Vnet, event.EventType, p.ClusterId),
				}
				err := svc.strCntClient.UpdateObjectUserPolicy(ctx, params)
				if err != nil {
					log.Error(err, "error updating principal from sds")
					return fmt.Errorf("error updating user")
				}
			}

			log.Info("updating bucket subnet status ")
			if _, err := svc.bucketAPIClient.UpdateBucketSubnetStatus(ctx, &pb.BucketSubnetStatusUpdateRequest{
				ResourceId:      event.Vnet.Metadata.ResourceId,
				CloudacccountId: event.Vnet.Metadata.CloudAccountId,
				VNetName:        event.Vnet.Metadata.Name,
				Status:          getSuccessStatus(event.EventType),
			}); err != nil {
				log.Error(err, "error updating principal status in storage api service")
				return fmt.Errorf("error updating principal status update ")
			}
		}
	}

	return nil
}

func getUpdatedPolicies(currentSpec []*pb.ObjectUserPermissionSpec, updatedVnet *pb.VNetPrivate, eventType pb.BucketSubnetEventStatus, clusterId string) []*stcnt_api.S3Principal_Policy {
	principalPolicies := []*stcnt_api.S3Principal_Policy{}

	allowSourceIpFilters := []string{}

	if eventType == pb.BucketSubnetEventStatus_E_ADDING {
		allowSourceIpFilters = append(allowSourceIpFilters, fmt.Sprintf("%s/%d", updatedVnet.Spec.Subnet, updatedVnet.Spec.PrefixLength))
	} else if eventType == pb.BucketSubnetEventStatus_E_DELETING {
		// curretly we are supporting only single subnet per bucket
		allowSourceIpFilters = append(allowSourceIpFilters, "0.0.0.0/27")
	}
	for _, inPolicy := range currentSpec {
		read, write, delete := mapPermissionsToS3(inPolicy.Permission)
		outPolicy := &stcnt_api.S3Principal_Policy{
			BucketId: &stcnt_api.BucketIdentifier{
				ClusterId: &stcnt_api.ClusterIdentifier{
					Uuid: clusterId,
				},
				Id: inPolicy.BucketId,
			},
			Prefix:  inPolicy.Prefix,
			Read:    read,
			Write:   write,
			Delete:  delete,
			Actions: mapActionsToS3(inPolicy.Actions),
			SourceIpFilter: &stcnt_api.S3Principal_Policy_SourceIpFilter{
				Allow: allowSourceIpFilters,
			},
		}
		principalPolicies = append(principalPolicies, outPolicy)
	}
	return principalPolicies
}

func getSuccessStatus(in pb.BucketSubnetEventStatus) pb.BucketSubnetEventStatus {
	switch in {
	case pb.BucketSubnetEventStatus_E_ADDING:
		return pb.BucketSubnetEventStatus_E_ADDED
	case pb.BucketSubnetEventStatus_E_DELETING:
		return pb.BucketSubnetEventStatus_E_DELETED
	default:
		return pb.BucketSubnetEventStatus_E_UNSPECIFIED
	}
}
