package server

import (
	"context"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func (fs *FilesystemServiceServer) CreateFilesystemOrgPrivate(ctx context.Context, in *pb.FilesystemOrgCreateRequestPrivate) (*pb.FilesystemOrgPrivate, error) {
	// initialize logger and start trace span
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.CreateFilesystemOrgPrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering filesystem org creation", "with input : ", in)
	defer logger.Info("returning from filesystem org creation")
	if in.Spec == nil || in.Spec.Prefix == "" {
		return nil, status.Errorf(codes.InvalidArgument, "prefix is required")
	}
	privateReq := pb.FilesystemCreateRequestPrivate{
		Metadata: in.Metadata,
		Spec:     in.Spec,
	}
	privateReq.Spec.FilesystemType = pb.FilesystemType_ComputeKubernetes
	privateReq.Spec.StorageClass = in.Spec.StorageClass
	privateReq.Spec.Prefix = in.Spec.Prefix
	// Process request size
	if in.Spec != nil && in.Spec.Request != nil && in.Spec.Request.Storage != "" {
		in.Spec.Request.Storage = utils.ProcesSize(in.Spec.Request.Storage)
	}
	fsPrivate, err := fs.createFilesystem(ctx, &privateReq)
	if err != nil {
		return nil, err
	}

	orgPrivate := pb.FilesystemOrgPrivate{
		Metadata: fsPrivate.Metadata,
		Spec:     fsPrivate.Spec,
		Status:   fsPrivate.Status,
	}
	return &orgPrivate, nil
}

func (fs *FilesystemServiceServer) UpdateFilesystemOrgPrivate(ctx context.Context, in *pb.FilesystemOrgUpdateRequestPrivate) (*pb.FilesystemOrgPrivate, error) {
	// initialize logger and start trace span
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.UpdateFilesystemOrgPrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering filesystem org update", "with input : ", in)
	defer logger.Info("returning from filesystem org update")
	if in.Spec == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Spec is required")
	}
	privateReq := pb.FilesystemUpdateRequestPrivate{
		Metadata: in.Metadata,
		Spec:     in.Spec,
	}
	privateReq.Spec.FilesystemType = pb.FilesystemType_ComputeKubernetes
	privateReq.Spec.StorageClass = in.Spec.StorageClass
	privateReq.Spec.Prefix = in.Spec.Prefix
	// Process request size
	if in.Spec != nil && in.Spec.Request != nil && in.Spec.Request.Storage != "" {
		in.Spec.Request.Storage = utils.ProcesSize(in.Spec.Request.Storage)
	}
	fsPrivate, err := fs.update(ctx, &privateReq)
	if err != nil {
		return nil, err
	}

	orgPrivate := pb.FilesystemOrgPrivate{
		Metadata: fsPrivate.Metadata,
		Spec:     fsPrivate.Spec,
		Status:   fsPrivate.Status,
	}
	return &orgPrivate, nil
}

func (fs *FilesystemServiceServer) GetFilesystemOrgPrivate(ctx context.Context, in *pb.FilesystemOrgGetRequestPrivate) (*pb.FilesystemOrgPrivate, error) {
	// initialize logger and start trace span
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.GetFilesystemOrgPrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("entering filesystem org get")
	defer logger.Info("returning from filesystem org get")
	inPrivate := pb.FilesystemGetRequestPrivate{
		Metadata: in.Metadata,
	}
	fsPrivate, err := fs.getFilesystem(ctx, &inPrivate)
	if err != nil {
		return nil, err
	}
	orgPrivate := pb.FilesystemOrgPrivate{
		Metadata: fsPrivate.Metadata,
		Spec:     fsPrivate.Spec,
		Status:   fsPrivate.Status,
	}
	return &orgPrivate, nil
}

func (fs *FilesystemServiceServer) DeleteFilesystemOrgPrivate(ctx context.Context, in *pb.FilesystemOrgDeleteRequestPrivate) (*emptypb.Empty, error) {
	// initialize logger and start trace span
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.DeleteFilesystemOrgPrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()

	logger.Info("entering filesystem org delete")
	defer logger.Info("returning from filesystem org delete")

	if in.Prefix == "" {
		return nil, status.Errorf(codes.InvalidArgument, "prefix is required")
	}
	deletePrivateRequest := pb.FilesystemDeleteRequestPrivate{
		Metadata: in.Metadata,
	}
	return fs.deleteFilesystem(ctx, &deletePrivateRequest, false)
}

func (fs *FilesystemServiceServer) ListFilesystemsInOrgPrivate(ctx context.Context, in *pb.FilesystemsInOrgListRequestPrivate) (*pb.FilesystemsInOrgListResponsePrivate, error) {
	// initialize logger and start trace span
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.ListFilesystemsInOrgPrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()

	nsCredentialsKMSPath := utils.GenerateKMSPath(in.Metadata.CloudAccountId, in.ClusterId, false)

	var name string
	switch x := in.Metadata.NameOrId.(type) {
	case *pb.FilesystemMetadataReference_Name:
		name = x.Name
	}
	fsQuery := &pb.FilesystemInOrgGetRequestPrivate{
		NamespaceCredsPath: nsCredentialsKMSPath,
		Name:               name,
		ClusterId:          in.ClusterId,
		CloudAccountId:     in.Metadata.CloudAccountId,
	}

	fsList, err := fs.schedulerClient.ListFilesystemInOrgs(ctx, fsQuery)
	if err != nil {
		logger.Error(err, "error listing filesystem in org")
		return nil, status.Errorf(codes.Internal, "filesystem fetch in org failed")
	}

	// Filter fsList to include only filesystems with a specific prefix
	filteredList := []*pb.FilesystemPrivate{}
	for _, fs := range fsList.Items {
		if strings.HasPrefix(fs.Metadata.Name, in.Prefix) {
			filteredList = append(filteredList, fs)
		}
	}
	response := &pb.FilesystemsInOrgListResponsePrivate{
		Items: filteredList,
	}

	defer logger.Info("returning from filesystems in org list")
	return response, nil
}

func (fs *FilesystemServiceServer) ListFilesystemOrgsPrivate(ctx context.Context, in *pb.FilesystemOrgsListRequestPrivate) (*pb.FilesystemOrgsResponsePrivate, error) {
	// initialize logger and start trace span
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("FilesystemServiceServer.ListFilesystemOrgsPrivate").WithValues(logkeys.CloudAccountId, in.Metadata.CloudAccountId).Start()
	defer span.End()

	fsQuery := &pb.FilesystemOrgsGetRequestPrivate{
		ClusterId: in.ClusterId,
	}

	orgList, err := fs.schedulerClient.ListFilesystemOrgs(ctx, fsQuery)
	if err != nil {
		logger.Error(err, "error list filesystem orgs")
		return nil, status.Errorf(codes.Internal, "error list filesystem orgs")
	}

	// Filter orgList to include only orgs with a specific prefix
	filteredList := []*pb.FilesystemOrgsPrivate{}
	for _, org := range orgList.Org {
		if strings.HasPrefix(org.Name, in.Prefix) {
			filteredList = append(filteredList, org)
		}
	}
	response := &pb.FilesystemOrgsResponsePrivate{
		Org: filteredList,
	}
	defer logger.Info("returning from filesystems orgs list")

	return response, nil

}
