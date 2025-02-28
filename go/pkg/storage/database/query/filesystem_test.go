package query

import (
	"context"
	"database/sql"
	"testing"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	db "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/database"

	//"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/database/query"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	sqlDb     *sql.DB
	txDb      *sql.Tx
	testDb    *manageddb.TestDb
	managedDb *manageddb.ManagedDb
	fsPrivate *pb.FilesystemPrivate
)

var (
	ctx = context.Background()
	//err error
)

func TestQuery(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Query Suite")
}

var _ = Describe("Query", func() {

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

		txDb, err = sqlDb.Begin()
		Expect(err).To(BeNil())
		Expect(txDb).NotTo(BeNil())

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
			}, //Meta
			Spec: &pb.FilesystemSpecPrivate{
				AvailabilityZone: "az1",
				Request: &pb.FilesystemCapacity{
					Storage: "2000000000",
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
		}
	})
	var _ = AfterSuite(func() {
		sqlDb.Close()
		txDb.Rollback()
	})

	Context("Testing updateFilesystemForDelete", func() {
		It("Should not succeed", func() {

			By("Updating fs in DB")
			err := updateFilesystemForDelete(ctx, txDb, fsPrivate)
			Expect(err).NotTo(BeNil())

		})
	})

	Context("Testing mapEventSqlToPb", func() {
		It("Should succeed", func() {
			rv := mapEventSqlToPb(AddingEventType)
			Expect(rv).To(Equal(pb.BucketSubnetEventStatus_E_ADDING))

			rv = mapEventSqlToPb(DeletingEventType)
			Expect(rv).To(Equal(pb.BucketSubnetEventStatus_E_DELETING))

			rv = mapEventSqlToPb(AddedEventType)
			Expect(rv).To(Equal(pb.BucketSubnetEventStatus_E_ADDED))

			rv = mapEventSqlToPb(DeletedEventType)
			Expect(rv).To(Equal(pb.BucketSubnetEventStatus_E_DELETED))

			rv = mapEventSqlToPb("")
			Expect(rv).To(Equal(pb.BucketSubnetEventStatus_E_UNSPECIFIED))
		})
	})

	Context("Testing mapEventPbToSql", func() {
		It("Should succeed", func() {
			rv := mapEventPbToSql(pb.BucketSubnetEventStatus_E_ADDED)
			Expect(rv).To(Equal(AddedEventType))

			rv = mapEventPbToSql(pb.BucketSubnetEventStatus_E_ADDING)
			Expect(rv).To(Equal(AddingEventType))

			rv = mapEventPbToSql(pb.BucketSubnetEventStatus_E_DELETED)
			Expect(rv).To(Equal(DeletedEventType))

			rv = mapEventPbToSql(pb.BucketSubnetEventStatus_E_DELETING)
			Expect(rv).To(Equal(DeletingEventType))

			rv = mapEventPbToSql(pb.BucketSubnetEventStatus_E_UNSPECIFIED)
			Expect(rv).To(Equal("UNSPECIFIED"))
		})
	})
})
