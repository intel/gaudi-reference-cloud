package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller"
	stcnt_api "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// StorageUserServiceServer is used to implement pb.UnimplementedFileStorageServiceServer
type StorageUserServiceServer struct {
	pb.UnimplementedFilesystemUserPrivateServiceServer
	pb.UnimplementedBucketUserPrivateServiceServer
	pb.UnimplementedBucketLifecyclePrivateServiceServer

	kmsClient    pb.StorageKMSPrivateServiceClient
	strCntClient *storagecontroller.StorageControllerClient
}

func NewStorageUserServiceServer(kmsClient pb.StorageKMSPrivateServiceClient,
	strCntCli *storagecontroller.StorageControllerClient) (*StorageUserServiceServer, error) {
	if kmsClient == nil {
		return nil, fmt.Errorf("storage kms client is required")
	}
	return &StorageUserServiceServer{
		kmsClient:    kmsClient,
		strCntClient: strCntCli,
	}, nil
}

func (storageUser *StorageUserServiceServer) CreateOrUpdate(ctx context.Context, in *pb.FilesystemUserCreateOrUpdateRequest) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("StorageUserServiceServer.CreateOrUpdate")

	logger.Info("entering filesystem user createOrUpdate for namespace", logkeys.Namespace, in.NamespaceName)
	defer logger.Info("returning from user createOrUpdate for namespace", logkeys.Namespace, in.NamespaceName)

	newUser := storagecontroller.User{
		Metadata: storagecontroller.UserMetadata{
			UUID:          in.ClusterUUID,
			NamespaceName: in.NamespaceName,
		},
		Properties: storagecontroller.UserProperties{
			NewUser:         in.UserName,
			NewUserPassword: in.NewUserPassword,
		},
	}
	nsCreds, err := readSecretsFromStorageKMS(ctx, storageUser.kmsClient, in.NamespaceCredsPath)
	if err != nil {
		return nil, status.Error(codes.Internal, "error reading from storage kms")
	}

	if user, found := nsCreds["username"]; found {
		newUser.Metadata.NamespaceUser = user
	}
	if password, found := nsCreds["password"]; found {
		newUser.Metadata.NamespacePassword = password
	}
	logger.Info("namespace credentials retrieved successfully")
	exists, err := storageUser.strCntClient.IsUserExists(ctx, newUser)
	if err != nil {
		logger.Error(err, "error checking if user exists")
		return nil, status.Errorf(codes.Internal, "error checking user")
	}

	if !exists {
		logger.Info("user does not exists, creating a new", logkeys.UserName, newUser.Properties.NewUser)
		err := storageUser.strCntClient.CreateUser(ctx, newUser)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "error creating user")
		}
		logger.Info("user created successfully", logkeys.UserName, newUser.Properties.NewUser)
	} else {
		logger.Info("user already exists, updating password", logkeys.UserName, newUser.Properties.NewUser)
		updateParam := storagecontroller.UpdateUserpassword{
			NamespaceUser:     newUser.Metadata.NamespaceUser,
			NamespaceName:     newUser.Metadata.NamespaceName,
			NamespacePassword: newUser.Metadata.NamespacePassword,
			UUID:              newUser.Metadata.UUID,
			UsertoBeUpdated:   newUser.Properties.NewUser,
			NewPassword:       newUser.Properties.NewUserPassword,
		}
		err := storageUser.strCntClient.UpdateUserPassword(ctx, updateParam)
		if err != nil {
			logger.Error(err, "error creating a new user")
			return nil, status.Errorf(codes.Internal, "error creating user")
		}
		logger.Info("user password updated successfully", logkeys.UserName, newUser.Properties.NewUser)
	}
	return &emptypb.Empty{}, nil
}

func (storageUser *StorageUserServiceServer) CreateOrGet(ctx context.Context, in *pb.FilesystemUserCreateOrGetRequest) (*pb.FilesystemUserResponsePrivate, error) {
	logger := log.FromContext(ctx).WithName("StorageUserServiceServer.CreateOrGet")

	logger.Info("entering filesystem user CreateOrGet for namespace", logkeys.Namespace, in.NamespaceName)
	defer logger.Info("returning from user CreateOrGet for namespace", logkeys.Namespace, in.NamespaceName)

	newUser := storagecontroller.User{
		Metadata: storagecontroller.UserMetadata{
			UUID:          in.ClusterUUID,
			NamespaceName: in.NamespaceName,
			NamespaceId:   in.NamespaceId,
		},
		Properties: storagecontroller.UserProperties{
			NewUser:         in.UserName,
			NewUserPassword: in.Password,
		},
	}
	userCredResponse := pb.FilesystemUserResponsePrivate{}
	logger.Info("calling storage kms service for get at path :", logkeys.SecretsPath, in.UserCredsPath)
	userCreds, err := readSecretsFromStorageKMS(ctx, storageUser.kmsClient, in.UserCredsPath)
	if err != nil {
		if strings.Contains(err.Error(), "no metadata at") {
			logger.Info("metadata not found, will creating new secrets later", "keyPath", in.UserCredsPath)
		} else if strings.Contains(err.Error(), "value type assertion failed") {
			logger.Info("secrets not found, will creating new secrets later", "keyPath", in.UserCredsPath)
		} else {
			return nil, status.Error(codes.Internal, "error reading from storage kms inside create or get for vast")
		}
	}

	if user, userFound := userCreds["username"]; userFound {
		if password, passwordFound := userCreds["password"]; passwordFound {
			userCredResponse.User = user
			userCredResponse.Password = password
			logger.Info("user credentials retrieved successfully from vault")
			// Already exists, return the user credentials
			return &userCredResponse, nil
		}
	}

	userResp := &stcnt_api.CreateUserResponse{}
	userResp, err = storageUser.strCntClient.CreateVastUser(ctx, newUser)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error creating user")
	}
	logger.Info("user did not exists, created a new vast user with name :", logkeys.UserName, newUser.Properties.NewUser)

	// storing the user credentials in vault
	secret := map[string]string{
		"username":    newUser.Properties.NewUser,
		"password":    newUser.Properties.NewUserPassword,
		"namespaceId": userResp.User.Id.NamespaceId.Id,
		"userId":      userResp.User.Id.Id,
	}
	err = putSecretsInStorageKMS(ctx, storageUser.kmsClient, in.UserCredsPath, secret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error storing user credentials")
	}
	logger.Info("user credentials stored successfully in vault for user : ", logkeys.UserName, newUser.Properties.NewUser)
	userCredResponse.User = newUser.Properties.NewUser
	userCredResponse.Password = newUser.Properties.NewUserPassword
	return &userCredResponse, nil
}

func (storageUser *StorageUserServiceServer) Delete(ctx context.Context, in *pb.FilesystemDeleteUserRequestPrivate) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("StorageUserServiceServer.Delete")

	logger.Info("entering filesystem user Delete for user", logkeys.UserName, in.UserName)
	defer logger.Info("returning from user Delete for user", logkeys.UserName, in.UserName)

	credsPath := "staas/tenant/" + in.CloudaccountId + "/" + in.ClusterId + "/" + in.UserName
	logger.Info("calling storage kms service for delete at path :", logkeys.SecretsPath, credsPath)
	var userCreds map[string]string
	userCreds, err := readSecretsFromStorageKMS(ctx, storageUser.kmsClient, credsPath)
	if err != nil {
		if strings.Contains(err.Error(), "value type assertion failed") {
			logger.Info("secrets not found in the path", "keyPath", credsPath)
			return nil, status.Errorf(codes.FailedPrecondition, "user details not found on vault")
		}
		return nil, status.Error(codes.Internal, "Error reading secrets while delete credentials")
	}
	// Check if namespaceId exists and is a string
	namespaceId, ok := userCreds["namespaceId"]
	if !ok || namespaceId == "" {
		// Handle the case where namespaceId is empty or not present
		return nil, status.Error(codes.Internal, "Error no namespaceId in vault credentials")
	}
	userId, ok := userCreds["userId"]
	if !ok || userId == "" {
		return nil, status.Error(codes.Internal, "Error no userId in vault credentials")
	}
	deleteVastUserRequest := storagecontroller.DeleteUserData{
		ClusterUUID: in.ClusterId,
		NamespaceID: namespaceId,
		UserID:      userId,
	}
	if _, userFound := userCreds["username"]; userFound {

		err = storageUser.strCntClient.DeleteVastUser(ctx, deleteVastUserRequest)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "error deleting the vast user")
		}
		logger.Info("vast user deleted successfully, now deleting secrets from kms", logkeys.UserName, in.UserName)
		err = deleteSecretsFromStorageKMS(ctx, storageUser.kmsClient, credsPath)
		if err != nil {
			if strings.Contains(err.Error(), "value type assertion failed") {
				logger.Info("secrets not found in the path", "keyPath", credsPath)
				return nil, status.Errorf(codes.FailedPrecondition, "user details not found on vault")
			}
			return nil, status.Errorf(codes.Internal, "unable to delete secrets from vault")
		}
	} else {
		return nil, status.Errorf(codes.Internal, "username not found in kms storage")
	}

	logger.Info("user deleted successfully from vault and sds", logkeys.UserName, in.UserName)
	return &emptypb.Empty{}, nil
}

func readSecretsFromStorageKMS(ctx context.Context, kmsClient pb.StorageKMSPrivateServiceClient, secretKeyPath string) (map[string]string, error) {
	logger := log.FromContext(ctx).WithName("readSecretsFromStorageKMS")
	logger.Info("calling storage kms service for get", logkeys.SecretsPath, secretKeyPath)

	request := pb.GetSecretRequest{
		KeyPath: secretKeyPath,
	}
	secretResp, err := kmsClient.Get(ctx, &request)
	if err != nil {
		logger.Error(err, "error reading secrets from storage kms")
		return nil, err
	}
	return secretResp.Secrets, nil
}

func deleteSecretsFromStorageKMS(ctx context.Context, kmsClient pb.StorageKMSPrivateServiceClient, secretKeyPath string) error {
	logger := log.FromContext(ctx).WithName("deleteSecretsFromStorageKMS")
	logger.Info("calling storage kms service for delete", logkeys.SecretsPath, secretKeyPath)

	request := pb.DeleteSecretRequest{
		KeyPath: secretKeyPath,
	}
	_, err := kmsClient.Delete(ctx, &request)
	if err != nil {
		logger.Error(err, "error deleting secrets from storage kms")
		return err
	}
	return nil
}

func putSecretsInStorageKMS(ctx context.Context, kmsClient pb.StorageKMSPrivateServiceClient, secretKeyPath string, userCreds map[string]string) error {
	logger := log.FromContext(ctx).WithName("putSecretsInStorageKMS")
	logger.Info("calling storage kms service for put", logkeys.SecretsPath, secretKeyPath)

	userStoreReq := pb.StoreSecretRequest{
		KeyPath: secretKeyPath,
		Secrets: userCreds,
	}
	_, err := kmsClient.Put(ctx, &userStoreReq)
	if err != nil {
		logger.Error(err, "error storing user credentials to storage kms")
		return fmt.Errorf("error storing user credentials to storage kms")
	}
	logger.Info("secrets stored successfully at location:", logkeys.SecretsPath, secretKeyPath)
	return nil
}

func (storageUser *StorageUserServiceServer) PingFileUserPrivate(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("StorageUserServiceServer.PingFileUserPrivate")
	logger.Info("entering filesystem private Ping")
	defer logger.Info("returning from filesystem private Ping")

	return &emptypb.Empty{}, nil
}

func (storageUser *StorageUserServiceServer) PingBucketUserPrivate(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("StorageUserServiceServer.PingBucketUserPrivate")

	logger.Info("entering filesystem private Ping")
	defer logger.Info("returning from filesystem private Ping")

	return &emptypb.Empty{}, nil
}

func (storageUser *StorageUserServiceServer) CreateBucketUser(ctx context.Context, in *pb.CreateBucketUserParams) (*pb.BucketPrincipal, error) {
	logger := log.FromContext(ctx).WithName("StorageUserServiceServer.CreateBucketUser").WithValues(logkeys.CloudAccountId, in.CloudAccountId)
	logger.Info("entering bucketUser create")
	defer logger.Info("returning from user create")

	principal := pb.BucketPrincipal{}

	objUser := storagecontroller.ObjectUserRequest{
		CloudAccountId: in.CloudAccountId,
		ClusterId:      in.CreateParams.ClusterUUID,
		UserName:       in.CreateParams.UserId,
		Password:       in.CreateParams.Password,
		Policies:       getS3PrincipalPolicy(in.CreateParams.Spec, in.CreateParams.SecurityGroup, in.CreateParams.ClusterUUID),
	}

	userCreated, err := storageUser.strCntClient.CreateObjectUser(ctx, objUser)
	if err != nil {
		logger.Error(err, "error creating principal from sds")
		return nil, fmt.Errorf("error creating user")
	}

	principal.ClusterId = userCreated.ClusterId
	principal.PrincipalId = userCreated.PrincipalId
	principal.Spec = in.CreateParams.GetSpec()
	principal.ClusterName = in.CreateParams.ClusterName
	principal.AccessEndpoint = in.CreateParams.AccessEndpoint
	return &principal, nil

}

func (storageUser *StorageUserServiceServer) UpdateBucketUserPolicy(ctx context.Context, in *pb.UpdateBucketUserPolicyParams) (*pb.BucketPrincipal, error) {
	logger := log.FromContext(ctx).WithName("StorageUserServiceServer.UpdateBucketUserPolicy").WithValues(logkeys.CloudAccountId, in.CloudAccountId)
	logger.Info("entering bucketUser update  policies")
	defer logger.Info("returning from principal update policies")

	principal := pb.BucketPrincipal{}
	params := storagecontroller.ObjectUserUpdateRequest{
		CloudAccountId: in.CloudAccountId,
		ClusterId:      in.UpdateParams.ClusterUUID,
		PrincipalId:    in.UpdateParams.PrincipalId,
		Policies:       getS3PrincipalPolicy(in.UpdateParams.Spec, in.UpdateParams.SecurityGroup, in.UpdateParams.ClusterUUID),
	}

	err := storageUser.strCntClient.UpdateObjectUserPolicy(ctx, params)
	if err != nil {
		logger.Error(err, "error updating principal from sds")
		return nil, fmt.Errorf("error updating principal")
	}
	principal.ClusterId = in.UpdateParams.ClusterUUID
	principal.PrincipalId = in.UpdateParams.PrincipalId
	principal.Spec = in.UpdateParams.Spec
	principal.AccessEndpoint = in.UpdateParams.AccessEndpoint
	principal.ClusterName = in.UpdateParams.ClusterName
	return &principal, nil
}

func (storageUser *StorageUserServiceServer) UpdateBucketUserCredentials(ctx context.Context, in *pb.UpdateBucketUserCredsParams) (*pb.BucketPrincipal, error) {
	logger := log.FromContext(ctx).WithName("StorageUserServiceServer.UpdateBucketUserCredentials").WithValues(logkeys.CloudAccountId, in.CloudAccountId)
	logger.Info("entering bucketUser update credentials")
	defer logger.Info("returning from principal update credentials")

	principal := pb.BucketPrincipal{}
	params := storagecontroller.ObjectUserUpdateRequest{
		CloudAccountId: in.CloudAccountId,
		ClusterId:      in.UpdateParams.ClusterUUID,
		PrincipalId:    in.UpdateParams.PrincipalId,
		UserName:       in.UpdateParams.UserId,
		Password:       in.UpdateParams.Password,
	}

	err := storageUser.strCntClient.UpdateObjectUserPass(ctx, params)
	if err != nil {
		logger.Error(err, "error updating principal credentials from sds")
		return nil, fmt.Errorf("error updating principal credentials ")
	}
	principal.ClusterId = in.UpdateParams.ClusterUUID
	principal.PrincipalId = in.UpdateParams.PrincipalId
	// NOTE: Not passing all the principal details here
	return &principal, nil
}

func (storageUser *StorageUserServiceServer) DeleteBucketUser(ctx context.Context, in *pb.DeleteBucketUserParams) (*emptypb.Empty, error) {
	logger := log.FromContext(ctx).WithName("StorageUserServiceServer.DeleteBucketUser").WithValues(logkeys.CloudAccountId, in.CloudAccountId)
	logger.Info("entering bucketUser delete")
	defer logger.Info("returning from principal delete")

	deleteParams := storagecontroller.ObjectUserData{
		CloudAccountId: in.CloudAccountId,
		ClusterId:      in.ClusterId,
		PrincipalId:    in.PrincipalId,
	}

	if err := storageUser.strCntClient.DeleteObjectUser(ctx, deleteParams); err != nil {
		logger.Error(err, "error deleting object principal from the sds")
		return nil, status.Errorf(codes.Internal, "error deleting principal from sds")
	}

	return &emptypb.Empty{}, nil
}

func getS3PrincipalPolicy(in []*pb.ObjectUserPermissionSpec, secGroup *pb.BucketSecurityGroup, clusterId string) []*stcnt_api.S3Principal_Policy {
	principalPolicies := []*stcnt_api.S3Principal_Policy{}

	allowSourceIpFilters := []string{}

	for _, subnet := range secGroup.NetworkFilterAllow {
		allowSourceIpFilters = append(allowSourceIpFilters, fmt.Sprintf("%s/%d", subnet.Subnet, subnet.PrefixLength))
	}
	for _, inPolicy := range in {
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
	fmt.Println("s3 principals ", principalPolicies)
	return principalPolicies
}

func mapPermissionsToS3(in []pb.BucketPermission) (bool, bool, bool) {
	var read, write, delete bool

	for _, perm := range in {
		if perm == pb.BucketPermission_ReadBucket {
			read = true
		}
		if perm == pb.BucketPermission_WriteBucket {
			write = true
		}
		if perm == pb.BucketPermission_DeleteBucket {
			delete = true
		}
	}
	return read, write, delete
}

func mapActionsToS3(inAction []pb.ObjectBucketActions) []stcnt_api.S3Principal_Policy_BucketActions {
	outActions := []stcnt_api.S3Principal_Policy_BucketActions{}

	for _, ia := range inAction {
		if ia == pb.ObjectBucketActions_GetBucketLocation {
			outActions = append(outActions, stcnt_api.S3Principal_Policy_BUCKET_ACTIONS_GET_BUCKET_LOCATION)
		}
		if ia == pb.ObjectBucketActions_GetBucketPolicy {
			outActions = append(outActions, stcnt_api.S3Principal_Policy_BUCKET_ACTIONS_GET_BUCKET_POLICY)
		}
		if ia == pb.ObjectBucketActions_GetBucketTagging {
			outActions = append(outActions, stcnt_api.S3Principal_Policy_BUCKET_ACTIONS_GET_BUCKET_TAGGING)
		}
		if ia == pb.ObjectBucketActions_ListBucket {
			outActions = append(outActions, stcnt_api.S3Principal_Policy_BUCKET_ACTIONS_LIST_BUCKET)
		}
		if ia == pb.ObjectBucketActions_ListBucketMultipartUploads {
			outActions = append(outActions, stcnt_api.S3Principal_Policy_BUCKET_ACTIONS_LIST_BUCKET_MULTIPART_UPLOADS)
		}
		if ia == pb.ObjectBucketActions_ListMultipartUploadParts {
			outActions = append(outActions, stcnt_api.S3Principal_Policy_BUCKET_ACTIONS_LIST_MULTIPART_UPLOAD_PARTS)
		}
	}
	return outActions
}

func (storageUser *StorageUserServiceServer) GetBucketCapacity(ctx context.Context, in *pb.BucketFilter) (*pb.BucketCapacity, error) {
	logger := log.FromContext(ctx).WithName("StorageUserServiceServer.GetBucketCapacity").WithValues(logkeys.ClusterId, in.ClusterId, logkeys.BucketId, in.BucketId)
	logger.Info("entering bucket capacity check")
	bucketCap := pb.BucketCapacity{}
	defer logger.Info("returning from principal update credentials", logkeys.BucketCapacity, bucketCap.Capacity)

	getBucketParams := storagecontroller.BucketFilter{
		ClusterId: in.ClusterId,
		BucketId:  in.BucketId,
	}

	bucket, err := storageUser.strCntClient.GetBucket(ctx, getBucketParams)
	if err != nil {
		logger.Error(err, "error in get bucket from the sds")
		return nil, status.Errorf(codes.Internal, "error getting bucket from sds")
	}

	bucketCap.Id = bucket.Metadata.BucketId
	bucketCap.Name = bucket.Metadata.Name
	bucketCap.Capacity = &pb.Capacity{
		TotalBytes:     bucket.Spec.Totalbytes,
		AvailableBytes: bucket.Spec.AvailableBytes,
	}

	return &bucketCap, nil
}
