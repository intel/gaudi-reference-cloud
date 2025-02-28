package server

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/database/query"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SecGroupMgrService struct {
	syncTicker *time.Ticker
	session    *sql.DB
}

const (
	SUBNET_SCANNER_SCHEDULE_INTERVAL = 2
)

func NewSecGroupMgrService(session *sql.DB) *SecGroupMgrService {
	return &SecGroupMgrService{
		syncTicker: time.NewTicker(time.Duration(SUBNET_SCANNER_SCHEDULE_INTERVAL) * time.Second),
		session:    session,
	}
}

func (s *SecGroupMgrService) StartSecurityGroupScanner(ctx context.Context) {
	// Start the scanner
	log := log.FromContext(ctx).WithName("SecGroupMgrService.StartSecurityGroupScanner")
	log.Info("storage subnet group scheduler")
	var latestVersion int64
	for {
		latestVersion = s.StartSchedulerLoop(ctx, latestVersion)
		tm := <-s.syncTicker.C
		if tm.IsZero() {
			return
		}
	}
}

func (s *SecGroupMgrService) StartSchedulerLoop(ctx context.Context, latestVersion int64) int64 {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageReplicatorService.Replicate").Start()
	defer span.End()

	dbSession := s.session
	if dbSession == nil {
		return 0
	}

	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return 0
	}
	defer tx.Commit()

	respStream, err := query.GetAllSubnetEvents(ctx, tx, latestVersion)
	if err != nil {
		logger.Error(err, "error reading requests ")
		return latestVersion
	}

	for _, event := range respStream {
		// process the request
		logger.Info("process event", "event", event)

		if event.EventType == pb.BucketSubnetEventStatus_E_ADDED ||
			event.EventType == pb.BucketSubnetEventStatus_E_ADDING {
			if err := addSecurityGroupHandler(ctx, tx, event); err != nil {
				logger.Error(err, "error adding security group")
				return latestVersion
			}

		} else if event.EventType == pb.BucketSubnetEventStatus_E_DELETED ||
			event.EventType == pb.BucketSubnetEventStatus_E_DELETING {
			if err := removeSecurityGroupHandler(ctx, tx, event); err != nil {
				logger.Error(err, "error removing security group")
				return latestVersion
			}
		}
		latestVersion = event.ResourceVerion
	}

	return latestVersion
}

func addSecurityGroupHandler(ctx context.Context, tx *sql.Tx, event *query.SubnetEvent) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageReplicatorService.addSecurityGroupHandler").Start()
	defer span.End()

	// get all resources for a given cloud account
	logger.Info("reading all filesystems for cloud account", "cloud_account_id", event.VNet.Metadata.CloudAccountId)
	//Using the filter type as unspecified to make sure we get all the filesystems both general and kubernetes
	fsPrivateList, err := query.GetFilesystemsByCloudaccountId(ctx, tx, event.VNet.Metadata.CloudAccountId, pb.FilesystemType_Unspecified, timestampInfinityStr)
	if err != nil {
		return fmt.Errorf("error reading filesystems")
	}
	logger.Info("# filesystems for cloudaccount", "filesystems", len(fsPrivateList))
	for _, fs := range fsPrivateList {
		update := true
		if fs.Spec.StorageClass != pb.FilesystemStorageClass_GeneralPurposeStd {
			continue
		}
		// security group is assigned to the filesystem, add new
		newSecGrp := &pb.VolumeNetworkGroup{
			Subnet:       event.VNet.Spec.Subnet,       // subnet id
			PrefixLength: event.VNet.Spec.PrefixLength, // subnet prefix length
			Gateway:      event.VNet.Spec.Gateway,      // subnet gateway
		}

		if fs.Spec.SecurityGroup != nil {
			// check if the subnet is already added
			for _, secGrp := range fs.Spec.SecurityGroup.NetworkFilterAllow {
				if reflect.DeepEqual(secGrp, newSecGrp) {
					logger.Info("subnet already added to security group", "subnet", event.VNet.Spec.Subnet)
					update = false
					break
				}
			}
		}
		if update {
			if fs.Spec.SecurityGroup == nil {
				// no security group is assigned to the filesystem, add new
				fs.Spec.SecurityGroup = &pb.VolumeSecurityGroup{
					NetworkFilterAllow: []*pb.VolumeNetworkGroup{newSecGrp},
				}
			} else {
				fs.Spec.SecurityGroup.NetworkFilterAllow = append(fs.Spec.SecurityGroup.NetworkFilterAllow, newSecGrp)
			}
			fs.Metadata.UpdateTimestamp = timestamppb.Now()
			logger.Info("updated security group", "security_group", fs.Spec.SecurityGroup)
			//update the filesystem record in the database
			if err := query.UpdateFilesystemRequest(ctx, tx, fs); err != nil {
				return fmt.Errorf("error updating filesystem after subnet add")
			}
		}
	}
	return nil
}

func removeSecurityGroupHandler(ctx context.Context, tx *sql.Tx, event *query.SubnetEvent) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StorageReplicatorService.removeSecurityGroupHandler").Start()
	defer span.End()

	// get all resources for a given cloud account
	logger.Info("reading all filesystems for cloud account", "cloud_account_id", event.VNet.Metadata.CloudAccountId)
	fsPrivateList, err := query.GetFilesystemsByCloudaccountId(ctx, tx, event.VNet.Metadata.CloudAccountId, pb.FilesystemType_ComputeGeneral, timestampInfinityStr)
	if err != nil {
		return fmt.Errorf("error reading filesystems")
	}

	for _, fs := range fsPrivateList {
		if fs.Spec.StorageClass != pb.FilesystemStorageClass_GeneralPurposeStd {
			continue
		}
		if fs.Spec.SecurityGroup == nil {
			return nil
		} else if len(fs.Spec.SecurityGroup.NetworkFilterAllow) == 1 && fs.Spec.SecurityGroup.NetworkFilterAllow[0].Subnet == event.VNet.Spec.Subnet {
			// single security group is assigned to the filesystem, remove it
			fs.Spec.SecurityGroup = nil
		} else {
			// TODO: multiple security groups are assigned, removed the one which is being deleted
			// Currently, we have single subnet per tenant
		}
		//Update the timestamp
		fs.Metadata.UpdateTimestamp = timestamppb.Now()
		//update the filesystem record in the database
		if err := query.UpdateFilesystemRequest(ctx, tx, fs); err != nil {
			return fmt.Errorf("error updating filesystem after subnet add")
		}
	}

	return nil
}
