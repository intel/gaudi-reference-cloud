// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	db "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/database"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	sqlDb              *sql.DB
	txDb               *sql.Tx
	testDb             *manageddb.TestDb
	managedDb          *manageddb.ManagedDb
	fsPrivate          *pb.FilesystemPrivate
	fsDelete           *pb.FilesystemDeleteRequest
	fsUpdate           *pb.FilesystemUpdateStatusRequest
	fsSchedule         *pb.FilesystemSchedule
	fsUpdateDeletion   *pb.FilesystemMetadataReference
	fsUpdateDel2       *pb.FilesystemMetadataReference
	bucket             *pb.ObjectBucketPrivate
	bucketUpdateStatus *pb.ObjectBucketStatusUpdateRequest
	vnet               *pb.VNetPrivate
	vnetDelete         *pb.BucketSubnetStatusUpdateRequest
)

func TestQuery(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Query Suite")
}

var _ = BeforeSuite(func() {
	ctx := context.Background()
	By("Starting database")
	testDb = &manageddb.TestDb{}
	var err error
	managedDb, err = testDb.Start(ctx)
	Expect(err).Should(Succeed())
	Expect(managedDb).ShouldNot(BeNil())
	Expect(managedDb.Migrate(ctx, db.Fsys, "migrations")).Should(Succeed())
	sqlDb, err = managedDb.Open(ctx)
	Expect(err).Should(Succeed())
	Expect(sqlDb).ShouldNot(BeNil())

	var version = new(string)
	*version = "v2"
	fsPrivate = &pb.FilesystemPrivate{
		Metadata: &pb.FilesystemMetadataPrivate{
			CloudAccountId:    "123456789012",
			Name:              "test",
			ResourceId:        "6787226a-2a55-4d6f-bae9-fa2a2ca2450a",
			ResourceVersion:   "1",
			Description:       "Sample Filesystem",
			Labels:            map[string]string{"key": "value"},
			CreationTimestamp: &timestamp.Timestamp{Seconds: 1637077200, Nanos: 0},
			DeletionTimestamp: &timestamp.Timestamp{Seconds: 1639612800, Nanos: 0},
			//DeletionTimestamp: nil,
		}, //Meta
		Spec: &pb.FilesystemSpecPrivate{
			AvailabilityZone: "az1",
			Request: &pb.FilesystemCapacity{
				Storage: "2GB",
			},
			MountProtocol: 0,
			Encrypted:     false,
			Scheduler: &pb.FilesystemSchedule{
				FilesystemName: "test",
				Cluster: &pb.AssignedCluster{
					ClusterName:    "1",
					ClusterAddr:    "1",
					ClusterUUID:    "1",
					ClusterVersion: version,
				},
				Namespace: &pb.AssignedNamespace{
					Name:            "123456789012",
					CredentialsPath: "/path/to/secret",
				},
			},
		}, //Spec
		Status: &pb.FilesystemStatusPrivate{
			Phase:   pb.FilesystemPhase_FSProvisioning,
			Message: "Filesystem is being provisioned",
			// Add relevant fields from your status
		}, // Status
	} //FSPrivate
	fsDelete = &pb.FilesystemDeleteRequest{
		Metadata: &pb.FilesystemMetadataReference{
			CloudAccountId:  "123456789012",
			ResourceVersion: "1",
		},
	}
	fsUpdate = &pb.FilesystemUpdateStatusRequest{
		Metadata: &pb.FilesystemIdReference{
			CloudAccountId:  "123456789012",
			ResourceId:      "6787226a-2a55-4d6f-bae9-fa2a2ca2450a",
			ResourceVersion: "1",
		},
		Status: &pb.FilesystemStatusPrivate{
			Message: "az1",
		},
	}
	name := &pb.FilesystemMetadataReference_Name{
		Name: "test",
	}
	//id := "6787226a-2a55-4d6f-bae9-fa2a2ca2450a"
	fsUpdateDeletion = &pb.FilesystemMetadataReference{
		CloudAccountId:  "123456789012",
		ResourceVersion: "1",
		NameOrId:        name,
	}
	Expect(fsUpdateDeletion).NotTo(BeNil())
	fsUpdateDel2 = &pb.FilesystemMetadataReference{
		CloudAccountId:  "",
		ResourceVersion: "1",
		NameOrId:        name,
	}
	var vrsn *string
	tmp := "heheh"
	vrsn = &tmp
	fsSchedule = &pb.FilesystemSchedule{
		FilesystemName: "test",
		Cluster: &pb.AssignedCluster{
			ClusterName:    "c1",
			ClusterAddr:    "localhost",
			ClusterUUID:    "BackendUUID1",
			ClusterVersion: vrsn,
		},
		Namespace: &pb.AssignedNamespace{
			Name:            "123456789012",
			CredentialsPath: "path/to/creds",
		},
	}
	bucket = &pb.ObjectBucketPrivate{
		Metadata: &pb.ObjectBucketMetadataPrivate{
			CloudAccountId:    "123456789012",
			Name:              "bucket1",
			ResourceId:        "934b5026-d346-78c8-fcd3-899852346509",
			BucketId:          "123456789012-bucket1",
			CreationTimestamp: &timestamp.Timestamp{Seconds: 1637077200, Nanos: 0},
			DeletionTimestamp: nil,
		},
		Spec: &pb.ObjectBucketSpecPrivate{
			AccessPolicy:     pb.BucketAccessPolicy_READ_WRITE,
			AvailabilityZone: "az1",
			Request: &pb.StorageCapacityRequest{
				Size: "10GB",
			},
		},
		Status: &pb.ObjectBucketStatus{
			Phase: pb.BucketPhase_BucketReady,
		},
	}
	bucketUpdateStatus = &pb.ObjectBucketStatusUpdateRequest{
		Metadata: &pb.ObjectBucketIdReference{
			CloudAccountId: "123456789012",
			ResourceId:     "934b5026-d346-78c8-fcd3-899852346509",
		},
		Status: &pb.ObjectBucketStatus{
			Phase:   pb.BucketPhase_BucketReady,
			Message: "",
		},
	}
	vnet = &pb.VNetPrivate{
		Metadata: &pb.VNetPrivate_Metadata{
			CloudAccountId: "123456789012",
			Name:           "default",
			ResourceId:     "8623ccaa-704e-4839-bc72-9a89daa20111",
		},
		Spec: &pb.VNetSpecPrivate{
			Region:           "us-region-1",
			AvailabilityZone: "az1",
			Subnet:           "0.0.0.0",
			PrefixLength:     0,
			Gateway:          "0.0.0.0",
		},
	}
	vnetDelete = &pb.BucketSubnetStatusUpdateRequest{
		ResourceId:      vnet.Metadata.ResourceId,
		CloudacccountId: "123456789012",
		VNetName:        "default",
		Status:          pb.BucketSubnetEventStatus_E_ADDED,
	}
	//rs := &pb. MockFilesystemPrivateService_SearchFilesystemRequestsServer{}
})
var _ = AfterSuite(func() {
	sqlDb.Close()
})

func NewMockFilesystemPrivateService_SearchFilesystemRequestsServer() pb.FilesystemPrivateService_SearchFilesystemRequestsServer {
	mockController := gomock.NewController(GinkgoT())
	rs := pb.NewMockFilesystemPrivateService_SearchFilesystemRequestsServer(mockController)
	rs.EXPECT().Context().Return(context.Background()).Times(1)
	return rs
}
