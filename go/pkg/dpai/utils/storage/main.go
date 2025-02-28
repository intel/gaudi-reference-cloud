// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package storage

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type Storage struct {
	CloudAccountID       string
	Conf                 *config.Config
	GrpcClientConn       *grpc.ClientConn
	ObjectStoreClient    pb.ObjectStorageServicePrivateClient
	ObjectStorePrincipal *pb.AccessPrincipal
}

func (s *Storage) GetStorageClient(conf *config.Config, cloudAccountId string) error {
	log.Println("Start : Inside Storage Client")
	s.Conf = conf

	if cloudAccountId == "" {
		return fmt.Errorf("input missing: please provide the cloudAccountId")
	}

	dialOptions := []grpc.DialOption{}
	log.Printf("Inside GetStorageClient : conf.GrpcAPIServerAddr: %s", conf.GrpcAPIServerAddr)

	grpcClientConn, err := grpcutil.NewClient(context.Background(), conf.GrpcAPIServerAddr, dialOptions...)
	if err != nil {
		return err
	}
	s.CloudAccountID = cloudAccountId
	s.GrpcClientConn = grpcClientConn
	s.ObjectStoreClient = pb.NewObjectStorageServicePrivateClient(grpcClientConn)
	log.Println("End: Inside Storage Client")
	log.Printf("s.ObjectStoreClient: %+v", s.ObjectStoreClient)
	result, err := s.ObjectStoreClient.PingPrivate(context.Background(), &emptypb.Empty{})
	if err != nil {
		log.Printf("error %s:", err)
		return err
	}
	log.Printf("s.ObjectStoreClient.ping: %+v", result)
	return nil
}

func (s *Storage) GetBucket(id string) (*pb.ObjectBucketPrivate, error) {
	log.Printf("Start : Getbucket: %s :", id)
	bucket, err := s.ObjectStoreClient.GetBucketPrivate(context.Background(), &pb.ObjectBucketGetPrivateRequest{
		Metadata: &pb.ObjectBucketMetadataRef{
			CloudAccountId: s.CloudAccountID,
			NameOrId: &pb.ObjectBucketMetadataRef_BucketId{
				BucketId: id,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return bucket, nil
}

func (s *Storage) IsValidAirflowBucket(id string) (bool, error) {
	log.Printf("Start : Is Valid bucket function: %s :", id)
	bucket, err := s.GetBucket(id)
	if err != nil {
		return false, err
	}
	if bucket.GetSpec().Versioned {
		return false, fmt.Errorf("version enabled")
	}
	log.Println("End : Is Valid bucket function")
	return true, nil
}

func (s *Storage) CreateObjectUser(args *pb.CreateObjectUserPrivateRequest, createCredentials bool) (*pb.ObjectUserPrivate, error) {
	log.Printf("Start : CreateObjectUser with params: %+v", args)
	user, err := s.ObjectStoreClient.CreateObjectUserPrivate(context.Background(), args)
	if err != nil {
		log.Printf("error %+v:", err)
		return nil, err
	}

	// if createCredentials {
	// 	user, err = s.UpdateObjectUserCredentials(user.Metadata.UserId)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	return user, nil

}

// func (s *Storage) UpdateObjectUserCredentials(userId string) (*pb.ObjectUserPrivate, error) {
// 	user, err := s.ObjectStoreClient.UpdateObjectUserCredentials(context.Background(), &pb.ObjectUserUpdateCredsRequest{
// 		Metadata: &pb.ObjectUserMetadataRef{
// 			CloudAccountId: s.CloudAccountID,
// 			NameOrId: &pb.ObjectUserMetadataRef_UserId{
// 				UserId: userId,
// 			},
// 		},
// 	})

// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create the credential. Error message: %+v", err)
// 	}
// 	return user, nil
// }

func (s *Storage) CreateAirflowObjectUser(bucketId string, userName string) (*pb.ObjectUserPrivate, error) {
	spec := pb.ObjectUserPermissionSpec{
		BucketId:   bucketId,
		Permission: []pb.BucketPermission{pb.BucketPermission_ReadBucket, pb.BucketPermission_WriteBucket},
		Actions: []pb.ObjectBucketActions{
			pb.ObjectBucketActions_GetBucketLocation,
			pb.ObjectBucketActions_GetBucketPolicy,
			pb.ObjectBucketActions_GetBucketTagging,
			pb.ObjectBucketActions_ListBucket,
			pb.ObjectBucketActions_ListBucketMultipartUploads,
			pb.ObjectBucketActions_ListMultipartUploadParts,
		},
	}
	var specs []*pb.ObjectUserPermissionSpec
	specs = append(specs, &spec)

	log.Printf("specs: %+v", specs)
	user, err := s.CreateObjectUser(&pb.CreateObjectUserPrivateRequest{
		Metadata: &pb.ObjectUserMetadataCreate{
			CloudAccountId: s.CloudAccountID,
			Name:           userName,
		},
		Spec: specs,
	}, true)
	if err != nil {
		return nil, err
	}

	// Create
	return user, nil
}
