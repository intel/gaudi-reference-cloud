// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/database/query"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/utils"
	idcutils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FilesystemServiceServer is used to implement pb.UnimplementedFileStorageServiceServer
type FilesystemServiceServer struct {
	pb.UnimplementedFileStorageServiceServer
	pb.UnimplementedFilesystemPrivateServiceServer
	pb.UnimplementedFilesystemOrgPrivateServiceServer
	pb.UnimplementedFilesystemStorageClusterPrivateServiceServer

	session                     *sql.DB
	cloudAccountServiceClient   v1.CloudAccountServiceClient
	productcatalogServiceClient v1.ProductCatalogServiceClient
	authzServiceClient          v1.AuthzServiceClient
	schedulerClient             v1.FilesystemSchedulerPrivateServiceClient
	kmsClient                   v1.StorageKMSPrivateServiceClient
	quotaServiceClient          *QuotaService
	userClient                  v1.FilesystemUserPrivateServiceClient
	wekaAgentClient             v1.WekaStatefulAgentPrivateServiceClient
	mu                          sync.Mutex
	fileProduct                 fileProductInfo
	deleteFilesystemMutex       DeleteFilesystemMutex
	gpVASTEnabled               bool
	cfg                         *Config
}

type DeleteFilesystemMutex struct {
	mu  sync.Mutex
	cfg *Config
}

const (
	kErrUniqueViolation  = "23505"
	timestampInfinityStr = "infinity"
)

type fileProductInfo struct {
	MinSize          int64
	MaxSize          int64
	UpdatedTimestamp time.Time
}

func NewFilesystemService(ctx context.Context, session *sql.DB, cloudAccountSvc v1.CloudAccountServiceClient,
	productcatalogSvc v1.ProductCatalogServiceClient, authzSvc v1.AuthzServiceClient,
	schedulerClient v1.FilesystemSchedulerPrivateServiceClient, kmsClient v1.StorageKMSPrivateServiceClient,
	quotaServiceClient *QuotaService,
	userClient v1.FilesystemUserPrivateServiceClient,
	wekaClient v1.WekaStatefulAgentPrivateServiceClient, cfg *Config) (*FilesystemServiceServer, error) {
	if session == nil {
		return nil, fmt.Errorf("db session is required")
	}

	fsSrv := FilesystemServiceServer{
		session:                     session,
		cloudAccountServiceClient:   cloudAccountSvc,
		schedulerClient:             schedulerClient,
		kmsClient:                   kmsClient,
		productcatalogServiceClient: productcatalogSvc,
		authzServiceClient:          authzSvc,
		quotaServiceClient:          quotaServiceClient,
		userClient:                  userClient,
		wekaAgentClient:             wekaClient,
		gpVASTEnabled:               cfg.GeneralPurposeVASTEnabled,
		cfg:                         cfg,
	}

	fsSrv.fileProduct = fileProductInfo{}
	if err := updateAvailableFileSizes(ctx, productcatalogSvc, &fsSrv.fileProduct); err != nil {
		return nil, fmt.Errorf("failed to update storage product details")
	}

	return &fsSrv, nil
}

func (fs *FilesystemServiceServer) Create(ctx context.Context, in *pb.FilesystemCreateRequest) (*pb.Filesystem, error) {
	//initialize logger and start trace span
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.Create").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()

	defer logger.Info("returning from filesystem creation")

	privateReq := convertFilesystemCreatePublicToPrivate(in)

	fsPrivate, err := fs.createFilesystem(ctx, privateReq)
	if err != nil {
		return nil, err
	}
	return convertFilesystemPrivateToPublic(fsPrivate, false), nil
}

func (fs *FilesystemServiceServer) createFilesystem(ctx context.Context, in *pb.FilesystemCreateRequestPrivate) (*pb.FilesystemPrivate, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.createFilesystem").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering filesystem creation")
	defer logger.Info("returning from filesystem creation")
	// if len(fs.fileProduct.availableSizes) == 0 {
	// 	if err := updateAvailableFileSizes(ctx, fs.productcatalogServiceClient, &fs.fileProduct); err != nil {
	// 		return nil, fmt.Errorf("failed to update storage product details")
	// 	}
	// }
	//in.Metadata.SkipProductCheck = true
	if err := isValidFilesystemCreateRequest(ctx, in, fs.fileProduct, in.Metadata.SkipProductCheck); err != nil {
		return nil, err
	}

	cloudAccountId := in.Metadata.CloudAccountId
	if err := cloudaccount.CheckValidId(cloudAccountId); err != nil {
		return nil, err
	}

	dbSession := fs.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}

	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "error startind db tx.")
	}
	defer tx.Rollback()

	reqSize := utils.ParseFileSize(in.Spec.Request.Storage)
	if reqSize == -1 {
		logger.Info("invalid input size arguments", "reqSize", in.Spec.Request.Storage)
		return nil, status.Error(codes.InvalidArgument, "invalid storage size")
	}

	if !in.Metadata.SkipQuotaCheck {
		// Check quota for this cloudaccount
		cloudAccount, err := fs.cloudAccountServiceClient.GetById(ctx, &v1.CloudAccountId{Id: in.Metadata.CloudAccountId})
		if err != nil {
			logger.Error(err, "error querying cloudaccount")
			return nil, status.Errorf(codes.FailedPrecondition, "error querying cloudaccount")
		}
		if !fs.quotaServiceClient.checkAndUpdateFileQuota(ctx, cloudAccount, reqSize, false) {
			logger.Info("quota check fail")
			return nil, status.Errorf(codes.FailedPrecondition, "quota check failed")
		}
	}

	// Get current cluster resources (cluster,namespace) assigned to
	// this cloudaccount
	fs.mu.Lock()
	defer fs.mu.Unlock()
	currAssign, err := query.GetFilesystemAccounts(ctx, tx, in.Metadata.CloudAccountId, timestampInfinityStr)
	if err != nil {
		logger.Info("error reading current resource assignments for the user")
		return nil, status.Errorf(codes.FailedPrecondition, "user account retrieval failed.")
	}

	fsName := in.Metadata.Name
	fsPrivate := pb.FilesystemPrivate{
		Metadata: &pb.FilesystemMetadataPrivate{
			CloudAccountId:    in.Metadata.CloudAccountId,
			ResourceId:        uuid.NewString(),
			Name:              in.Metadata.Name,
			Description:       in.Metadata.Description,
			CreationTimestamp: timestamppb.Now(),
			SkipQuotaCheck:    in.Metadata.SkipQuotaCheck,
			SkipProductCheck:  in.Metadata.SkipProductCheck,
		},
		Spec: &pb.FilesystemSpecPrivate{
			Request: &pb.FilesystemCapacity{
				Storage: in.Spec.Request.Storage,
			},
			Encrypted:        true, // Enabled by default
			FilesystemType:   in.Spec.FilesystemType,
			StorageClass:     in.Spec.StorageClass,
			AvailabilityZone: in.Spec.AvailabilityZone,
			Prefix:           in.Spec.Prefix,
			Scheduler:        &pb.FilesystemSchedule{},
		},
	}

	fsReq := pb.FilesystemScheduleRequest{}
	fsReq.CloudaccountId = in.Metadata.CloudAccountId
	fsReq.Assignments = currAssign.Assignments
	fsReq.RequestSpec = in.Spec
	fsReq.RequestSpec.StorageClass = in.Spec.StorageClass
	fsSched, err := fs.scheduleFilestorage(ctx, &fsReq)
	if err != nil {
		logger.Error(err, "error scheduling filesystem to cluster")
		fs.quotaServiceClient.decFileQuota(ctx, in.Metadata.CloudAccountId, reqSize, false)
		return nil, status.Errorf(codes.Internal, "filesystem scheduling failed")
	}
	fsPrivate.Spec.Scheduler = &pb.FilesystemSchedule{}
	fsPrivate.Status = &pb.FilesystemStatusPrivate{}

	if fsPrivate.Spec.StorageClass == pb.FilesystemStorageClass_GeneralPurpose || fsPrivate.Spec.StorageClass == pb.FilesystemStorageClass_AIOptimized {
		nsCredentialsKMSPath := utils.GenerateKMSPath(in.Metadata.CloudAccountId, fsSched.Schedule.ClusterUUID, false)
		userCredentialsKMSPath := utils.GenerateKMSPath(in.Metadata.CloudAccountId, fsSched.Schedule.ClusterUUID, true)
		nsAdminUser := utils.GenerateFilesystemNamespaceUser(in.Metadata.CloudAccountId)

		if fsSched.NewSchedule {
			var nsAdminPwd string
			nsAdminPwd, err = utils.GenerateRandomPassword()
			if err != nil {
				logger.Error(err, "error generating ns admin password")
			}

			nsCreds := map[string]string{
				"username": nsAdminUser,
				"password": nsAdminPwd,
			}

			if err := fs.storeSecretsToStorageKMS(ctx,
				nsCredentialsKMSPath, userCredentialsKMSPath,
				nsCreds); err != nil {
				logger.Error(err, "error storing filesystem account credentials into kms")
				fs.quotaServiceClient.decFileQuota(ctx, in.Metadata.CloudAccountId, reqSize, false)
				return nil, status.Errorf(codes.Internal, "kms transaction failed")
			}
		}
		fsPrivate.Spec.Scheduler.Namespace = &pb.AssignedNamespace{
			Name:            nsAdminUser,
			CredentialsPath: nsCredentialsKMSPath,
		}

	} else if fsPrivate.Spec.StorageClass == pb.FilesystemStorageClass_GeneralPurposeStd {
		fsPrivate.Spec.Scheduler.Namespace = &pb.AssignedNamespace{
			Name: utils.GenerateVASTNamespaceName(in.Metadata.CloudAccountId),
		}
		fsPrivate.Spec.VolumePath = in.Spec.VolumePath
		vnetPrivate, err := query.GetBucketSubnetByAccount(ctx, tx, in.Metadata.CloudAccountId)
		if err != nil {
			// soft handling of the error here
			logger.Info("error reading bucket subnet from database")
		}
		if vnetPrivate != nil {
			fsPrivate.Spec.SecurityGroup = &v1.VolumeSecurityGroup{
				NetworkFilterAllow: []*v1.VolumeNetworkGroup{{
					Subnet:       vnetPrivate.Spec.Subnet,
					PrefixLength: vnetPrivate.Spec.PrefixLength,
					Gateway:      vnetPrivate.Spec.Gateway,
				},
				},
			}
		}
	}

	fsPrivate.Spec.Scheduler.Cluster = &pb.AssignedCluster{
		ClusterName:    fsSched.Schedule.ClusterName,
		ClusterAddr:    fsSched.Schedule.ClusterAddr,
		ClusterUUID:    fsSched.Schedule.ClusterUUID,
		ClusterVersion: &fsSched.Schedule.ClusterVersion,
	}

	if fsSched.NewSchedule {
		// store filesystem account information if new account is created
		if err := query.StoreFilesystemAccount(ctx, tx, in.Metadata.CloudAccountId, fsPrivate.Spec.Scheduler); err != nil {
			logger.Error(err, "error storing filesystem schedule into db")
			fs.quotaServiceClient.decFileQuota(ctx, in.Metadata.CloudAccountId, reqSize, false)
			return nil, status.Errorf(codes.Internal, "database transaction failed")
		}
	}
	// store filesystem request
	if err := query.StoreFilesystemRequest(ctx, tx, &fsPrivate); err != nil {
		fs.quotaServiceClient.decFileQuota(ctx, in.Metadata.CloudAccountId, reqSize, false)
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) && pgErr.Code == kErrUniqueViolation {
			return nil, status.Error(codes.AlreadyExists, "insert: filesystem name "+fsName+" already exists")
		}
		logger.Error(err, "error storing filesystem request into db")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}
	if err := tx.Commit(); err != nil {
		logger.Error(err, "error commiting transaction")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}

	return &fsPrivate, nil
}

func (fs *FilesystemServiceServer) Get(ctx context.Context, in *pb.FilesystemGetRequest) (*pb.Filesystem, error) {

	if idcutils.IsValidCloudAccountId(in.Metadata.CloudAccountId) {
		return nil, status.Error(codes.InvalidArgument, "invalid cloudaccount for getting filesystem")
	}
	inPrivate := (*pb.FilesystemGetRequestPrivate)(in)
	fsPrivate, err := fs.getFilesystem(ctx, inPrivate)
	if err != nil {
		return nil, err
	}
	fsPublic := convertFilesystemPrivateToPublic(fsPrivate, true)

	if fs.cfg.AuthzEnabled {
		_, err := fs.lookupResourcePermission(ctx, in.Metadata.CloudAccountId, "get", []string{fsPrivate.Metadata.ResourceId}, false)
		if err != nil {
			return nil, err
		}
	}

	return fsPublic, nil
}
func (fs *FilesystemServiceServer) getFilesystem(ctx context.Context, in *pb.FilesystemGetRequestPrivate) (*pb.FilesystemPrivate, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.getFilesystem").
		WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId, logkeys.ResourceId, in.Metadata.GetResourceId()).Start()
	defer span.End()

	logger.Info("entering filesystem get")
	defer logger.Info("returning from filesystem get")
	if err := isValidFilesystemGetRequest(ctx, in.Metadata); err != nil {
		return nil, err
	}
	dbSession := fs.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "error starting db tx.")
	}

	fsPrivate := &pb.FilesystemPrivate{}
	if in.Metadata.GetResourceId() != "" {
		fsPrivate, err = query.GetFilesystemByResourceId(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetResourceId(), timestampInfinityStr)
	} else if in.Metadata.GetName() != "" {
		fsPrivate, err = query.GetFilesystemByName(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetName(), timestampInfinityStr)
	}
	if err != nil {
		return nil, err
	}

	// Populate cluster information
	if err := fs.populateClusterInfo(ctx, fsPrivate); err != nil {
		logger.Error(err, "error populating cluster info in get filesystem")
	}

	return fsPrivate, nil
}

func (fs *FilesystemServiceServer) Search(ctx context.Context, in *pb.FilesystemSearchRequest) (*pb.FilesystemSearchResponse, error) {
	//initialize logger and start trace span
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.Search").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()

	if idcutils.IsValidCloudAccountId(in.Metadata.CloudAccountId) {
		return nil, status.Error(codes.InvalidArgument, "invalid cloudaccount for searching filesystem")
	}
	if err := isValidFilesystemSearchRequest(ctx, in); err != nil {
		return nil, err
	}

	dbSession := fs.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}

	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "error startind db tx.")
	}
	defer tx.Rollback()
	fsPrivateList, err := query.GetFilesystemsByCloudaccountId(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.FilterType, timestampInfinityStr)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "error reading filesystems")
	}

	authzResourceIds := []string{}
	authzResourceIdMap := make(map[string]bool)
	if fs.cfg.AuthzEnabled {
		for idx := 0; idx < len(fsPrivateList); idx++ {
			authzResourceIds = append(authzResourceIds, fsPrivateList[idx].Metadata.ResourceId)
		}
		authzLookupResponse, err := fs.lookupResourcePermission(ctx, in.Metadata.CloudAccountId, "get", authzResourceIds, true)
		if err != nil {
			logger.Error(err, "error authz lookup")
			return nil, err
		}
		for _, resourceId := range authzLookupResponse.ResourceIds {
			authzResourceIdMap[resourceId] = true
		}
	}

	resp := pb.FilesystemSearchResponse{}

	for idx := 0; idx < len(fsPrivateList); idx++ {
		// this means that authz didn't find permission for this resourceId so will skip it in the response
		if _, exists := authzResourceIdMap[fsPrivateList[idx].Metadata.ResourceId]; fs.cfg.AuthzEnabled && !exists {
			continue
		}
		fsPublic := convertFilesystemPrivateToPublic(fsPrivateList[idx], true)
		resp.Items = append(resp.Items, fsPublic)
	}
	return &resp, nil
}

func (fs *FilesystemServiceServer) Update(ctx context.Context, in *pb.FilesystemUpdateRequest) (*pb.Filesystem, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.Update").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId,
		logkeys.ResourceId, in.Metadata.GetResourceId()).Start()
	defer span.End()

	logger.Info("entering filesystem Update for ", "cloudaccountid", in.Metadata.CloudAccountId)
	defer logger.Info("returning from filesystem Update for ", "cloudaccountid", in.Metadata.CloudAccountId)

	if idcutils.IsValidCloudAccountId(in.Metadata.CloudAccountId) {
		return nil, status.Error(codes.InvalidArgument, "invalid cloudaccount for updating filesystem")
	}
	resourceId := in.Metadata.GetResourceId()
	// update by name
	if resourceId == "" {
		inPrivateReq := &pb.FilesystemGetRequestPrivate{
			Metadata: &pb.FilesystemMetadataReference{
				CloudAccountId: in.Metadata.CloudAccountId,
				NameOrId: &pb.FilesystemMetadataReference_Name{
					Name: in.Metadata.GetName(),
				},
			},
		}
		// fetch fs to get resourceId
		fsPrivate, err := fs.getFilesystem(ctx, inPrivateReq)
		if err != nil {
			return nil, err
		}
		// set resourceId
		resourceId = fsPrivate.Metadata.ResourceId
	}

	if fs.cfg.AuthzEnabled {
		_, err := fs.lookupResourcePermission(ctx, in.Metadata.CloudAccountId, "update", []string{resourceId}, false)
		if err != nil {
			return nil, err
		}
	}
	privateReq := convertFilesystemUpdatePublicToPrivate(in)

	fsPrivate, err := fs.update(ctx, privateReq)
	if err != nil {
		return nil, err
	}

	return convertFilesystemPrivateToPublic(fsPrivate, false), nil
}

func (fs *FilesystemServiceServer) GetUser(ctx context.Context, in *pb.FilesystemGetUserRequest) (*pb.FilesystemGetUserResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.GetUser").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId,
		logkeys.ResourceId, in.Metadata.GetResourceId()).Start()
	defer span.End()
	logger.Info("entering filesystem getUser")
	defer logger.Info("returning from filesystem getUser")
	inPrivateReq := (*pb.FilesystemGetRequestPrivate)(in)
	fsPrivate, err := fs.getFilesystem(ctx, inPrivateReq)
	if err != nil {
		return nil, err
	}
	if fs.cfg.AuthzEnabled {
		_, err := fs.lookupResourcePermission(ctx, in.Metadata.CloudAccountId, "getuser", []string{fsPrivate.Metadata.GetResourceId()}, false)
		if err != nil {
			return nil, err
		}
	}

	inPrivate := (*pb.FilesystemGetUserRequestPrivate)(in)
	creds, err := fs.getUser(ctx, inPrivate)
	if err != nil {
		return nil, err
	}
	resp := (*pb.FilesystemGetUserResponse)(creds)
	return resp, nil
}

func (fs *FilesystemServiceServer) getUser(ctx context.Context, in *pb.FilesystemGetUserRequestPrivate) (*pb.FilesystemGetUserResponsePrivate, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.getUser").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId,
		logkeys.ResourceId, in.Metadata.GetResourceId()).Start()
	defer span.End()

	logger.Info("entering filesystem getUser")
	defer logger.Info("returning from filesystem getUser")
	if err := isValidFilesystemGetUserRequest(ctx, in.Metadata); err != nil {
		return nil, err
	}

	// first check if that filesystem exists
	dbSession := fs.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "error startind db tx.")
	}

	fsPrivate := &pb.FilesystemPrivate{}
	if in.Metadata.GetResourceId() != "" {
		fsPrivate, err = query.GetFilesystemByResourceId(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetResourceId(), timestampInfinityStr)
	} else if in.Metadata.GetName() != "" {
		fsPrivate, err = query.GetFilesystemByName(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetName(), timestampInfinityStr)
	}
	if err != nil {
		return nil, err
	}
	if fsPrivate == nil {
		return nil, status.Errorf(codes.Internal, "error retrieving filesystem record")
	}

	nsUserPwd, err := utils.GenerateRandomPassword()
	if err != nil {
		logger.Error(err, "error generating ns user password")
	}

	logger.Info("details for debugging", logkeys.FSScheduler, fsPrivate.Spec.Scheduler)
	userName := utils.GenerateFilesystemUser(in.Metadata.CloudAccountId)
	requestParams := pb.FilesystemUserCreateOrUpdateRequest{
		ClusterUUID:        fsPrivate.Spec.Scheduler.Cluster.ClusterUUID,
		NamespaceName:      fsPrivate.Spec.Scheduler.Namespace.Name,
		NamespaceCredsPath: fsPrivate.Spec.Scheduler.Namespace.CredentialsPath,
		UserName:           userName,
		NewUserPassword:    nsUserPwd,
	}

	tries := 3
	for tries > 0 {
		_, err = fs.userClient.CreateOrUpdate(ctx, &requestParams)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				logger.Error(err, "ContextDeadline Exceeded or cancelled: true", logkeys.Message, err.Error())
				tries -= 1
				time.Sleep(2 * time.Second)
				continue
			} else {
				logger.Error(err, "error creating or updating user")
				return nil, status.Errorf(codes.Internal, "error updating user")
			}
		} else {
			break
		}
	}
	logger.Info("user created or updated succesfully", logkeys.UserName, userName)
	resp := pb.FilesystemGetUserResponsePrivate{
		User:     userName,
		Password: nsUserPwd,
	}
	return &resp, nil
}

func (fs *FilesystemServiceServer) getUserPrivate(ctx context.Context, in *pb.FilesystemGetUserRequestPrivate) (*pb.FilesystemGetUserResponsePrivate, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.getUserPrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId,
		logkeys.ResourceId, in.Metadata.GetResourceId()).Start()
	defer span.End()

	if err := isValidFilesystemGetUserRequest(ctx, in.Metadata); err != nil {
		return nil, err
	}

	// first check if that filesystem exists
	dbSession := fs.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "error startind db tx.")
	}

	fsPrivate := &pb.FilesystemPrivate{}
	if in.Metadata.GetName() != "" {
		fsPrivate, err = query.GetFilesystemByName(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetName(), timestampInfinityStr)
	} else {
		return nil, status.Errorf(codes.Internal, "error name not found")
	}
	if fsPrivate == nil {
		return nil, status.Errorf(codes.Internal, "error retrieving filesystem record")
	}

	// check if volume is ready, only then we can create the user

	if fsPrivate.Status == nil || fsPrivate.Status.Phase != pb.FilesystemPhase_FSReady {
		return nil, status.Errorf(codes.FailedPrecondition, "volume is not ready")
	}
	fsUserPwd, err := utils.GenerateRandomPassword()
	if err != nil {
		logger.Error(err, "error generating vast fs user password")
		return nil, status.Errorf(codes.Internal, "error creating user")
	}
	userName := utils.GenerateFilesystemUser(in.Metadata.CloudAccountId) + "-" + in.Metadata.GetName() //one user per fs name
	requestParams := pb.FilesystemUserCreateOrGetRequest{
		ClusterUUID:   fsPrivate.Spec.Scheduler.Cluster.ClusterUUID,
		NamespaceName: fsPrivate.Spec.Scheduler.Namespace.Name,
		UserCredsPath: "staas/tenant/" + in.Metadata.CloudAccountId + "/" + fsPrivate.Spec.Scheduler.Cluster.ClusterUUID + "/" + userName,
		UserName:      userName,
		Password:      fsUserPwd,
		NamespaceId:   strconv.FormatInt(fsPrivate.Status.VolumeIdentifiers.TenantId, 10),
	}
	var response *pb.FilesystemUserResponsePrivate

	tries := 3
	for tries > 0 {
		response, err = fs.userClient.CreateOrGet(ctx, &requestParams)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				logger.Error(err, "ContextDeadline Exceeded or cancelled: true", logkeys.Message, err.Error())
				tries -= 1
				time.Sleep(2 * time.Second)
				continue
			} else {
				logger.Error(err, "error creating or updating user")
				return nil, status.Errorf(codes.Internal, "error updating user")
			}
		} else {
			break
		}
	}
	logger.Info("user created or fetched succesfully", logkeys.UserName, userName)
	resp := pb.FilesystemGetUserResponsePrivate{
		User:     response.User,
		Password: response.Password,
	}
	return &resp, nil
}

func (fs *FilesystemServiceServer) Delete(ctx context.Context, in *pb.FilesystemDeleteRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.Delete").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId,
		logkeys.ResourceId, in.Metadata.GetResourceId()).Start()
	defer span.End()
	logger.Info("entering filesystem Delete")
	defer logger.Info("returning from filesystem Delete")

	if idcutils.IsValidCloudAccountId(in.Metadata.CloudAccountId) {
		return nil, status.Error(codes.InvalidArgument, "invalid cloudaccount for deleting filesystem")
	}

	if fs.cfg.AuthzEnabled {
		_, err := fs.lookupResourcePermission(ctx, in.Metadata.CloudAccountId, "delete", []string{in.Metadata.GetResourceId()}, false)
		if err != nil {
			return nil, err
		}
	}

	inPrivate := (*pb.FilesystemDeleteRequestPrivate)(in)
	return fs.deleteFilesystem(ctx, inPrivate, true)
}

func (fs *FilesystemServiceServer) deleteFilesystem(ctx context.Context, in *pb.FilesystemDeleteRequestPrivate, quotaReclaim bool) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.deleteFilesystem").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId,
		logkeys.ResourceId, in.Metadata.GetResourceId()).Start()
	defer span.End()
	fs.mu.Lock()
	defer fs.mu.Unlock()
	logger.Info("entering filesystem delete")
	defer logger.Info("returning from filesystem delete")
	if err := isValidFilesystemDeleteRequest(ctx, in.Metadata); err != nil {
		return nil, err
	}

	dbSession := fs.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	txOptions := &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	}

	tx, err := dbSession.BeginTx(ctx, txOptions)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "error startind db tx.")
	}
	defer tx.Rollback()
	fsPrivate := &pb.FilesystemPrivate{}
	if in.Metadata.GetResourceId() != "" {
		fsPrivate, err = query.GetFilesystemByResourceId(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetResourceId(), timestampInfinityStr)
	} else if in.Metadata.GetName() != "" {
		fsPrivate, err = query.GetFilesystemByName(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetName(), timestampInfinityStr)
	}
	if err != nil {
		return nil, err
	}

	// Check if filesystem is in provisioning state,
	// filesystem can not be deleted in `provisioning` state
	if fsPrivate != nil && fsPrivate.Status != nil && fsPrivate.Status.Phase == pb.FilesystemPhase_FSProvisioning {
		logger.Info("Filesystem is in provisioning state, can not be deleted")
		return &emptypb.Empty{}, status.Errorf(codes.FailedPrecondition, "filesystem in provisioning state can not be deleted")
	}

	// Check if the deletionTimestamp is set
	if fsPrivate != nil && fsPrivate.Metadata != nil && fsPrivate.Metadata.DeletionTimestamp != nil {
		logger.Info("Filesystem has a deletion timestamp set already")
		// Handle the case where the filesystem has a deletion timestamp set already to avoid double deletion
		// This is a no-op and returning success and not error
		return &emptypb.Empty{}, nil
	}
	reclaimedSize, err := query.UpdateFilesystemForDeletion(ctx, tx, in.Metadata)
	if err != nil {
		logger.Error(err, "error updating filesystem for deletion")
		return &emptypb.Empty{}, err
	}
	// we will not be updating quota for private apis
	// quota set and reclaim is only for public requests
	if quotaReclaim {
		if ok := fs.quotaServiceClient.decFileQuota(ctx, in.Metadata.CloudAccountId, reclaimedSize, false); !ok {
			logger.Info("quota update on deletion failed", logkeys.Request, in)
		} else {
			logger.Info("quota updated to reclaim deletion", logkeys.Request, in)
		}
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err, "error commiting db for filesystem_namespace_user_account transaction")
		return nil, status.Errorf(codes.Internal, "db transaction failed for filesystem_namespace_user_account")
	}

	if fs.cfg.AuthzEnabled {
		if fsPrivate == nil || fsPrivate.Metadata == nil {
			return nil, status.Errorf(codes.Internal, "missing fsPrivate metadata during Authz check")
		}
		_, err := fs.authzServiceClient.RemoveResourceFromCloudAccountRole(ctx,
			&v1.CloudAccountRoleResourceRequest{
				CloudAccountId: in.Metadata.CloudAccountId,
				ResourceId:     fsPrivate.Metadata.ResourceId,
				ResourceType:   "filestorage"})
		if err != nil {
			logger.Error(err, "error authz when removing resource from cloudAccountRole", "resourceId", fsPrivate.Metadata.ResourceId, "resourceType", "filestorage")
			return nil, status.Errorf(codes.Internal, "storage %v couldn't be removed from cloudAccountRole resourceType %v",
				fsPrivate.Metadata.ResourceId, "filestorage")
		}
	}

	return &emptypb.Empty{}, nil
}

func (fs *FilesystemServiceServer) Ping(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("FilesystemServiceServer.Ping")

	logger.Info("entering filesystem Ping")
	defer logger.Info("returning from filesystem Ping")
	return &emptypb.Empty{}, nil
}

func (fs *FilesystemServiceServer) scheduleFilestorage(ctx context.Context,
	currAssign *pb.FilesystemScheduleRequest) (*pb.FilesystemScheduleResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.getFilesystemTarget").WithValues(logkeys.CloudAccountId, currAssign.CloudaccountId).Start()
	defer span.End()
	logger.Info("discovering filesystem target")

	schedResp, err := fs.schedulerClient.ScheduleFile(ctx, currAssign)
	if err != nil {
		logger.Error(err, "error scheduling request")
		return nil, fmt.Errorf("error scheduling request")
	}
	logger.Info("scheduler response", logkeys.Response, schedResp)
	return schedResp, nil
}

func (fs *FilesystemServiceServer) UpdateStatus(ctx context.Context, in *pb.FilesystemUpdateStatusRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.UpdateStatus").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId,
		logkeys.ResourceId, in.Metadata.GetResourceId()).Start()
	defer span.End()

	logger.Info("entering filesystem Update", logkeys.Status, in.Status)
	defer logger.Info("returning from filesystem Update")
	if idcutils.IsValidCloudAccountId(in.Metadata.CloudAccountId) {
		return nil, status.Error(codes.InvalidArgument, "invalid cloudaccount for updating filesystem status")
	}
	dbSession := fs.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "error startind db tx.")
	}
	defer tx.Rollback()

	if err := query.UpdateFilesystemState(ctx, tx, in); err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err, "error commiting transaction")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}

	if in.Status.Phase == pb.FilesystemPhase_FSFailed {
		getReqPrivate := pb.FilesystemGetRequestPrivate{
			Metadata: &v1.FilesystemMetadataReference{
				CloudAccountId: in.Metadata.CloudAccountId,
				NameOrId: &v1.FilesystemMetadataReference_Name{
					// ResourceId is used as name
					Name: in.Metadata.ResourceId,
				},
			},
		}
		fsPrivate, err := fs.getFilesystem(ctx, &getReqPrivate)
		if err != nil {
			logger.Error(err, "failed to read the filesystem, skipping quota update")
			return &emptypb.Empty{}, nil
		}
		if fsPrivate.Metadata.ClientType == pb.APIClientTypes_Private {
			logger.Info("skipping quota update for private filesystem")
			return &emptypb.Empty{}, nil
		}
		reclaimedSize := utils.ParseFileSize(fsPrivate.Spec.Request.Storage)
		if reclaimedSize == -1 {
			logger.Info("invalid size argument", "reclaimedSize", fsPrivate.Spec.Request.Storage)
			return nil, status.Error(codes.InvalidArgument, "invalid storage size")
		}
		if ok := fs.quotaServiceClient.decFileQuota(ctx, in.Metadata.CloudAccountId, reclaimedSize, false); !ok {
			logger.Info("quota update on failed state for filesystem", logkeys.ResourceId, in.Metadata.ResourceId)
		} else {
			logger.Info("quota updated to reclaim failed filesystem", logkeys.ResourceId, in.Metadata.ResourceId)
		}
	}
	return &emptypb.Empty{}, nil
}

func (fs *FilesystemServiceServer) SearchFilesystemRequests(in *pb.FilesystemSearchStreamPrivateRequest, rs pb.FilesystemPrivateService_SearchFilesystemRequestsServer) error {
	dbSession := fs.session
	if dbSession == nil {
		return status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	tx, err := dbSession.BeginTx(rs.Context(), nil)
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "error startind db tx.")
	}
	defer tx.Rollback()

	return query.GetFilesystemsRequests(tx, in.ResourceVersion, timestampInfinityStr, rs)
}

func (fs *FilesystemServiceServer) RemoveFinalizer(ctx context.Context, in *pb.FilesystemRemoveFinalizerRequest) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.RemoveFinalizers").
		WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId, logkeys.ResourceId, in.Metadata.GetResourceId()).Start()
	defer span.End()
	logger.Info("entering filesystem remove finalizer")
	defer logger.Info("returning from filesystem remove finalizer ")
	dbSession := fs.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}
	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "error startind db tx.")
	}
	defer tx.Rollback()

	fsPrivate, err := query.GetFilesystemByName(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetResourceId(), timestampInfinityStr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error fetching filesystem from db")
	}
	if err := query.UpdateFilesystemDeletionTime(ctx, tx, in.Metadata.CloudAccountId,
		in.Metadata.ResourceId); err != nil {
		logger.Error(err, "error updating deletion timestamp")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}

	logger.Info("filesystem deletion timestamp updated", "fsPrivate object is ", fsPrivate)

	orgExists := false
	//fsSched.Schedule.ClusterUUID is the cluster id
	requestOrg := &pb.FilesystemOrgsIsExistsRequestPrivate{
		ClusterId: fsPrivate.Spec.Scheduler.Cluster.ClusterUUID,
		Name:      fsPrivate.Spec.Scheduler.Namespace.Name,
	}
	resp, err := fs.schedulerClient.IsOrgExists(ctx, requestOrg)
	if err != nil {
		//If the SDS fails with 502, we assume that org exists in the backend and not delete the last entry from the table, since defensive is better.
		logger.Error(err, "error querying org in backend, silently proceeding")
		orgExists = true //defensive to not delete entry in the table
	} else {
		orgExists = resp.Exists
	}

	if !orgExists {
		logger.Info("Org doesn't exist in the backend, deleting the last entry from the table")
		// Delete entry from filesystem_namespace_user_account table as this is the last active filesystem or iks namespace for the user.
		if err := query.DeleteFilesystemAccount(ctx, tx, in.Metadata.CloudAccountId, fsPrivate.Spec.Scheduler.Cluster.ClusterUUID, timestampInfinityStr); err != nil {
			logger.Error(err, "error during deleting filesystem namespace account entry for cluster")
			return nil, status.Errorf(codes.Internal, "db transaction failed")
		}
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err, "error commiting transaction")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}

	return &emptypb.Empty{}, nil
}

func (fs *FilesystemServiceServer) PingPrivate(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("FilesystemServiceServer.Ping")

	logger.Info("entering filesystem private Ping")
	defer logger.Info("returning from filesystem private Ping")
	return &emptypb.Empty{}, nil
}

func (fs *FilesystemServiceServer) populateClusterInfo(ctx context.Context, fsPrivate *pb.FilesystemPrivate) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.populateClusterInfo").WithValues(logkeys.CloudAccountId, fsPrivate.Metadata.CloudAccountId).Start()
	defer span.End()

	// Get cluster information using ListClusters API
	listClustersReq := &pb.ListClusterRequest{}
	clusterStream, err := fs.schedulerClient.ListClusters(ctx, listClustersReq)
	if err != nil {
		logger.Error(err, "error querying cluster info")
		return err
	}

	// Iterate over the cluster stream to find the relevant cluster
	for {
		clusterResp, err := clusterStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Error(err, "error receiving cluster info")
			return err
		}
		// Check if the cluster matches the one in the filesystem spec
		if clusterResp.Labels != nil && len(clusterResp.Labels) > 0 && clusterResp.Labels["VASTBackend"] == fsPrivate.Spec.Scheduler.Cluster.ClusterAddr {
			if fsPrivate.Status.ClusterInfo == nil {
				fsPrivate.Status.ClusterInfo = make(map[string]string)
			}
			fsPrivate.Status.ClusterInfo = clusterResp.Labels
			break
		}
	}

	return nil
}

func (fs *FilesystemServiceServer) CreatePrivate(ctx context.Context, in *pb.FilesystemCreateRequestPrivate) (*pb.FilesystemPrivate, error) {
	// Initialize logger and start trace span
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.CreatePrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering filesystem create private")
	defer logger.Info("returning from filesystem create private")

	// Process request size
	if in == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "input cannot be empty")
	}
	if in != nil && in.Spec != nil && in.Spec.Request != nil && in.Spec.Request.Storage != "" {
		in.Spec.Request.Storage = utils.ProcesSize(in.Spec.Request.Storage)
	}

	filesystemPrivate, err := fs.createFilesystem(ctx, in)
	if err != nil {
		return nil, err
	}

	// Populate cluster information
	if err := fs.populateClusterInfo(ctx, filesystemPrivate); err != nil {
		logger.Error(err, "error populating cluster info")
	}

	return filesystemPrivate, nil
}

func (fs *FilesystemServiceServer) GetPrivate(ctx context.Context, in *pb.FilesystemGetRequestPrivate) (*pb.FilesystemPrivate, error) {
	return fs.getFilesystem(ctx, in)
}

func (fs *FilesystemServiceServer) DeletePrivate(ctx context.Context, in *pb.FilesystemDeleteRequestPrivate) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.DeletePrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()

	logger.Info("entering filesystem delete private")
	defer logger.Info("returning from filesystem delete private")
	return fs.deleteFilesystem(ctx, in, false)
}

func (fs *FilesystemServiceServer) GetUserPrivate(ctx context.Context, in *pb.FilesystemGetUserRequestPrivate) (*pb.FilesystemGetUserResponsePrivate, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.GetUserPrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()

	logger.Info("entering filesystem getUser private")
	defer logger.Info("returning from filesystem getUser private")
	resp, err := fs.getUser(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (fs *FilesystemServiceServer) CreateorGetUserPrivate(ctx context.Context, in *pb.FilesystemGetUserRequestPrivate) (*pb.FilesystemGetUserResponsePrivate, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.CreateorGetUserPrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()

	logger.Info("entering filesystem CreateorGetUserPrivate private")
	defer logger.Info("returning from filesystem CreateorGetUserPrivate private")
	resp, err := fs.getUserPrivate(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (fs *FilesystemServiceServer) DeleteUserPrivate(ctx context.Context, in *pb.FilesystemDeleteUserRequestPrivate) (*emptypb.Empty, error) {
	ctx, _, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.DeleteUserPrivate").WithValues(logkeys.CloudAccountId, in.CloudaccountId).Start()
	defer span.End()

	_, err := fs.userClient.Delete(ctx, in)
	if err != nil {
		if strings.Contains(err.Error(), "user details not found on vault") {
			return nil, status.Errorf(codes.FailedPrecondition, "user details not found on vault")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete user")
	}
	return &emptypb.Empty{}, nil
}

func (fs *FilesystemServiceServer) update(ctx context.Context, in *pb.FilesystemUpdateRequestPrivate) (*pb.FilesystemPrivate, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.update").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()

	logger.Info("entering filesystem update")

	defer logger.Info("returning from filesystem update")
	if err := isValidFilesystemUpdateRequest(ctx, in, fs.fileProduct, in.Metadata.SkipProductCheck); err != nil {
		return nil, err
	}
	dbSession := fs.session
	if dbSession == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "no database connection found.")
	}

	tx, err := dbSession.BeginTx(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "error startind db tx.")
	}
	defer tx.Rollback()

	fsPrivate := &pb.FilesystemPrivate{}
	if in.Metadata.GetResourceId() != "" {
		fsPrivate, err = query.GetFilesystemByResourceId(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetResourceId(), timestampInfinityStr)
	} else if in.Metadata.GetName() != "" {
		fsPrivate, err = query.GetFilesystemByName(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.GetName(), timestampInfinityStr)
	}

	if err != nil {
		return nil, err
	}

	reqSize := utils.ParseFileSize(in.Spec.Request.Storage)
	if reqSize == -1 {
		logger.Info("invalid input size arguments", "reqSize", in.Spec.Request.Storage)
		return nil, status.Error(codes.InvalidArgument, "invalid storage size")
	}

	existingSize := utils.ParseFileSize(fsPrivate.Spec.Request.Storage)
	if existingSize == -1 {
		logger.Info("invalid size argument", "existingSize", fsPrivate.Spec.Request.Storage)
		return nil, status.Error(codes.InvalidArgument, "invalid storage size")
	}

	sizeDiff := (reqSize - existingSize)

	if sizeDiff <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "only size extension is allowed for file storage")
	}
	if !fsPrivate.Metadata.SkipQuotaCheck {
		// Check quota for this cloudaccount
		logger.Info("quota check inside update")
		cloudAccount, err := fs.cloudAccountServiceClient.GetById(ctx, &v1.CloudAccountId{Id: in.Metadata.CloudAccountId})
		if err != nil {
			logger.Error(err, "error querying cloudaccount")
			return nil, status.Errorf(codes.FailedPrecondition, "error querying cloudaccount")
		}
		if !fs.quotaServiceClient.checkAndUpdateFileQuota(ctx, cloudAccount, sizeDiff, true) {
			logger.Info("quota check fail")
			return nil, status.Errorf(codes.FailedPrecondition, "quota check failed")
		}
	}

	fsPrivate.Spec.Request.Storage = in.Spec.Request.Storage
	fsPrivate.Metadata.UpdateTimestamp = timestamppb.Now()
	// store filesystem request
	if err := query.UpdateFilesystemRequest(ctx, tx, fsPrivate); err != nil {
		fs.quotaServiceClient.decFileQuota(ctx, in.Metadata.CloudAccountId, sizeDiff, true)
		logger.Error(err, "error storing updated record to db")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err, "error commiting transaction")
		return nil, status.Errorf(codes.Internal, "database transaction failed")
	}

	return fsPrivate, nil

}

func (fs *FilesystemServiceServer) UpdatePrivate(ctx context.Context, in *pb.FilesystemUpdateRequestPrivate) (*pb.FilesystemPrivate, error) {
	ctx, _, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.UpdatePrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	return fs.update(ctx, in)
}
