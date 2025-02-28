package query

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	insertBucketSubnet = `
		insert into bucket_network_security_group 
			(resource_id, cloud_account_id, vnetName, event_status, updated_timestamp, value) 
		values ($1, $2, $3, $4, $5, $6)
		on conflict DO nothing
	`

	getBucketSubnetByCloudAccount = `
		select value
		from bucket_network_security_group
		where cloud_account_id = $1 
		and event_status in ('ADDING', 'ADDED')
	`

	getAllBucketSubnetByCloudAccount = `
		select value
		from bucket_network_security_group
		where cloud_account_id = $1 
		and vnetname = $2
	`

	getAllSubnetEvents = `
		select cloud_account_id, event_status, resource_version, value
			from bucket_network_security_group
			where resource_version > $1 
			order by id asc
	`

	deleteBucketSubnet = `
		update bucket_network_security_group
		set    resource_version = nextval('vnet_resource_version_seq'),
			event_status = $1,
			updated_timestamp = $2
			where cloud_account_id = $3
			and vnetname = $4
	`

	deleteBucketSubnetFromDB = `
		delete from  bucket_network_security_group
			where cloud_account_id = $1
			and vnetname = $2
	`

	getAllBucketSubnetEvents = `
		select objectuser.value, secgroup.value, secgroup.event_status
		from object_user as objectuser, bucket_network_security_group as secgroup
		where secgroup.cloud_account_id = objectuser.cloud_account_id
			and objectuser.deleted_timestamp = $1 
			and secgroup.event_status in ('ADDING', 'DELETING')
	`

	updateStatusForSubnet = `
		update bucket_network_security_group
		set    resource_version = nextval('vnet_resource_version_seq'),
			event_status = $1,  
			updated_timestamp = $2
			where resource_id = $3
			and cloud_account_id = $4
			and vnetname = $5
	`

	updateSubnetValue = `
		update bucket_network_security_group
		set    resource_version = nextval('vnet_resource_version_seq'),
			event_status = $1,  
			updated_timestamp = $2,
			value = $3
			where resource_id = $4
			and cloud_account_id = $5
			and vnetname = $6
	`

	AddingEventType = "ADDING"

	DeletingEventType = "DELETING"

	AddedEventType = "ADDED"

	DeletedEventType = "DELETED"
)

type SubnetEvent struct {
	VNet           *pb.VNetPrivate
	EventType      pb.BucketSubnetEventStatus
	ResourceVerion int64
}

func StoreBucketSubnet(ctx context.Context, tx *sql.Tx, vnet *pb.VNetPrivate) error {
	logger := log.FromContext(ctx).WithName("StoreBucketSubnet").
		WithValues(logkeys.ResourceId, vnet.Metadata.ResourceId, logkeys.CloudAccountId, vnet.Metadata.CloudAccountId, "vnet", vnet)

	logger.Info("begin bucket subnet record insertion")

	jsonVal, err := json.MarshalIndent(vnet, "", "    ")
	if err != nil {
		return fmt.Errorf("json marshaling: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		insertBucketSubnet,
		vnet.Metadata.ResourceId,
		vnet.Metadata.CloudAccountId,
		vnet.Metadata.Name,
		AddingEventType,
		time.Now(),
		string(jsonVal))
	if err != nil {
		logger.Error(err, "error inserting subnet record")
		return err
	}
	logger.Info("bucket subnet record inserted successfully")

	return nil
}

func UpdateBucketSubnet(ctx context.Context, tx *sql.Tx, vnet *pb.VNetPrivate) error {
	logger := log.FromContext(ctx).WithName("UpdateBucketSubnet").
		WithValues(logkeys.ResourceId, vnet.Metadata.ResourceId, logkeys.CloudAccountId, vnet.Metadata.CloudAccountId, "vnet", vnet)

	logger.Info("begin bucket subnet record update")

	jsonVal, err := json.MarshalIndent(vnet, "", "    ")
	if err != nil {
		return fmt.Errorf("json marshaling: %w", err)
	}
	_, err = tx.ExecContext(ctx,
		updateSubnetValue,
		AddingEventType,
		time.Now(),
		string(jsonVal),
		vnet.Metadata.ResourceId,
		vnet.Metadata.CloudAccountId,
		vnet.Metadata.Name,
	)
	if err != nil {
		logger.Error(err, "error updating subnet record")
		return err
	}
	logger.Info("bucket subnet records updated successfully")

	return nil
}

func CheckSubnetExists(ctx context.Context, tx *sql.Tx, vnet *pb.VNetPrivate) (bool, error) {
	logger := log.FromContext(ctx).WithName("CheckExists")
	rows, err := tx.QueryContext(ctx, getAllBucketSubnetByCloudAccount, vnet.Metadata.CloudAccountId, vnet.Metadata.Name)
	if err != nil {
		logger.Error(err, "error searching buckets record in db")
		return false, status.Errorf(codes.Internal, "buckets record search failed")
	}
	defer rows.Close()
	if rows.Next() {
		return true, nil
	}
	return false, nil
}

func DeleteBucketSubnet(ctx context.Context, tx *sql.Tx, releaseReq *pb.VNetReleaseSubnetRequest) error {
	logger := log.FromContext(ctx).WithName("DeleteBucketSubnet")
	logger.Info("begin bucket subnet release request", logkeys.ResourceId, releaseReq)

	result, err := tx.ExecContext(ctx,
		deleteBucketSubnet,

		DeletingEventType,
		time.Now(),
		releaseReq.VNetReference.CloudAccountId,
		releaseReq.VNetReference.Name,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	logger.Info("debug info", logkeys.NumAffectedRows, rowsAffected)

	if rowsAffected < 1 {
		return status.Error(codes.FailedPrecondition, "no records updated; possible update conflict")
	}

	return nil
}

func GetBucketSubnetByAccount(ctx context.Context, tx *sql.Tx, cloudaccountId string) (*pb.VNetPrivate, error) {
	logger := log.FromContext(ctx).WithName("GetBucketSubnetByAccount")
	logger.Info("begin bucket subnet get request", logkeys.CloudAccountId, cloudaccountId)

	dataBuf := []byte{}
	vnetPrivate := pb.VNetPrivate{}
	row := tx.QueryRowContext(ctx, getBucketSubnetByCloudAccount, cloudaccountId)
	switch err := row.Scan(&dataBuf); err {
	case sql.ErrNoRows:
		logger.Info("no records found ", logkeys.CloudAccountId, cloudaccountId)
		return nil, status.Errorf(codes.NotFound, "no matching records found")
	case nil:
		err := json.Unmarshal([]byte(dataBuf), &vnetPrivate)
		if err != nil {
			logger.Error(err, "Error Unmarshalling JSON")
			return nil, status.Errorf(codes.Internal, "bucket subnet record search failed")
		}
	default:
		logger.Error(err, "error searching bucket subnet record in db")
		return nil, status.Errorf(codes.Internal, "bucket subnet record find failed")
	}
	return &vnetPrivate, nil
}

func GetAllBucketSubnetEvents(ctx context.Context, tx *sql.Tx) ([]*pb.BucketSubnetUpdateEvent, error) {
	logger := log.FromContext(ctx).WithName("GetAllBucketSubnetEvents")
	logger.Info("begin bucket subnet get all events request")

	subnetEvents := []*pb.BucketSubnetUpdateEvent{}

	rows, err := tx.QueryContext(ctx, getAllBucketSubnetEvents, timestampInfinityStr)
	if err != nil {
		logger.Error(err, "error searching buckets record in db")
		return nil, status.Errorf(codes.Internal, "buckets record search failed")
	}
	defer rows.Close()
	for rows.Next() {
		userDataBuf := []byte{}
		vnetDataBuf := []byte{}
		vnetPrivate := pb.VNetPrivate{}
		userPrivate := pb.ObjectUserPrivate{}
		eventType := ""
		currEvent := pb.BucketSubnetUpdateEvent{}

		if err := rows.Scan(&userDataBuf, &vnetDataBuf, &eventType); err != nil {
			// FIXME: Handle this failure
			logger.Info("error reading result row, continue...", logkeys.Error, err)
		}
		err := json.Unmarshal([]byte(userDataBuf), &userPrivate)
		if err != nil {
			logger.Error(err, "Error Unmarshalling JSON")
			return nil, status.Errorf(codes.Internal, "bucket subnet record search failed")
		}
		err = json.Unmarshal([]byte(vnetDataBuf), &vnetPrivate)
		if err != nil {
			logger.Error(err, "Error Unmarshalling JSON")
			return nil, status.Errorf(codes.Internal, "bucket subnet record search failed")
		}

		currEvent.EventType = mapEventSqlToPb(eventType)
		currEvent.Vnet = &vnetPrivate
		if userPrivate.Status != nil && userPrivate.Status.Principal != nil {
			principal := pb.BucketPrincipal{
				ClusterId:   userPrivate.Status.Principal.Cluster.ClusterId,
				PrincipalId: userPrivate.Status.Principal.PrincipalId,
				Spec:        userPrivate.Spec,
			}
			currEvent.Principals = append(currEvent.Principals, &principal)
		}
		subnetEvents = append(subnetEvents, &currEvent)
	}

	return subnetEvents, nil
}

func UpdateStatusForSubnet(ctx context.Context, tx *sql.Tx, subnetStatus *pb.BucketSubnetStatusUpdateRequest) error {
	logger := log.FromContext(ctx).WithName("UpdateStatusForSubnet")
	logger.Info("begin bucket subnet update status request", logkeys.ResourceId, subnetStatus.ResourceId)

	result, err := tx.ExecContext(ctx,
		updateStatusForSubnet,
		mapEventPbToSql(subnetStatus.Status),
		time.Now(),
		subnetStatus.ResourceId,
		subnetStatus.CloudacccountId,
		subnetStatus.VNetName,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	logger.Info("debug info", logkeys.NumAffectedRows, rowsAffected)

	if rowsAffected < 1 {
		return status.Error(codes.FailedPrecondition, "no records updated; possible update conflict")
	}

	return nil
}

func DeleteBucketSubnetFromDB(ctx context.Context, tx *sql.Tx, subnetStatus *pb.BucketSubnetStatusUpdateRequest) error {
	logger := log.FromContext(ctx).WithName("DeleteBucketSubnetFromDB")
	logger.Info("begin bucket subnet delete status request", logkeys.ResourceId, subnetStatus.ResourceId)

	result, err := tx.ExecContext(ctx,
		deleteBucketSubnetFromDB,
		subnetStatus.CloudacccountId,
		subnetStatus.VNetName,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	logger.Info("debug", logkeys.NumAffectedRows, rowsAffected)

	if rowsAffected < 1 {
		return status.Error(codes.FailedPrecondition, "no records updated; possible update conflict")
	}

	return nil
}

func GetAllSubnetEvents(ctx context.Context, tx *sql.Tx, resourceVersion int64) ([]*SubnetEvent, error) {
	logger := log.FromContext(ctx).WithName("GetAllSubnetEvents")

	subnetEvents := []*SubnetEvent{}

	rows, err := tx.QueryContext(ctx, getAllSubnetEvents, resourceVersion)
	if err != nil {
		logger.Error(err, "error searching subnet record in db")
		return nil, status.Errorf(codes.Internal, "subnet search record failed")
	}
	defer rows.Close()
	eventMapByAccount := make(map[string]*SubnetEvent)
	for rows.Next() {
		event := SubnetEvent{}
		dataBuf := []byte{}
		vnetPrivate := pb.VNetPrivate{}
		eventType := ""
		cloudAccountId := ""
		resourceVersion := int64(0)
		if err := rows.Scan(&cloudAccountId, &eventType, &resourceVersion, &dataBuf); err != nil {
			logger.Error(err, "Error reading the row")
			return nil, status.Errorf(codes.Internal, "subnet record search failed")
		}
		err := json.Unmarshal([]byte(dataBuf), &vnetPrivate)
		if err != nil {
			logger.Error(err, "Error Unmarshalling JSON")
			return nil, status.Errorf(codes.Internal, "subnet record search failed")
		}
		event.EventType = mapEventSqlToPb(eventType)
		event.VNet = &vnetPrivate
		event.ResourceVerion = resourceVersion
		eventMapByAccount[cloudAccountId] = &event
	}
	for _, event := range eventMapByAccount {
		subnetEvents = append(subnetEvents, event)
	}

	return subnetEvents, nil
}

func mapEventSqlToPb(event string) pb.BucketSubnetEventStatus {
	switch event {
	case AddingEventType:
		return pb.BucketSubnetEventStatus_E_ADDING
	case DeletingEventType:
		return pb.BucketSubnetEventStatus_E_DELETING
	case AddedEventType:
		return pb.BucketSubnetEventStatus_E_ADDED
	case DeletedEventType:
		return pb.BucketSubnetEventStatus_E_DELETED
	default:
		return pb.BucketSubnetEventStatus_E_UNSPECIFIED
	}
}

func mapEventPbToSql(status pb.BucketSubnetEventStatus) string {
	switch status {
	case pb.BucketSubnetEventStatus_E_ADDED:
		return AddedEventType
	case pb.BucketSubnetEventStatus_E_ADDING:
		return AddingEventType
	case pb.BucketSubnetEventStatus_E_DELETED:
		return DeletedEventType
	case pb.BucketSubnetEventStatus_E_DELETING:
		return DeletingEventType
	default:
		return "UNSPECIFIED"
	}
}
