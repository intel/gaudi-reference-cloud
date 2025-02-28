// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/codes"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/config"
	db_query "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/db_query_constants"
	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/iks_utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/query"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/server"
	clusterv1alpha "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var clusterName = uuid.New().String()[0:11]
var clusterUuid string

var errorBaseDefault = "rpc error: code = Unknown desc = "
var errorBaseNotFound = "rpc error: code = NotFound desc = "
var errorBaseAlreadyExists = "rpc error: code = AlreadyExists desc = "
var errorBasePermissionDenied = "rpc error: code = PermissionDenied desc = "
var errorBaseFailedPrecondition = "rpc error: code = FailedPrecondition desc = "
var errorBaseInvalidArgument = "rpc error: code = InvalidArgument desc = "

var desc = "Clusterdescription"
var svccidr = "192.160.10.20"
var clscidr = "192.168.30.30"
var clusterdns = "testdns"
var clusterCreate = pb.ClusterRequest{
	CloudAccountId: "iks_user",
	Name:           clusterName,
	Description:    &desc,
	K8Sversionname: "1.28",
	Runtimename:    "Containerd",
}

const getLatestClusterRev = `
		SELECT desiredspec_json
		FROM public.clusterrev
		WHERE cluster_id = (
			SELECT cluster_id
			FROM public.cluster
			WHERE unique_id = $1
		)
		ORDER BY clusterrev_id desc
		LIMIT 1
	`

func clearDB() error {
	sqlDB, err := managedDb.Open(context.Background())
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec(`TRUNCATE TABLE public.nodegroup CASCADE`)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec(`TRUNCATE TABLE public.clusterrev CASCADE`)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec(`TRUNCATE TABLE public.provisioninglog CASCADE`)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec(`TRUNCATE TABLE public.vipdetails CASCADE`)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec(`TRUNCATE TABLE public.vip CASCADE`)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec(`TRUNCATE TABLE public.cluster CASCADE`)
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	return nil
}

/* CLUSTERS UNIT TEST */
func setActionableState(uuid string) error {

	sqlDB, err := managedDb.Open(context.Background())
	_, err = sqlDB.Exec(`
		UPDATE clusterrev
		SET change_applied = true
		WHERE cluster_id = (select cluster_id from cluster where unique_id = $1)`,
		uuid,
	)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec(`
		UPDATE cluster
		SET clusterstate_name = 'Active'
		WHERE unique_id = $1
		RETURNING cluster_id`,
		uuid,
	)
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	return nil
}

func setDeletingState(uuid string) error {
	sqlDB, err := managedDb.Open(context.Background())
	_, err = sqlDB.Exec(`
		UPDATE clusterrev
		SET change_applied = true
		WHERE cluster_id = (select cluster_id from cluster where unique_id = $1)`,
		uuid,
	)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec(`
		UPDATE cluster
		SET clusterstate_name = 'Deleting'
		WHERE unique_id = $1
		RETURNING cluster_id`,
		uuid,
	)
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	return nil
}

func setNodeGroupsActionableState(clusteruuid string) error {
	sqlDB, err := managedDb.Open(context.Background())
	_, err = sqlDB.Exec(`
		UPDATE nodegroup
		SET nodegroupstate_name = 'Active'
		WHERE cluster_id = (select cluster_id from cluster where unique_id = $1)`,
		clusteruuid,
	)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec(`
		UPDATE cluster
		SET clusterstate_name = 'Active'
		WHERE unique_id = $1
		RETURNING cluster_id`,
		clusteruuid,
	)
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	return nil
}

func setNodeGroupsDeletingState(clusteruuid string) error {
	sqlDB, err := managedDb.Open(context.Background())
	_, err = sqlDB.Exec(`
		UPDATE nodegroup
		SET nodegroupstate_name = 'Deleting'
		WHERE cluster_id = (select cluster_id from cluster where unique_id = $1)`,
		clusteruuid,
	)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec(`
		UPDATE cluster
		SET clusterstate_name = 'Active'
		WHERE unique_id = $1
		RETURNING cluster_id`,
		clusteruuid,
	)
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	return nil
}

func setVipDeletingState(clusteruuid string) error {
	sqlDB, err := managedDb.Open(context.Background())
	_, err = sqlDB.Exec(`
		UPDATE vip
		SET vipstate_name = 'Deleting'
		WHERE cluster_id = (select cluster_id from cluster where unique_id = $1)`,
		clusteruuid,
	)
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	return nil
}

func setVipActiveState(clusteruuid string) error {
	sqlDB, err := managedDb.Open(context.Background())
	_, err = sqlDB.Exec(`
		UPDATE vip
		SET vipstate_name = 'Active'
		WHERE cluster_id = (select cluster_id from cluster where unique_id = $1)`,
		clusteruuid,
	)
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	return nil
}

func setK8sVersionActiveAndDepricate(newK8sVersionName string, oldK8sVersionName string) error {
	sqlDB, err := managedDb.Open(context.Background())
	_, err = sqlDB.Exec(`
			UPDATE k8sversion
			SET lifecyclestate_id = '1'
			WHERE k8sversion_name = $1
		`,
		newK8sVersionName,
	)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec(`
			UPDATE k8sversion
			SET lifecyclestate_id = '2'
			WHERE k8sversion_name = $1
		`,
		oldK8sVersionName,
	)
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	return nil
}

func setnodescounttomax(nodegroupid string) error {
	sqlDB, err := managedDb.Open(context.Background())
	_, err = sqlDB.Exec(`
			UPDATE nodegroup
			SET nodecount = '10'
			WHERE unique_id = $1
		`,
		nodegroupid,
	)
	if err != nil {
		return err
	}

	defer sqlDB.Close()
	return nil
}

func setnodecountback(nodegroupid string) error {
	sqlDB, err := managedDb.Open(context.Background())
	_, err = sqlDB.Exec(`
			UPDATE nodegroup
			SET nodecount = '1'
			WHERE unique_id = $1
		`,
		nodegroupid,
	)
	if err != nil {
		return err
	}

	defer sqlDB.Close()
	return nil
}
func setRestrictedToFalse() error {
	sqlDB, err := managedDb.Open(context.Background())
	_, err = sqlDB.Exec(`
			update defaultconfig
			set value=false
			where name='restrict_create_cluster';
		`,
	)
	if err != nil {
		return err
	}

	defer sqlDB.Close()
	return nil
}

func testsixnodegroups(clusterUuid string) error {
	err := setActionableState(clusterUuid)
	if err != nil {
		return err
	}

	sqlDB, err := managedDb.Open(context.Background())

	var clusterId string
	err = sqlDB.QueryRow("SELECT cluster_id FROM public.cluster WHERE unique_id = $1", clusterUuid).Scan(&clusterId)
	if err != nil {
		return err
	}

	k8sVersion := "1.27.11"
	k8sOsImageInstance_name := "iks-vm-u22-cd-wk-1-27-11-v20240227"

	query := `INSERT INTO public.nodegroup (cluster_id,k8sversion_name,nodegroupstate_name,nodegrouptype_name,runtime_name,unique_id,name,description,statedetails,instancetype_name,osimageinstance_name,upgstrategydrainbefdel,upgstrategymaxnodes,lifecyclestate_id,createddate,vnets,nodegrouptype)
	VALUES ($1, $2 ,'Active','Worker','Containerd',$3,$4,'testng','stateddetails','vm-spr-tny',$5,'t','10','1','2023-10-05 22:09:26.813616','[{"availabilityzonename":"us-region-1a","networkinterfacevnetname":"us-region-1a-default"}]', 'gp')`

	_, err = sqlDB.Exec(query, clusterId, k8sVersion, "unique_id_1", "nodegroup0", k8sOsImageInstance_name)
	if err != nil {
		return err
	}

	err = setActionableState(clusterUuid)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec(query, clusterId, k8sVersion, "unique_id_2", "nodegroup2", k8sOsImageInstance_name)
	if err != nil {
		return err
	}

	err = setActionableState(clusterUuid)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec(query, clusterId, k8sVersion, "unique_id_3", "nodegroup2", k8sOsImageInstance_name)
	if err != nil {
		return err
	}
	err = setActionableState(clusterUuid)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec(query, clusterId, k8sVersion, "unique_id_4", "nodegroup3", k8sOsImageInstance_name)
	if err != nil {
		return err
	}
	err = setActionableState(clusterUuid)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec(query, clusterId, k8sVersion, "unique_id_5", "nodegroup4", k8sOsImageInstance_name)
	if err != nil {
		return err
	}

	defer sqlDB.Close()
	return nil
}

func deletenodegroups(clusterUuid string) error {
	sqlDB, err := managedDb.Open(context.Background())
	var clusterId string
	err = sqlDB.QueryRow("SELECT cluster_id FROM public.cluster WHERE unique_id = $1", clusterUuid).Scan(&clusterId)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec(`DELETE from public.nodegroup WHERE cluster_id = $1 and nodegrouptype_name = 'Worker'`, clusterId)
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	return nil
}

func deletevips(clusterUuid string) error {
	sqlDB, err := managedDb.Open(context.Background())
	var clusterId string
	err = sqlDB.QueryRow("SELECT cluster_id FROM public.cluster WHERE unique_id = $1", clusterUuid).Scan(&clusterId)
	if err != nil {
		return err
	}

	rows, err := sqlDB.Query("SELECT vip_id FROM public.vip v WHERE v.cluster_id =$1 and owner = 'customer'", clusterId)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var vipId int32
		err = rows.Scan(&vipId)
		// Delete Vip Details
		if err != nil {
			return err
		}
		_, err = sqlDB.Exec("DELETE FROM public.vipdetails v WHERE v.vip_id = $1", vipId)
		if err != nil {
			return err
		}
	}
	_, err = sqlDB.Exec(`DELETE from public.vip WHERE cluster_id = $1 and owner = 'customer'`, clusterId)
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	return nil
}

func setControlPlaneToK8sVersion(clusterUuid string, newK8sVersionName string) error {
	sqlDB, err := managedDb.Open(context.Background())
	_, err = sqlDB.Exec(`
			update public.nodegroup
			set k8sversion_name = $2
			where nodegrouptype_name = 'ControlPlane' and cluster_id = (select cluster_id from public.cluster where unique_id = $1)
		`,
		clusterUuid,
		newK8sVersionName,
	)
	if err != nil {
		return err
	}

	defer sqlDB.Close()
	return nil
}

func setControlPlaneUpgrade(uuid string) error {
	sqlDB, err := managedDb.Open(context.Background())
	_, err = sqlDB.Exec(`
		UPDATE nodegroup
		SET nodegroupstate_name = 'Active'
		WHERE cluster_id = (select cluster_id from cluster where unique_id = $1)`,
		uuid,
	)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec(`
		UPDATE cluster
		SET clusterstate_name = 'Active'
		WHERE unique_id = $1
		RETURNING cluster_id`,
		uuid,
	)
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	return nil
}

func setAllowCreateStorageTrue(cloudAccountID string) error {
	sqlDB, err := managedDb.Open(context.Background())
	if err != nil {
		return err
	}

	_, err = sqlDB.Exec(`INSERT INTO public.cloudaccountextraspec (cloudaccount_id, allow_create_storage, active_account_create_cluster, maxclusters_override, maxclusterng_override, maxclusterilb_override, maxclustervm_override, maxnodegroupvm_override) VALUES($1, true, true,3,5,2,50,10)`, cloudAccountID)
	if err != nil {
		return err
	}

	defer sqlDB.Close()
	return nil
}

/*
func TestDBCleanupAAA(t *testing.T) {
	err := removeClusterFromDB("cl-3ccbd8fe-80")
	if err != nil {
		t.Fatalf("Could not clean up entry %v", err.Error())
	}
}
*/

func removeClusterFromDB(uuid string) error {
	sqlDB, err := managedDb.Open(context.Background())
	if err != nil {
		return err
	}
	var clusterId string
	err = sqlDB.QueryRow("SELECT cluster_id FROM public.cluster WHERE unique_id = $1", uuid).Scan(&clusterId)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec("DELETE from public.k8snode n where n.cluster_id = $1", clusterId)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec("DELETE from public.nodegroup n where n.cluster_id = $1", clusterId)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec("DELETE from public.cluster_extraconfig n where n.cluster_id = $1", clusterId)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec("DELETE from public.clusterrev n where n.cluster_id = $1", clusterId)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec("DELETE from public.provisioninglog n where n.cluster_id = $1", clusterId)
	if err != nil {
		return err
	}
	rows, err := sqlDB.Query("SELECT vip_id FROM public.vip v WHERE v.cluster_id =$1", clusterId)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var vipId int32
		err = rows.Scan(&vipId)
		// Delete Vip Details
		if err != nil {
			return err
		}
		_, err = sqlDB.Exec("DELETE FROM public.vipdetails v WHERE v.vip_id = $1", vipId)
		if err != nil {
			return err
		}
	}
	_, err = sqlDB.Exec("DELETE from public.vip v where v.cluster_id = $1", clusterId)
	if err != nil {
		return err
	}

	_, err = sqlDB.Exec("DELETE from public.storage v where v.cluster_id = $1", clusterId)
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec("DELETE from public.cluster n where n.cluster_id = $1", clusterId)
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	return nil
}

func InitSshMockClient(ctx context.Context, t *testing.T) (*server.Server, error) {
	file, e := os.Create("./encryption_keys")
	if e != nil {
		t.Fatalf("Failed to initialize Mock compute Client")
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	_, err := fmt.Fprint(w, `{"data":{"2":"abcdefghijklmnopqrstuvwxyz123456", "1":"abcdefghijklmnopqrstuvwxyz123456"}, "metadata":{"created_time":"2023-10-25T19:27:45.560705679Z","custom_metadata":null,"deletion_time":"","destroyed":false,"version":1}}`)
	if err != nil {
		t.Fatalf("Failed to initialize Mock compute Client")
	}
	w.Flush()
	sshclient := NewMockSshServiceClient(ctx, t)
	cfg := config.Config{
		ListenPort:     443,
		EncryptionKeys: "./encryption_keys",
	}
	iksSrv, err := server.NewIksService(sqlDb, nil, sshclient, nil, nil, nil, cfg)
	if err != nil {
		t.Fatalf("Failed to initialize Mock compute Client")
	}
	return iksSrv, nil
}
func InitVnetMockClient(ctx context.Context, t *testing.T) (*server.Server, error) {
	vnetclient := NewMockVnetServiceClient(ctx, t)
	computeclient := NewMockComputeServiceClient(ctx, t)
	cfg := config.Config{
		ListenPort:     443,
		EncryptionKeys: "./encryption_keys",
	}
	iksSrv, err := server.NewIksService(sqlDb, computeclient, nil, vnetclient, nil, nil, cfg)
	if err != nil {
		t.Fatalf("Failed to initialize Mock compute Client")
	}
	return iksSrv, nil
}

func InitComputeMockClient(ctx context.Context, t *testing.T) (*server.Server, error) {
	computeClient := NewMockComputeServiceClient(ctx, t)

	cfg := config.Config{
		ListenPort:     443,
		EncryptionKeys: "./encryption_keys",
	}
	iksSrv, err := server.NewIksService(sqlDb, computeClient, nil, nil, nil, nil, cfg)
	if err != nil {
		t.Fatalf("Failed to initialize Mock compute Client")
	}
	return iksSrv, nil
}

func InitProductCatalogClient(ctx context.Context, t *testing.T) (*server.Server, error) {
	productCatalogServiceClient := NewMockProductCatalogServiceClient(ctx, t)
	// computeclient := NewMockComputeServiceClient(ctx, t)
	cfg := config.Config{
		ListenPort:     443,
		EncryptionKeys: "./encryption_keys",
	}
	iksSrv, err := server.NewIksService(sqlDb, nil, nil, nil, productCatalogServiceClient, nil, cfg)
	if err != nil {
		t.Fatalf("Failed to initialize Mock product catalog Client")
	}
	return iksSrv, nil
}

func InitInstanceServiceMockClient(t *testing.T, mockDelMethod bool, c codes.Code, errMsg string) (*server.Server, error) {
	return server.NewIksService(
		sqlDb,
		nil,
		nil,
		nil,
		nil,
		NewMockInstanceServiceClient(t, mockDelMethod, c, errMsg),
		config.Config{
			ListenPort:     443,
			EncryptionKeys: "./encryption_keys",
		},
	)
}

func TestCreateClusterRestrictions(t *testing.T) {
	clearDB()

	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.CreateNewCluster").WithValues(logkeys.ClusterName, clusterCreate.Name, logkeys.CloudAccountId, clusterCreate.CloudAccountId).Start()
	defer span.End()

	// Setup mock
	iksSrv, err := InitSshMockClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with ssh mocking failed %v", err.Error())
	}
	// Test
	_, err = iksSrv.CreateNewCluster(context.Background(), &clusterCreate)
	if err == nil {
		t.Fatalf("Expected an restriction error, but none occured %v", err.Error())
	}

	// Check
	expectedErrorMessage := errorBasePermissionDenied + "Due to restrictions, we are currently not allowing non-approved users to provision clusters."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected an restriction error: %s, recieved: %v", expectedErrorMessage, err.Error())
	}
}

func TestCreateClusterLimit(t *testing.T) {
	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.CreateNewCluster").WithValues(logkeys.ClusterName, clusterCreate.Name, logkeys.CloudAccountId, clusterCreate.CloudAccountId).Start()
	defer span.End()

	// Setup mock
	iksSrv, err := InitSshMockClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with ssh mocking failed %v", err.Error())
	}

	err = setRestrictedToFalse()
	if err != nil {
		t.Fatalf("Could not set restricted env to true")
	}

	// Test
	clusterCreate1 := pb.ClusterRequest{
		CloudAccountId: clusterCreate.CloudAccountId,
		Name:           "cluster1",
		Description:    clusterCreate.Description,
		K8Sversionname: clusterCreate.K8Sversionname,
		Runtimename:    clusterCreate.Runtimename,
	}

	_, err = iksSrv.CreateNewCluster(context.Background(), &clusterCreate1)
	if err != nil {
		t.Fatal(err)
	}

	clusterCreate2 := pb.ClusterRequest{
		CloudAccountId: clusterCreate.CloudAccountId,
		Name:           "cluster2",
		Description:    clusterCreate.Description,
		K8Sversionname: clusterCreate.K8Sversionname,
		Runtimename:    clusterCreate.Runtimename,
	}

	_, err = iksSrv.CreateNewCluster(context.Background(), &clusterCreate2)
	if err != nil {
		t.Fatal(err)
	}

	// Test
	errCount := 0
	expectedError := errorBasePermissionDenied + "Can not create more than 3 clusters for this cloud account"
	expectedErrCount := 1 // Expecting to get exactly one error due to quota limit

	var wg sync.WaitGroup
	wg.Add(2)

	errCh := make(chan error, 2)

	go func() {
		defer wg.Done()
		clusterCreate3 := pb.ClusterRequest{
			CloudAccountId: clusterCreate.CloudAccountId,
			Name:           "cluster3",
			Description:    clusterCreate.Description,
			K8Sversionname: clusterCreate.K8Sversionname,
			Runtimename:    clusterCreate.Runtimename,
		}

		_, err = iksSrv.CreateNewCluster(context.Background(), &clusterCreate3)
		if err != nil {
			if expectedError != err.Error() { // only count error or no error is expected
				errCh <- fmt.Errorf("create cluster record: Expected count error, received: %v", err.Error())
				return
			}
			errCount++ // Got expected error, increment error count
		}
	}()

	go func() {
		defer wg.Done()
		clusterCreate4 := pb.ClusterRequest{
			CloudAccountId: clusterCreate.CloudAccountId,
			Name:           "cluster4",
			Description:    clusterCreate.Description,
			K8Sversionname: clusterCreate.K8Sversionname,
			Runtimename:    clusterCreate.Runtimename,
		}

		_, err = iksSrv.CreateNewCluster(context.Background(), &clusterCreate4)
		if err != nil {
			if expectedError != err.Error() { // only count error or no error is expected
				errCh <- fmt.Errorf("Create cluster record: Expected count error, received: %v", err.Error())
				return
			}
			errCount++ // Got expected error, increment error count
		}
	}()

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatal(err)
	}

	// Check that we got exactly one error
	if errCount != expectedErrCount {
		t.Fatalf("Expected %d errors, but got %d", expectedErrCount, errCount)
	}
}

func TestCreateClusterImiK8sVersionFailureForIksUser(t *testing.T) {
	clearDB()
	// Setup mock

	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.CreateNewCluster").WithValues(logkeys.ClusterName, clusterCreate.Name, logkeys.CloudAccountId, clusterCreate.CloudAccountId).Start()
	defer span.End()

	iksSrv, err := InitSshMockClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with ssh mocking failed")
	}

	err = setRestrictedToFalse()
	if err != nil {
		t.Fatalf("Could not set restricted env to true")
	}
	// Set up
	original := clusterCreate.K8Sversionname
	clusterCreate.K8Sversionname = "1.1"
	// Test
	_, err = iksSrv.CreateNewCluster(context.Background(), &clusterCreate)
	clusterCreate.K8Sversionname = original
	if err == nil {
		t.Fatalf("Expected an imi mismatch mismatch error, but none occured")
	}
	// Check
	expectedErrorMessage := errorBaseFailedPrecondition + "IMI version not compatible"
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected an imi mismatch mismatch error, recieved: %v", err.Error())
	}
}
func TestCreateClusterImiRuntimeFailureForIksUser(t *testing.T) {
	// Set up
	// Setup mock
	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.CreateNewCluster").WithValues(logkeys.ClusterName, clusterCreate.Name, logkeys.CloudAccountId, clusterCreate.CloudAccountId).Start()
	defer span.End()
	iksSrv, err := InitSshMockClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with ssh mocking failed")
	}
	original := clusterCreate.Runtimename
	clusterCreate.Runtimename = "xxx"
	// Test
	_, err = iksSrv.CreateNewCluster(context.Background(), &clusterCreate)
	clusterCreate.Runtimename = original
	if err == nil {
		t.Fatalf("Expected an imi runtime failure error, but none occured")
	}
	// Check
	expectedErrorMessage := errorBaseFailedPrecondition + "IMI version not compatible"
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected an imi runtime failure error, recieved: %v", err.Error())
	}
}

/*
func TestCreateClusterSuccessForRkeAndValidateDefaults(t *testing.T) {
	// SET UP
	ogUser := clusterCreate.CloudAccountId
	ogK8sVersion := clusterCreate.K8Sversionname
	ogAnnotations := clusterCreate.Annotations
	clusterCreate.CloudAccountId = "rke2_user"
	clusterCreate.K8Sversionname = "v1.27.2+rke2r1"
	clusterCreate.Annotations = []*pb.Annotations{
		{
			Key:   "key1",
			Value: "value1",
		},
	}
	// Setup mock
	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("IKS.ClusterCreate").WithValues("name", clusterCreate.Name).Start()
	defer span.End()

	iksSrv, err := InitSshMockClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with ssh mocking failed")
	}

	// setup encryption key
	file, e := os.Create("./encryption_keys")
	if e != nil {
		t.Fatalf("can not create encryption keys")
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	_, err = fmt.Fprint(w, `{"data":{"2":"abcdefghijklmnopqrstuvwxyz123456", "1":"abcdefghijklmnopqrstuvwxyz123456"}, "metadata":{"created_time":"2023-10-25T19:27:45.560705679Z","custom_metadata":null,"deletion_time":"","destroyed":false,"version":1}}`)
	w.Flush()

	sqlDB, err := managedDb.Open(context.Background())
	_, err = sqlDB.Exec(`INSERT INTO public.cloudaccountextraspec (cloudaccount_id, provider_name) VALUES('rke2_user', 'rke2')`)
	if err != nil {
		t.Fatalf("Could not clean up entry: %v", err)
	}
	// TEST
	res, err := iksSrv.CreateNewCluster(context.Background(), &clusterCreate)
	clusterCreate.CloudAccountId = ogUser
	clusterCreate.K8Sversionname = ogK8sVersion
	clusterCreate.Annotations = ogAnnotations
	if err != nil {
		t.Fatalf("Create cluster record: %v", err)
	}
	// CHECK
	sqlDB, err = managedDb.Open(context.Background())
	var (
		state    string
		name     string
		provider string
	)
	sqlDB.QueryRow("SELECT clusterstate_name, provider_name, name FROM public.cluster WHERE unique_id = $1", res.Uuid).Scan(
		&state,
		&provider,
		&name,
	)
	var clusterCidr string
	sqlDB.QueryRow("SELECT value FROM public.defaultconfig WHERE name = 'cluster_cidr'").Scan(
		&clusterCidr,
	)
	var (
		desiredJson []byte
		clusterCrd  clusterv1alpha.Cluster
	)
	err = sqlDB.QueryRow(getLatestClusterRev,
		res.Uuid,
	).Scan(&desiredJson)
	defer sqlDB.Close()
	if err != nil {
		t.Fatalf("Should not error: %v", err.Error())
	}
	err = json.Unmarshal([]byte(desiredJson), &clusterCrd)
	if err != nil {
		t.Fatalf("Should not error: %v", err.Error())
	}
	err = removeClusterFromDB(res.Uuid)
	if err != nil {
		t.Fatalf("Could not clean up entry")
	}
	if state != "Pending" || provider != "rke2" || name != clusterCreate.Name {
		t.Fatalf("Create cluster record: Expected DB values do not match")
	}
	if clusterCrd.Spec.Network.ClusterDNS != clusterCidr {
		t.Fatalf("Create cluster record: Expected Default Netowrk values do not match")
	}
	defer sqlDB.Close()
}
*/

func TestCreateClusterSuccessForIksAndPassedNetworkValues(t *testing.T) {
	// Setup mock

	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.CreateNewCluster").WithValues(logkeys.ClusterName, clusterCreate.Name, logkeys.CloudAccountId, clusterCreate.CloudAccountId).Start()
	defer span.End()

	iksSrv, err := InitSshMockClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with ssh mocking failed")
	}
	// TEST
	res, err := iksSrv.CreateNewCluster(context.Background(), &clusterCreate)
	if err != nil {
		t.Fatalf("Create cluster record: %v", err)
	}
	// CHECK
	clusterUuid = res.Uuid
	sqlDB, err := managedDb.Open(context.Background())
	var (
		state    string
		name     string
		provider string
	)
	sqlDB.QueryRow("SELECT clusterstate_name, provider_name, name FROM public.cluster WHERE unique_id = $1", res.Uuid).Scan(
		&state,
		&provider,
		&name,
	)
	var (
		desiredJson []byte
		clusterCrd  clusterv1alpha.Cluster
	)
	err = sqlDB.QueryRow(getLatestClusterRev,
		res.Uuid,
	).Scan(&desiredJson)
	if err != nil {
		t.Fatalf("Should not error: %v", err.Error())
	}
	err = json.Unmarshal([]byte(desiredJson), &clusterCrd)
	if err != nil {
		t.Fatalf("Should not error: %v", err.Error())
	}
	if state != "Pending" || provider != "iks" || name != clusterCreate.Name {
		t.Fatalf("Create cluster record: Expected DB values do not match")
	}
	defer sqlDB.Close()
}

func TestCreateClusterFailureOnUniqueName(t *testing.T) {
	// Setup mock
	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.CreateNewCluster").WithValues(logkeys.ClusterName, clusterCreate.Name, logkeys.CloudAccountId, clusterCreate.CloudAccountId).Start()
	defer span.End()

	iksSrv, err := InitSshMockClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with ssh mocking failed")
	}
	// Test
	_, err = iksSrv.CreateNewCluster(context.Background(), &clusterCreate)
	if err == nil {
		t.Fatalf("Create cluster record: Expected error, none occured")
	}
	// Check
	expectedError := errorBaseAlreadyExists + "Cluster name already in use"
	if expectedError != err.Error() {
		t.Fatalf("Create cluster record: Expected name error, received: %v", err.Error())
	}
}

func TestCreateClusterWithControlPlaneSshkey(t *testing.T) {
	// Setup mock
	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.CreateNewCluster").WithValues(logkeys.ClusterName, clusterCreate.Name, logkeys.CloudAccountId, clusterCreate.CloudAccountId).Start()
	defer span.End()

	iksSrv, err := InitSshMockClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with ssh mocking failed")
	}
	// SET UP
	serviceCIDR := "10.43.1.0/16"
	clusterCIDR := "10.43.1.0/16"
	clusterDNS := "10.43.1.1"
	region := "region"
	clusterCPWithSSHKeyName := "clusterCPWithSshKey"
	clusterCreate.Network = &pb.Network{
		Servicecidr: &serviceCIDR,
		Clustercidr: &clusterCIDR,
		Clusterdns:  &clusterDNS,
		Region:      region,
	}
	clusterCreate.Name = clusterCPWithSSHKeyName
	// Test
	resp, err := iksSrv.CreateNewCluster(context.Background(), &clusterCreate)
	if err != nil {
		t.Fatalf("could not create clustser %v", err)
	}

	// Replace "cl" with "cp" in the beginning of resp.Uuid
	clusterID := strings.Replace(resp.Uuid, "cl", "cp", 1)

	// CHECK
	sqlDB, err := managedDb.Open(context.Background())
	var sshkey []byte

	sqlDB.QueryRow("SELECT sshkey FROM nodegroup WHERE unique_id = $1", clusterID).Scan(
		&sshkey,
	)
	if sshkey == nil {
		t.Fatalf("Create cluster record: ssh key for cp can not be nil")
	}
	defer sqlDB.Close()
}

/* NODEGROUPS UNIT TESTS */
var uri = "https://"
var nodeGroupCreate = pb.CreateNodeGroupRequest{
	Clusteruuid:    "uuid",
	Name:           "nodegroup1",
	Description:    &desc,
	Instancetypeid: "vm-spr-tny",
	Count:          2,
	CloudAccountId: "iks_user",
	Vnets: []*pb.Vnet{ // Must matche return value in NewMockVnetServiceClient return function
		{
			Availabilityzonename:     "us-dev-1a",
			Networkinterfacevnetname: "us-dev-1a-default",
		},
	},
	Userdataurl: &uri,
}

func TestCreateNodegroupClusterNotFound(t *testing.T) {
	// Initialize compute client mocking

	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.CreateNodegroup").WithValues(logkeys.NodeGroupName, nodeGroupCreate.Name, logkeys.CloudAccountId, nodeGroupCreate.CloudAccountId).Start()
	defer span.End()

	iksSrv, err := InitVnetMockClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with compute mocking failed")
	}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	rows := sqlmock.NewRows([]string{"cluster_id", "clusterType"}).AddRow(1, "generalpurpose")
	mock.ExpectQuery(db_query.GetClusterTypeQuery).WillReturnRows(rows)

	// Test
	_, err = iksSrv.CreateNodeGroup(context.Background(), &nodeGroupCreate)
	if err == nil {
		t.Fatalf("Expected Error when none recieved")
	}

	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + nodeGroupCreate.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Create cluster record: Expected name error, received: %v", err.Error())
	}
}
func TestCreateNodegroupNoPermissions(t *testing.T) {
	// Initialize compute client mocking
	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.CreateNodegroup").WithValues(logkeys.NodeGroupName, nodeGroupCreate.Name, logkeys.CloudAccountId, "fake_user").Start()
	defer span.End()

	iksSrv, err := InitVnetMockClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with compute mocking failed")
	}

	// Set up
	nodeGroupCreate.Clusteruuid = clusterUuid
	original := nodeGroupCreate.CloudAccountId
	nodeGroupCreate.CloudAccountId = "fake_user"
	// Test
	_, err = iksSrv.CreateNodeGroup(context.Background(), &nodeGroupCreate)
	nodeGroupCreate.CloudAccountId = original
	if err == nil {
		t.Fatalf("Expected Error when none recieved")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + clusterUuid
	if expectedError != err.Error() {
		t.Fatalf("Create cluster record: Expected name error, received: %v", err.Error())
	}

}

func TestCreateNodegroupNotInActionableState(t *testing.T) {
	// Initialize compute client mocking
	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.CreateNodegroup").WithValues(logkeys.NodeGroupName, nodeGroupCreate.Name, logkeys.CloudAccountId, nodeGroupCreate.CloudAccountId).Start()
	defer span.End()

	iksSrv, err := InitVnetMockClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with compute mocking failed")
	}

	// Set up
	nodeGroupCreate.Clusteruuid = clusterUuid
	// Test
	_, err = iksSrv.CreateNodeGroup(context.Background(), &nodeGroupCreate)
	if err == nil {
		t.Fatalf("Expected Error when none recieved")
	}
	// Check
	expectedError := errorBaseFailedPrecondition + "Cluster not in actionable state"
	if expectedError != err.Error() {
		t.Fatalf("Create cluster record: Expected name error, received: %v", err.Error())
	}

}

func TestCreateNodeGroupNotInComputeInstanceType(t *testing.T) {
	// Initialize compute client mocking
	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.CreateNodegroup").WithValues(logkeys.NodeGroupName, nodeGroupCreate.Name, logkeys.CloudAccountId, nodeGroupCreate.CloudAccountId).Start()
	defer span.End()

	iksSrv, err := InitComputeMockClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with compute mocking failed")
	}

	//Set up
	nodeGroupCreate.Instancetypeid = "bm-icp-gaudi"

	// Test
	_, err = iksSrv.CreateNodeGroup(context.Background(), &nodeGroupCreate)
	if err == nil {
		t.Fatalf("Expected Error when none recieved")
	}
	// Check
	expectedError := "Instance types does not match to compute instance types"
	if expectedError != err.Error() {
		t.Fatalf("Create Nodegroup record: Expected instance types error, received: %v", err.Error())
	}
}

func TestCreateNodeGroupNotInIKSActiveInstanceType(t *testing.T) {
	// Initialize compute client mocking
	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.CreateNodegroup").WithValues(logkeys.NodeGroupName, nodeGroupCreate.Name, logkeys.CloudAccountId, nodeGroupCreate.CloudAccountId).Start()
	defer span.End()

	iksSrv, err := InitVnetMockClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with compute mocking failed")
	}

	// Set up
	nodeGroupCreate.Instancetypeid = "vm-spr-lrg"

	// Test
	_, err = iksSrv.CreateNodeGroup(context.Background(), &nodeGroupCreate)
	if err == nil {
		t.Fatalf("Expected Error when none recieved")
	}
	// Check
	expectedError := "Instance Type is not supported by iks"
	if expectedError != err.Error() {
		t.Fatalf("Create Nodegroup record: Expected instance types error, received: %v", err.Error())
	}
}

func TestCreateNodeGroupCountValidation(t *testing.T) {
	// Set up
	nodeGroupCreate.Instancetypeid = "vm-spr-tny"

	err := setActionableState(clusterUuid)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
	err = testsixnodegroups(clusterUuid)
	if err != nil {
		t.Fatalf("testsixnodegroups failed: %v", err.Error())
	}

	// Initialize compute client mocking
	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.CreateNodegroup").WithValues(logkeys.NodeGroupName, nodeGroupCreate.Name, logkeys.CloudAccountId, nodeGroupCreate.CloudAccountId).Start()
	defer span.End()

	iksSrv, err := InitVnetMockClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with compute mocking failed")
	}

	// Set up
	nodeGroupCreate.Name = "nodegroup1"

	// Test
	_, err = iksSrv.CreateNodeGroup(context.Background(), &nodeGroupCreate)
	if err == nil {
		t.Fatalf("Expected Error when none recieved")
	}
	// Check
	expectedError := errorBasePermissionDenied + "Can not create more than 5 nodegroups for this cluster"
	if expectedError != err.Error() {
		t.Fatalf("Create Nodegroup record: Expected create nodegroup error, received: %v", err.Error())
	}
}

func TestCreateNodegroupSuccessAndCheckDesiredJsonAndDefaultValues(t *testing.T) {
	err := deletenodegroups(clusterUuid)
	if err != nil {
		t.Fatalf("delete nodegroups failed")
	}
	// Initialize compute client mocking
	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.CreateNodegroup").WithValues(logkeys.NodeGroupName, nodeGroupCreate.Name, logkeys.CloudAccountId, nodeGroupCreate.CloudAccountId).Start()
	defer span.End()

	iksSrv, err := InitVnetMockClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with compute mocking failed")
	}
	// Set up
	nodeGroupCreate.Name = "nodegroup1"
	nodeGroupCreate.Clusteruuid = clusterUuid
	nodeGroupCreate.Instancetypeid = "vm-spr-tny"
	err = setActionableState(clusterUuid)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
	// Test
	res, err := iksSrv.CreateNodeGroup(context.Background(), &nodeGroupCreate)
	if err != nil {
		t.Fatalf("Should not error: %v", err.Error())
	}
	// Check
	sqlDB, err := managedDb.Open(context.Background())
	var (
		desiredJson []byte
		clusterCrd  clusterv1alpha.Cluster
	)
	err = sqlDB.QueryRow(getLatestClusterRev,
		res.Clusteruuid,
	).Scan(&desiredJson)
	defer sqlDB.Close()
	if err != nil {
		t.Fatalf("Should not error: %v", err.Error())
	}
	err = json.Unmarshal([]byte(desiredJson), &clusterCrd)
	if err != nil {
		t.Fatalf("Should not error: %v", err.Error())
	}
	if clusterCrd.Spec.Nodegroups == nil || len(clusterCrd.Spec.Nodegroups) == 0 {
		t.Fatalf("Nodegroups should not be empty")
	}
	if clusterCrd.Spec.Nodegroups[0].Name != res.Nodegroupuuid {
		t.Fatalf("Nodegroups UUID does not match")
	}
	if clusterCrd.Spec.Nodegroups[0].Count != 2 {
		t.Fatalf("Nodegroup count does not match")
	}
}

/* VIPS UNIT TESTS*/
var vipCreate = pb.VipCreateRequest{
	Clusteruuid:    "uuid-1",
	Name:           "vip1",
	Description:    "Description",
	Port:           443,
	Viptype:        "private",
	CloudAccountId: "iks_user",
}

func TestCreateVipClusterNotFound(t *testing.T) {
	// SET UP
	// TEST
	_, err := client.CreateNewVip(context.Background(), &vipCreate)
	if err == nil {
		t.Fatalf("Expected Error when none recieved")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + vipCreate.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Create cluster record: Expected name error, received: %v", err.Error())
	}
}
func TestCreateVipPermissions(t *testing.T) {
	// SET UP
	vipCreate.Clusteruuid = clusterUuid
	original := vipCreate.CloudAccountId
	vipCreate.CloudAccountId = "fake_user"
	// TEST
	_, err := client.CreateNewVip(context.Background(), &vipCreate)
	vipCreate.CloudAccountId = original
	if err == nil {
		t.Fatalf("Expected Error when none recieved")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + clusterUuid
	if expectedError != err.Error() {
		t.Fatalf("Create cluster record: Expected not found error, received: %v", err.Error())
	}
}
func TestCreateVipWrongVipType(t *testing.T) {
	// SET UP
	vipCreate.Clusteruuid = clusterUuid
	vipCreate.Viptype = "fake_vip"
	// TEST
	_, err := client.CreateNewVip(context.Background(), &vipCreate)
	if err == nil {
		t.Fatalf("Expected Error when none recieved")
	}
	// Check
	expectedError := errorBaseInvalidArgument + "Vip type should be 'public' or 'private'"
	if expectedError != err.Error() {
		t.Fatalf("Create cluster record: Expected name error, received: %v", err.Error())
	}
}
func TestCreateVipActionable(t *testing.T) {
	// SET UP
	vipCreate.Clusteruuid = clusterUuid
	original := vipCreate.CloudAccountId
	vipCreate.CloudAccountId = "iks_user"
	vipCreate.Viptype = "public"
	// TEST
	_, err := client.CreateNewVip(context.Background(), &vipCreate)
	vipCreate.CloudAccountId = original
	if err == nil {
		t.Fatalf("Expected Error when none recieved")
	}
	// Check
	expectedError := errorBaseFailedPrecondition + "Cluster not in actionable state"
	if expectedError != err.Error() {
		t.Fatalf("Create cluster record: Expected name error, received: %v", err.Error())
	}
}

func TestCreatePublicVipValidateCount(t *testing.T) {
	// SET UP
	vipCreate.Clusteruuid = clusterUuid
	errCount := 0

	err := setActionableState(clusterUuid)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}

	expectedError := errorBasePermissionDenied + "Can not create more than 2 User Vips for this cluster"
	expectedErrCount := 1 // Expecting to get exactly one error due to quota limit

	// TEST
	_, err = client.CreateNewVip(context.Background(), &vipCreate)
	if err != nil {
		t.Fatalf("Create vip failed")
	}

	err = setActionableState(clusterUuid)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}

	var wg sync.WaitGroup
	wg.Add(2)

	errCh := make(chan error, 2)

	// Create 2 more Vips in parallel
	// Testing in parallel in order to verify the case when multiple create vip requests are made near the same time and all of them wrongly succeed despite one should hit quota limit
	// That's may happen if vip table is not locked properly for reading vip count (see:  https://jira.devtools.intel.com/browse/IDCK8S-9)

	vip2Create := pb.VipCreateRequest{
		Clusteruuid:    vipCreate.Clusteruuid,
		Name:           "vip2",
		Description:    "Description",
		Port:           443,
		Viptype:        "private",
		CloudAccountId: "iks_user",
	}

	vip3Create := pb.VipCreateRequest{
		Clusteruuid:    vipCreate.Clusteruuid,
		Name:           "vip3",
		Description:    "Description",
		Port:           443,
		Viptype:        "private",
		CloudAccountId: "iks_user",
	}
	go func() {
		defer wg.Done()
		var err error
		_, err = client.CreateNewVip(context.Background(), &vip2Create)
		// Check
		if err != nil {
			if expectedError != err.Error() { // only count error or no error is expected
				errCh <- fmt.Errorf("create vip record: Expected count error, received: %v", err.Error())
				return
			} else {
				errCount++ // Got expected error, increment error count
			}
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		_, err = client.CreateNewVip(context.Background(), &vip3Create)
		// Check
		if err != nil {
			if expectedError != err.Error() { // only count error or no error is expected
				errCh <- fmt.Errorf("create vip record: Expected count error, received: %v", err.Error())
			} else {
				errCount++ // Got expected error, increment error count
			}
		}
	}()

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatal(err)
	}

	// Check that we got exactly one error
	if errCount != expectedErrCount {
		t.Fatalf("Create vip record: Expected error count: %d, received: %d", expectedErrCount, errCount)
	}
}

func TestCreatePublicVipValidateDuplicate(t *testing.T) {
	// SET UP
	err := deletevips(clusterUuid)
	if err != nil {
		t.Fatalf("Could not delete vips: %v", err.Error())
	}

	errCount := 0

	expectedError := errorBaseFailedPrecondition + "Vip name already in use"
	expectedErrCount := 1 // Expecting to get exactly one error due to duplicate vip name

	// Test
	err = setActionableState(clusterUuid)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}

	var wg sync.WaitGroup
	wg.Add(2)

	errCh := make(chan error, 2)

	// Create 2 more Vips in parallel
	// Testing in parallel in order to verify the case when multiple create vip requests are made near the same time and all of them wrongly succeed despite one should hit quota limit
	// That's may happen if vip table is not locked properly for validating vip name (see:  https://jira.devtools.intel.com/browse/IDCK8S-9)
	vipDuplicate := pb.VipCreateRequest{
		Clusteruuid:    vipCreate.Clusteruuid,
		Name:           "duplicateName",
		Description:    "Description",
		Port:           443,
		Viptype:        "private",
		CloudAccountId: "iks_user",
	}
	go func() {
		defer wg.Done()
		_, err = client.CreateNewVip(context.Background(), &vipDuplicate)
		// Check
		if err != nil {
			if expectedError != err.Error() { // only duplicate name error or no error is expected
				errCh <- fmt.Errorf("create vip record: Expected duplicate name error, received: %v", err.Error())
				return
			} else {
				errCount++ // Got expected error, increment error count
			}
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		_, err = client.CreateNewVip(context.Background(), &vipDuplicate)
		// Check
		if err != nil {
			if expectedError != err.Error() { // only duplicate name error or no error is expected
				errCh <- fmt.Errorf("create vip record: Expected duplicate name error, received: %v", err.Error())
			} else {
				errCount++ // Got expected error, increment error count
			}
		}
	}()

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatal(err)
	}

	// Check that we got exactly one error
	if errCount != expectedErrCount {
		t.Fatalf("Duplicate  vip name: Expected error count: %d, received: %d", expectedErrCount, errCount)
	}
}

func TestCreateVipSuccess(t *testing.T) {
	err := deletevips(clusterUuid)
	if err != nil {
		t.Fatalf("Could not delete vips: %v", err.Error())
	}
	// SET UP
	vipCreate.Clusteruuid = clusterUuid
	vipCreate.Viptype = "public"
	err = setActionableState(clusterUuid)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}

	// TEST
	res, err := client.CreateNewVip(context.Background(), &vipCreate)
	if err != nil {
		t.Fatalf("Should not error: %v", err.Error())
	}
	sqlDB, err := managedDb.Open(context.Background())

	// CHECK CRD
	var (
		desiredJson []byte
		clusterCrd  clusterv1alpha.Cluster
	)
	err = sqlDB.QueryRow(getLatestClusterRev,
		clusterUuid,
	).Scan(&desiredJson)
	if err != nil {
		t.Fatalf("Should not error: %v", err.Error())
	}
	err = json.Unmarshal([]byte(desiredJson), &clusterCrd)
	if err != nil {
		t.Fatalf("Should not error: %v", err.Error())
	}
	added := false
	for _, ilb := range clusterCrd.Spec.ILBS {
		if ilb.Name == vipCreate.Name {
			added = true
			break
		}
	}
	if !added {
		t.Fatalf("Vip %v was not added to CRD", vipCreate.Name)
	}
	// CHECK DB
	count := 0
	err = sqlDB.QueryRow(`SELECT COUNT(*) FROM public.vip WHERE vip_id = $1`,
		res.Vipid,
	).Scan(&count)
	if count != 1 {
		t.Fatalf("Vip %v was not added to Database", vipCreate.Name)
	}
}

//Storage tests

var clusterStorageRequest = pb.ClusterStorageRequest{
	Clusteruuid:    "empty",
	CloudAccountId: "iks_user",
	Enablestorage:  true,
	Storagesize:    "1TB",
}

func TestMockEnableClusterStorageClusterNotFound(t *testing.T) {
	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.EnableClusterStorage").WithValues(logkeys.ClusterId, clusterUuid, logkeys.CloudAccountId, clusterStorageRequest.CloudAccountId).Start()
	defer span.End()

	iksSrv, err := InitProductCatalogClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with product catalog mocking failed")
	}

	// Test
	_, err = iksSrv.EnableClusterStorage(context.Background(), &clusterStorageRequest)
	if err == nil {
		t.Fatalf("Expected Error when none recieved")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: empty"
	if expectedError != err.Error() {
		t.Fatalf(errorBaseNotFound+"Enable Cluster Storage: Expected Cluster not found error, received: %v", err.Error())
	}

}

func TestMockEnableClusterStorageFailureRestriction(t *testing.T) {
	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.EnableClusterStorage").WithValues(logkeys.ClusterId, clusterUuid, logkeys.CloudAccountId, clusterStorageRequest.CloudAccountId).Start()
	defer span.End()

	iksSrv, err := InitProductCatalogClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with product catalog mocking failed")
	}
	// Setup
	clusterStorageRequest.Clusteruuid = clusterUuid

	// Test
	_, err = iksSrv.EnableClusterStorage(context.Background(), &clusterStorageRequest)
	if err == nil {
		t.Fatalf("Expected Error when none recieved")
	}
	// Check
	expectedError := errorBasePermissionDenied + "Due to storage restrictions, we are currently not allowing non-approved users to use storage."
	if expectedError != err.Error() {
		t.Fatalf("Enable Cluster Storage: Expected Storage Restriction Error, received: %v", err.Error())
	}

}

func TestMockEnableClusterStorageSuccess(t *testing.T) {
	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.EnableClusterStorage").WithValues(logkeys.ClusterId, clusterUuid, logkeys.CloudAccountId, clusterStorageRequest.CloudAccountId).Start()
	defer span.End()

	iksSrv, err := InitProductCatalogClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with product catalog mocking failed")
	}

	// Setup
	clusterStorageRequest.Clusteruuid = clusterUuid

	err = setAllowCreateStorageTrue(clusterStorageRequest.CloudAccountId)
	if err != nil {
		t.Fatalf("Cannot Set Enable Storage to true, received: %v", err.Error())
	}

	err = setActionableState(clusterUuid)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}

	// Test
	storageStatus, err := iksSrv.EnableClusterStorage(context.Background(), &clusterStorageRequest)
	if err != nil {
		t.Fatalf("Expected no error, received: %v", err.Error())
	}
	// Check
	expectedSize := "1TB"
	if expectedSize != storageStatus.Size {
		t.Fatalf("Enable Cluster Storage: Expected Storage Size of 1TB, received: %s", storageStatus.Size)
	}

}

var updateClusterStorageRequest = pb.ClusterStorageUpdateRequest{
	Clusteruuid:    "",
	CloudAccountId: "iks_user",
	Storagesize:    "1TB",
}

func TestMockUpdateClusterStorageEqualStorageFailure(t *testing.T) {
	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.UpdateClusterStorage").WithValues(logkeys.ClusterId, clusterUuid, logkeys.CloudAccountId, clusterStorageRequest.CloudAccountId).Start()
	defer span.End()

	iksSrv, err := InitProductCatalogClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with product catalog mocking failed")
	}

	// Setup
	updateClusterStorageRequest.Clusteruuid = clusterUuid

	// Test
	_, err = iksSrv.UpdateClusterStorage(context.Background(), &updateClusterStorageRequest)
	if err == nil {
		t.Fatalf("Expected Error when none recieved, received: %v", err.Error())
	}
	// Check
	expectedError := "Requested storage size is equal to the current storage size. New storage size should be greater than current storage size."
	if expectedError != err.Error() {
		t.Fatalf("Update Cluster Storage: Expected size restriction error, received: %s", err.Error())
	}

}

func TestMockUpdateClusterStorageSuccess(t *testing.T) {
	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.UpdateClusterStorage").WithValues(logkeys.ClusterId, clusterUuid, logkeys.CloudAccountId, clusterStorageRequest.CloudAccountId).Start()
	defer span.End()

	iksSrv, err := InitProductCatalogClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with product catalog mocking failed")
	}

	// Setup
	updateClusterStorageRequest.Clusteruuid = clusterUuid
	updateClusterStorageRequest.Storagesize = "2TB"

	err = setActionableState(clusterUuid)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}

	// Test
	storageStatus, err := iksSrv.UpdateClusterStorage(context.Background(), &updateClusterStorageRequest)
	if err != nil {
		t.Fatalf("Expected no error, received: %v", err.Error())
	}
	// Check
	expectedSize := "2TB"
	if expectedSize != storageStatus.Size {
		t.Fatalf("Update Cluster Storage: Expected Storage Size of 2TB, received: %s", storageStatus.Size)
	}

}

/* CLEAN UP UNIT TESTS */
func TestDBCleanup(t *testing.T) {
	err := removeClusterFromDB(clusterUuid)
	if err != nil {
		t.Fatalf("Could not clean up entry %v", err.Error())
	}
}

/* MOCK TESTS */
func mockCreateClusterValues(mock sqlmock.Sqlmock, stopLocation int) {
	rows := sqlmock.NewRows([]string{"restric_create_cluster"}).AddRow(false)
	mock.ExpectQuery(utils.GetDefaultConfigUtilsQuery).WillReturnRows(rows)
	rows = sqlmock.NewRows([]string{"unique_id", "clusterstate_name"}).AddRow("1", "Deleted")
	mock.ExpectQuery(query.GetClustersStatesByName).WillReturnRows(rows)
	if stopLocation == 2 {
		return
	}
	rows = sqlmock.NewRows([]string{"provider_name"}).AddRow("iks")
	mock.ExpectQuery(query.GetCloudAccountProvider).WillReturnRows(rows)
	if stopLocation == 3 {
		return
	}
	fakeValues := `[
		{"name": "region", "value": "us-region-1"},
		{"name": "vnet", "value": "us-region-1a"},
		{"name": "availabilityzone", "value": "us-region-1a-default"},
		{"name": "networkinterfacename", "value": "eth0"},
		{"name": "ilb_environment", "value": "8"},
		{"name": "ilb_customer_environment", "value": "9"},
		{"name": "ilb_ipprotocol", "value": "tcp"},
		{"name": "ilb_usergroup", "value": "1545"},
		{"name": "ilb_customer_usergroup", "value": "15455"},
		{"name": "ilb_apiservername", "value": "apiserver"},
		{"name": "ilb_apiserverport", "value": "6443"},
		{"name": "ilb_public_apiserverport", "value": "6443"},
		{"name": "ilb_apiserverpool_port", "value": "6443"},
		{"name": "ilb_public_apiserverpool_port", "value": "6443"},
		{"name": "ilb_etcdservername", "value": "etcd"},
		{"name": "ilb_etcdport", "value": "2379"},
		{"name": "ilb_etcdpool_port", "value": "2380"},
		{"name": "ilb_konnectivityname", "value": "konnectivity"},
		{"name": "ilb_konnectivityport", "value": "8132"},
		{"name": "ilb_konnectivitypool_port", "value": "8132"},
		{"name": "ilb_loadbalancingmode", "value": "least-connections-member"},
		{"name": "ilb_minactivemembers", "value": "1"},
		{"name": "ilb_monitor", "value": "i_tcp"},
		{"name": "ilb_memberConnectionLimit", "value": "0"},
		{"name": "ilb_memberPriorityGroup", "value": "0"},
		{"name": "ilb_memberratio", "value": "1"},
		{"name": "ilb_memberadminstate", "value": "enabled"},
		{"name": "ilb_persist", "value": null},
		{"name": "networkservice_cidr", "value": "100.66.0.0/16"},
		{"name": "networkpod_cidr", "value": "100.68.0.0/16"},
		{"name": "cluster_cidr", "value": "100.66.0.10"},
		{"name": "ilb_allowed_ports","value": "80,443"},
		{"name": "max_cust_cluster_ilb","value": "2"},
		{"name": "max_cluster_ng", "value": "5"},
		{"name": "max_cluster_vm", "value": "50"},
		{"name": "max_nodegroup_vm","value": "10"},
		{"name": "max_cluster","value": "3"}
	]`
	rows = sqlmock.NewRows([]string{"defaultvalues"}).AddRow(fakeValues)
	mock.ExpectQuery(`SELECT to_jsonb(array_agg(t)) as defaultvalues FROM public.defaultconfig t`).WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"maxclusters_override", "maxclusterng_override", "maxclusterilb_oberride", "maxclustervm_override", "maxnodegroupvm_override"}).AddRow(-1, -1, -1, -1, -1)
	mock.ExpectQuery(utils.GetCloudAccountExtraSpecMaxValues).WillReturnRows(rows)

	if stopLocation == 5 {
		return
	}
	rows = sqlmock.NewRows([]string{"instncetype_name", "nodeprovider_name"}).AddRow("vm-spr-tny", "Compute")
	mock.ExpectQuery(query.GetDefaultCpInstanceTypeAndNodeProvider).WillReturnRows(rows)
	rows = sqlmock.NewRows([]string{"osimage_name"}).AddRow("Ubuntu-22-04")
	mock.ExpectQuery(query.GetDefaultCpOsImageQuery).WillReturnRows(rows)
	if stopLocation == 4 {
		return
	}
	rows = sqlmock.NewRows([]string{"cp_osimageinstance_name", "provider_name"}).AddRow("cpOsImageInstance1", "providerName")
	mock.ExpectQuery(query.GetDefaultCpOsImageInstance + " AND ks.minor_version = $5").WillReturnRows(rows)
	if stopLocation == 6 {
		return
	}
	rows = sqlmock.NewRows([]string{"imiartifact"}).AddRow("imiartifact")
	mock.ExpectQuery(query.GetImiArtifactQuery).WillReturnRows(rows)
	if stopLocation == 8 {
		return
	}

	rows = sqlmock.NewRows([]string{"name", "install_type", "artifact_repo"}).AddRow("name1", "type1", "artifact_repo1")
	mock.ExpectQuery(utils.GetAddonsQuery).WillReturnRows(rows)
	mock.ExpectBegin()
	//rows = sqlmock.NewRows([]string{"nodegroup_id"})
	//mock.ExpectQuery(query.GetNodeGroupCountByName)

	mock.ExpectExec(query.LockClusterTable).WillReturnResult(sqlmock.NewResult(0, 0))

	rows = sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(query.GetClusterCounts).WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(utils.GetClusterUuidCheck).WillReturnRows(rows)
	if stopLocation == 9 {
		return
	}
	rows = sqlmock.NewRows([]string{"cluster_id"}).AddRow(123)
	mock.ExpectQuery(query.InsertClusterRecordQuery).WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{})
	mock.ExpectQuery(query.InsertSshkeyQuery).WillReturnRows(rows)
	if stopLocation == 10 {
		return
	}

	rows = sqlmock.NewRows([]string{"nodegroup_id"}).AddRow(123)
	mock.ExpectQuery(query.InsertControlPlaneQuery).WillReturnRows(rows)
	if stopLocation == 11 {
		return
	}
	rows = sqlmock.NewRows([]string{"vipprovider_name"}).AddRow("vipProvider1")
	mock.ExpectQuery(query.GetVipProviderDefaults).WillReturnRows(rows)
	if stopLocation == 12 {
		return
	}
	rows = sqlmock.NewRows([]string{"vip_id"}).AddRow(1)
	mock.ExpectQuery(query.InsertVipQuery).WillReturnRows(rows)
	if stopLocation == 13 {
		return
	}
	rows = sqlmock.NewRows([]string{"vip_id"}).AddRow(1)
	mock.ExpectQuery(query.InsertVipDetailsQuery).WillReturnRows(rows)
	rows = sqlmock.NewRows([]string{"vip_id"}).AddRow(2)
	mock.ExpectQuery(query.InsertVipQuery).WillReturnRows(rows)
	rows = sqlmock.NewRows([]string{"vip_id"}).AddRow(2)
	mock.ExpectQuery(query.InsertVipDetailsQuery).WillReturnRows(rows)
	rows = sqlmock.NewRows([]string{"vip_id"}).AddRow(3)
	mock.ExpectQuery(query.InsertVipQuery).WillReturnRows(rows)
	rows = sqlmock.NewRows([]string{"vip_id"}).AddRow(3)
	mock.ExpectQuery(query.InsertVipDetailsQuery).WillReturnRows(rows)
	rows = sqlmock.NewRows([]string{"vip_id"}).AddRow(4)
	mock.ExpectQuery(query.InsertVipQuery).WillReturnRows(rows)
	rows = sqlmock.NewRows([]string{"vip_id"}).AddRow(4)
	mock.ExpectQuery(query.InsertVipDetailsQuery).WillReturnRows(rows)
	if stopLocation == 14 {
		return
	}
	rows = sqlmock.NewRows([]string{"clusterrev_id"}).AddRow(123)
	mock.ExpectQuery(query.InsertRevQuery).WillReturnRows(rows)
	if stopLocation == 15 {
		return
	}
	rows = sqlmock.NewRows([]string{"provisioninglog_id"})
	mock.ExpectQuery(query.InsertProvisioningQuery).WillReturnRows(rows)
	mock.ExpectCommit()

}

func mockCreateFirewallValues(mock sqlmock.Sqlmock, stopLocation int) {
	if stopLocation == 1 {
		return
	}
	rows := sqlmock.NewRows([]string{"vip_id", "viptype_name"}).AddRow(59, "Public")
	mock.ExpectQuery(query.GetvipidForFirewallQuery).WillReturnRows(rows)

	if stopLocation == 2 {
		return
	}
	rows = sqlmock.NewRows([]string{"vip_ip", "port"}).AddRow(59, 443)
	mock.ExpectQuery(query.GetVipPortQuery).WillReturnRows(rows)

	if stopLocation == 3 {
		return
	}
	rows = sqlmock.NewRows([]string{"vip_ip"}).AddRow("10.20.30.40")
	mock.ExpectQuery(query.GetdestipQuery).WillReturnRows(rows)
	mock.ExpectCommit()
}

func TestMockCreateGetClusterStatesByName(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	rows := sqlmock.NewRows([]string{"restric_create_cluster"}).AddRow(false)
	mock.ExpectQuery(utils.GetDefaultConfigUtilsQuery).WillReturnRows(rows)
	mock.ExpectQuery(query.GetClustersStatesByName)
	// SET UP
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, nil, "test-key", "test")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected an generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}

}

func TestMockCreateGetClusterStatesByNameScan(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 2)
	// SET UP
	clusterCreate.Name = "test-cluster"
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, nil, "test-key", "test")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}

func TestMockCreateGetCloudAccountProvider(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 2)
	mock.ExpectQuery(query.GetCloudAccountProvider)
	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, nil, "test-key", "test")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}
func TestMockCreateGetDefaults(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 3)
	mock.ExpectQuery(`SELECT to_jsonb(array_agg(t)) as defaultvalues FROM public.defaultconfig t`)

	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, nil, "test-key", "test")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}

func TestMockCreateGetDefaultsParse1(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 3)
	fakeValues := `[{"name": "region", "value": "us-region-1"}, {"name": "vnet", "value": "us-region-1a"}, {"name": "availabilityzone", "value": "us-region-1a-default"}, {"name": "networkinterfacename", "value": "eth0"}, {"name": "ilb_ipprotocol", "value": "tcp"}, {"name": "ilb_usergroup", "value": "1545"}, {"name": "ilb_apiservername", "value": "apiserver"}, {"name": "ilb_apiserverport", "value": "6443"}, {"name": "ilb_etcdservername", "value": "etcd"}, {"name": "ilb_etcdport", "value": "2379"}, {"name": "ilb_konnectivityname", "value": "konnectivity"}, {"name": "ilb_konnectivityport", "value": "8132"}, {"name": "ilb_loadbalancingmode", "value": "least-connections-member"}, {"name": "ilb_minactivemembers", "value": "1"}, {"name": "ilb_monitor", "value": "i_tcp"}, {"name": "ilb_memberConnectionLimit", "value": "0"}, {"name": "ilb_memberPriorityGroup", "value": "0"}, {"name": "ilb_memberratio", "value": "1"}, {"name": "ilb_memberadminstate", "value": "enabled"}, {"name": "ilb_persist", "value": null}, {"name": "networkservice_cidr", "value": "100.66.0.0/16"}, {"name": "networkpod_cidr", "value": "100.68.0.0/16"}, {"name": "cluster_cidr", "value": "100.66.0.10"}]`
	rows3 := sqlmock.NewRows([]string{"defaultvalues"}).
		AddRow(fakeValues)
	mock.ExpectQuery(`SELECT to_jsonb(array_agg(t)) as defaultvalues FROM public.defaultconfig t`).WillReturnRows(rows3)
	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, nil, "test-key", "test")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}

func TestMockCreateGetDefaultsParse2(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 3)
	fakeValues := `[{"name": "region", "value": "us-region-1"}, {"name": "vnet", "value": "us-region-1a"}, {"name": "availabilityzone", "value": "us-region-1a-default"}, {"name": "networkinterfacename", "value": "eth0"}, {"name": "ilb_environment", "value": "8"}, {"name": "ilb_ipprotocol", "value": "tcp"}, {"name": "ilb_apiservername", "value": "apiserver"}, {"name": "ilb_apiserverport", "value": "6443"}, {"name": "ilb_etcdservername", "value": "etcd"}, {"name": "ilb_etcdport", "value": "2379"}, {"name": "ilb_konnectivityname", "value": "konnectivity"}, {"name": "ilb_konnectivityport", "value": "8132"}, {"name": "ilb_loadbalancingmode", "value": "least-connections-member"}, {"name": "ilb_minactivemembers", "value": "1"}, {"name": "ilb_monitor", "value": "i_tcp"}, {"name": "ilb_memberConnectionLimit", "value": "0"}, {"name": "ilb_memberPriorityGroup", "value": "0"}, {"name": "ilb_memberratio", "value": "1"}, {"name": "ilb_memberadminstate", "value": "enabled"}, {"name": "ilb_persist", "value": null}, {"name": "networkservice_cidr", "value": "100.66.0.0/16"}, {"name": "networkpod_cidr", "value": "100.68.0.0/16"}, {"name": "cluster_cidr", "value": "100.66.0.10"}]`
	rows3 := sqlmock.NewRows([]string{"defaultvalues"}).
		AddRow(fakeValues)
	mock.ExpectQuery(`SELECT to_jsonb(array_agg(t)) as defaultvalues FROM public.defaultconfig t`).WillReturnRows(rows3)
	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, nil, "test-key", "test")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}

func TestMockCreateGetDefaultsParse3(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 3)
	fakeValues := `[{"name": "region", "value": "us-region-1"}, {"name": "vnet", "value": "us-region-1a"}, {"name": "availabilityzone", "value": "us-region-1a-default"}, {"name": "networkinterfacename", "value": "eth0"}, {"name": "ilb_environment", "value": "8"}, {"name": "ilb_ipprotocol", "value": "tcp"}, {"name": "ilb_usergroup", "value": "1545"}, {"name": "ilb_apiservername", "value": "apiserver"}, {"name": "ilb_apiserverport", "value": "6443"}, {"name": "ilb_etcdservername", "value": "etcd"}, {"name": "ilb_etcdport", "value": "2379"}, {"name": "ilb_konnectivityname", "value": "konnectivity"}, {"name": "ilb_konnectivityport", "value": "8132"}, {"name": "ilb_loadbalancingmode", "value": "least-connections-member"}, {"name": "ilb_monitor", "value": "i_tcp"}, {"name": "ilb_memberPriorityGroup", "value": "0"}, {"name": "ilb_memberratio", "value": "1"}, {"name": "ilb_memberadminstate", "value": "enabled"}, {"name": "ilb_persist", "value": null}, {"name": "networkservice_cidr", "value": "100.66.0.0/16"}, {"name": "networkpod_cidr", "value": "100.68.0.0/16"}, {"name": "cluster_cidr", "value": "100.66.0.10"}]`
	rows3 := sqlmock.NewRows([]string{"defaultvalues"}).
		AddRow(fakeValues)
	mock.ExpectQuery(`SELECT to_jsonb(array_agg(t)) as defaultvalues FROM public.defaultconfig t`).WillReturnRows(rows3)
	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, nil, "test-key", "test")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}

func TestMockCreateGetDefaultsParse4(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 3)
	fakeValues := `[{"name": "region", "value": "us-region-1"}, {"name": "vnet", "value": "us-region-1a"}, {"name": "availabilityzone", "value": "us-region-1a-default"}, {"name": "networkinterfacename", "value": "eth0"}, {"name": "ilb_environment", "value": "8"}, {"name": "ilb_ipprotocol", "value": "tcp"}, {"name": "ilb_usergroup", "value": "1545"}, {"name": "ilb_apiservername", "value": "apiserver"}, {"name": "ilb_apiserverport", "value": "6443"}, {"name": "ilb_etcdservername", "value": "etcd"}, {"name": "ilb_etcdport", "value": "2379"}, {"name": "ilb_konnectivityname", "value": "konnectivity"}, {"name": "ilb_konnectivityport", "value": "8132"}, {"name": "ilb_loadbalancingmode", "value": "least-connections-member"}, {"name": "ilb_minactivemembers", "value": "1"}, {"name": "ilb_monitor", "value": "i_tcp"}, {"name": "ilb_memberPriorityGroup", "value": "0"}, {"name": "ilb_memberratio", "value": "1"}, {"name": "ilb_memberadminstate", "value": "enabled"}, {"name": "ilb_persist", "value": null}, {"name": "networkservice_cidr", "value": "100.66.0.0/16"}, {"name": "networkpod_cidr", "value": "100.68.0.0/16"}, {"name": "cluster_cidr", "value": "100.66.0.10"}]`
	rows3 := sqlmock.NewRows([]string{"defaultvalues"}).
		AddRow(fakeValues)
	mock.ExpectQuery(`SELECT to_jsonb(array_agg(t)) as defaultvalues FROM public.defaultconfig t`).WillReturnRows(rows3)
	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, nil, "test-key", "test")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}

func TestMockCreateGetDefaultsParse5(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 3)
	fakeValues := `[{"name": "region", "value": "us-region-1"}, {"name": "vnet", "value": "us-region-1a"}, {"name": "availabilityzone", "value": "us-region-1a-default"}, {"name": "networkinterfacename", "value": "eth0"}, {"name": "ilb_environment", "value": "8"}, {"name": "ilb_ipprotocol", "value": "tcp"}, {"name": "ilb_usergroup", "value": "1545"}, {"name": "ilb_apiservername", "value": "apiserver"}, {"name": "ilb_apiserverport", "value": "6443"}, {"name": "ilb_etcdservername", "value": "etcd"}, {"name": "ilb_etcdport", "value": "2379"}, {"name": "ilb_konnectivityname", "value": "konnectivity"}, {"name": "ilb_loadbalancingmode", "value": "least-connections-member"}, {"name": "ilb_minactivemembers", "value": "1"}, {"name": "ilb_monitor", "value": "i_tcp"}, {"name": "ilb_memberConnectionLimit", "value": "0"}, {"name": "ilb_memberratio", "value": "1"}, {"name": "ilb_memberadminstate", "value": "enabled"}, {"name": "ilb_persist", "value": null}, {"name": "networkservice_cidr", "value": "100.66.0.0/16"}, {"name": "networkpod_cidr", "value": "100.68.0.0/16"}, {"name": "cluster_cidr", "value": "100.66.0.10"}]`
	rows3 := sqlmock.NewRows([]string{"defaultvalues"}).
		AddRow(fakeValues)
	mock.ExpectQuery(`SELECT to_jsonb(array_agg(t)) as defaultvalues FROM public.defaultconfig t`).WillReturnRows(rows3)
	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, nil, "test-key", "test")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}

func TestMockCreateGetDefaultsParse6(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 3)
	fakeValues := `[{"name": "region", "value": "us-region-1"}, {"name": "vnet", "value": "us-region-1a"}, {"name": "availabilityzone", "value": "us-region-1a-default"}, {"name": "networkinterfacename", "value": "eth0"}, {"name": "ilb_environment", "value": "8"}, {"name": "ilb_ipprotocol", "value": "tcp"}, {"name": "ilb_usergroup", "value": "1545"}, {"name": "ilb_apiservername", "value": "apiserver"}, {"name": "ilb_apiserverport", "value": "6443"}, {"name": "ilb_etcdservername", "value": "etcd"}, {"name": "ilb_etcdport", "value": "2379"}, {"name": "ilb_konnectivityname", "value": "konnectivity"}, {"name": "ilb_konnectivityport", "value": "8132"}, {"name": "ilb_loadbalancingmode", "value": "least-connections-member"}, {"name": "ilb_minactivemembers", "value": "1"}, {"name": "ilb_monitor", "value": "i_tcp"}, {"name": "ilb_memberConnectionLimit", "value": "0"}, {"name": "ilb_memberPriorityGroup", "value": "0"}, {"name": "ilb_memberadminstate", "value": "enabled"}, {"name": "ilb_persist", "value": null}, {"name": "networkservice_cidr", "value": "100.66.0.0/16"}, {"name": "networkpod_cidr", "value": "100.68.0.0/16"}, {"name": "cluster_cidr", "value": "100.66.0.10"}]`
	rows3 := sqlmock.NewRows([]string{"defaultvalues"}).
		AddRow(fakeValues)
	mock.ExpectQuery(`SELECT to_jsonb(array_agg(t)) as defaultvalues FROM public.defaultconfig t`).WillReturnRows(rows3)

	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, nil, "test-key", "test")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}
func TestMockCreateGetDefaultsParse7(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 3)
	fakeValues := `[{"name": "region", "value": "us-region-1"}, {"name": "vnet", "value": "us-region-1a"}, {"name": "availabilityzone", "value": "us-region-1a-default"}, {"name": "networkinterfacename", "value": "eth0"}, {"name": "ilb_environment", "value": "8"}, {"name": "ilb_ipprotocol", "value": "tcp"}, {"name": "ilb_usergroup", "value": "1545"}, {"name": "ilb_apiservername", "value": "apiserver"}, {"name": "ilb_apiserverport", "value": "6443"}, {"name": "ilb_etcdservername", "value": "etcd"}, {"name": "ilb_konnectivityname", "value": "konnectivity"}, {"name": "ilb_loadbalancingmode", "value": "least-connections-member"}, {"name": "ilb_minactivemembers", "value": "1"}, {"name": "ilb_monitor", "value": "i_tcp"}, {"name": "ilb_memberConnectionLimit", "value": "0"}, {"name": "ilb_memberPriorityGroup", "value": "0"}, {"name": "ilb_memberratio", "value": "1"}, {"name": "ilb_memberadminstate", "value": "enabled"}, {"name": "ilb_persist", "value": null}, {"name": "networkservice_cidr", "value": "100.66.0.0/16"}, {"name": "networkpod_cidr", "value": "100.68.0.0/16"}, {"name": "cluster_cidr", "value": "100.66.0.10"}]`
	rows3 := sqlmock.NewRows([]string{"defaultvalues"}).
		AddRow(fakeValues)
	mock.ExpectQuery(`SELECT to_jsonb(array_agg(t)) as defaultvalues FROM public.defaultconfig t`).WillReturnRows(rows3)
	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, nil, "test-key", "test")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}
func TestMockCreateGetDefaultsParse8(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 3)
	fakeValues := `[{"name": "region", "value": "us-region-1"}, {"name": "vnet", "value": "us-region-1a"}, {"name": "availabilityzone", "value": "us-region-1a-default"}, {"name": "networkinterfacename", "value": "eth0"}, {"name": "ilb_environment", "value": "8"}, {"name": "ilb_ipprotocol", "value": "tcp"}, {"name": "ilb_usergroup", "value": "1545"}, {"name": "ilb_apiservername", "value": "apiserver"}, {"name": "ilb_etcdservername", "value": "etcd"}, {"name": "ilb_etcdport", "value": "2379"}, {"name": "ilb_konnectivityname", "value": "konnectivity"}, {"name": "ilb_loadbalancingmode", "value": "least-connections-member"}, {"name": "ilb_minactivemembers", "value": "1"}, {"name": "ilb_monitor", "value": "i_tcp"}, {"name": "ilb_memberConnectionLimit", "value": "0"}, {"name": "ilb_memberPriorityGroup", "value": "0"}, {"name": "ilb_memberratio", "value": "1"}, {"name": "ilb_memberadminstate", "value": "enabled"}, {"name": "ilb_persist", "value": null}, {"name": "networkservice_cidr", "value": "100.66.0.0/16"}, {"name": "networkpod_cidr", "value": "100.68.0.0/16"}, {"name": "cluster_cidr", "value": "100.66.0.10"}]`
	rows3 := sqlmock.NewRows([]string{"defaultvalues"}).
		AddRow(fakeValues)
	mock.ExpectQuery(`SELECT to_jsonb(array_agg(t)) as defaultvalues FROM public.defaultconfig t`).WillReturnRows(rows3)
	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, nil, "test-key", "test")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}
func TestMockCreateGetDefaultsParse9(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 3)
	fakeValues := `[{"name": "region", "value": "us-region-1"}, {"name": "vnet", "value": "us-region-1a"}, {"name": "availabilityzone", "value": "us-region-1a-default"}, {"name": "networkinterfacename", "value": "eth0"}, {"name": "ilb_environment", "value": "8"}, {"name": "ilb_ipprotocol", "value": "tcp"}, {"name": "ilb_usergroup", "value": "1545"}, {"name": "ilb_apiservername", "value": "apiserver"}, {"name": "ilb_apiserverport", "value": "6443"}, {"name": "ilb_etcdservername", "value": "etcd"}, {"name": "ilb_etcdport", "value": "2379"}, {"name": "ilb_konnectivityname", "value": "konnectivity"}, {"name": "ilb_loadbalancingmode", "value": "least-connections-member"}, {"name": "ilb_minactivemembers", "value": "1"}, {"name": "ilb_monitor", "value": "i_tcp"}, {"name": "ilb_memberConnectionLimit", "value": "0"}, {"name": "ilb_memberPriorityGroup", "value": "0"}, {"name": "ilb_memberratio", "value": "1"}, {"name": "ilb_memberadminstate", "value": "enabled"}, {"name": "ilb_persist", "value": null}, {"name": "networkservice_cidr", "value": "100.66.0.0/16"}, {"name": "networkpod_cidr", "value": "100.68.0.0/16"}, {"name": "cluster_cidr", "value": "100.66.0.10"}]`
	rows3 := sqlmock.NewRows([]string{"defaultvalues"}).
		AddRow(fakeValues)
	mock.ExpectQuery(`SELECT to_jsonb(array_agg(t)) as defaultvalues FROM public.defaultconfig t`).WillReturnRows(rows3)
	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, nil, "test-key", "test")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}

func TestMockCreateGetDefaultCpOsImageInstance(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 4)
	mock.ExpectQuery(query.GetDefaultCpOsImageInstance + " AND ks.minor_version = $5")
	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, nil, "test-key", "test")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}

func TestMockCreateGetDefaultCpInstanceAndNodeProvider(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 5)
	mock.ExpectQuery(query.GetDefaultCpInstanceTypeAndNodeProvider)
	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, nil, "test-key", "test")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}

func TestMockCreategetImiArtifactQuery(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 6)
	mock.ExpectQuery(query.GetImiArtifactQuery)
	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, nil, "test-key", "test")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}

/*
func TestMockCreateGetAddonsfork8sVersionQuery(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 7)
	mock.ExpectQuery(`SELECT addonversion_name FROM public.addoncompatibilityk8s where k8sversion_name = $1`)
	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, nil, "test-key", "test")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}
*/

func TestMockCreateGetDefaultAddonsQuery(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 8)
	mock.ExpectQuery(utils.GetAddonsQuery)
	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, nil, "test-key", "test")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}

func TestMockCreateInsertClusterRecordQuery(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 9)
	mock.ExpectQuery(query.InsertClusterRecordQuery)
	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, nil, "test-key", "test")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}

func TestMockCreateInsertControlPlane(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 10)
	mock.ExpectQuery(query.InsertControlPlaneQuery)
	mock.ExpectRollback()

	// Setup Encryption Key
	file, e := os.Create("./encryption_keys")
	if e != nil {
		t.Fatalf("can not create encryption keys")
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	_, err = fmt.Fprint(w, `{"data":{"2":"abcdefghijklmnopqrstuvwxyz123456", "1":"abcdefghijklmnopqrstuvwxyz123456"}, "metadata":{"created_time":"2023-10-25T19:27:45.560705679Z","custom_metadata":null,"deletion_time":"","destroyed":false,"version":1}}`)
	w.Flush()

	// Setup private and public key
	pubkey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCvrYjVgtH+OVLNscdMO67t0DJVx9NO0/fFp1D1L1kQCr20izUel5uBunByJbtq5/DK6GCWAB287zj1JKTltpCcGFjArjKGHQoF6ipDHhJUUDP/vCzRnAyC9Dd8/pAZMYUDip1L2foZPO/MrWBDQ4zlFN8Ef7HaZE39E/OecgDMJo8BTxzaLCqEGStAowYdPbF13gFfe7BQTfnhddZXwf0gAz7LrNFMmvDDNEslCbFUq5jwhN27jYW+y2hfDGsWjNyXBxmdlO7H3KjepPjEADZU7BOuhSiopk7iwDFg2lyh1Qy3MUUA2SN+4C2gwxkiEfIb1s1nScN8Pvu1mM4rMKXtxmY0JieW3eVkvSZ/q7TyJtfb0dj1BULsdTfjascbEaQtYCzrH/ol8HFiexd/doAt/SadQKDs4+GfLDeFekc8MhsROLWPvFVhHb1NwlKqj6uyKeFxjT3FT7hBlSwube4I6M5khr35aiPhNz8yf37tdbbMCAvUV9YHdAQNJ8Jpegs6f0td9N7Q8ogAIYYrXRvpbcFex9jOnHJv0RZg+BIdZ4cKNwSKWr2oKFJvJ6/fzsuFcYTrBDE9bKq+ou58066SI4YQmHvGFgGqr9x6yvEvoKktV6ikQiqVful2TrPWT/byRAiXO8bvAu3SP2+k/JoY+sPVsBjIw/9Hjfm6MlZG6w==\n"
	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, []byte(pubkey), "test-key", "./encryption_keys")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}

func TestMockCreateGetVipProviderDefaults(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 11)
	mock.ExpectQuery(query.GetVipProviderDefaults)

	// Setup Encryption Key
	file, e := os.Create("./encryption_keys")
	if e != nil {
		t.Fatalf("can not create encryption keys")
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	_, err = fmt.Fprint(w, `{"data":{"2":"abcdefghijklmnopqrstuvwxyz123456", "1":"abcdefghijklmnopqrstuvwxyz123456"}, "metadata":{"created_time":"2023-10-25T19:27:45.560705679Z","custom_metadata":null,"deletion_time":"","destroyed":false,"version":1}}`)
	w.Flush()

	// Setup private and public key
	pubkey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCvrYjVgtH+OVLNscdMO67t0DJVx9NO0/fFp1D1L1kQCr20izUel5uBunByJbtq5/DK6GCWAB287zj1JKTltpCcGFjArjKGHQoF6ipDHhJUUDP/vCzRnAyC9Dd8/pAZMYUDip1L2foZPO/MrWBDQ4zlFN8Ef7HaZE39E/OecgDMJo8BTxzaLCqEGStAowYdPbF13gFfe7BQTfnhddZXwf0gAz7LrNFMmvDDNEslCbFUq5jwhN27jYW+y2hfDGsWjNyXBxmdlO7H3KjepPjEADZU7BOuhSiopk7iwDFg2lyh1Qy3MUUA2SN+4C2gwxkiEfIb1s1nScN8Pvu1mM4rMKXtxmY0JieW3eVkvSZ/q7TyJtfb0dj1BULsdTfjascbEaQtYCzrH/ol8HFiexd/doAt/SadQKDs4+GfLDeFekc8MhsROLWPvFVhHb1NwlKqj6uyKeFxjT3FT7hBlSwube4I6M5khr35aiPhNz8yf37tdbbMCAvUV9YHdAQNJ8Jpegs6f0td9N7Q8ogAIYYrXRvpbcFex9jOnHJv0RZg+BIdZ4cKNwSKWr2oKFJvJ6/fzsuFcYTrBDE9bKq+ou58066SI4YQmHvGFgGqr9x6yvEvoKktV6ikQiqVful2TrPWT/byRAiXO8bvAu3SP2+k/JoY+sPVsBjIw/9Hjfm6MlZG6w==\n"

	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, []byte(pubkey), "test-key", "./encryption_keys")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}

func TestMockCreateInsertVipQuery(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 12)
	mock.ExpectQuery(query.InsertVipQuery)

	// Setup Encryption Key
	file, e := os.Create("./encryption_keys")
	if e != nil {
		t.Fatalf("can not create encryption keys")
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	_, err = fmt.Fprint(w, `{"data":{"2":"abcdefghijklmnopqrstuvwxyz123456", "1":"abcdefghijklmnopqrstuvwxyz123456"}, "metadata":{"created_time":"2023-10-25T19:27:45.560705679Z","custom_metadata":null,"deletion_time":"","destroyed":false,"version":1}}`)
	w.Flush()

	// Setup private and public key
	pubkey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCvrYjVgtH+OVLNscdMO67t0DJVx9NO0/fFp1D1L1kQCr20izUel5uBunByJbtq5/DK6GCWAB287zj1JKTltpCcGFjArjKGHQoF6ipDHhJUUDP/vCzRnAyC9Dd8/pAZMYUDip1L2foZPO/MrWBDQ4zlFN8Ef7HaZE39E/OecgDMJo8BTxzaLCqEGStAowYdPbF13gFfe7BQTfnhddZXwf0gAz7LrNFMmvDDNEslCbFUq5jwhN27jYW+y2hfDGsWjNyXBxmdlO7H3KjepPjEADZU7BOuhSiopk7iwDFg2lyh1Qy3MUUA2SN+4C2gwxkiEfIb1s1nScN8Pvu1mM4rMKXtxmY0JieW3eVkvSZ/q7TyJtfb0dj1BULsdTfjascbEaQtYCzrH/ol8HFiexd/doAt/SadQKDs4+GfLDeFekc8MhsROLWPvFVhHb1NwlKqj6uyKeFxjT3FT7hBlSwube4I6M5khr35aiPhNz8yf37tdbbMCAvUV9YHdAQNJ8Jpegs6f0td9N7Q8ogAIYYrXRvpbcFex9jOnHJv0RZg+BIdZ4cKNwSKWr2oKFJvJ6/fzsuFcYTrBDE9bKq+ou58066SI4YQmHvGFgGqr9x6yvEvoKktV6ikQiqVful2TrPWT/byRAiXO8bvAu3SP2+k/JoY+sPVsBjIw/9Hjfm6MlZG6w==\n"

	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, []byte(pubkey), "test-key", "./encryption_keys")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}

func TestMockCreateInsertVipDetailsQuery(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 13)
	mock.ExpectQuery(query.InsertVipDetailsQuery)
	// Setup Encryption Key
	file, e := os.Create("./encryption_keys")
	if e != nil {
		t.Fatalf("can not create encryption keys")
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	_, err = fmt.Fprint(w, `{"data":{"2":"abcdefghijklmnopqrstuvwxyz123456", "1":"abcdefghijklmnopqrstuvwxyz123456"}, "metadata":{"created_time":"2023-10-25T19:27:45.560705679Z","custom_metadata":null,"deletion_time":"","destroyed":false,"version":1}}`)
	w.Flush()

	// Setup private and public key
	pubkey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCvrYjVgtH+OVLNscdMO67t0DJVx9NO0/fFp1D1L1kQCr20izUel5uBunByJbtq5/DK6GCWAB287zj1JKTltpCcGFjArjKGHQoF6ipDHhJUUDP/vCzRnAyC9Dd8/pAZMYUDip1L2foZPO/MrWBDQ4zlFN8Ef7HaZE39E/OecgDMJo8BTxzaLCqEGStAowYdPbF13gFfe7BQTfnhddZXwf0gAz7LrNFMmvDDNEslCbFUq5jwhN27jYW+y2hfDGsWjNyXBxmdlO7H3KjepPjEADZU7BOuhSiopk7iwDFg2lyh1Qy3MUUA2SN+4C2gwxkiEfIb1s1nScN8Pvu1mM4rMKXtxmY0JieW3eVkvSZ/q7TyJtfb0dj1BULsdTfjascbEaQtYCzrH/ol8HFiexd/doAt/SadQKDs4+GfLDeFekc8MhsROLWPvFVhHb1NwlKqj6uyKeFxjT3FT7hBlSwube4I6M5khr35aiPhNz8yf37tdbbMCAvUV9YHdAQNJ8Jpegs6f0td9N7Q8ogAIYYrXRvpbcFex9jOnHJv0RZg+BIdZ4cKNwSKWr2oKFJvJ6/fzsuFcYTrBDE9bKq+ou58066SI4YQmHvGFgGqr9x6yvEvoKktV6ikQiqVful2TrPWT/byRAiXO8bvAu3SP2+k/JoY+sPVsBjIw/9Hjfm6MlZG6w==\n"

	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, []byte(pubkey), "test-key", "./encryption_keys")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}

func TestMockCreateInsertRevQuery(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 14)
	mock.ExpectQuery(query.InsertRevQuery)

	// Setup Encryption Key
	file, e := os.Create("./encryption_keys")
	if e != nil {
		t.Fatalf("can not create encryption keys")
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	_, err = fmt.Fprint(w, `{"data":{"2":"abcdefghijklmnopqrstuvwxyz123456", "1":"abcdefghijklmnopqrstuvwxyz123456"}, "metadata":{"created_time":"2023-10-25T19:27:45.560705679Z","custom_metadata":null,"deletion_time":"","destroyed":false,"version":1}}`)
	w.Flush()

	// Setup private and public key
	pubkey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCvrYjVgtH+OVLNscdMO67t0DJVx9NO0/fFp1D1L1kQCr20izUel5uBunByJbtq5/DK6GCWAB287zj1JKTltpCcGFjArjKGHQoF6ipDHhJUUDP/vCzRnAyC9Dd8/pAZMYUDip1L2foZPO/MrWBDQ4zlFN8Ef7HaZE39E/OecgDMJo8BTxzaLCqEGStAowYdPbF13gFfe7BQTfnhddZXwf0gAz7LrNFMmvDDNEslCbFUq5jwhN27jYW+y2hfDGsWjNyXBxmdlO7H3KjepPjEADZU7BOuhSiopk7iwDFg2lyh1Qy3MUUA2SN+4C2gwxkiEfIb1s1nScN8Pvu1mM4rMKXtxmY0JieW3eVkvSZ/q7TyJtfb0dj1BULsdTfjascbEaQtYCzrH/ol8HFiexd/doAt/SadQKDs4+GfLDeFekc8MhsROLWPvFVhHb1NwlKqj6uyKeFxjT3FT7hBlSwube4I6M5khr35aiPhNz8yf37tdbbMCAvUV9YHdAQNJ8Jpegs6f0td9N7Q8ogAIYYrXRvpbcFex9jOnHJv0RZg+BIdZ4cKNwSKWr2oKFJvJ6/fzsuFcYTrBDE9bKq+ou58066SI4YQmHvGFgGqr9x6yvEvoKktV6ikQiqVful2TrPWT/byRAiXO8bvAu3SP2+k/JoY+sPVsBjIw/9Hjfm6MlZG6w==\n"

	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, []byte(pubkey), "test-key", "./encryption_keys")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}
func TestMockCreateInsertProvisioningQuery(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, 15)
	mock.ExpectQuery(query.InsertProvisioningQuery)

	// Setup Encryption Key
	file, e := os.Create("./encryption_keys")
	if e != nil {
		t.Fatalf("can not create encryption keys")
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	_, err = fmt.Fprint(w, `{"data":{"2":"abcdefghijklmnopqrstuvwxyz123456", "1":"abcdefghijklmnopqrstuvwxyz123456"}, "metadata":{"created_time":"2023-10-25T19:27:45.560705679Z","custom_metadata":null,"deletion_time":"","destroyed":false,"version":1}}`)
	w.Flush()

	// Setup private and public key
	pubkey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCvrYjVgtH+OVLNscdMO67t0DJVx9NO0/fFp1D1L1kQCr20izUel5uBunByJbtq5/DK6GCWAB287zj1JKTltpCcGFjArjKGHQoF6ipDHhJUUDP/vCzRnAyC9Dd8/pAZMYUDip1L2foZPO/MrWBDQ4zlFN8Ef7HaZE39E/OecgDMJo8BTxzaLCqEGStAowYdPbF13gFfe7BQTfnhddZXwf0gAz7LrNFMmvDDNEslCbFUq5jwhN27jYW+y2hfDGsWjNyXBxmdlO7H3KjepPjEADZU7BOuhSiopk7iwDFg2lyh1Qy3MUUA2SN+4C2gwxkiEfIb1s1nScN8Pvu1mM4rMKXtxmY0JieW3eVkvSZ/q7TyJtfb0dj1BULsdTfjascbEaQtYCzrH/ol8HFiexd/doAt/SadQKDs4+GfLDeFekc8MhsROLWPvFVhHb1NwlKqj6uyKeFxjT3FT7hBlSwube4I6M5khr35aiPhNz8yf37tdbbMCAvUV9YHdAQNJ8Jpegs6f0td9N7Q8ogAIYYrXRvpbcFex9jOnHJv0RZg+BIdZ4cKNwSKWr2oKFJvJ6/fzsuFcYTrBDE9bKq+ou58066SI4YQmHvGFgGqr9x6yvEvoKktV6ikQiqVful2TrPWT/byRAiXO8bvAu3SP2+k/JoY+sPVsBjIw/9Hjfm6MlZG6w==\n"

	// RUN
	_, err = query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, []byte(pubkey), "test-key", "./encryption_keys")
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not create Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}
func TestMockCreateClusterCommit(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateClusterValues(mock, -1)
	// Setup Encryption Key
	file, e := os.Create("./encryption_keys")
	if e != nil {
		t.Fatalf("can not create encryption keys")
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	_, err = fmt.Fprint(w, `{"data":{"2":"abcdefghijklmnopqrstuvwxyz123456", "1":"abcdefghijklmnopqrstuvwxyz123456"}, "metadata":{"created_time":"2023-10-25T19:27:45.560705679Z","custom_metadata":null,"deletion_time":"","destroyed":false,"version":1}}`)
	w.Flush()

	// Setup private and public key
	pubkey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCvrYjVgtH+OVLNscdMO67t0DJVx9NO0/fFp1D1L1kQCr20izUel5uBunByJbtq5/DK6GCWAB287zj1JKTltpCcGFjArjKGHQoF6ipDHhJUUDP/vCzRnAyC9Dd8/pAZMYUDip1L2foZPO/MrWBDQ4zlFN8Ef7HaZE39E/OecgDMJo8BTxzaLCqEGStAowYdPbF13gFfe7BQTfnhddZXwf0gAz7LrNFMmvDDNEslCbFUq5jwhN27jYW+y2hfDGsWjNyXBxmdlO7H3KjepPjEADZU7BOuhSiopk7iwDFg2lyh1Qy3MUUA2SN+4C2gwxkiEfIb1s1nScN8Pvu1mM4rMKXtxmY0JieW3eVkvSZ/q7TyJtfb0dj1BULsdTfjascbEaQtYCzrH/ol8HFiexd/doAt/SadQKDs4+GfLDeFekc8MhsROLWPvFVhHb1NwlKqj6uyKeFxjT3FT7hBlSwube4I6M5khr35aiPhNz8yf37tdbbMCAvUV9YHdAQNJ8Jpegs6f0td9N7Q8ogAIYYrXRvpbcFex9jOnHJv0RZg+BIdZ4cKNwSKWr2oKFJvJ6/fzsuFcYTrBDE9bKq+ou58066SI4YQmHvGFgGqr9x6yvEvoKktV6ikQiqVful2TrPWT/byRAiXO8bvAu3SP2+k/JoY+sPVsBjIw/9Hjfm6MlZG6w==\n"

	// RUN
	res, err := query.CreateClusterRecord(context.Background(), db, &clusterCreate, nil, []byte(pubkey), "test-key", "./encryption_keys")
	// RUN
	if err != nil {
		t.Fatalf("Should not get Error, recieved: %v", err)
	}
	// CHECK
	if res.Name != clusterCreate.Name {
		t.Fatalf("Not expeceted values")
	}
}
