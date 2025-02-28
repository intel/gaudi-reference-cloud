// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/utils"
	idcutils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func isValidFilesystemDeleteRequest(ctx context.Context, md *pb.FilesystemMetadataReference) error {
	logger := log.FromContext(ctx).WithName("isValidFilesystemGetRequest")
	logger.Info("validating filesystem delete request")

	if md.CloudAccountId == "" || (md.GetName() == "" && md.GetResourceId() == "") {
		return status.Error(codes.InvalidArgument, "missing input arguments")
	}
	// Validate resourceId.
	if md.GetResourceId() != "" {
		if _, err := uuid.Parse(md.GetResourceId()); err != nil {
			return status.Error(codes.InvalidArgument, "invalid resourceId")
		}
	}
	if md.GetName() != "" {
		if err := utils.ValidateResourceName(md.GetName()); err != nil {
			return err
		}
	}
	return nil
}

func (fs *FilesystemServiceServer) readSecretsFromStorageKMS(ctx context.Context,
	secretKeyPath string) (map[string]string, error) {
	logger := log.FromContext(ctx).WithName("FilesystemServiceServer.readSecretsFromStorageKMS")
	logger.Info("calling storage kms service for get", logkeys.SecretsPath, secretKeyPath)

	request := pb.GetSecretRequest{
		KeyPath: secretKeyPath,
	}
	secretResp, err := fs.kmsClient.Get(ctx, &request)
	if err != nil {
		logger.Error(err, "error reading secrets from storage kms")
		return nil, fmt.Errorf("error reading secrets from storage kms")
	}
	return secretResp.Secrets, nil
}

func (fs *FilesystemServiceServer) storeSecretsToStorageKMS(ctx context.Context,
	nsCredsPath, userCredsPath string, nsCreds map[string]string) error {
	logger := log.FromContext(ctx).WithName("FilesystemServiceServer.storeSecretsToStorageKMS")

	nsStoreReq := pb.StoreSecretRequest{
		KeyPath: nsCredsPath,
		Secrets: nsCreds,
	}
	_, err := fs.kmsClient.Put(ctx, &nsStoreReq)
	if err != nil {
		logger.Error(err, "error storing namespace credentials to storage kms")
		return fmt.Errorf("error storing credentials to storage kms")
	}

	logger.Info("kms credentials stored successfully", logkeys.NsCredsPath, nsCredsPath, logkeys.UserCredsPath, userCredsPath)

	return nil
}

func isValidFilesystemUpdateRequest(ctx context.Context,
	request *pb.FilesystemUpdateRequestPrivate,
	storageProduct fileProductInfo,
	skipProductCheck bool) error {
	logger := log.FromContext(ctx).WithName("isValidFilesystemUpdateRequest")
	logger.Info("validating filesystem update request")

	// Validate input.
	if request.Metadata == nil {
		return status.Error(codes.InvalidArgument, "missing metadata")
	}
	if request.Spec == nil {
		return status.Error(codes.InvalidArgument, "missing spec")
	}
	if request.Metadata.Name == "" && request.Metadata.ResourceId == "" {
		return status.Error(codes.InvalidArgument, "missing input arguments")
	}
	if request.Metadata.Name != "" {
		if err := utils.ValidateResourceName(request.Metadata.Name); err != nil {
			return err
		}
	}

	if request.Spec.Request == nil ||
		request.Spec.Request.Storage == "" {
		return status.Error(codes.InvalidArgument, "missing specs arguments")
	}

	reqSize := utils.ParseFileSize(request.Spec.Request.Storage)
	if reqSize == -1 {
		logger.Info("invalid input size arguments", logkeys.Size, request.Spec.Request.Storage)
		return status.Error(codes.InvalidArgument, "invalid storage size")
	}

	unit := utils.ParseUnit(request.Spec.Request.Storage)
	if !skipProductCheck {
		if reqSize < storageProduct.MinSize*unit || reqSize > storageProduct.MaxSize*unit {
			logger.Info("invalid input size arguments", logkeys.Size, request.Spec.Request.Storage)
			return status.Error(codes.InvalidArgument, "invalid storage size is outside allowed range")
		}
	}
	return nil
}

func isValidFilesystemCreateRequest(ctx context.Context,
	request *pb.FilesystemCreateRequestPrivate,
	storageProduct fileProductInfo,
	skipProductCheck bool) error {
	logger := log.FromContext(ctx).WithName("isValidFilesystemCreateRequest")
	logger.Info("validating filesystem create request")

	// Validate input.
	if request.Metadata == nil {
		return status.Error(codes.InvalidArgument, "missing metadata")
	}
	if request.Spec == nil {
		return status.Error(codes.InvalidArgument, "missing spec")
	}

	// Calculate name if not provided.
	if request.Metadata.Name == "" {
		return status.Error(codes.InvalidArgument, "missing name")
	}
	name := request.Metadata.Name

	if err := utils.ValidateResourceName(name); err != nil {
		return err
	}

	if err := hasValidRequestParams(request.Metadata.CloudAccountId, request.Spec.AvailabilityZone); err != nil {
		return err
	}

	if request.Spec.Request == nil ||
		request.Spec.Request.Storage == "" {
		return status.Error(codes.InvalidArgument, "missing specs arguments")
	}

	reqSize := utils.ParseFileSize(request.Spec.Request.Storage)
	if reqSize == -1 {
		logger.Info("invalid input size arguments", logkeys.Size, request.Spec.Request.Storage)
		return status.Error(codes.InvalidArgument, "invalid storage size")
	}
	unit := utils.ParseUnit(request.Spec.Request.Storage)
	if !skipProductCheck {
		if reqSize < storageProduct.MinSize*unit || reqSize > storageProduct.MaxSize*unit {
			logger.Info("invalid input size arguments", logkeys.Size, request.Spec.Request.Storage)
			return status.Error(codes.InvalidArgument, "invalid storage size is outside allowed range")
		}
	}
	return nil
}

func isValidFilesystemSearchRequest(ctx context.Context, request *pb.FilesystemSearchRequest) error {
	logger := log.FromContext(ctx).WithName("isValidFilesystemSearchRequest")
	logger.Info("validating filesystem get request")

	if request.Metadata.CloudAccountId == "" {
		return status.Error(codes.InvalidArgument, "missing metadata")
	}
	return nil
}

func isValidFilesystemGetRequest(ctx context.Context, md *pb.FilesystemMetadataReference) error {
	logger := log.FromContext(ctx).WithName("isValidFilesystemGetRequest")
	logger.Info("validating filesystem get request")

	if md.CloudAccountId == "" || (md.GetName() == "" && md.GetResourceId() == "") {
		return status.Error(codes.InvalidArgument, "missing input arguments")
	}
	// Validate resourceId.
	if md.GetResourceId() != "" {
		if _, err := uuid.Parse(md.GetResourceId()); err != nil {
			return status.Error(codes.InvalidArgument, "invalid resourceId")
		}
	}
	if md.GetName() != "" {
		if err := utils.ValidateResourceName(md.GetName()); err != nil {
			return err
		}
	}
	return nil
}

func updateAvailableFileSizes(ctx context.Context, producatcatalogClient v1.ProductCatalogServiceClient,
	fsProd *fileProductInfo) error {
	logger := log.FromContext(ctx).WithName("updateAvailableFileSizes")
	logger.Info("updating storage product configs")

	fileProductName := "storage-file"
	productFilter := pb.ProductFilter{
		Name: &fileProductName,
	}
	productResponse, err := producatcatalogClient.AdminRead(context.Background(), &productFilter)
	if err != nil {
		logger.Error(err, "error reading from product catalog")
		return err
	}
	logger.Info("product catalog response", logkeys.Response, productResponse)

	for _, product := range productResponse.Products {
		val, found := product.Metadata["volume.size.min"]
		if !found {
			return fmt.Errorf("product metadata not found")
		}
		min, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fmt.Errorf("error parsing product size")
		}
		fsProd.MinSize = min

		val, found = product.Metadata["volume.size.max"]
		if !found {
			return fmt.Errorf("product metadata not found")
		}
		max, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fmt.Errorf("error parsing product size")
		}
		fsProd.MaxSize = max
	}
	fsProd.UpdatedTimestamp = time.Now()

	return nil
}

func isValidFilesystemGetUserRequest(ctx context.Context, request *pb.FilesystemMetadataReference) error {
	logger := log.FromContext(ctx).WithName("isValidFilesystemGetUserRequest")
	logger.Info("validating filesystem get user request")

	if request.CloudAccountId == "" || (request.GetName() == "" && request.GetResourceId() == "") {
		return status.Error(codes.InvalidArgument, "missing input arguments")
	}
	// Validate resourceId.
	if request.GetResourceId() != "" {
		if _, err := uuid.Parse(request.GetResourceId()); err != nil {
			return status.Error(codes.InvalidArgument, "invalid resourceId")
		}
	}
	if request.GetName() != "" {
		if err := utils.ValidateResourceName(request.GetName()); err != nil {
			return err
		}
	}
	return nil
}

func (fs *FilesystemServiceServer) lookupResourcePermission(ctx context.Context, cloudaccount string, action string, resourceIds []string, search bool) (*v1.LookupResponse, error) {
	logger := log.FromContext(ctx).WithName("lookupResourceWithAuthz")
	logger.Info("check filesystem permission with Authz")

	email, enterpriseId, groups, err := grpcutil.ExtractEmailEnterpriseIDAndGroups(ctx)
	if err != nil {
		logger.Info("failed to extract token info from context")
	}
	if email == nil || enterpriseId == nil {
		cloudAccountInfo, err := fs.cloudAccountServiceClient.GetById(ctx, &v1.CloudAccountId{Id: cloudaccount})
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "required cloud account info could not be fetched")
		}
		email = &cloudAccountInfo.Name
		enterpriseId = &cloudAccountInfo.Oid
	}

	authzLookupResponse, err := fs.authzServiceClient.LookupInternal(ctx, &v1.LookupRequestInternal{
		CloudAccountId: cloudaccount,
		ResourceType:   "filestorage",
		Action:         action,
		ResourceIds:    resourceIds,
		User: &v1.UserIdentification{
			Email:        *email,
			EnterpriseId: *enterpriseId,
			Groups:       groups,
		},
	})

	if err != nil {
		logger.Error(err, "error authz lookup")
		return nil, status.Errorf(codes.Internal, "error authz lookup")
	}

	logger.Info("lookup", "action", action, "resource", "filestorage")
	logger.Info("request resourceIds", "sent", resourceIds)
	logger.Info("Response resourceIds", "allowed", authzLookupResponse.ResourceIds)

	if len(authzLookupResponse.ResourceIds) == 0 && !search {
		logger.Error(err, "error permission deny for this resource")
		return nil, status.Errorf(codes.PermissionDenied, "permission denied for this resource")
	}
	return authzLookupResponse, nil
}

func hasValidRequestParams(cloudAccountId, availabilityZone string) error {
	// TODO: change the name of IsValidCloudAccountId or logic inside that function for readability
	invalidCloudAccount := idcutils.IsValidCloudAccountId(cloudAccountId)
	if invalidCloudAccount {
		return status.Error(codes.InvalidArgument, "invalid cloudaccount")
	}
	validAvailibilityZone := idcutils.IsValidAvailibilityZone(availabilityZone)
	if !validAvailibilityZone {
		return status.Error(codes.InvalidArgument, "invalid availibility zone")
	}
	return nil
}
