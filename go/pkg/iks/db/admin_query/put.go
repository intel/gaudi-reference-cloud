// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package admin_query

import (
	"context"
	"database/sql"
	"google.golang.org/grpc/codes"
	grpc_status "google.golang.org/grpc/status"
	"strconv"
	"time"

	"github.com/blang/semver/v4"
	"github.com/golang/protobuf/ptypes/empty"
	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/iks_utils"
	clusterv1alpha "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

const (
	putimiquery = `
	Update public.osimageinstance SET osimage_name = $1 ,runtime_name = $2, provider_name = $3, imiartifact = $4, k8sversion_name = $5 , lifecyclestate_id = $6 , created_date = $7 , runtimeversion_name = $8, nodegrouptype_name = $9, instancetypecategory = $10, instancetypefamiliy = $11
	where osimageinstance_name = $12`

	putcomponentquery = `
	Update public.osimageinstancecomponent SET component_name = $1 , version = $2 , artifact_repo = $3
	where osimageinstance_name = $4
	`
	putInstanceTypeQuery = `
	Update public.instancetype SET memory = $1, cpu = $2, nodeprovider_name = $3, storage = $4 , lifecyclestate_id = $5 , displayname = $6 , imi_override = $7, description = $8, instancecategory = $9, instancetypefamiliy = $10
	where instancetype_name = $11`

	checkK8sVersionQuery = `
		SELECT
			k8sversion.lifecyclestate_id,
			k8scompatibility.provider_name
		FROM k8sversion
			JOIN k8scompatibility ON k8scompatibility.k8sversion_name = k8sversion.k8sversion_name
		WHERE k8sversion.k8sversion_name = $1`

	getLifecycleStateQuery = `SELECT lifecyclestate_id FROM lifecyclestate WHERE name = $1`

	updateK8sVersionStateQuery   = `UPDATE k8sversion SET lifecyclestate_id = $1 WHERE k8sversion_name = $2`
	updateK8sVersionCpImiQuery   = `UPDATE k8scompatibility SET cp_osimageinstance_name = $1 WHERE k8sversion_name = $2`
	updateK8sVersionWorkImiQuery = `UPDATE k8scompatibility SET wrk_osimageinstance_name = $1 WHERE k8sversion_name = $2`
	updateNodegroupImiQuery      = `
		UPDATE public.nodegroup
		SET k8sversion_name = $2, osimageinstance_name = $3
		WHERE unique_id = $1
	`
	getImiFromK8sVersion = `
		SELECT
			CASE WHEN $1 = 'ControlPlane'
				THEN kc.cp_osimageinstance_name
				ELSE kc.wrk_osimageinstance_name
			END
		FROM k8scompatibility AS kc
		WHERE kc.provider_name = $2 AND kc.runtime_name = $3 AND kc.instancetype_name=$4 AND kc.osimage_name=$5 AND kc.k8sversion_name = $6
	`
	insertRevQuery = `
 		INSERT INTO public.clusterrev (cluster_id, currentspec_json, desiredspec_json, component_typegrp,
			component_typename, currentdata, desireddata, timestamp, change_applied)
 		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING clusterrev_id
 	`
	updateClusterStateQuery = `
		UPDATE public.cluster
		SET clusterstate_name = $2
		WHERE cluster_id = $1
	`
	updateNodegroupStateQuery = `
		UPDATE public.nodegroup
		SET nodegroupstate_name = $2
		WHERE unique_id = $1
	`

	updateCloudAccountApproveListQuery = `
	UPDATE public.cloudaccountextraspec
	SET active_account_create_cluster = $2, allow_create_storage = $3,
	maxclusters_override = CASE WHEN $4 = 0 THEN NULL ELSE $4 END,
	maxclusterng_override = CASE WHEN $5 = 0 THEN NULL ELSE $5 END,
	maxclusterilb_override = CASE WHEN $6 = 0 THEN NULL ELSE $6 END,
	maxclustervm_override = CASE WHEN $7 = 0 THEN NULL ELSE $7 END,
	maxnodegroupvm_override = CASE WHEN $8 = 0 THEN NULL ELSE $8 END
	WHERE cloudaccount_id = $1
	`

	updateWorkerIMIInstanceTypeToK8sCompatibilityQuery = `
	UPDATE public.k8scompatibility
	SET wrk_osimageinstance_name = $6
	WHERE runtime_name = $1 AND k8sversion_name = $2 AND osimage_name = $3 AND provider_name = $4 AND instancetype_name = $5
	`

	updateCPIMIInstanceTypeToK8sCompatibilityQuery = `
	UPDATE public.k8scompatibility
	SET cp_osimageinstance_name = $5
	WHERE runtime_name = $1 AND k8sversion_name = $2 AND osimage_name = $3 AND provider_name = $4`

	getInstanceTypesExistenceQuery = `
	SELECT i.instancetype_name
	FROM public.instancetype i
	INNER JOIN lifecyclestate li
	ON li.lifecyclestate_id = i.lifecyclestate_id
	WHERE i.instancetype_name = $1 AND i.lifecyclestate_id = (SELECT lifecyclestate_id from public.lifecyclestate WHERE name = 'Active');
	`

	insertK8sCompatibilityQuery = `INSERT INTO k8scompatibility(runtime_name,k8sversion_name,osimage_name,cp_osimageinstance_name,wrk_osimageinstance_name,provider_name,instancetype_name,lifecyclestate_id) VALUES($1, $2, $3, $4, $5, $6, $7, $8)`

	getK8sCompatibilityQueryByInstanceType = `
	SELECT instancetype_name FROM k8scompatibility 
	WHERE runtime_name = $1 AND k8sversion_name = $2 AND osimage_name = $3 AND wrk_osimageinstance_name = $4 AND provider_name = $5 AND instancetype_name = $6
	`
)

func PutInstanceType(ctx context.Context, dbconn *sql.DB, record *pb.UpdateInstanceTypeRequest) (*pb.InstanceTypeResponse, error) {
	friendlyMessage := "PutInstanceType.UnexpectedError"
	failedFunction := "PutInstanceType."
	returnError := &pb.InstanceTypeResponse{}

	// Validate Node Provider Existence
	err := dbconn.QueryRowContext(ctx, GetNodeProviderNameQuery, record.Nodeprovidername).Scan(&providerName, &providerState)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, grpc_status.Error(codes.NotFound, "No Node Provider exists under the given name")
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetProviderNameQuery", friendlyMessage+err.Error())
	}
	if providerState != "1" {
		return returnError, grpc_status.Error(codes.NotFound, "No Node Provider exists under the given name")
	}

	// Validate Instance Name Existence
	name = ""
	err = dbconn.QueryRowContext(ctx, GetInstanceTypeNameQuery, record.Name).Scan(&name)
	if err != sql.ErrNoRows && err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetInstanceTypeNameQuery", friendlyMessage+err.Error())
	}
	if name == "" {
		return returnError, grpc_status.Error(codes.NotFound, "No Instance Type exists under the given name")
	}

	// BEGIN TRANSACTION
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage+err.Error())
	}

	// UPDATE IMI QUERY
	_, err = tx.QueryContext(ctx, putInstanceTypeQuery, record.GetMemory(), record.GetCpu(), record.GetNodeprovidername(), record.GetStorage(), record.GetStatus(), record.GetDisplayname(), record.GetImioverride(), record.GetDescription(), record.Category, record.Family, record.Name)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"putInstanceTypeQuery", friendlyMessage+err.Error())
	}

	// COMMIT
	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"putInstanceTypeQuery", friendlyMessage+err.Error())
	}

	return &pb.InstanceTypeResponse{
		Instancetypename: record.Name,
		Memory:           record.Memory,
		Cpu:              record.Memory,
		Nodeprovidername: record.Nodeprovidername,
		Storage:          record.Storage,
		Status:           record.Status,
		Displayname:      record.Displayname,
		Imioverride:      record.Imioverride,
		Description:      record.Description,
		Category:         record.Category,
		Family:           record.Family,
		IksDB:            true,
	}, nil
}

func PutIMI(ctx context.Context, dbconn *sql.DB, record *pb.UpdateIMIRequest) (*pb.IMIResponse, error) {
	friendlyMessage := "PutIMI.UnexpectedError"
	failedFunction := "PutIMI."
	returnError := &pb.IMIResponse{}

	// // GET COMPONENTS
	// var res Result
	// reqcomponents := record.GetComponents()
	// err := json.Unmarshal([]byte(data), &res)
	// if err != nil {
	// 	return returnError, utils.ErrorHandler(ctx, err, failedFunction+"unmarshal.data", friendlyMessage+err.Error())
	// }

	// // PARSE COMPONENTS
	// lrescomponents := len(res.Data.Components)
	// datamap := make(map[int]Component, lrescomponents)
	// for i, c := range res.Data.Components {
	// 	datamap[i] = c
	// }
	// for i, c := range reqcomponents {
	// 	val, _ := datamap[i]
	// 	if val.Name != c.Name || val.Version != c.Version || val.Url != c.Artifact {
	// 		return returnError, errors.New("Components name doesn not match with upstream Name")
	// 	}
	// }
	// if record.Upstreamreleasename != res.Data.ReleaseId {
	// 	return returnError, errors.New("Release ID doesn not match with upstream")
	// }

	// Validate Os Image Existence
	err := dbconn.QueryRowContext(ctx, GetOsImageNameQuery, record.Os).Scan(&name, &osImageState)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, grpc_status.Error(codes.NotFound, "No Os Image exists under the given name")
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

	// BEGIN TRANSACTION
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage+err.Error())
	}

	// UPDATE IMI QUERY
	_, err = tx.QueryContext(ctx, putimiquery, record.GetOs(), record.GetRuntime(), record.GetProvider(), record.GetArtifact(), record.GetUpstreamreleasename(), record.GetState(), time.Now(), runtimeVersionName, record.Type, record.Category, record.Family, record.Name)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"putimiquery", friendlyMessage+err.Error())
	}

	// DELETE AND INSERT COMPONENTS
	_, err = tx.QueryContext(ctx, deletecomponentquery, record.Name)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"deletecomponentquery", friendlyMessage+err.Error())
	}
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

	// COMMIT
	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"putcomponentquery", friendlyMessage+err.Error())
	}

	return &pb.IMIResponse{
		Name:                record.GetName(),
		Upstreamreleasename: record.GetUpstreamreleasename(), // update this to the k8sversion name
		Provider:            record.GetProvider(),
		Type:                record.GetType(),
		Runtime:             record.GetRuntime(),
		Os:                  record.GetOs(),
		State:               record.GetState(),
		Components:          record.GetComponents(),
		Bootstraprepo:       record.GetBootstraprepo(),
		Artifact:            record.GetArtifact(),
	}, nil
}

func PutK8SVersion(ctx context.Context, dbconn *sql.DB, record *pb.UpdateK8SRequest) (*pb.K8SversionResponse, error) {
	friendlyMessage := "Could not Put k8svversion. Please try again."
	failedFunction := "PutK8SVersion."
	returnError := &pb.K8SversionResponse{}

	// Get k8sversion if exists
	var lifecycleState int64
	err := dbconn.QueryRowContext(ctx, checkK8sVersionQuery, record.Name).Scan(&lifecycleState, &providerName)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, grpc_status.Error(codes.NotFound, "No K8s Version exists under the given name")
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"checkK8sVersionQuery", friendlyMessage+err.Error())
	}

	// Start the transaction
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage+err.Error())
	}

	// Validate CP OS Image Instance Name Uniqueness
	if record.GetCpimi() != "" {
		name = ""
		err = tx.QueryRowContext(ctx, GetOsImageInstanceNameQuery, record.Cpimi).Scan(&name)

		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			if err == sql.ErrNoRows {
				return returnError, grpc_status.Error(codes.NotFound, "No CP IMI exists under the given name")
			}
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetOsImageInstanceNameQuery", friendlyMessage+err.Error())
		}

		// Update CP Imi
		_, err = tx.ExecContext(ctx, updateK8sVersionCpImiQuery, record.Cpimi, record.Name)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"updateK8sVersionCpImiQuery", friendlyMessage+err.Error())
		}
	}

	if record.GetWorkimi() != "" {
		// Validate Worker OS Image Instance Name Uniqueness
		name = ""
		err = tx.QueryRowContext(ctx, GetOsImageInstanceNameQuery, record.Workimi).Scan(&name)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			if err == sql.ErrNoRows {
				return returnError, grpc_status.Error(codes.NotFound, "No Worker IMI exists under the given name")
			}
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetOsImageInstanceNameQuery", friendlyMessage+err.Error())
		}

		// Update CP Imi
		_, err = tx.ExecContext(ctx, updateK8sVersionWorkImiQuery, record.Workimi, record.Name)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"updateK8sVersionworkImiQuery", friendlyMessage+err.Error())
		}
	}

	// Update K8s Version
	if record.GetState() != "" {
		var newLifecyleState int64
		err = tx.QueryRowContext(ctx, getLifecycleStateQuery, record.State).Scan(&newLifecyleState)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			if err == sql.ErrNoRows {
				return returnError, grpc_status.Error(codes.NotFound, "No Lifecycle State exists under the given name")
			}
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getLifecycleStateQuery", friendlyMessage+err.Error())
		}

		// Set previous version state to Archived if this version new state is Active
		if lifecycleState == lifecyclestateStaged && newLifecyleState == lifecyclestateActive {
			// Get all other versions for this version provider
			rows, err := dbconn.QueryContext(ctx, getK8sActiveVersionsForProviderQuery, providerName)
			if err != nil {
				if errtx := tx.Rollback(); errtx != nil {
					return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
				}
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getK8sActiveVersionForProviderQuery", friendlyMessage+err.Error())
			}
			defer rows.Close()

			k8sVersionNameMap := make(map[string]string, 0)
			k8sVersions := make([]semver.Version, 0)
			for rows.Next() {
				var versionName, cpImi, wrkImi string
				err = rows.Scan(&versionName, &cpImi, &wrkImi)
				if err != nil {
					if errtx := tx.Rollback(); errtx != nil {
						return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
					}
					return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getK8sActiveVersionForProviderQuery.rows.scan", friendlyMessage+err.Error())
				}

				v, err := parseVersions(versionName)
				if err != nil {
					if errtx := tx.Rollback(); errtx != nil {
						return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
					}
					return returnError, utils.ErrorHandler(ctx, err, failedFunction+"parseVersions.versionName", friendlyMessage+err.Error())
				}
				k8sVersionNameMap[v[0].String()] = versionName
				k8sVersions = append(k8sVersions, v[0])
			}
			semver.Sort(k8sVersions)

			// Looks for previous version
			cvArr, err := parseVersions(record.Name)
			if err != nil {
				if errtx := tx.Rollback(); errtx != nil {
					return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
				}
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"parseVersions.record.Name", friendlyMessage+err.Error())
			}
			currentVer := cvArr[0]

			for i := range k8sVersions {
				// Current version reached
				prevVer := k8sVersions[i]

				if prevVer.Major == currentVer.Major && prevVer.Minor == currentVer.Minor && prevVer.Patch < currentVer.Patch {
					// Updates old version to state archived
					_, err = tx.ExecContext(ctx, updateK8sVersionStateQuery, lifecyclestateArchived, k8sVersionNameMap[prevVer.String()])
					if err != nil {
						if errtx := tx.Rollback(); errtx != nil {
							return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
						}
						return returnError, utils.ErrorHandler(ctx, err, failedFunction+"updateK8sVersionStateQueryu", friendlyMessage+err.Error())
					}
					break
				}
			}
		}

		// UPDATE K8S VERSION
		_, err = tx.ExecContext(ctx, updateK8sVersionStateQuery, newLifecyleState, record.Name)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"updateK8sVersionStateQuery", friendlyMessage+err.Error())
		}

	}

	// COMMIT
	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"tx.commit", friendlyMessage+err.Error())
	}

	return GetK8SVersion(ctx, dbconn, &pb.GetK8SRequest{Name: record.Name})
}

func UpgradeClusterControlPlane(ctx context.Context, dbconn *sql.DB, record *pb.UpgradeControlPlaneRequest) (*empty.Empty, error) {
	friendlyMessage := "Could not Upgrade Cluster Control Plane. Please try again."
	failedFunction := "UpgradeClusterControlPlane"
	returnError := &empty.Empty{}
	returnValue := &empty.Empty{}

	/* VALIDATIONS */
	// Validate Cluster Existance
	var clusterId int32
	clusterId, err := utils.ValidateClusterExistance(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterExistance", friendlyMessage+err.Error())
	}
	if clusterId == -1 {
		return returnError, grpc_status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid)
	}
	// Validate cluster is in actionable state
	actionableState, err := utils.ValidaterClusterActionable(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterActionable", friendlyMessage+err.Error())
	}
	if !actionableState {
		return returnError, grpc_status.Error(codes.FailedPrecondition, "Cluster is not in an actionable state")
	}

	/* GET INFO NEEDED TO UPDATE THE CONTROL PLANE VERSION */
	// Get the available Patch k8sversions
	var upgradeVersion string
	var upgradeImi string
	availableVersionsKeys, availableVersionsImi, err := utils.GetAvailableControlPlaneImiUpgrades(ctx, dbconn, record.Clusteruuid, record.Nodegroupuuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetAvailableControlPlaneImiUpgrades", friendlyMessage+err.Error())
	}
	if len(availableVersionsImi) == 0 {
		return returnError, grpc_status.Error(codes.FailedPrecondition, "No available upgrades for the control plane")
	}
	if record.K8Sversionname != nil {
		// Get Cluster Compat information
		clusterProvider, _, clusterRuntime, clusterInstanceType, clusterOsImage, err := utils.GetClusterCompatDetails(ctx, dbconn, record.Clusteruuid)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetClusterCompatDetails", friendlyMessage+err.Error())
		}
		// Get the imi for the k8sversion
		err = dbconn.QueryRowContext(ctx, getImiFromK8sVersion,
			cpNodegroupType,
			clusterProvider,
			clusterRuntime,
			clusterInstanceType,
			clusterOsImage,
			record.K8Sversionname,
		).Scan(&upgradeImi)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ScanImiFromK8sVersion", friendlyMessage+err.Error())
		}
		// Set the upgrade variables
		upgradeVersion = *record.K8Sversionname
	} else {
		// Set the upgrade variables
		upgradeVersion = availableVersionsKeys[len(availableVersionsImi)-1]
		upgradeImi = availableVersionsImi[len(availableVersionsImi)-1]
	}
	// Get the Artifiact for the new IMI
	imiArtifactRepo, err := utils.GetImiArtifact(ctx, dbconn, upgradeImi)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetImiArtifact", friendlyMessage+err.Error())
	}
	// Get the default Addons for the latest version
	defaultAddons, err := utils.GetDefaultAddons(ctx, dbconn, upgradeVersion)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetDefaultAddons", friendlyMessage)
	}

	/* UPGRADE THE CONTROL PLANE IMI DB ENTRY AND CRD*/
	// Begin the transaction
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage+err.Error())
	}

	/* UPGRADE THE CONTROL PLANE IMI */
	// update the Cluster's ControlPlane with new k8sversion
	_, err = tx.Exec(updateNodegroupImiQuery,
		record.Nodegroupuuid,
		upgradeVersion,
		upgradeImi,
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"updateNodegroupImiQuery", friendlyMessage+err.Error())
	}

	// Get current json from REV table
	currentJson, clusterCrd, err := utils.GetLatestClusterRev(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetLatestClusterRev", friendlyMessage+err.Error())
	}

	// Make changes to the Control Plane CRD
	clusterCrd.Spec.KubernetesVersion = upgradeVersion
	clusterCrd.Spec.InstanceIMI = imiArtifactRepo

	// Update the CRD for Addons
	var newAddons []clusterv1alpha.AddonTemplateSpec
	for _, addon := range defaultAddons {
		addonCrd := clusterv1alpha.AddonTemplateSpec{
			Name:     addon.Name,
			Type:     clusterv1alpha.AddonType(addon.Type),
			Artifact: addon.Artifact,
		}
		newAddons = append(newAddons, addonCrd)
	}
	clusterCrd.Spec.Addons = newAddons

	// Create new cluster rev table entry
	var revversion string
	err = tx.QueryRowContext(ctx, insertRevQuery,
		clusterId,
		currentJson,
		clusterCrd,
		"test", // ?? DEFAULT VALUES
		"test", // ?? DEFAULT VALUES
		"test", // ?? DEFAULT VALUES
		"test", // ?? DEFAULT VALUES
		time.Now(),
		false,
	).Scan(&revversion)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"insertRevQuery", friendlyMessage+err.Error())
	}

	/* UPDATE STATES TO PENDING*/
	// Update the cluster state
	_, err = tx.QueryContext(ctx, updateClusterStateQuery,
		clusterId,
		"Pending",
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"updateClusterStateQuery", friendlyMessage+err.Error())
	}
	// Update the nodegroup state
	_, err = tx.QueryContext(ctx, updateNodegroupStateQuery,
		record.Nodegroupuuid,
		"Updating",
	)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"updateNodegroupStateQuery", friendlyMessage+err.Error())
	}

	/* COMMIT TRANSACTION */
	// close the transaction with a Commit() or Rollback() method on the resulting Tx variable.
	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"tx.commit", friendlyMessage+err.Error())
	}

	/* GET NODEGROUP STATUS */
	return returnValue, nil
}

func PutCloudAccountApproveList(ctx context.Context, dbconn *sql.DB, record *pb.CloudAccountApproveListRequest) (*pb.CloudAccountApproveList, error) {
	// Start the transaction
	friendlyMessage := "PutCloudAccountApproveList.UnexpectedError"
	failedFunction := "PutCloudAccountApproveList."
	returnError := &pb.CloudAccountApproveList{}

	cloudAccountID = record.GetCloudaccountId()
	if cloudAccountID == "" {
		return returnError, grpc_status.Error(codes.InvalidArgument, "PutCloudAccountApproveList.CloudAccountID: Invalid Input Request")
	}

	// Start the transaction
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage+err.Error())
	}

	// Validate Cloud Account Existence
	var count int32
	err = tx.QueryRowContext(ctx, GetCloudAccountApproveListCountQuery, record.CloudaccountId).Scan(&count)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetCloudAccountApproveListQuery", friendlyMessage+err.Error())
	}
	if count != 1 {
		return returnError, grpc_status.Error(codes.NotFound, "No Cloud Account exists under the given ID")
	}

	/* Get available cloud account approve list details */
	var accountId string
	var status bool
	var storageStatus bool
	err = dbconn.QueryRowContext(ctx, getCloudAccountsApproveListByIDQuery, record.CloudaccountId).Scan(&accountId, &status, &storageStatus)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getCloudAccountsApproveListQuery", friendlyMessage+err.Error())
	}

	if accountId == record.CloudaccountId && !status && storageStatus {
		return returnError, grpc_status.Error(codes.PermissionDenied, "Cloud Account is Disabled for the user. Cannot Enable Storage. Please Activate Cloud Account First")
	}

	/* Get override details for the specific cloud account and update them accordingly */
	maxClustersPerCloudAccountRequest := 0
	maxNodegroupsPerClusterRequest := 0
	maxIlbsPerClusterRequest := 0
	maxNodesPerClusterRequest := 0
	maxNodesPerNodegroupRequest := 0
	defaultvalues, err := utils.GetDefaultValues(ctx, dbconn)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetDefaultValues", friendlyMessage)
	}
	_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, maxIlbsPerCluster, maxNodegroupsPerCluster, max_nodegroup_vm, _, _, cloudAccountMaxClusters, err := utils.ConvDefaultsToInt(ctx, defaultvalues)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"convDefaultsToInt", friendlyMessage)
	}
	max_cluster_vm, err := strconv.Atoi(defaultvalues["max_cluster_vm"])
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"convDefaultsToInt", friendlyMessage)
	}
	limits := &pb.ResourceLimits{
		Maxclusterpercloudaccount: int32(cloudAccountMaxClusters),
		Maxnodegroupspercluster:   int32(maxNodegroupsPerCluster),
		Maxvipspercluster:         int32(maxIlbsPerCluster),
		Maxnodespernodegroup:      int32(max_nodegroup_vm),
		Maxclustervm:              int32(max_cluster_vm),
	}
	if record.MaxclustersOverride > 0 && record.MaxclustersOverride >= limits.Maxclusterpercloudaccount {
		maxClustersPerCloudAccountRequest = int(record.MaxclustersOverride)
	}
	if record.MaxclusterngOverride > 0 && record.MaxclusterngOverride >= limits.Maxnodegroupspercluster {
		maxNodegroupsPerClusterRequest = int(record.MaxclusterngOverride)
	}
	if record.MaxclusterilbOverride > 0 && record.MaxclusterilbOverride >= limits.Maxvipspercluster {
		maxIlbsPerClusterRequest = int(record.MaxclusterilbOverride)
	}
	if record.MaxclustervmOverride > 0 && record.MaxclustervmOverride >= limits.Maxclustervm {
		maxNodesPerClusterRequest = int(record.MaxclustervmOverride)
	}
	if record.MaxnodegroupvmOverride > 0 && record.MaxnodegroupvmOverride >= limits.Maxnodespernodegroup {
		maxNodesPerNodegroupRequest = int(record.MaxnodegroupvmOverride)
	}

	/* Update cloud account approve list details */
	_, err = tx.ExecContext(ctx, updateCloudAccountApproveListQuery, record.CloudaccountId, record.Status, record.EnableStorage, maxClustersPerCloudAccountRequest, maxNodegroupsPerClusterRequest, maxIlbsPerClusterRequest, maxNodesPerClusterRequest, maxNodesPerNodegroupRequest)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"updateCloudAccountApproveListQuery", friendlyMessage+err.Error())
	}

	/* COMMIT TRANSACTION */
	// close the transaction with a Commit() or Rollback() method on the resulting Tx variable.
	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"tx.Commit", friendlyMessage+err.Error())
	}

	returnValue := &pb.CloudAccountApproveList{
		Account:      record.CloudaccountId,
		Providername: iksProvider,
		Status:       record.Status,
	}

	return returnValue, nil
}

// Update will involve to update to any new IMIs Patch versions into k8scompatibility table
func UpdateIMIToK8sCompatibility(ctx context.Context, dbconn *sql.DB, record *pb.IMIInstanceTypeK8SRequest) (*pb.IMIInstanceTypeK8SResponse, error) {
	friendlyMessage := "Could not Create IMI Instance Type to K8S. Please try again."
	failedFunction := "UpdateIMIToK8sCompatibility."
	returnError := &pb.IMIInstanceTypeK8SResponse{}

	// Validate Os Image Existence
	name = ""
	err := dbconn.QueryRowContext(ctx, GetOsImageNameQuery, record.Os).Scan(&name, &osImageState)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, grpc_status.Error(codes.NotFound, "No Os Image exists under the given name")
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

	// Validate OS Image Instance Name Existence
	name = ""
	err = dbconn.QueryRowContext(ctx, GetOsImageInstanceNameQuery, record.Name).Scan(&name)
	if err != sql.ErrNoRows && err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetOsImageInstanceNameQuery", friendlyMessage+err.Error())
	}
	if name == "" {
		return returnError, grpc_status.Error(codes.NotFound, "Worker IMI name doesnot exist")
	}

	// Start the transaction
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage+err.Error())
	}

	if record.Type == imiWorkerType {
		for _, instanceType := range record.Instancetypes {
			// Update IMI
			_, err = tx.ExecContext(ctx, updateWorkerIMIInstanceTypeToK8sCompatibilityQuery,
				record.Runtime,
				record.Upstreamreleasename,
				record.Os,
				record.Provider,
				instanceType,
				record.Name,
			)
			if err != nil {
				if errtx := tx.Rollback(); errtx != nil {
					return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TrasactionRollbackFailed", friendlyMessage+errtx.Error())
				}
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"updateIMIInstanceTypeToK8sCompatibilityQuery", friendlyMessage+err.Error())
			}
		}
	}

	if record.Type == imiControlPlaneType {
		// Update IMI
		_, err = tx.ExecContext(ctx, updateCPIMIInstanceTypeToK8sCompatibilityQuery,
			record.Runtime,
			record.Upstreamreleasename,
			record.Os,
			record.Provider,
			record.Name,
		)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TrasactionRollbackFailed", friendlyMessage+errtx.Error())
			}
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"updateIMIInstanceTypeToK8sCompatibilityQuery", friendlyMessage+err.Error())
		}
	}

	/* COMMIT TRANSACTION */
	// close the transaction with a Commit() or Rollback() method on the resulting Tx variable.
	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"tx.Commit", friendlyMessage+err.Error())
	}

	imiInstancetypeK8sUpdateResponse := &pb.IMIInstanceTypeK8SResponse{
		Osimageinstance:     record.Name,
		Upstreamreleasename: record.Upstreamreleasename,
		Provider:            record.Provider,
		Type:                record.Type,
		Runtime:             record.Runtime,
		Os:                  record.Os,
		State:               record.State,
	}

	return imiInstancetypeK8sUpdateResponse, nil
}

// Update will involve to update to any new Instance Types versions into k8scompatibility table
func UpdateInstanceTypeToK8sCompatibility(ctx context.Context, dbconn *sql.DB, record *pb.InstanceTypeIMIK8SRequest) (*pb.InstanceTypeIMIK8SResponse, error) {
	friendlyMessage := "Could not Create Instance Type to K8S. Please try again."
	failedFunction := "UpdateInstanceTypeToK8sCompatibility."
	returnError := &pb.InstanceTypeIMIK8SResponse{}

	// Validate InstanceType Existence && Active State
	instanceName := ""
	err := dbconn.QueryRowContext(ctx, getInstanceTypesExistenceQuery, record.Name).Scan(&instanceName)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, grpc_status.Error(codes.FailedPrecondition, "No Instance Type Exist under the given name or maybe Inactive.")
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getInstanceTypesExistenceQuery", friendlyMessage+err.Error())
	}

	if len(record.Instacetypeimik8Scompatibilityresponse) <= 0 {
		return returnError, grpc_status.Error(codes.NotFound, "No IMIs to tag to the given instance types.")
	}

	// Start the transaction
	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage+err.Error())
	}

	lifeCycleState := "1"
	for _, imiRecord := range record.Instacetypeimik8Scompatibilityresponse {
		if record.Category == imiRecord.Category && record.Family == imiRecord.Family && imiRecord.Type == imiWorkerType {
			// Validate Runtime Existence
			err = dbconn.QueryRowContext(ctx, GetRuntimeVersionNameQuery, imiRecord.Runtime).Scan(&runtimeVersionName, &runtimeName)
			if err != nil {
				if err == sql.ErrNoRows {
					return returnError, grpc_status.Error(codes.NotFound, "No Runtime exists under the given name")
				}
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetRuntimeVersionNameQuery", friendlyMessage+err.Error())
			}

			// Validate Provider Existence
			err = dbconn.QueryRowContext(ctx, GetProviderNameQuery, imiRecord.Provider).Scan(&providerName, &providerState)
			if err != nil {
				if err == sql.ErrNoRows {
					return returnError, grpc_status.Error(codes.NotFound, "No Provider exists under the given name")
				}
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetProviderNameQuery", friendlyMessage+err.Error())
			}

			if providerState != "1" {
				return returnError, grpc_status.Error(codes.NotFound, "No Provider exists under the given name")
			}

			// Validate Worker OS Image Instance Name Existence
			workerOsImageName := ""
			err = dbconn.QueryRowContext(ctx, GetOsImageInstanceNameQuery, imiRecord.Name).Scan(&workerOsImageName)
			if err != sql.ErrNoRows && err != nil {
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetOsImageInstanceNameQuery", friendlyMessage+err.Error())
			}
			if workerOsImageName == "" {
				return returnError, grpc_status.Error(codes.NotFound, "Worker IMI name doesn't exist")
			}

			// Validate Controlplane OS Image Instance Name Existence
			cpOsImageName := ""
			err = dbconn.QueryRowContext(ctx, GetOsImageInstanceNameQuery, imiRecord.Cposimageinstances).Scan(&cpOsImageName)
			if err != sql.ErrNoRows && err != nil {
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetOsImageInstanceNameQuery", friendlyMessage+err.Error())
			}
			if cpOsImageName == "" {
				return returnError, grpc_status.Error(codes.NotFound, "Control Plane IMI name doesn't exist")
			}

			// Validate InstanceType Existence in K8S Compatibility table
			instanceTypeName := ""
			err = dbconn.QueryRowContext(ctx, getK8sCompatibilityQueryByInstanceType,
				imiRecord.Runtime,
				imiRecord.Upstreamreleasename,
				imiRecord.Os,
				imiRecord.Name,
				imiRecord.Provider,
				record.Name).Scan(&instanceTypeName)
			if err != nil {
				if err == sql.ErrNoRows {
					// Insert InstanceType to K8sCompatibility table
					_, err = tx.ExecContext(ctx, insertK8sCompatibilityQuery,
						imiRecord.Runtime,
						imiRecord.Upstreamreleasename,
						imiRecord.Os,
						imiRecord.Cposimageinstances,
						imiRecord.Name,
						imiRecord.Provider,
						record.Name,
						lifeCycleState,
					)
					if err != nil {
						if errtx := tx.Rollback(); errtx != nil {
							return returnError, utils.ErrorHandler(ctx, errtx, failedFunction+"TrasactionRollbackFailed", friendlyMessage+errtx.Error())
						}
						return returnError, utils.ErrorHandler(ctx, err, failedFunction+"insertK8sCompatibilityQuery", friendlyMessage+err.Error())
					}
				} else {
					return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getK8sCompatibilityQueryByInstanceType", friendlyMessage+err.Error())
				}
			}
			if instanceTypeName != "" {
				return returnError, grpc_status.Error(codes.AlreadyExists, "Instance Type already Present. Please update the instance type to k8scompatibility on IMIs Tab")
			}
		}
	}

	/* COMMIT TRANSACTION */
	// close the transaction with a Commit() or Rollback() method on the resulting Tx variable.
	err = tx.Commit()
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"tx.Commit", friendlyMessage+err.Error())
	}

	instancetypeIMIK8sUpdateResponse := &pb.InstanceTypeIMIK8SResponse{
		Name:             record.Name,
		Nodeprovidername: record.Nodeprovidername,
		Category:         record.Category,
		Family:           record.Family,
	}

	return instancetypeIMIK8sUpdateResponse, nil
}
