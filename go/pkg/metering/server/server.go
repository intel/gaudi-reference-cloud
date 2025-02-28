// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/db/query"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/utils/strings/slices"
)

const InvalidCloudAccountId = "INVALID_CLOUD_ACCOUNT_ID"

// MeteringServer is used to implement helloworld.GreeterServer.
type MeteringServer struct {
	v1.UnsafeMeteringServiceServer
	session           *sql.DB
	validServiceTypes []string
}

func NewMeteringService(session *sql.DB, validServiceTypes []string) (*MeteringServer, error) {
	if session == nil {
		return nil, fmt.Errorf("db session is required")
	}
	return &MeteringServer{
		session: session,
		// productcatalog and billing enabled serviceTypes
		validServiceTypes: validServiceTypes,
	}, nil
}

func (srv *MeteringServer) CreateInvalidRecords(ctx context.Context, createInvalidMeeringRecords *v1.CreateInvalidMeteringRecords) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("MeteringServer.CreateInvalidRecords").Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	log.Info("Entering create invalid metering records")
	defer log.Info("Returning from create invalid metering records")
	dbSession := srv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found for creation of invalid metering records")
	}

	// todo: think of incorporation of partial success
	for _, createInvalidMeteringRecord := range createInvalidMeeringRecords.CreateInvalidMeteringRecords {
		if utils.IsValidRecordId(string(createInvalidMeteringRecord.RecordId)) ||
			utils.IsValidCloudAccountId(createInvalidMeteringRecord.CloudAccountId) ||
			utils.IsValidResourceId(createInvalidMeteringRecord.ResourceId) ||
			utils.IsValidTransactionId(createInvalidMeteringRecord.TransactionId) ||
			utils.IsValidRegion(createInvalidMeteringRecord.Region) ||
			(createInvalidMeteringRecord.Timestamp != nil && utils.IsValidTimestamp(createInvalidMeteringRecord.Timestamp)) {
			log.Info("invalid input arguments, ignoring invalid records creation")
			return nil, status.Errorf(codes.InvalidArgument, "invalid input arguments, ignoring invalid records creation")
		}
	}

	// todo: handle partial failures.
	for _, createInvalidMeteringRecord := range createInvalidMeeringRecords.CreateInvalidMeteringRecords {
		if err := query.CreateInvalidMeteringRec(ctx, dbSession, createInvalidMeteringRecord); err != nil {
			log.Error(err, "failed to create invalid metering record")
			return nil, err
		}
	}

	return &emptypb.Empty{}, nil
}

func (srv *MeteringServer) Create(ctx context.Context, r *v1.UsageCreate) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("MeteringServer.Create").
		WithValues("cloudAccountId", r.GetCloudAccountId(), "resourceId", r.GetResourceId(), "transactionId", r.GetTransactionId()).
		Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	dbSession := srv.session
	if dbSession == nil {
		err := status.Errorf(codes.FailedPrecondition, "no database connection found")
		log.Error(err, "no database connection found")
		return &emptypb.Empty{}, nil
		//return nil, status.Errorf(codes.FailedPrecondition, "no database connection found")
	}

	if utils.IsValidCloudAccountId(r.CloudAccountId) ||
		utils.IsValidResourceId(r.ResourceId) ||
		utils.IsValidTransactionId(r.TransactionId) ||
		utils.IsValidTimestamp(r.Timestamp) ||
		r.Properties == nil {
		err := status.Errorf(codes.InvalidArgument, "invalid input arguments")
		log.Error(err, "invalid input arguments, ignoring record creation")
		return &emptypb.Empty{}, nil
		//return nil, status.Errorf(codes.InvalidArgument, "invalid input arguments, ignoring record creation")
	}

	// todo: these will move to mandatory arguments.
	const serviceTypeKey = "serviceType"
	const regionKey = "region"

	if _, foundRegionKey := r.Properties[regionKey]; !foundRegionKey {
		err := status.Errorf(codes.InvalidArgument, "missing region, ignoring record creation")
		log.Error(err, "missing region, ignoring record creation")

		if err := query.CreateInvalidUsageRecord(ctx, dbSession, r); err != nil {
			log.Error(err, "failed to record invalid metering record for missing region")
			return &emptypb.Empty{}, nil
			//return nil, status.Errorf(codes.Internal, "failed to record invalid metering record for missing region")
		}

		return &emptypb.Empty{}, nil
		//return nil, status.Errorf(codes.InvalidArgument, "missing region, ignoring record creation")
	} else {
		var regionRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]*$`)

		if r.Properties[regionKey] == "" || !regionRegex.MatchString(r.Properties[regionKey]) {
			if err := query.CreateInvalidUsageRecord(ctx, dbSession, r); err != nil {
				log.Error(err, "failed to record invalid metering record for invalid region value", "region", r.Properties[regionKey])
				//return nil, status.Errorf(codes.Internal, "failed to record invalid metering record for invalid region value")
				return &emptypb.Empty{}, nil
			}
			err := status.Errorf(codes.InvalidArgument, "invalid region value '%s', ignoring record creation", r.Properties[regionKey])
			log.Error(err, "invalid region value, ignoring record creation", "region", r.Properties[regionKey])
			return &emptypb.Empty{}, nil
			//return nil, status.Errorf(codes.InvalidArgument, "invalid region value, ignoring record creation")
		}
	}

	// have to disable this as time metric is not standardized.
	/**if _, foundRunningSecondsKey := r.Properties[runningSecondsKey]; !foundRunningSecondsKey {
		log.Info("missing running seconds, ignoring record creation")
		if err := query.CreateInvalidUsageRecord(ctx, dbSession, r); err != nil {
			return &emptypb.Empty{}, nil
			//return nil, status.Errorf(codes.Internal, "failed to record invalid metering record for missing running seconds")
		}
		return &emptypb.Empty{}, nil
		//return nil, status.Errorf(codes.InvalidArgument, "missing running seconds, ignoring record creation")
	} else {
		if r.Properties[runningSecondsKey] == "" {
			if err := query.CreateInvalidUsageRecord(ctx, dbSession, r); err != nil {
				return &emptypb.Empty{}, nil
				//return nil, status.Errorf(codes.Internal, "failed to record invalid metering record for invalid running seconds value")
			}
			return &emptypb.Empty{}, nil
			//return nil, status.Errorf(codes.InvalidArgument, "invalid running seconds value, ignoring record creation")
		}
	}**/

	// check metering record properties for invalid serviceType
	if r.Properties != nil {
		if r.Properties[serviceTypeKey] != "" {
			if !slices.Contains(srv.validServiceTypes, r.Properties["serviceType"]) {
				if err := query.CreateInvalidUsageRecord(ctx, dbSession, r); err != nil {
					log.Error(err, "failed to record invalid metering record for service type", "serviceType", r.Properties["serviceType"])
					return &emptypb.Empty{}, nil
					//return nil, status.Errorf(codes.Internal, "failed to record invalid metering record for service type")
				}
				err := status.Errorf(codes.InvalidArgument, "service type '%s' not supported", r.Properties["serviceType"])
				log.Error(err, "service type not supported, ignoring record creation", "serviceType", r.Properties["serviceType"])
				return &emptypb.Empty{}, nil
				//return nil, status.Errorf(codes.Internal, "service type not supported")
			}
		}
	}

	if err := query.CreateUsageRecord(ctx, dbSession, r); err != nil {
		log.Error(err, "error creating metering record")
		return &emptypb.Empty{}, nil
		//return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (srv *MeteringServer) SearchInvalid(f *v1.InvalidMeteringRecordFilter, rs v1.MeteringService_SearchInvalidServer) error {
	log := log.FromContext(rs.Context()).WithName("MeteringServer.SearchInvalid")

	log.Info("entering metering invalid record search ")
	defer log.Info("returning from metering invalid record search")
	dbSession := srv.session
	if dbSession == nil {
		return status.Errorf(codes.FailedPrecondition, "no database connection found")
	}

	if (f.CloudAccountId != nil && utils.IsValidCloudAccountId(*f.CloudAccountId)) ||
		(f.RecordId != nil && utils.IsValidRecordId(*f.RecordId)) ||
		(f.Region != nil && utils.IsValidRegion(*f.Region)) ||
		(f.ResourceId != nil && utils.IsValidResourceId(*f.ResourceId)) ||
		(f.TransactionId != nil && utils.IsValidTransactionId(*f.TransactionId)) ||
		(f.StartTime != nil && utils.IsValidTimestamp(f.StartTime)) ||
		(f.EndTime != nil && utils.IsValidTimestamp(f.EndTime)) {
		log.Info("invalid input arguments, ignoring invalid record search")
		return status.Errorf(codes.InvalidArgument, "invalid input arguments, ignoring invalid record search")
	}
	if err := query.SearchInvalid(rs, dbSession, f); err != nil {
		log.Error(err, "error completing invalid record search query with input filters")
		return err
	}

	return nil
}

func (srv *MeteringServer) Search(f *v1.UsageFilter, rs v1.MeteringService_SearchServer) error {
	_, log, span := obs.LogAndSpanFromContextOrGlobal(rs.Context()).WithName("MeteringServer.Search").
		WithValues("cloudAccountId", f.GetCloudAccountId(), "resourceId", f.GetResourceId(), "reported", f.GetReported(), "startTime", f.GetStartTime(), "endTime", f.GetEndTime()).
		Start()

	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	dbSession := srv.session
	if dbSession == nil {
		return status.Errorf(codes.FailedPrecondition, "no database connection found")
	}

	if (f.CloudAccountId != nil && utils.IsValidCloudAccountId(*f.CloudAccountId)) ||
		(f.ResourceId != nil && utils.IsValidResourceId(*f.ResourceId)) ||
		(f.TransactionId != nil && utils.IsValidTransactionId(*f.TransactionId)) ||
		(f.StartTime != nil && utils.IsValidTimestamp(f.StartTime)) ||
		(f.EndTime != nil && utils.IsValidTimestamp(f.EndTime)) {
		log.Info("invalid input arguments, ignoring record search")
		return status.Errorf(codes.InvalidArgument, "invalid input arguments, ignoring record search")
	}
	if err := query.Search(rs, dbSession, f); err != nil {
		log.Error(err, "error completing search query with input filters")
		return err
	}

	return nil
}

func (srv *MeteringServer) FindPrevious(ctx context.Context, in *v1.UsagePrevious) (*v1.Usage, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("MeteringServer.FindPrevious").Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	log.Info("FindPrevious", "entering metering record findPrevious for: ", in.ResourceId)
	defer log.Info("FindPrevious", "returning from metering record findPrevious for: ", in.ResourceId)
	dbSession := srv.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found")
	}

	if utils.IsValidResourceId(in.ResourceId) || in.Id == 0 {
		log.Info("invalid input arguments, ignoring find request")
		return nil, status.Errorf(codes.InvalidArgument, "invalid input arguments, ignoring find previous")
	}

	return query.FindPreviousRecord(ctx, dbSession, in.ResourceId, in.GetId())
}

func (srv *MeteringServer) Update(ctx context.Context, in *v1.UsageUpdate) (*emptypb.Empty, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("MeteringServer.Update").WithValues("usageId", in.GetId(), "reported", in.GetReported()).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	log.Info("Update", "entering metering record update for: ", in.Id)
	defer log.Info("Update", "returning from metering record update for: ", in.Id)
	dbSession := srv.session
	if dbSession == nil {
		return nil, fmt.Errorf("no database connection found")
	}
	if in.Id == nil || len(in.Id) == 0 || strconv.FormatBool(in.Reported) == "" {
		log.Info("invalid input arguments, ignoring record update")
		return &emptypb.Empty{}, status.Errorf(codes.InvalidArgument, "invalid input arguments, ignoring record update")
	}
	if err := query.UpdateUsageRecord(ctx, dbSession, in); err != nil {
		log.Error(err, "error updating metering record")
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

func (srv *MeteringServer) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("MeteringServer.Ping")
	log.Info("Ping")
	return &emptypb.Empty{}, nil
}

func (srv *MeteringServer) ValidateMeteringFilter(ctx context.Context, filter *v1.MeteringFilter) error {
	_, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("MeteringServer.ValidateMeteringFilter").
		WithValues("cloudAccountId", filter.GetCloudAccountId(), "reported", filter.GetReported(), "startTime", filter.GetStartTime(), "endTime", filter.GetEndTime()).
		Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	if (filter.CloudAccountId != nil && utils.IsValidCloudAccountId(*filter.CloudAccountId)) ||
		(filter.StartTime != nil && utils.IsValidTimestamp(filter.StartTime)) ||
		(filter.EndTime != nil && utils.IsValidTimestamp(filter.EndTime)) {
		log.Info("invalid filter")
		return errors.New("invalid filter")
	}

	return nil
}

// Eventually we need to change the usage apis in metering service as metering APIs.
// Following the existing approach to be changed later.
func (srv *MeteringServer) SearchResourceMeteringRecords(ctx context.Context, filter *v1.MeteringFilter) (*v1.ResourceMeteringRecordsList, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("MeteringServer.SearchResourceMeteringRecords").
		WithValues("cloudAccountId", filter.GetCloudAccountId(), "reported", filter.GetReported(), "startTime", filter.GetStartTime(), "endTime", filter.GetEndTime()).
		Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	dbSession := srv.session
	if dbSession == nil {
		err := status.Errorf(codes.FailedPrecondition, "no database connection found")
		log.Error(err, "no database connection found")
		return nil, err
	}

	if err := srv.ValidateMeteringFilter(ctx, filter); err != nil {
		log.Info("invalid input arguments, ignoring search")
		return nil, status.Errorf(codes.InvalidArgument, "invalid input arguments, ignoring search")
	}

	resourceMeteringRecords, err := query.GetResourceMeteringRecords(ctx, dbSession, filter)

	if err != nil {
		log.Error(err, "failed to get resource metering records")
		return nil, status.Errorf(codes.Internal, "failed to get metering records for cloud account")
	}

	return resourceMeteringRecords, nil
}

func (srv *MeteringServer) SearchResourceMeteringRecordsAsStream(filter *v1.MeteringFilter,
	rs v1.MeteringService_SearchResourceMeteringRecordsAsStreamServer) error {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(rs.Context()).WithName("MeteringServer.SearchResourceMeteringRecordsAsStream").
		WithValues("cloudAccountId", filter.GetCloudAccountId(), "reported", filter.GetReported(), "startTime", filter.GetStartTime(), "endTime", filter.GetEndTime()).
		Start()

	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	dbSession := srv.session
	if dbSession == nil {
		return status.Errorf(codes.FailedPrecondition, "no database connection found")
	}

	if err := srv.ValidateMeteringFilter(ctx, filter); err != nil {
		log.Info("invalid input arguments, ignoring search")
		return status.Errorf(codes.InvalidArgument, "invalid input arguments, ignoring search")
	}

	if err := query.GetResourceMeteringRecordsAsStream(rs, dbSession, filter); err != nil {
		log.Error(err, "error completing search query with input filters")
		return err
	}

	return nil
}

func (srv *MeteringServer) IsMeteringRecordAvailable(ctx context.Context, filter *v1.MeteringAvailableFilter) (*v1.MeteringAvailableResponse, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("MeteringServer.IsMeteringRecordAvailable").
		WithValues("cloudAccountId", filter.CloudAccountId).Start()
	defer span.End()
	log.V(9).Info("BEGIN")
	defer log.V(9).Info("END")

	dbSession := srv.session
	if dbSession == nil {
		err := status.Errorf(codes.FailedPrecondition, "no database connection found")
		log.Error(err, "no database connection found")
		return nil, err
	}
	if filter.MeteringDuration < 0 {
		log.Info("invalid input arguments, ignoring IsMeteringRecordAvailable")
		return nil, status.Errorf(codes.InvalidArgument, "invalid input arguments, ignoring IsMeteringRecordAvailable")
	}

	calculatedTimeFmt := timestamppb.New(time.Now().UTC().AddDate(0, int(filter.MeteringDuration)*-1, 0))
	if (filter.CloudAccountId != "" && utils.IsValidCloudAccountId(filter.CloudAccountId)) ||
		(calculatedTimeFmt != nil && utils.IsValidTimestamp(calculatedTimeFmt)) {
		log.Info("invalid input arguments, ignoring IsMeteringRecordAvailable")
		return nil, status.Errorf(codes.InvalidArgument, "invalid input arguments, ignoring IsMeteringRecordAvailable")
	}

	recordRecv, err := query.MeteringRecordAvailable(ctx, dbSession, filter.CloudAccountId, calculatedTimeFmt)
	if err != nil {
		log.Error(err, "error completing search query with input filters")
		return nil, err
	}
	log.Info("calculated time for metering records", "calculatedTimeFmt", calculatedTimeFmt.AsTime())
	if recordRecv.CloudAccountId == filter.CloudAccountId &&
		(recordRecv.Timestamp.AsTime().Equal(calculatedTimeFmt.AsTime()) || recordRecv.Timestamp.AsTime().After(calculatedTimeFmt.AsTime())) {
		log.Info("found matching record after given timestamp", "cloudAccountId", recordRecv.CloudAccountId, "timestamp", recordRecv.Timestamp.AsTime())
		return &pb.MeteringAvailableResponse{MeteringDataAvailable: true}, nil
	}
	log.Info("no matching record found after given timestamp", "cloudAccountId", recordRecv.CloudAccountId)
	return &pb.MeteringAvailableResponse{MeteringDataAvailable: false}, nil
}
