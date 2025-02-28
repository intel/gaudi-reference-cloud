// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package admin_query

import (
	"context"
	"database/sql"
	"fmt"
	"google.golang.org/grpc/codes"
	grpc_status "google.golang.org/grpc/status"
	"strings"

	"github.com/blang/semver/v4"
	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/iks_utils"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

const (
	// Validation Queries
	GetOsImageInstanceNameQuery = `
		SELECT osimageinstance_name
		FROM public.osimageinstance
		WHERE osimageinstance_name = $1;
	`

	GetOsImageNameQuery = `
		SELECT osimage_name, lifecyclestate_id
		FROM public.osimage
		WHERE osimage_name = $1;
	`

	GetRuntimeVersionNameQuery = `
		SELECT runtimeversion_name, runtime_name
		FROM public.runtimeversion
		WHERE runtime_name = $1;
	`

	GetCloudAccountApproveListCountQuery = `
		SELECT count(*)
		FROM public.cloudaccountextraspec
		WHERE cloudaccount_id = $1;
	`

	GetProviderNameQuery = `
		SELECT provider_name, lifecyclestate_id
		FROM public.provider
		WHERE provider_name = $1;
	`

	GetInstanceTypeNameQuery = `
	SELECT instancetype_name
	FROM public.instancetype
	WHERE instancetype_name = $1;
	`

	GetNodeProviderNameQuery = `
	SELECT nodeprovider_name, lifecyclestate_id
	FROM public.nodeprovider
	WHERE nodeprovider_name = $1;
	`

	ValidateLifeCycleQuery = `
	SELECT lifecyclestate_id FROM public.lifecycle`

	// insert queries
	InsertIMIQuery = `
		INSERT INTO public.osimageinstance(
		osimageinstance_name, osimage_name, runtimeversion_name, created_date, nodegrouptype_name, lifecyclestate_id, 
		runtime_name, provider_name, k8sversion_name, imiartifact, instancetypecategory, instancetypefamiliy)
		VALUES ($1, $2, $3, Now(),$4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING osimageinstance_name;
 	`

	InsertIMIComponentQuery = `
		INSERT INTO public.osimageinstancecomponent(
		component_name, osimageinstance_name, version, artifact_repo)
		VALUES ($1, $2, $3, $4)
		RETURNING component_name;
  	`

	InsertInstanceTypeQuery = `
		INSERT INTO public.instancetype(
		instancetype_name, memory, cpu, nodeprovider_name, storage, lifecyclestate_id, 
		displayname, imi_override, description, instancecategory, instancetypefamiliy)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING instancetype_name;
   	`

	InsertK8sVersionQuery = `INSERT INTO k8sversion VALUES ($1, $2, $3, $4, $5)`

	InsertK8sCompatibilityQuery = `INSERT INTO k8scompatibility VALUES ($1, $2, $3, $4, $5, $6)`

	InsertCloudAccountApproveListQuery = `INSERT INTO public.cloudaccountextraspec(
		cloudaccount_id,provider_name, active_account_create_cluster, allow_create_storage, maxclusters_override, maxclusterng_override, maxclusterilb_override, maxclustervm_override, maxnodegroupvm_override) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING cloudaccount_id,provider_name,active_account_create_cluster, allow_create_storage;
		`

	GetDefaultOsImageNameQuery = `SELECT osimage_name FROM osimage WHERE lifecyclestate_id = 1 LIMIT 1;`
)

type Result struct {
	Data UpstreamData `json:"result"`
}
type UpstreamData struct {
	ReleaseId    string      `json:"releaseId"`
	Vendor       string      `json:"vendor"`
	License      string      `json:"license"`
	Purl         string      `json:"purl"`
	Releasets    string      `json:"releaseTimestamp"`
	EosTimestamp string      `json:"eosTimestamp"`
	EolTimestamp string      `json:"eolTimestamp"`
	Components   []Component `json:"components"`
}

type Component struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Url     string `json:"purl"`
	Sha     string `json:"sha256"`
}

var (
	name                 string
	osImageInstanceState string
	osImageState         string
	runtimeVersionName   string
	runtimeName          string
	providerName         string
	providerState        string
	insertedName         string
	componentName        string
	cloudAccountID       string
)

// CreateInstanceType to insert record into instancetype table
func CreateInstanceTypes(ctx context.Context, dbconn *sql.DB, record *pb.CreateInstanceTypeRequest) (*pb.InstanceTypeResponse, error) {
	friendlyMessage := "Could not Create InstanceType. Please try again."
	failedFunction := "CreateInstanceType."
	returnError := &pb.InstanceTypeResponse{}

	// Validate Instance Name Uniqueness
	name = ""
	err := dbconn.QueryRowContext(ctx, GetInstanceTypeNameQuery, record.Instancetypename).Scan(&name)
	if err != sql.ErrNoRows && err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetInstanceTypeNameQuery", friendlyMessage+err.Error())
	}
	if name != "" {
		return returnError, grpc_status.Error(codes.AlreadyExists, "Instance Type name already in use")
	}
	// Start the transaction
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage+err.Error())
	}

	// Insert IMI
	err = tx.QueryRowContext(ctx, InsertInstanceTypeQuery,
		record.Instancetypename,
		record.Memory,
		record.Cpu,
		record.Nodeprovidername,
		record.Storage,
		record.Status,
		record.Displayname,
		record.Imioverride,
		record.Description,
		record.Category,
		record.Family,
	).Scan(&insertedName)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TrasactionRollbackFailed", friendlyMessage+errtx.Error())
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"InsertInstanceTypeQuery", friendlyMessage+err.Error())
	}

	/* COMMIT TRANSACTION */
	// close the transaction with a Commit() or Rollback() method on the resulting Tx variable.

	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"tx.Commit", friendlyMessage+err.Error())
	}

	instanceTypeCreateResponse := &pb.InstanceTypeResponse{
		Instancetypename: record.Instancetypename,
		Memory:           record.Memory,
		Cpu:              record.Cpu,
		Nodeprovidername: record.Nodeprovidername,
		Storage:          record.Storage,
		Status:           record.Status,
		Displayname:      record.Displayname,
		Description:      record.Description,
		Category:         record.Category,
		Family:           record.Family,
		IksDB:            true,
	}

	return instanceTypeCreateResponse, nil
}

// CreateIMI to insert record into osimageinstance table
func CreateIMI(ctx context.Context, dbconn *sql.DB, record *pb.IMIRequest) (*pb.IMIResponse, error) {
	friendlyMessage := "Could not Create IMI. Please try again."
	failedFunction := "CreateImi."
	returnError := &pb.IMIResponse{}

	// //Validate Upstream Release
	// var res Result
	// reqcomponents := record.GetComponents()
	// err := json.Unmarshal([]byte(data), &res)
	// if err != nil {
	// 	return returnError, utils.ErrorHandler(ctx, err, failedFunction+"unmarshal.data", friendlyMessage+err.Error())
	// }
	// if record.Upstreamreleasename != res.Data.ReleaseId {
	// 	return returnError, errors.New("Release ID does not match with upstream")
	// }

	// // Validate upstream component
	// lrescomponents := len(res.Data.Components)
	// datamap := make(map[int]Component, lrescomponents)
	// for i, c := range res.Data.Components {
	// 	datamap[i] = c
	// }
	// for i, c := range reqcomponents {
	// 	val, _ := datamap[i]
	// 	if val.Name != c.Name || val.Version != c.Version || val.Url != c.Artifact {
	// 		return returnError, errors.New("Components name does not match with upstream Name")
	// 	}
	// }

	// Validate Os Image Existence
	name = ""
	err := dbconn.QueryRowContext(ctx, GetOsImageNameQuery, record.Os).Scan(&name, &osImageState)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, grpc_status.Error(codes.NotFound, "No OS Image exists under the given name")
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetOsImageNameQuery", friendlyMessage+err.Error())
	}

	// Validate Runtime Existence
	err = dbconn.QueryRowContext(ctx, GetRuntimeVersionNameQuery, record.Runtime).Scan(&runtimeVersionName, &runtimeName)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, grpc_status.Error(codes.NotFound, "No Runtime exists under the given name")
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetRuntimeVersionNameQuery", friendlyMessage+err.Error())
	}

	// Validate Provider Existence
	err = dbconn.QueryRowContext(ctx, GetProviderNameQuery, record.Provider).Scan(&providerName, &providerState)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, grpc_status.Error(codes.NotFound, "No Provider exists under the given name")
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetProviderNameQuery", friendlyMessage+err.Error())
	}

	if providerState != "1" {
		return returnError, grpc_status.Error(codes.NotFound, "No Provider exists under the given name")
	}

	// Validate OS Image Instance Name Uniqueness
	name = ""
	err = dbconn.QueryRowContext(ctx, GetOsImageInstanceNameQuery, record.Name).Scan(&name)
	if err != sql.ErrNoRows && err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetOsImageInstanceNameQuery", friendlyMessage+err.Error())
	}
	if name != "" {
		return returnError, grpc_status.Error(codes.AlreadyExists, "IMI name already in use")
	}
	// Start the transaction
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage+err.Error())
	}

	// Insert IMI
	err = tx.QueryRowContext(ctx, InsertIMIQuery,
		record.Name,
		record.Os,
		runtimeVersionName,
		record.Type,
		record.State,
		record.Runtime,
		record.Provider,
		record.Upstreamreleasename,
		record.Artifact,
		record.Category,
		record.Family,
	).Scan(&insertedName)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TrasactionRollbackFailed", friendlyMessage+errtx.Error())
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"InsertIMIQuery", friendlyMessage+err.Error())
	}

	// Insert IMI Component
	for i := 0; i < len(record.Components); i++ {
		err := tx.QueryRowContext(ctx, InsertIMIComponentQuery,
			record.Components[i].Name,
			record.Name,
			record.Components[i].Version,
			record.Components[i].Artifact,
		).Scan(&componentName)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TrasactionRollbackFailed", friendlyMessage+errtx.Error())
			}
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"InsertIMIComponentQuery", friendlyMessage+err.Error())
		}
	}

	/* COMMIT TRANSACTION */
	// close the transaction with a Commit() or Rollback() method on the resulting Tx variable.

	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"tx.Commit", friendlyMessage+err.Error())
	}

	imiCreateResponse := &pb.IMIResponse{
		Name:                record.Name,
		Upstreamreleasename: record.Upstreamreleasename,
		Provider:            record.Provider,
		Type:                record.Type,
		Runtime:             record.Runtime,
		Os:                  record.Os,
		State:               record.State,
		Components:          record.Components,
		Bootstraprepo:       record.Bootstraprepo,
		Artifact:            record.Artifact,
	}

	return imiCreateResponse, nil
}

func CreateK8SVersion(ctx context.Context, dbconn *sql.DB, record *pb.Createk8SversionRequest) (*pb.K8SversionResponse, error) {
	friendlyMessage := "Could not Put k8svversion. Please try again."
	failedFunction := "PutK8SVersion."
	returnError := &pb.K8SversionResponse{}

	version, err := semver.ParseTolerant(record.Name)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ParseTolerant", friendlyMessage+err.Error())
	}

	// Get Default Os Image
	var osImageName string
	err = dbconn.QueryRowContext(ctx, GetDefaultOsImageNameQuery).Scan(&osImageName)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, grpc_status.Error(codes.NotFound, "No Default OS Image exists")
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetDefaultOsImageNameQuery", friendlyMessage+err.Error())
	}

	// Validate Runtime Existence
	err = dbconn.QueryRowContext(ctx, GetRuntimeVersionNameQuery, record.Runtime).Scan(&runtimeVersionName, &runtimeName)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, grpc_status.Error(codes.NotFound, "No Runtime exists under the given name")
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetRuntimeVersioNameQuery", friendlyMessage+err.Error())
	}

	// Validate Provider Existence
	err = dbconn.QueryRowContext(ctx, GetProviderNameQuery, record.Provider).Scan(&providerName, &providerState)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, grpc_status.Error(codes.NotFound, "No Provider exists under the given name")
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetProviderNameQuery", friendlyMessage+err.Error())
	}
	if providerState != "1" {
		return returnError, grpc_status.Error(codes.NotFound, "No Provider exists under the given name")
	}

	// Validate CP OS Image Instance Name Uniqueness
	name = ""
	err = dbconn.QueryRowContext(ctx, GetOsImageInstanceNameQuery, record.Cpimi).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, grpc_status.Error(codes.NotFound, "No CP IMI exists under the given name")
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetOsImageInstanceNameQuery.CPimi", friendlyMessage+err.Error())
	}

	// Validate Worker OS Image Instance Name Uniqueness
	name = ""
	err = dbconn.QueryRowContext(ctx, GetOsImageInstanceNameQuery, record.Workimi).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, grpc_status.Error(codes.NotFound, "No Worker IMI exists under the given name")
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetOsImageInstanceNameQuery.Workimi", friendlyMessage+err.Error())
	}

	// Start the transaction
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage+err.Error())
	}

	// Insert K8s Version
	var formatedVersion string
	if record.Provider == iksProvider {
		formatedVersion = version.FinalizeVersion()
	} else {
		formatedVersion = record.Name
		if !strings.HasPrefix(formatedVersion, "v") {
			formatedVersion = "v" + formatedVersion
		}
	}
	_, err = tx.ExecContext(ctx, InsertK8sVersionQuery,
		formatedVersion,
		lifecyclestateStaged,
		fmt.Sprintf("%d.%d", version.Major, version.Minor),
		fmt.Sprintf("%d", version.Major),
		false,
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TrasactionRollbackFailed", friendlyMessage+errtx.Error())
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"InsertK8sVersionQuery", friendlyMessage+err.Error())
	}

	// Insert K8s Compatibility Version
	_, err = tx.ExecContext(ctx, InsertK8sCompatibilityQuery,
		record.Runtime,
		formatedVersion,
		osImageName,
		record.Cpimi,
		record.Workimi,
		record.Provider,
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TrasactionRollbackFailed", friendlyMessage+errtx.Error())
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"insertK8sCompatabilityQuery", friendlyMessage+err.Error())
	}

	// COMMIT
	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"tx.Commit", friendlyMessage+err.Error())
	}

	k8sVersionCreateResponse := &pb.K8SversionResponse{
		Name:        formatedVersion,
		Runtime:     record.Runtime,
		Releasename: formatedVersion,
		Provider:    record.Provider,
		Cpimi:       record.Cpimi,
		Workimi:     record.Workimi,
		State:       "Staged",
		Minor:       fmt.Sprintf("%d.%d", version.Major, version.Minor),
		Major:       fmt.Sprintf("%d", version.Major),
	}

	return k8sVersionCreateResponse, nil
}

func CreateCloudAccountApproveList(ctx context.Context, dbconn *sql.DB, record *pb.CloudAccountApproveListRequest) (*pb.CloudAccountApproveList, error) {
	// Start the transaction
	friendlyMessage := "CreateCloudAccountApproveList.UnexpectedError"
	failedFunction := "CreateCloudAccountApproveList."
	returnError := &pb.CloudAccountApproveList{}
	returnValue := &pb.CloudAccountApproveList{}

	cloudAccountID = record.GetCloudaccountId()
	if cloudAccountID == "" {
		return returnError, grpc_status.Error(codes.InvalidArgument, "CreateCloudAccountApproveList.CloudAccountID: Invalid Input Request")
	}

	// Validate Cloud Account Existence
	var count int32
	err := dbconn.QueryRowContext(ctx, GetCloudAccountApproveListCountQuery, record.CloudaccountId).Scan(&count)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetCloudAccountApproveListQuery", friendlyMessage+err.Error())
	}
	if count >= 1 {
		return returnError, grpc_status.Error(codes.AlreadyExists, "Cloud Account already in use")
	}

	// Start the transaction
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage+err.Error())
	}

	/* Insert cloud account approve list details */
	err = tx.QueryRowContext(ctx, InsertCloudAccountApproveListQuery, record.CloudaccountId, iksProvider, record.Status, record.EnableStorage, record.MaxclustersOverride, record.MaxclusterngOverride, record.MaxclusterilbOverride, record.MaxclustervmOverride, record.MaxnodegroupvmOverride).Scan(
		&returnValue.Account,
		&returnValue.Providername,
		&returnValue.Status,
		&returnValue.EnableStorage,
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"InsertCloudAccountApproveListQuery", friendlyMessage+err.Error())
	}

	/* COMMIT TRANSACTION */
	// close the transaction with a Commit() or Rollback() method on the resulting Tx variable.

	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"tx.Commit", friendlyMessage+err.Error())
	}

	return returnValue, nil
}
