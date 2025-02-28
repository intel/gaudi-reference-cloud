package billing_intel_comms_test

import (
	"context"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing/db"
	billingIntel "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_intel"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"google.golang.org/grpc"
)

type TestService struct {
	billingIntel.Service
	testDB manageddb.TestDb
}

var Test TestService

func (ts *TestService) Init(ctx context.Context, cfg *billingIntel.Config,
	resolver grpcutil.Resolver, grpcServer *grpc.Server) error {

	var err error
	log := log.FromContext(ctx)
	ts.Mdb, err = ts.testDB.Start(ctx)
	if err != nil {
		return fmt.Errorf("testDb.Start: %m", err)
	}
	if err := ts.Service.Init(ctx, cfg, resolver, grpcServer); err != nil {
		fmt.Print(err)
	}

	if err := ts.Mdb.Migrate(ctx, db.MigrationsFs, db.MigrationsDir); err != nil {
		log.Error(err, "error migrating database")
		return err
	}
	log.Info("successfully migrated database model")

	return nil
}
func EmbedService(ctx context.Context) {
	grpcutil.AddTestService[*billingIntel.Config](&TestService{}, &billingIntel.Config{})
}
