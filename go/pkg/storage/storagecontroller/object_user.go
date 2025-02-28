package storagecontroller

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	stcnt_api "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storagecontroller/api"
	"google.golang.org/grpc/metadata"
)

type ObjectUserRequest struct {
	CloudAccountId string
	ClusterId      string
	UserName       string
	Password       string
	Policies       []*stcnt_api.S3Principal_Policy
}

type ObjectUserResponse struct {
	ClusterId   string
	PrincipalId string
	UserName    string
	Password    string
	Policies    []*stcnt_api.S3Principal_Policy
}

type ObjectUserUpdateRequest struct {
	CloudAccountId string
	ClusterId      string
	PrincipalId    string
	UserName       string
	Password       string
	Policies       []*stcnt_api.S3Principal_Policy
}

type ObjectUserData struct {
	CloudAccountId string
	ClusterId      string
	PrincipalId    string
}

// Create object service user
func (client *StorageControllerClient) CreateObjectUser(ctx context.Context, in ObjectUserRequest) (*ObjectUserResponse, error) {
	logger := log.FromContext(ctx).WithName("StorageControllerClient.CreateObjectUser")
	logger.Info("inside CreateObjectUser func")

	// validate input
	if in.ClusterId == "" {
		logger.Info("empty cluster id")
		return nil, fmt.Errorf("cluster id cannot be empty")
	}
	if in.UserName == "" {
		logger.Info("empty object user name")
		return nil, fmt.Errorf("objectUser name cannot be empty")
	}
	if in.Password == "" {
		logger.Info("empty password")
		return nil, fmt.Errorf("password cannot be empty")
	}
	if in.Policies == nil {
		logger.Info("nil policies")
		return nil, fmt.Errorf("no policies supplied")
	}

	md := metadata.Pairs(
		"X-Cloud-Account", in.CloudAccountId,
	)
	// Create a context with metadata
	ctx = metadata.NewOutgoingContext(ctx, md)

	// make request call
	user, err := client.S3ServiceClient.CreateS3Principal(ctx, &stcnt_api.CreateS3PrincipalRequest{
		ClusterId: &stcnt_api.ClusterIdentifier{
			Uuid: in.ClusterId,
		},
		Name:        in.UserName,
		Credentials: in.Password,
	})
	// check for error
	if err != nil {
		logger.Error(err, "error in creating object user in the controller")
		return nil, err
	}
	logger.Info("UpdateS3PrincipalPolicies sds input params:", logkeys.CloudAccountId, in.CloudAccountId, logkeys.ClusterId, in.ClusterId, logkeys.Input, in.Policies)
	// make request call
	_, err = client.S3ServiceClient.UpdateS3PrincipalPolicies(ctx, &stcnt_api.UpdateS3PrincipalPoliciesRequest{
		PrincipalId: &stcnt_api.S3PrincipalIdentifier{
			ClusterId: &stcnt_api.ClusterIdentifier{
				Uuid: in.ClusterId,
			},
			Id: user.S3Principal.Id.Id,
		},
		Policies: in.Policies,
	})
	// check for error
	if err != nil {
		logger.Error(err, "error in creating object principal policies in the controller")
		return nil, err
	}
	// format output
	res := &ObjectUserResponse{
		ClusterId:   user.S3Principal.Id.ClusterId.Uuid,
		PrincipalId: user.S3Principal.Id.Id,
		UserName:    user.S3Principal.Name,
		Policies:    user.S3Principal.Policies,
	}
	return res, nil
}

// Delete object user
func (client *StorageControllerClient) DeleteObjectUser(ctx context.Context, in ObjectUserData) error {
	logger := log.FromContext(ctx).WithName("StorageControllerClient.DeleteObjectUser")
	logger.Info("inside DeleteObjectUser func")

	// validate input
	if in.ClusterId == "" {
		logger.Info("empty cluster id")
		return fmt.Errorf("cluster id cannot be empty")
	}
	if in.PrincipalId == "" {
		logger.Info("empty principal id ")
		return fmt.Errorf("user id cannot be empty")
	}

	md := metadata.Pairs(
		"X-Cloud-Account", in.CloudAccountId,
	)
	// Create a context with metadata
	ctx = metadata.NewOutgoingContext(ctx, md)

	// make request call
	logger.Info("DeleteS3Principal sds input params:", logkeys.CloudAccountId, in.CloudAccountId, logkeys.ClusterId, in.ClusterId, logkeys.PrincipalId, in.PrincipalId)
	_, err := client.S3ServiceClient.DeleteS3Principal(ctx, &stcnt_api.DeleteS3PrincipalRequest{
		PrincipalId: &stcnt_api.S3PrincipalIdentifier{
			ClusterId: &stcnt_api.ClusterIdentifier{
				Uuid: in.ClusterId,
			},
			Id: in.PrincipalId,
		},
	})
	// check for error
	if err != nil {
		logger.Error(err, "error in deleting object principal in controller")
		return err
	}

	return nil
}

// Get object service user
func (client *StorageControllerClient) GetObjectUser(ctx context.Context, in ObjectUserData) (*ObjectUserResponse, error) {
	logger := log.FromContext(ctx).WithName("StorageControllerClient.GetObjectUser")
	logger.Info("inside GetObjectUser func")

	// validate inputs
	if in.ClusterId == "" {
		logger.Info("empty clusterId ")
		return nil, fmt.Errorf("cluster id cannot be empty")
	}
	if in.PrincipalId == "" {
		logger.Info("empty principal id")
		return nil, fmt.Errorf("principal id cannot be empty")
	}

	// make request call
	logger.Info("GetS3Principal sds input params:", logkeys.CloudAccountId, in.CloudAccountId, logkeys.ClusterId, in.ClusterId, logkeys.PrincipalId, in.PrincipalId)
	res, err := client.S3ServiceClient.GetS3Principal(ctx, &stcnt_api.GetS3PrincipalRequest{
		PrincipalId: &stcnt_api.S3PrincipalIdentifier{
			ClusterId: &stcnt_api.ClusterIdentifier{
				Uuid: in.ClusterId,
			},
			Id: in.PrincipalId,
		},
	})
	// check for error
	if err != nil {
		logger.Error(err, "error while finding object principal in controller")
		return nil, err
	}
	// format output
	objUser := &ObjectUserResponse{
		ClusterId:   res.S3Principal.Id.ClusterId.Uuid,
		PrincipalId: res.S3Principal.Id.Id,
		UserName:    res.S3Principal.Name,
		Policies:    res.S3Principal.Policies,
	}
	return objUser, nil
}

// Update object user
func (client *StorageControllerClient) UpdateObjectUserPolicy(ctx context.Context, in ObjectUserUpdateRequest) error {
	logger := log.FromContext(ctx).WithName("StorageControllerClient.UpdateObjectUserPolicy")
	logger.Info("inside UpdateObjectUserPolicy func")

	// validate input
	if in.ClusterId == "" {
		logger.Info("empty cluster id ")
		return fmt.Errorf("cluster id cannot be empty")
	}
	if in.PrincipalId == "" {
		logger.Info("empty principal id")
		return fmt.Errorf("UserId cannot be empty")
	}
	if in.Policies == nil || len(in.Policies) == 0 {
		logger.Info("nil policies")
		return fmt.Errorf("no policies supplied")
	}
	request := stcnt_api.UpdateS3PrincipalPoliciesRequest{
		PrincipalId: &stcnt_api.S3PrincipalIdentifier{
			ClusterId: &stcnt_api.ClusterIdentifier{
				Uuid: in.ClusterId,
			},
			Id: in.PrincipalId,
		},
		Policies: in.Policies,
	}

	md := metadata.Pairs(
		"X-Cloud-Account", in.CloudAccountId,
	)
	logger.Info("Updating principal policy:", logkeys.CloudAccountId, in.CloudAccountId, logkeys.Input, in.Policies)
	// Create a context with metadata
	ctx = metadata.NewOutgoingContext(ctx, md)
	// make request call
	logger.Info("UpdateS3PrincipalPolicies sds input params:", logkeys.CloudAccountId, in.CloudAccountId, logkeys.ClusterId, in.ClusterId, logkeys.Input, in.Policies)
	_, err := client.S3ServiceClient.UpdateS3PrincipalPolicies(ctx, &request)
	// check for error
	if err != nil {
		logger.Error(err, "error in updating objectUser in the controller ")
		return err
	}

	return nil
}

// Set/Update object user password
func (client *StorageControllerClient) UpdateObjectUserPass(ctx context.Context, in ObjectUserUpdateRequest) error {
	logger := log.FromContext(ctx).WithName("StorageControllerClient.UpdateObjectUserPass")
	logger.Info("inside UpdateObjectUserPass func")

	// validate inputs
	if in.ClusterId == "" {
		logger.Info("empty cluster id ")
		return fmt.Errorf("cluster id cannot be empty")
	}
	if in.PrincipalId == "" {
		logger.Info("empty principal id")
		return fmt.Errorf("principalId cannot be empty")
	}
	if in.Password == "" {
		logger.Info("empty password")
		return fmt.Errorf("password cannot be empty")
	}

	md := metadata.Pairs(
		"X-Cloud-Account", in.CloudAccountId,
	)
	// Create a context with metadata
	ctx = metadata.NewOutgoingContext(ctx, md)
	// make the request call
	logger.Info("SetS3PrincipalCredentials sds input params:", logkeys.CloudAccountId, in.CloudAccountId, logkeys.ClusterId, in.ClusterId, logkeys.Input, in.Policies)
	_, err := client.S3ServiceClient.SetS3PrincipalCredentials(ctx, &stcnt_api.SetS3PrincipalCredentialsRequest{
		PrincipalId: &stcnt_api.S3PrincipalIdentifier{
			ClusterId: &stcnt_api.ClusterIdentifier{
				Uuid: in.ClusterId,
			},
			Id: in.PrincipalId,
		},
		Credentials: in.Password,
	})
	// check for error
	if err != nil {
		logger.Error(err, "error while setting object principal password in controller")
		return err
	}

	return nil
}
