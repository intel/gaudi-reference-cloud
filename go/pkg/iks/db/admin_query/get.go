// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package admin_query

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/codes"
	grpc_status "google.golang.org/grpc/status"
	"strconv"

	"github.com/blang/semver/v4"
	query "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/db_query_constants"
	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/iks_utils"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

const (
	GetIMIsQuery = `
	SELECT osimageinstance.osimageinstance_name, osimageinstance.osimage_name, osimageinstance.runtime_name, nodegrouptype_name, osimageinstance.provider_name, osimageinstance.k8sversion_name,
    lifecyclestate.name, imiartifact,
	CASE WHEN instancetypecategory IS NULL THEN '' ELSE instancetypecategory END AS instancetypecategory,
	CASE WHEN instancetypefamiliy IS NULL THEN '' ELSE instancetypefamiliy END AS instancetypefamiliy,
	CASE WHEN COUNT(k8scompatibility.osimage_name) > 0 THEN true ELSE false END AS is_active_imi
    FROM public.osimageinstance
    INNER JOIN public.lifecyclestate
    ON lifecyclestate.lifecyclestate_id = osimageinstance.lifecyclestate_id
	LEFT JOIN public.k8scompatibility
	ON k8scompatibility.wrk_osimageinstance_name = osimageinstance.osimageinstance_name OR k8scompatibility.cp_osimageinstance_name = osimageinstance.osimageinstance_name
    GROUP BY  osimageinstance.osimageinstance_name,  osimageinstance.osimage_name, osimageinstance.runtime_name, nodegrouptype_name, osimageinstance.provider_name, bootstrap_repo, osimageinstance.k8sversion_name,
    lifecyclestate.name, imiartifact;
	`

	GetIMIQuery = `
	SELECT osimageinstance.osimage_name, osimageinstance.runtime_name, nodegrouptype_name, osimageinstance.provider_name, osimageinstance.k8sversion_name,
    lifecyclestate.name, imiartifact,
	CASE WHEN instancetypecategory IS NULL THEN '' ELSE instancetypecategory END AS instancetypecategory,
	CASE WHEN instancetypefamiliy IS NULL THEN '' ELSE instancetypefamiliy END AS instancetypefamiliy,
	CASE WHEN COUNT(k8scompatibility.osimage_name) > 0 THEN true ELSE false END AS is_active_imi
    FROM public.osimageinstance 
    INNER JOIN public.lifecyclestate
    ON lifecyclestate.lifecyclestate_id = osimageinstance.lifecyclestate_id
	LEFT JOIN public.k8scompatibility
	ON k8scompatibility.wrk_osimageinstance_name = osimageinstance.osimageinstance_name OR k8scompatibility.cp_osimageinstance_name = osimageinstance.osimageinstance_name
	where osimageinstance.osimageinstance_name = $1
    GROUP BY  osimageinstance.osimageinstance_name,  osimageinstance.osimage_name, osimageinstance.runtime_name, nodegrouptype_name, osimageinstance.provider_name, bootstrap_repo, osimageinstance.k8sversion_name,
    lifecyclestate.name, imiartifact;
	`

	GetIMIComponentsQuery = `
	SELECT component_name, version, artifact_repo FROM public.osimageinstancecomponent WHERE osimageinstance_name = $1;
	`

	GetIMIQueryByFamilyAndCategory = `
	SELECT osimageinstance.osimageinstance_name, CASE WHEN COUNT(osimageinstancecomponent.component_name) > 0 THEN json_agg(json_build_object('name',component_name, 'version', version, 'artifact', artifact_repo)) ELSE '[]' END AS components, 
    osimageinstance.osimage_name, osimageinstance.runtime_name, nodegrouptype_name, osimageinstance.provider_name, osimageinstance.k8sversion_name,
    lifecyclestate.name, imiartifact,
	CASE WHEN instancetypecategory IS NULL THEN '' ELSE instancetypecategory END AS instancetypecategory,
	CASE WHEN instancetypefamiliy IS NULL THEN '' ELSE instancetypefamiliy END AS instancetypefamiliy
    FROM public.osimageinstance 
    INNER JOIN public.lifecyclestate
    ON lifecyclestate.lifecyclestate_id = osimageinstance.lifecyclestate_id
    LEFT JOIN public.osimageinstancecomponent
    ON osimageinstancecomponent.osimageinstance_name = osimageinstance.osimageinstance_name
	LEFT JOIN public.k8scompatibility
	ON k8scompatibility.wrk_osimageinstance_name = osimageinstance.osimageinstance_name OR k8scompatibility.cp_osimageinstance_name = osimageinstance.osimageinstance_name
	where instancetypecategory = $1 AND instancetypefamiliy = $2 AND osimageinstance.lifecyclestate_id = (SELECT lifecyclestate_id from public.lifecyclestate WHERE name = 'Active')
    GROUP BY  osimageinstance.osimageinstance_name,  osimageinstance.osimage_name, osimageinstance.runtime_name, nodegrouptype_name, osimageinstance.provider_name, bootstrap_repo, osimageinstance.k8sversion_name,
    lifecyclestate.name, imiartifact;
	`

	GetIMIsInfoQuery = `
	WITH osimage_cte AS (
			SELECT json_agg(json_build_object('osimage', osimage_name)) AS osimage 
			FROM public.osimage oi 
			INNER JOIN public.lifecyclestate li ON li.lifecyclestate_id = oi.lifecyclestate_id 
			WHERE li.name = 'Active'), 
		provider_cte AS (
			SELECT json_agg(json_build_object('provider', provider_name)) AS provider 
			FROM public.provider p INNER JOIN public.lifecyclestate li ON li.lifecyclestate_id = p.lifecyclestate_id 
			WHERE li.name = 'Active'), 
		runtime_cte AS (
			SELECT json_agg(json_build_object('runtime', runtime_name)) AS runtime 
			FROM public.runtime), 
		state_cte AS (
			SELECT json_agg(json_build_object('state', name)) AS state FROM public.lifecyclestate) 
		SELECT osimage.osimage, provider.provider, runtime_cte.runtime, state_cte.state 
		FROM osimage_cte osimage 
		CROSS JOIN provider_cte provider 
		CROSS JOIN runtime_cte 
		CROSS JOIN state_cte;
	`
	GetInstanceTypesQuery = `
	SELECT 
		i.instancetype_name,
		i.memory,
		i.cpu,
		i.nodeprovider_name,
		i.storage,
		li.name AS status,
		coalesce(i.displayname, i.instancetype_name) AS displayname,
		i.imi_override,
		CASE WHEN i.instancecategory IS NULL THEN '' ELSE instancecategory END AS instancecategory,
		CASE WHEN i.instancetypefamiliy IS NULL THEN '' ELSE instancetypefamiliy END AS instancetypefamiliy,
		CASE WHEN i.description IS NULL THEN '' ELSE description END AS description
		FROM instancetype i
		INNER JOIN lifecyclestate li
		ON li.lifecyclestate_id = i.lifecyclestate_id;
	`
	GetActiveInstanceTypesQuery = `
	SELECT 
		i.instancetype_name,
		i.memory,
		i.cpu,
		i.nodeprovider_name,
		i.storage,
		li.name AS status,
		coalesce(i.displayname, i.instancetype_name) AS displayname,
		i.imi_override,
		CASE WHEN i.instancecategory IS NULL THEN '' ELSE instancecategory END AS instancecategory,
		CASE WHEN i.instancetypefamiliy IS NULL THEN '' ELSE instancetypefamiliy END AS instancetypefamiliy,
		CASE WHEN i.description IS NULL THEN '' ELSE description END AS description
		FROM instancetype i
		INNER JOIN lifecyclestate li
		ON li.lifecyclestate_id = i.lifecyclestate_id
		WHERE i.lifecyclestate_id = (SELECT lifecyclestate_id from public.lifecyclestate WHERE name = 'Active');
	`

	GetInstanceTypeQuery = `
	SELECT 
		i.instancetype_name,
		i.memory,
		i.cpu,
		i.nodeprovider_name,
		i.storage,
		li.name AS status,
		coalesce(i.displayname, i.instancetype_name) AS displayname,
		i.imi_override,
		CASE WHEN i.instancecategory IS NULL THEN '' ELSE instancecategory END AS instancecategory,
		CASE WHEN i.instancetypefamiliy IS NULL THEN '' ELSE instancetypefamiliy END AS instancetypefamiliy,
		CASE WHEN i.description IS NULL THEN '' ELSE description END AS description
		FROM instancetype i
		INNER JOIN lifecyclestate li
		ON li.lifecyclestate_id = i.lifecyclestate_id
		WHERE i.instancetype_name = $1;
	`

	GetOsimageInstanceK8sCompatibilityQueryByInstanceName = `
	SELECT wrk_osimageinstance_name FROM public.k8scompatibility where runtime_name = $1 AND k8sversion_name = $2 AND osimage_name = $3 AND provider_name = $4 AND instancetype_name = $5`

	GetOsimageInstanceK8sCompatibilityQuery = `
	SELECT wrk_osimageinstance_name FROM public.k8scompatibility where runtime_name = $1 AND k8sversion_name = $2 AND osimage_name = $3 AND provider_name = $4 AND instancetype_name = $5 AND wrk_osimageinstance_name =$6`

	GetControlPlaneOsimageInstanceByK8sVersionQuery = `
	SELECT osimageinstance_name FROM public.osimageinstance where k8sversion_name = $1 AND nodegrouptype_name = $2 ORDER BY created_date DESC LIMIT 1`

	GetWorkerPlaneOsimageInstanceByK8sVersionQuery = `
	SELECT osimageinstance_name FROM public.osimageinstance where k8sversion_name = $1 AND nodegrouptype_name = $2 AND instancetypecategory = $3 AND instancetypefamiliy = $4`

	GetCountComponentQuery = `
	SELECT count(component_name) FROM public.osimageinstancecomponent where osimageinstance_name = $1`

	GetComponentQuery = `
	SELECT component_name , version , artifact_repo FROM public.osimageinstancecomponent where osimageinstance_name = $1`

	GetSshkeys = `
	SELECT cluster_ssh_key, cluster_ssh_pub_key,encryptionkey_id,nonce FROM public.cluster_extraconfig where cluster_id = $1`

	clusterExistanceQuery = `
		SELECT count(*) 
		FROM public.cluster 
		WHERE unique_id = $1
	`

	k8sExistanceQuery = `
	SELECT k8sversion_name 
	FROM public.k8sversion 
	WHERE k8sversion_name = $1
	`

	k8sExistanceAndActiveQuery = `
	SELECT k8sversion_name 
	FROM public.k8sversion 
	WHERE k8sversion_name = $1 AND lifecyclestate_id = (SELECT lifecyclestate_id from public.lifecyclestate WHERE name = 'Active')
	`

	getClusterControlPlaneUuidQuery = `
		SELECT n.unique_id
	 	FROM public.nodegroup n
		WHERE n.nodegrouptype_name = 'ControlPlane' AND n.cluster_id = (
			SELECT c.cluster_id 
			FROM public.cluster  c
			WHERE c.unique_id = $1
		)
	`

	getClusterQuery = `
		SELECT 
			cl.name, 
			cl.provider_name, 
			cl.cloudaccount_id,
			cl.region_name,
			ng.k8sversion_name,
			ng.name,
			nd.created_date,
			cl.networkservice_cidr,
			cl.networkpod_cidr,
			cl.clustertype
		FROM cluster AS cl
		JOIN nodegroup AS ng ON ng.cluster_id = cl.cluster_id AND ng.nodegrouptype_name = 'ControlPlane'
		LEFT JOIN k8snode AS nd ON ng.nodegroup_id = nd.nodegroup_id
		WHERE cl.unique_id = $1
		ORDER BY nd.created_date DESC
		LIMIT 1
	`
	getNodeGroupNodesQuery = `
		SELECT
			n.k8snode_name,
			n.ip_address,
			COALESCE(n.dns_name, ''),
			COALESCE(n.osimageinstance_name, ''),
			n.k8snodestate_name,
			n.kubernetes_status,
			n.created_date,
			n.weka_storage_client_id,
			n.weka_storage_status,
			n.weka_storage_custom_status,
			n.weka_storage_message
		FROM public.k8snode n
		WHERE n.nodegroup_id = (
			SELECT nodegroup_id
			FROM public.nodegroup
			WHERE unique_id = $1
		)
	`

	getClusterNodeGroupsQuery = `
		SELECT 
			ng.unique_id,
			ng.nodegroup_id,
			ng.name,
			ng.nodecount,
			ng.k8sversion_name,
			ng.nodegroupstate_name,
			ng.osimageinstance_name,
			ng.instancetype_name,
			ng.sshkey,
			ng.nodegrouptype_name,
			ng.nodegrouptype
		FROM nodegroup ng
		WHERE ng.cluster_id = (SELECT c.cluster_id FROM public.cluster c WHERE c.unique_id = $1)
	`

	getClusterAddonsQuery = `
		SELECT 
			ad.name,
			ad.version,
			cad.clusteraddonstate_name, 
			cad.addonargs,
			ad.tags,
			ad.artifact_repo
		FROM clusteraddonversion AS cad
		JOIN addonversion AS ad ON ad.addonversion_name = cad.addonversion_name
		JOIN cluster AS cl ON cl.cluster_id = cad.cluster_id
		WHERE cl.unique_id = $1
	`

	getClusterSnapshotsQuery = `
	SELECT
		ss.snapshot_name,
		ss.schedule_type,
		ss.snapshotstate_name,
		ss.created_date,
		ss.snapshotfile
	FROM snapshot AS ss 
	JOIN cluster as cl on cl.cluster_id = ss.cluster_id
	WHERE cl.unique_id = $1
	`

	getK8sActiveVersionsForProviderQuery = `
		SELECT 
			v.k8sversion_name,
			c.cp_osimageinstance_name,
			c.wrk_osimageinstance_name
		FROM k8sversion AS v
		JOIN k8scompatibility AS c ON c.k8sversion_name = v.k8sversion_name
		WHERE v.lifecyclestate_id = 1 AND c.provider_name = $1
	`

	getClustersQuery = `
		SELECT 
			cl.unique_id,
			cl.name, 
			cl.clusterstate_name,
			cl.provider_name,
			cl.cloudaccount_id,
			ng.k8sversion_name,
			cl.created_date,
			cl.clustertype
		FROM cluster AS cl
		JOIN nodegroup AS ng 
			ON ng.cluster_id = cl.cluster_id AND ng.nodegrouptype_name = 'ControlPlane'
	`

	getActiveClustersQuery = `
	SELECT 
		count(*)
	FROM cluster AS cl
	JOIN nodegroup AS ng 
		ON ng.cluster_id = cl.cluster_id AND ng.nodegrouptype_name = 'ControlPlane' AND ng.k8sversion_name = $1 AND cl.clusterstate_name = 'Active'
`

	getAllK8sActiveVersionsQuery = `
		SELECT 
			v.k8sversion_name,
			c.provider_name,
			c.cp_osimageinstance_name,
			c.wrk_osimageinstance_name
		FROM k8sversion AS v
		JOIN k8scompatibility AS c ON c.k8sversion_name = v.k8sversion_name
		WHERE v.lifecyclestate_id = 1
	`

	getK8sVersionsQuery = `
		SELECT 
			c.cp_osimageinstance_name,
			c.k8sversion_name,
			c.provider_name,
			c.k8sversion_name,
			c.wrk_osimageinstance_name,
			c.runtime_name,
			l.name,
			v.major_version,
			v.minor_version
		FROM k8scompatibility AS c
		JOIN k8sversion AS v ON c.k8sversion_name = v.k8sversion_name
		JOIN lifecyclestate AS l on l.lifecyclestate_id = v.lifecyclestate_id
	`

	getK8sVersionQuery = `
		SELECT 
			c.cp_osimageinstance_name,
			c.k8sversion_name,
			c.provider_name,
			c.k8sversion_name,
			c.wrk_osimageinstance_name,
			c.runtime_name,
			l.name,
			v.major_version,
			v.minor_version
		FROM k8scompatibility AS c
		JOIN k8sversion AS v ON c.k8sversion_name = v.k8sversion_name
		JOIN lifecyclestate AS l on l.lifecyclestate_id = v.lifecyclestate_id
		WHERE c.k8sversion_name = $1
	`

	getCloudAccountsApproveListQuery = `
		SELECT 
			cloudaccount_id,
			provider_name,
			active_account_create_cluster,
			allow_create_storage,
			CASE WHEN maxclusters_override IS NULL THEN 0 ELSE maxclusters_override END AS maxclusters_override,
  			CASE WHEN maxclusterng_override IS NULL THEN 0 ELSE maxclusterng_override END AS maxclusterng_override,
  			CASE WHEN maxclusterilb_override IS NULL THEN 0 ELSE maxclusterilb_override END AS maxclusterilb_override,
  			CASE WHEN maxnodegroupvm_override IS NULL THEN 0 ELSE maxnodegroupvm_override END AS maxnodegroupvm_override,
  			CASE WHEN maxclustervm_override IS NULL THEN 0 ELSE maxclustervm_override END AS maxclustervm_override
		FROM public.cloudaccountextraspec
	`

	getCloudAccountsApproveListByIDQuery = `
	SELECT 
		cloudaccount_id,
		active_account_create_cluster,
		allow_create_storage
	FROM public.cloudaccountextraspec
	WHERE cloudaccount_id = $1; 
	`

	getlbIdsQuery = `
	SELECT vip_id FROM public.vip where cluster_id = $1`

	getlbIdQuery = `
	SELECT v.vip_id , d.vip_name, v.vipstate_name , v.vip_ip , d.port, d.pool_port, v.viptype_name, v.created_date
	FROM public.vip v
		INNER JOIN public.vipdetails d ON 
		v.vip_id = d.vip_id
	WHERE v.cluster_id = $1 AND v.vip_id = $2`
)

const (
	iksProvider            = "iks"
	rke2Provider           = "rke2"
	cpNodegroupType        = "ControlPlane"
	wkNodegroupType        = "Worker"
	ngStatusUpdating       = "Updating"
	imiWorkerType          = "worker"
	imiControlPlaneType    = "controlplane"
	lifecyclestateActive   = 1
	lifecyclestateArchived = 2
	lifecyclestateStaged   = 3
)

var (
	imiName                  string
	imiProvider              string
	imiType                  string
	imiState                 string
	imiComponents            string
	imiRuntime               string
	imiBootstrapRepo         string
	imiUpstreamReleaseName   string
	imiComponentName         string
	imiComponentVersion      string
	imiComponentArtifactRepo string
	imiArtifactRepo          string
	imiOs                    string
	imiCategory              string
	imiFamily                string
	k8sversion               string
	isactiveimi              bool
)

// Instance Type Response
var (
	instanceTypeName string
	memory           int32
	cpu              int32
	nodeProviderName string
	storage          int32
	status           string
	displayName      string
	imiOverride      bool
	category         string
	family           string
	description      string
)

func GetInstanceTypes(ctx context.Context, dbconn *sql.DB) (*pb.GetInstanceTypesResponse, error) {
	friendlyMessage := "GetInstanceTypes.UnexpectedError"
	failedFunction := "GetInstanceTypes."
	returnError := &pb.GetInstanceTypesResponse{}

	// GET INSTANCETYPES
	var instanceTypeResponse []*pb.InstanceTypeResponse
	rows, err := dbconn.QueryContext(ctx, GetInstanceTypesQuery)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetInstanceTypesQuery", friendlyMessage+err.Error())
	}
	defer rows.Close()

	// PARSE ROWSE
	for rows.Next() {
		err := rows.Scan(&instanceTypeName, &memory, &cpu, &nodeProviderName, &storage, &status, &displayName, &imiOverride, &category, &family, &description)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetInstanceTypesQuery.rows.scan", friendlyMessage+err.Error())
		}
		instanceType := &pb.InstanceTypeResponse{
			Instancetypename: instanceTypeName,
			Memory:           memory,
			Cpu:              cpu,
			Nodeprovidername: nodeProviderName,
			Storage:          storage,
			Status:           status,
			Displayname:      displayName,
			Imioverride:      imiOverride,
			Category:         category,
			Family:           family,
			Description:      description,
			IksDB:            true,
		}
		instanceTypeResponse = append(instanceTypeResponse, instanceType)
	}

	return &pb.GetInstanceTypesResponse{
		InstanceTypeResponse: instanceTypeResponse,
	}, nil
}

func GetInstanceType(ctx context.Context, dbconn *sql.DB, record *pb.GetInstanceTypeRequest) (*pb.InstanceTypeResponse, error) {
	friendlyMessage := "GetInstanceType.UnexpectedError"
	failedFunction := "GetInstanceType."
	returnError := &pb.InstanceTypeResponse{}

	// GET INSTANCETYPE
	err := dbconn.QueryRowContext(ctx, GetInstanceTypeQuery, record.Name).
		Scan(&instanceTypeName, &memory, &cpu, &nodeProviderName, &storage, &status, &displayName, &imiOverride, &category, &family, &description)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, grpc_status.Errorf(codes.NotFound, "Instance Type not found: %s", record.Name)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetInstanceTypeQuery", friendlyMessage+err.Error())
	}

	// GET IMIS By Family and Category
	var imisResponse []*pb.IMIResponse
	var imisK8sCompatibilityResponse []*pb.IMIResponse
	rows, err := dbconn.QueryContext(ctx, GetIMIQueryByFamilyAndCategory, category, family)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetIMIQueryByFamilyAndCategory", friendlyMessage+err.Error())
	}
	defer rows.Close()

	// PARSE ROWSE
	for rows.Next() {
		err := rows.Scan(&imiName, &imiComponents, &imiOs, &imiRuntime, &imiType, &imiProvider, &imiUpstreamReleaseName, &imiState, &imiArtifactRepo, &imiCategory, &imiFamily)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetIMIsQuery.rows.scan", friendlyMessage+err.Error())
		}
		if imiCategory == category && imiFamily == family {
			var components []*pb.Component
			err = json.Unmarshal([]byte(imiComponents), &components)
			if err != nil {
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"unmarshal.imiComponents", friendlyMessage+err.Error())
			}

			//Validate if K8s is Active in K8sVersion Table
			var k8sversion string
			isactivek8s := true
			err = dbconn.QueryRowContext(ctx, k8sExistanceAndActiveQuery, imiUpstreamReleaseName).Scan(&k8sversion)
			if err != nil {
				if err == sql.ErrNoRows {
					// Get Active Cluster Count for Archived K8s Versions
					activeClusterWithArchivedK8sClusterCount := 0
					activeClusterErr := dbconn.QueryRowContext(ctx, getActiveClustersQuery, imiUpstreamReleaseName).Scan(&activeClusterWithArchivedK8sClusterCount)
					if activeClusterErr != nil {
						return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getActiveClustersQuery", friendlyMessage+err.Error())
					}
					if activeClusterWithArchivedK8sClusterCount > 0 {
						isactivek8s = true
					} else {
						isactivek8s = false
					}
				} else {
					return returnError, utils.ErrorHandler(ctx, err, failedFunction+"k8sExistanceAndActiveQuery", friendlyMessage+err.Error())
				}
			}

			// Get all Control Plane Images associated for the k8sversion
			var controlPlaneOsImageInstances []string
			if imiType == imiWorkerType {
				cp_rows, err := dbconn.QueryContext(ctx, GetControlPlaneOsimageInstanceByK8sVersionQuery, imiUpstreamReleaseName, imiControlPlaneType)
				if err != nil {
					return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetCPOsimageImageInstanceByK8sVersionQuery", friendlyMessage+err.Error())
				}
				defer cp_rows.Close()
				// PARSE ROWSE
				for cp_rows.Next() {
					var osimageinstancename string
					err := cp_rows.Scan(&osimageinstancename)
					if err != nil {
						return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetControlPlaneOsimageImageInstanceByK8sVersionQuery.rows.scan", friendlyMessage+err.Error())
					}
					controlPlaneOsImageInstances = append(controlPlaneOsImageInstances, osimageinstancename)
				}
			}

			imi := &pb.IMIResponse{
				Name:                imiName,
				Upstreamreleasename: imiUpstreamReleaseName,
				Provider:            imiProvider,
				Type:                imiType,
				Runtime:             imiRuntime,
				Os:                  imiOs,
				State:               imiState,
				Components:          components,
				Artifact:            imiArtifactRepo,
				Category:            imiCategory,
				Family:              imiFamily,
				Isk8SActive:         isactivek8s,
				Cposimageinstances:  controlPlaneOsImageInstances,
			}

			// Get IMIs that are not associated to instance types in k8scompability
			if imiType == imiWorkerType {
				workerosimageinstance := ""
				err := dbconn.QueryRowContext(ctx, GetOsimageInstanceK8sCompatibilityQueryByInstanceName, imiRuntime, imiUpstreamReleaseName, imiOs, imiProvider, instanceTypeName).Scan(&workerosimageinstance)
				if err != nil {
					if err == sql.ErrNoRows {
						if isactivek8s {
							imi.Iscompatabilityactiveimi = false
							imisK8sCompatibilityResponse = append(imisK8sCompatibilityResponse, imi)
						}
					} else {
						return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetOsimageInstanceK8sCompatibilityQuery", friendlyMessage+err.Error())
					}
				}
				if workerosimageinstance != "" && workerosimageinstance == imiName {
					imi.Iscompatabilityactiveimi = true
				}
			}
			imisResponse = append(imisResponse, imi)
		}
	}

	return &pb.InstanceTypeResponse{
		Instancetypename:                       instanceTypeName,
		Memory:                                 memory,
		Cpu:                                    cpu,
		Nodeprovidername:                       nodeProviderName,
		Storage:                                storage,
		Status:                                 status,
		Displayname:                            displayName,
		Imioverride:                            imiOverride,
		Category:                               category,
		Family:                                 family,
		IksDB:                                  true,
		Description:                            description,
		ImiResponse:                            imisResponse,
		Instacetypeimik8Scompatibilityresponse: imisK8sCompatibilityResponse,
	}, nil
}

func GetIMIsInfo(ctx context.Context, dbconn *sql.DB) (*pb.GetIMIsInfoResponse, error) {
	friendlyMessage := "GetIMIsInfo.UnexpectedError"
	failedFunction := "GetIMIsInfo."
	returnError := &pb.GetIMIsInfoResponse{}

	var imisInfo *pb.GetIMIsInfoResponse
	rows, err := dbconn.QueryContext(ctx, GetIMIsInfoQuery)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetIMIsInfoQuery", friendlyMessage+err.Error())
	}
	defer rows.Close()

	// PARSE ROWSE
	for rows.Next() {
		err := rows.Scan(&imiOs, &imiProvider, &imiRuntime, &imiState)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetIMIsInfoQuery.rows.scan", friendlyMessage+err.Error())
		}
		var osimages []*pb.OSImages
		err = json.Unmarshal([]byte(imiOs), &osimages)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"unmarshal.imiOs", friendlyMessage+err.Error())
		}
		var providers []*pb.Providers
		err = json.Unmarshal([]byte(imiProvider), &providers)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"unmarshal.imiProvider", friendlyMessage+err.Error())
		}
		var runtimes []*pb.Runtimes
		err = json.Unmarshal([]byte(imiRuntime), &runtimes)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"unmarshal.imiRuntime", friendlyMessage+err.Error())
		}
		var states []*pb.States
		err = json.Unmarshal([]byte(imiState), &states)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"unmarshal.imiState", friendlyMessage+err.Error())
		}
		imisInfo = &pb.GetIMIsInfoResponse{
			Runtime:  runtimes,
			Osimage:  osimages,
			Provider: providers,
			State:    states,
		}
	}
	return imisInfo, nil
}

func GetIMIs(ctx context.Context, dbconn *sql.DB) (*pb.GetIMIResponse, error) {
	friendlyMessage := "GetIMIs.UnexpectedError"
	failedFunction := "GetIMI."
	returnError := &pb.GetIMIResponse{}

	// GET IMIS
	var imis []*pb.IMIResponse
	rows, err := dbconn.QueryContext(ctx, GetIMIsQuery)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetIMIsQuery", friendlyMessage+err.Error())
	}
	defer rows.Close()

	// PARSE ROWSE
	for rows.Next() {
		err := rows.Scan(&imiName, &imiOs, &imiRuntime, &imiType, &imiProvider, &imiUpstreamReleaseName, &imiState, &imiArtifactRepo, &imiCategory, &imiFamily, &isactiveimi)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetIMIsQuery.rows.scan", friendlyMessage+err.Error())
		}

		// Get IMI Components Data
		var components []*pb.Component
		imiComponentsRows, err := dbconn.QueryContext(ctx, GetIMIComponentsQuery, imiName)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetIMIsQuery", friendlyMessage+err.Error())
		}
		defer imiComponentsRows.Close()

		for imiComponentsRows.Next() {
			err := imiComponentsRows.Scan(&imiComponentName, &imiComponentVersion, &imiComponentArtifactRepo)
			if err != nil {
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetIMIsQuery.rows.scan", friendlyMessage+err.Error())
			}
			imiComponent := pb.Component{
				Name:     imiComponentName,
				Version:  imiComponentVersion,
				Artifact: imiComponentArtifactRepo,
			}
			components = append(components, &imiComponent)
		}

		//Validate if K8s is Present in K8sVersion Table
		isactivek8s := true
		var k8sversion string
		err = dbconn.QueryRowContext(ctx, k8sExistanceQuery, imiUpstreamReleaseName).Scan(&k8sversion)
		if err != nil {
			if err == sql.ErrNoRows {
				isactivek8s = false
			} else {
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetIMIQuery", friendlyMessage+err.Error())
			}
		}

		imi := &pb.IMIResponse{
			Name:                     imiName,
			Upstreamreleasename:      imiUpstreamReleaseName,
			Provider:                 imiProvider,
			Type:                     imiType,
			Runtime:                  imiRuntime,
			Os:                       imiOs,
			State:                    imiState,
			Components:               components,
			Artifact:                 imiArtifactRepo,
			Category:                 imiCategory,
			Family:                   imiFamily,
			Iscompatabilityactiveimi: isactiveimi,
			Isk8SActive:              isactivek8s,
		}

		imis = append(imis, imi)
	}

	return &pb.GetIMIResponse{
		Imiresponse: imis,
	}, nil
}

func GetIMI(ctx context.Context, dbconn *sql.DB, record *pb.GetIMIRequest) (*pb.IMIResponse, error) {
	friendlyMessage := "GetIMI.UnexpectedError"
	failedFunction := "GetIMI."
	returnError := &pb.IMIResponse{}

	// Get IMI Data
	err := dbconn.QueryRowContext(ctx, GetIMIQuery, record.Name).Scan(&imiOs, &imiRuntime, &imiType, &imiProvider, &imiUpstreamReleaseName, &imiState, &imiArtifactRepo, &imiCategory, &imiFamily, &isactiveimi)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, grpc_status.Errorf(codes.NotFound, "IMI not found: %s", record.Name)
		}
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetIMIQuery", friendlyMessage+err.Error())
	}

	// Get IMI Components Data
	var components []*pb.Component
	imiComponentsRows, err := dbconn.QueryContext(ctx, GetIMIComponentsQuery, record.Name)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetIMIsQuery", friendlyMessage+err.Error())
	}
	defer imiComponentsRows.Close()

	for imiComponentsRows.Next() {
		err := imiComponentsRows.Scan(&imiComponentName, &imiComponentVersion, &imiComponentArtifactRepo)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetIMIsQuery.rows.scan", friendlyMessage+err.Error())
		}
		imiComponent := pb.Component{
			Name:     imiComponentName,
			Version:  imiComponentVersion,
			Artifact: imiComponentArtifactRepo,
		}
		components = append(components, &imiComponent)
	}

	//Validate if K8s is Present in K8sVersion Table
	isactivek8s := true
	var k8sversion string
	err = dbconn.QueryRowContext(ctx, k8sExistanceQuery, imiUpstreamReleaseName).Scan(&k8sversion)
	if err != nil {
		if err == sql.ErrNoRows {
			isactivek8s = false
		} else {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetIMIQuery", friendlyMessage+err.Error())
		}
	}

	// Get All Instance Types associated by imi Name, Category and k8scompatibility
	var instanceTypeResponse []*pb.InstanceTypeResponse
	var instanceTypek8scompatibilityResponse []*pb.InstanceTypeResponse
	rows, err := dbconn.QueryContext(ctx, GetActiveInstanceTypesQuery)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetInstanceTypesQuery", friendlyMessage+err.Error())
	}
	defer rows.Close()

	// PARSE ROWS
	for rows.Next() {
		err := rows.Scan(&instanceTypeName, &memory, &cpu, &nodeProviderName, &storage, &status, &displayName, &imiOverride, &category, &family, &description)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetInstanceTypeByNameAndCategoryQuery.rows.scan", friendlyMessage+err.Error())
		}
		if imiCategory == category && imiFamily == family {
			instanceType := &pb.InstanceTypeResponse{
				Instancetypename: instanceTypeName,
				Nodeprovidername: nodeProviderName,
				Status:           status,
				Displayname:      displayName,
				Imioverride:      imiOverride,
				Category:         category,
				Family:           family,
				Description:      description,
				IksDB:            true,
			}

			// Get Instance Types that are not associated to worker osimageinstance in k8scompability
			instanceType.Iscompatabilityactiveinstance = true
			if imiType == imiWorkerType && isactivek8s {
				workerosimageinstance := ""
				err := dbconn.QueryRowContext(ctx, GetOsimageInstanceK8sCompatibilityQueryByInstanceName, imiRuntime, imiUpstreamReleaseName, imiOs, imiProvider, instanceTypeName).Scan(&workerosimageinstance)
				if err != nil {
					if err == sql.ErrNoRows {
						instanceType.Iscompatabilityactiveinstance = false
					} else {
						return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetOsimageInstanceK8sCompatibilityQuery", friendlyMessage+err.Error())
					}
				}
				if workerosimageinstance != "" && workerosimageinstance != record.Name {
					instanceType.Iscompatabilityactiveinstance = false
					instanceTypek8scompatibilityResponse = append(instanceTypek8scompatibilityResponse, instanceType)
				}
			}
			instanceTypeResponse = append(instanceTypeResponse, instanceType)
		}
	}

	// Get all Control Plane Images associated for the k8sversion
	var controlPlaneOsImageInstances []string
	if imiType == imiWorkerType && isactivek8s {
		rows, err = dbconn.QueryContext(ctx, GetControlPlaneOsimageInstanceByK8sVersionQuery, imiUpstreamReleaseName, imiControlPlaneType)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetCPOsimageImageInstanceByK8sVersionQuery", friendlyMessage+err.Error())
		}
		defer rows.Close()
		// PARSE ROWSE
		for rows.Next() {
			var osimageinstancename string
			err := rows.Scan(&osimageinstancename)
			if err != nil {
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetControlPlaneOsimageImageInstanceByK8sVersionQuery.rows.scan", friendlyMessage+err.Error())
			}
			controlPlaneOsImageInstances = append(controlPlaneOsImageInstances, osimageinstancename)
		}
	}

	return &pb.IMIResponse{
		Name:                                   record.Name,
		Upstreamreleasename:                    imiUpstreamReleaseName, // update this to the k8sversion name
		Provider:                               imiProvider,
		Type:                                   imiType,
		Runtime:                                imiRuntime,
		Os:                                     imiOs,
		State:                                  imiState,
		Components:                             components, // chnage this to components
		Artifact:                               imiArtifactRepo,
		Category:                               imiCategory,
		Family:                                 imiFamily,
		Iscompatabilityactiveimi:               isactiveimi,
		InstanceTypeResponse:                   instanceTypeResponse,
		Instacetypeimik8Scompatibilityresponse: instanceTypek8scompatibilityResponse,
		Cposimageinstances:                     controlPlaneOsImageInstances,
		Isk8SActive:                            isactivek8s,
	}, nil
}

func GetCluster(ctx context.Context, dbconn *sql.DB, record *pb.AdminClusterID) (*pb.GetClusterAdmin, error) {
	friendlyMessage := "GetCluster.UnexpectedError"
	failedFunction := "GetCluster."
	returnError := &pb.GetClusterAdmin{}
	returnValue := &pb.GetClusterAdmin{
		Network: &pb.NetworkAdmin{},
	}

	/* VALIDATE CLUSTER EXISTANCE */
	var clusterId int32
	clusterId, err := utils.ValidateClusterExistance(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterExistance", friendlyMessage+err.Error())
	}
	if clusterId == -1 {
		return returnError, grpc_status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid)
	}

	/* GET CLUSTER INFO FROM DB*/
	var cpName string
	var nodeCreateDtm sql.NullTime
	err = dbconn.QueryRowContext(ctx, getClusterQuery,
		record.Clusteruuid,
	).Scan(
		&returnValue.Name,
		&returnValue.Provider,
		&returnValue.Account,
		&returnValue.Region,
		&returnValue.K8Sversion,
		&cpName,
		&nodeCreateDtm,
		&returnValue.Network.Servicecidr,
		&returnValue.Network.Clustercidr,
		&returnValue.Clustertype,
	)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getClusterQuery", friendlyMessage+err.Error())
	}
	if nodeCreateDtm.Valid {
		returnValue.Certsexpiring = append(returnValue.Certsexpiring, &pb.ClusterCerts{
			Cpname:         cpName,
			Certexpirydate: nodeCreateDtm.Time.AddDate(1, 0, 0).String(),
		})
	}
	/* GET AVAILABLE UPGRADES FOR CLUSTER */
	availableK8sUpgrades, err := utils.GetAvailableClusterVersionUpgrades(ctx, dbconn, record.Clusteruuid)
	returnValue.K8Supgradeavailable = len(availableK8sUpgrades) > 0
	returnValue.K8Supgradeversions = availableK8sUpgrades

	/* GET CLUSTER STORAGE */
	var getClustersStorageStatus []*pb.ClusterStorageStatus
	getClustersStorageStatus, err = utils.GetClusterStorageStatus(ctx, dbconn, clusterId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"utils.GetClusterStorageStatus", friendlyMessage)
	}
	returnValue.Storages = getClustersStorageStatus

	/* GET NODEGROUPS */
	rows, err := dbconn.QueryContext(ctx, getClusterNodeGroupsQuery, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getClusterNodeGroupsQuery", friendlyMessage+err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		// Get nodegroup info
		nodeGroup := &pb.Nodegroup{}
		var nodegroupTypeName, nodeGroupUuid string
		var sshkeys []byte
		err = rows.Scan(
			&nodeGroupUuid,
			&nodeGroup.Id,
			&nodeGroup.Name,
			&nodeGroup.Count,
			&nodeGroup.Releaseversion,
			&nodeGroup.Status,
			&nodeGroup.Imi,
			&nodeGroup.Instancetype,
			&sshkeys,
			&nodegroupTypeName,
			&nodeGroup.Nodegrouptype,
		)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getClusterNodeGroupsQuery.rows.scan", friendlyMessage+err.Error())
		}

		err = json.Unmarshal(sshkeys, &nodeGroup.Sshkey)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getSshKey.unmarshal", friendlyMessage+err.Error())
		}
		/* GET NODES FOR NODEGROUP*/
		summary := &pb.NodegroupSummary{}
		nodes, err := dbconn.QueryContext(ctx, getNodeGroupNodesQuery, nodeGroupUuid)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getNodeGroupNodesQuery", friendlyMessage+err.Error())
		}
		defer nodes.Close()
		for nodes.Next() {
			// Scan into Node return
			node := &pb.Node{
				WekaStorage: &pb.WekaStorageStatusAdmin{},
			}
			var wekaStorageClientId string
			var wekaStorageStatus string
			var wekaStorageCustomStatus string
			var wekaStorageMessage string
			err = nodes.Scan(&node.Name, &node.Ipaddress, &node.Dnsname, &node.Imi, &node.State, &node.Status, &node.Createddate, &wekaStorageClientId, &wekaStorageStatus, &wekaStorageCustomStatus, &wekaStorageMessage)
			if err != nil {
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getNodeGroupNodesQuery.nodes.scan", friendlyMessage+err.Error())
			}

			node.WekaStorage.ClientId = wekaStorageClientId
			node.WekaStorage.Status = wekaStorageStatus
			node.WekaStorage.CustomStatus = wekaStorageCustomStatus
			node.WekaStorage.Message = wekaStorageMessage

			// Fetching the Node Summary Status
			switch node.State {
			case "Active":
				summary.Activenodes++
			case "Creating", "Updating":
				summary.Provisioningnodes++
			case "Error":
				summary.Errornodes++
			case "Deleting":
				summary.Deletingnodes++
			default:
				summary.Errornodes++
			}

			nodeGroup.Nodes = append(nodeGroup.Nodes, node)
		}

		// Get Available Upgrade for nodegroup
		var availableImiUpgrades []string
		if nodegroupTypeName == cpNodegroupType {
			_, availableImiUpgrades, err = utils.GetAvailableControlPlaneImiUpgrades(ctx, dbconn, record.Clusteruuid, nodeGroupUuid)
		} else if nodegroupTypeName == wkNodegroupType {
			_, availableImiUpgrades, err = utils.GetAvailableWorkerImiUpgrades(ctx, dbconn, record.Clusteruuid, nodeGroupUuid)
		}
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetAvailableXXXImIUpgrades."+nodegroupTypeName, friendlyMessage+err.Error())
		}
		nodeGroup.Imiupgradeavailable = len(availableImiUpgrades) > 0
		nodeGroup.Imiupgradeversions = append(nodeGroup.Imiupgradeversions, availableImiUpgrades...)
		nodeGroup.Nodegroupsummary = summary
		nodeGroup.Nodegrouptypename = nodegroupTypeName
		nodeGroup.Nodgroupuuid = nodeGroupUuid

		// Append to response
		returnValue.Nodegroups = append(returnValue.Nodegroups, nodeGroup)
	}

	/* GET ADDONS */
	rows, err = dbconn.QueryContext(ctx, getClusterAddonsQuery, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getClusterAddonsQuery", friendlyMessage+err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		// Get Addon info
		addon := pb.AddOnAdmin{}
		var args, tags sql.NullString
		var artifacts sql.NullString
		err = rows.Scan(
			&addon.Name,
			&addon.Version,
			&addon.State,
			&args,
			&tags,
			&artifacts,
		)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getClusterAddonsQuery.rows.scan", friendlyMessage+err.Error())
		}

		// Get Args
		if args.Valid {
			var jsonArr []*pb.Keyvaluepair
			err = json.Unmarshal([]byte(args.String), &jsonArr)
			if err != nil {
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"unmarshal.args.String", friendlyMessage+err.Error())
			}
			addon.Args = jsonArr
		}

		// Get Tags
		if tags.Valid {
			var jsonArr []*pb.Keyvaluepair
			err = json.Unmarshal([]byte(tags.String), &jsonArr)
			if err != nil {
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"unmarshal.tags.String", friendlyMessage+err.Error())
			}
			addon.Tags = jsonArr
		}

		// Append to response
		returnValue.Addons = append(returnError.Addons, &addon)
	}

	/* GET SNAPSHOTS */
	rows, err = dbconn.QueryContext(ctx, getClusterSnapshotsQuery, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getClsuterSnapshotsQuery", friendlyMessage+err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		// Get snapshots info
		snapshot := pb.Snapshot{}
		err = rows.Scan(
			&snapshot.Name,
			&snapshot.Type,
			&snapshot.State,
			&snapshot.Created,
			&snapshot.Filename,
		)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getClsuterSnapshotsQuery.rows.scan", friendlyMessage+err.Error())
		}

		returnValue.Snapshot = append(returnValue.Snapshot, &snapshot)
	}
	return returnValue, nil
}

func GetClusters(ctx context.Context, dbconn *sql.DB) (*pb.GetClustersAdmin, error) {
	friendlyMessage := "GetClusters.UnexpectedError"
	failedFunction := "GetClusters."
	returnError := &pb.GetClustersAdmin{}
	returnValue := &pb.GetClustersAdmin{}

	/* Get clusters from DB */
	rows, err := dbconn.QueryContext(ctx, getClustersQuery)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getClustersQuery", friendlyMessage+err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		// Scan each entry
		cluster := pb.ClustersResponseAdmin{}
		err = rows.Scan(
			&cluster.Uuid,
			&cluster.Name,
			&cluster.State,
			&cluster.Provider,
			&cluster.Account,
			&cluster.K8Sversion,
			&cluster.Createddate,
			&cluster.Clustertype,
		)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getClustersQuery.rows.scan", friendlyMessage+err.Error())
		}

		/* CHECK AVAILABLE CLUSTER MINOR VERSIONS */
		availableVersions, err := utils.GetAvailableClusterVersionUpgrades(ctx, dbconn, cluster.Uuid)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getClusterAvailableVersions", friendlyMessage+err.Error())
		}
		cluster.K8Supgradeavailable = len(availableVersions) > 0
		cluster.K8Supgradeversions = availableVersions

		/* GET CONTROL PLAN UPGRADES */
		// Get control plane UUID
		var cpUuid string
		err = dbconn.QueryRowContext(ctx, getClusterControlPlaneUuidQuery, cluster.Uuid).Scan(&cpUuid)
		// Get Available CP upgrades
		_, availableVersionsImi, err := utils.GetAvailableControlPlaneImiUpgrades(ctx, dbconn, cluster.Uuid, cpUuid)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetAvailableControlPlaneImiUpgrades", friendlyMessage+err.Error())
		}
		cluster.Cpupgradeavailable = len(availableVersionsImi) > 0
		cluster.Cpupgradeversions = availableVersionsImi

		returnValue.Response = append(returnValue.Response, &cluster)
	}

	return returnValue, nil
}

func getClusterAvailableVersions(cluserProvider string, currentVersion string, clusterActiveVersions []string) ([]string, error) {
	var availableVersions []string

	pCurrentVersion, err := parseVersions(currentVersion)
	if err != nil {
		return make([]string, 0), err
	}

	pClusterActiveVersions, err := parseVersions(clusterActiveVersions...)
	if err != nil {
		return make([]string, 0), err
	}
	semver.Sort(pClusterActiveVersions)

	if cluserProvider == iksProvider {
		for _, p := range pClusterActiveVersions {
			if pCurrentVersion[0].Minor < p.Minor {
				availableVersions = append(availableVersions, fmt.Sprintf("%d.%d", p.Major, p.Minor))
				break
			}
		}
	} else if cluserProvider == rke2Provider {
		for _, p := range pClusterActiveVersions {
			if pCurrentVersion[0].Minor == p.Minor && pCurrentVersion[0].Patch < p.Patch {
				availableVersions = append(availableVersions, p.String())
			} else if pCurrentVersion[0].Minor < p.Minor {
				availableVersions = append(availableVersions, p.String())
				break
			}
		}
	}

	return availableVersions, nil
}

func GetK8SVersions(ctx context.Context, dbconn *sql.DB) (*pb.GetK8SVersionResponse, error) {
	friendlyMessage := "GetK8SVersions.UnexpectedError"
	failedFunction := "GetK8sVersions."
	returnError := &pb.GetK8SVersionResponse{}
	returnValue := &pb.GetK8SVersionResponse{}

	/* Get available k8s versions */
	rows, err := dbconn.QueryContext(ctx, getK8sVersionsQuery)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getK8sVersionQuery", friendlyMessage+err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		version := pb.K8SversionResponse{}
		err = rows.Scan(
			&version.Cpimi,
			&version.Name,
			&version.Provider,
			&version.Releasename,
			&version.Workimi,
			&version.Runtime,
			&version.State,
			&version.Major,
			&version.Minor,
		)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getK8sVersionQuery.scan", friendlyMessage+err.Error())
		}

		returnValue.K8Sversions = append(returnValue.K8Sversions, &version)
	}

	return returnValue, nil
}

func GetK8SVersion(ctx context.Context, dbconn *sql.DB, req *pb.GetK8SRequest) (*pb.K8SversionResponse, error) {
	friendlyMessage := "GetK8SVersion.UnexpectedError"
	failedFunction := "GetK8sVersion."
	returnError := &pb.K8SversionResponse{}
	returnValue := &pb.K8SversionResponse{}

	/* Get available k8s versions */
	err := dbconn.QueryRowContext(ctx, getK8sVersionQuery, req.Name).Scan(
		&returnValue.Cpimi,
		&returnValue.Name,
		&returnValue.Provider,
		&returnValue.Releasename,
		&returnValue.Workimi,
		&returnValue.Runtime,
		&returnValue.State,
		&returnValue.Major,
		&returnValue.Minor,
	)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getK8sVersionQuery", friendlyMessage+err.Error())
	}

	return returnValue, nil
}

/* Converts semver defined version strings into Version structs */
func parseVersions(versions ...string) ([]semver.Version, error) {
	var parsedVersions []semver.Version

	for _, v := range versions {
		parsedVer, err := semver.ParseTolerant(v)
		if err != nil {
			return nil, err
		}

		parsedVersions = append(parsedVersions, parsedVer)
	}

	return parsedVersions, nil
}

func GetControlPlaneSSHKeys(ctx context.Context, dbconn *sql.DB, record *pb.AdminClusterID, filepath string) (*pb.ClusterSSHKeys, error) {
	// Start the transaction
	friendlyMessage := "GetControlPlaneSSHKeys.UnexpectedError"
	failedFunction := "GetControlPlaneSSH."
	returnError := &pb.ClusterSSHKeys{}

	/* VALIDATE CLUSTER EXISTANCE */
	var clusterId int32
	clusterId, err := utils.ValidateClusterExistance(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterExistance", friendlyMessage+err.Error())
	}
	if clusterId == -1 {
		return returnError, grpc_status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid)
	}

	/* GET SECRETS FROM DB */
	var sshprivatekey, sshpubkey, nonce string
	var encryptionKeyId int32
	err = dbconn.QueryRowContext(ctx, GetSshkeys, clusterId).Scan(&sshprivatekey, &sshpubkey, &encryptionKeyId, &nonce)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetClusterSecrets", friendlyMessage)
	}
	decodedNonce, err := utils.Base64DecodeString(ctx, nonce)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"DecodeNonce", friendlyMessage)
	}

	/* DECODE SSH Keys */
	encryptionKeyBytes, err := utils.GetSpecificEncryptionKey(ctx, filepath, encryptionKeyId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetEncryptionKeys", friendlyMessage)
	}
	sshprivateEnc, err := utils.AesDecryptSecret(ctx, sshprivatekey, encryptionKeyBytes, decodedNonce)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"AesDecryption.SshPrivateKey", friendlyMessage)
	}
	sshpubEnc, err := utils.AesDecryptSecret(ctx, sshpubkey, encryptionKeyBytes, decodedNonce)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"AesDecryption.SshPublicKey", friendlyMessage)
	}
	return &pb.ClusterSSHKeys{Clusterid: clusterId, Sshprivatekey: sshprivateEnc, Sshpublickey: sshpubEnc}, nil
}

func GetCloudAccountApproveList(ctx context.Context, dbconn *sql.DB) (*pb.CloudAccountApproveListResponse, error) {
	// Start the transaction
	friendlyMessage := "GetCloudAccountApproveList.UnexpectedError"
	failedFunction := "GetCloudAccountApproveList."
	returnError := &pb.CloudAccountApproveListResponse{}
	returnValue := &pb.CloudAccountApproveListResponse{}

	/* Get available cloud account approve list details */
	rows, err := dbconn.QueryContext(ctx, getCloudAccountsApproveListQuery)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getCloudAccountsApproveListQuery", friendlyMessage+err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		approvalList := pb.CloudAccountApproveList{}
		err = rows.Scan(
			&approvalList.Account,
			&approvalList.Providername,
			&approvalList.Status,
			&approvalList.EnableStorage,
			&approvalList.MaxclustersOverride,
			&approvalList.MaxclusterngOverride,
			&approvalList.MaxclusterilbOverride,
			&approvalList.MaxnodegroupvmOverride,
			&approvalList.MaxclustervmOverride,
		)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getCloudAccountsApproveListQuery.scan", friendlyMessage+err.Error())
		}

		returnValue.ApproveListResponse = append(returnValue.ApproveListResponse, &approvalList)
	}

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

	returnValue.Existingresourcelimits = limits

	return returnValue, nil
}

func GetLoadBalancers(ctx context.Context, dbconn *sql.DB, record *pb.AdminClusterID) (*pb.LoadBalancers, error) {

	friendlyMessage := "Could not get LoadBalancers. Please try again"
	failedFunction := "GetLoadBalancers."
	returnError := &pb.LoadBalancers{}
	returnValue := &pb.LoadBalancers{
		Lbresponses: []*pb.LoadbalancerResponse{},
	}
	/* VALIDATE CLUSTER EXISTANCE */
	var clusterId int32
	clusterId, err := utils.ValidateClusterExistance(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterExistance", friendlyMessage)
	}
	if clusterId == -1 {
		return returnError, grpc_status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid)
	}

	/* GET LB IDs */
	rows, err := dbconn.QueryContext(ctx, getlbIdsQuery, clusterId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getlbIdsQuery", friendlyMessage)
	}
	defer rows.Close()
	for rows.Next() {
		lbId := &pb.GetLbRequest{
			Clusteruuid: record.Clusteruuid,
			Lbid:        0,
		}
		err = rows.Scan(&lbId.Lbid)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getlbIdsQuery.rows.scan", friendlyMessage)
		}
		lbResponse, err := GetLoadBalancer(ctx, dbconn, lbId)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetLoadBalancer", friendlyMessage)
		}
		returnValue.Lbresponses = append(returnValue.Lbresponses, lbResponse)
	}
	return returnValue, nil
}

func GetLoadBalancer(ctx context.Context, dbconn *sql.DB, record *pb.GetLbRequest) (*pb.LoadbalancerResponse, error) {
	friendlyMessage := "Could not get LoadBalancer. Please try again"
	failedFunction := "GetLoadBalancer."

	returnError := &pb.LoadbalancerResponse{}
	returnValue := &pb.LoadbalancerResponse{
		Lbid: 0,
		Lb: &pb.Loadbalancer{
			Lbname:        "",
			Vip:           "",
			Status:        "",
			Backendports:  []string{},
			Frontendportd: []string{},
			Viptype:       "",
			Nodegrouptype: "",
			Createddate:   "",
		},
	}

	/* VALIDATE CLUSTER AND VIP EXISTANCE */
	clusterId, vipId, err := utils.ValidateVipExistance(ctx, dbconn, record.Clusteruuid, record.Lbid, false)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateVipExistance", friendlyMessage)
	}
	if clusterId == -1 || vipId == -1 {
		return returnError, grpc_status.Errorf(codes.NotFound, "Vip not found in Cluster: %s", record.Clusteruuid)
	}

	/* GET LoadBalancer */
	rows, err := dbconn.QueryContext(ctx, getlbIdQuery, clusterId, record.Lbid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getlbIdQuery", friendlyMessage)
	}
	defer rows.Close()
	for rows.Next() {
		var frontEndPort, backEndPort string
		err = rows.Scan(&returnValue.Lbid, &returnValue.Lb.Lbname, &returnValue.Lb.Status,
			&returnValue.Lb.Vip, &frontEndPort, &backEndPort,
			&returnValue.Lb.Viptype, &returnValue.Lb.Createddate,
		)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"getlbIdQuery.rows.query", friendlyMessage)
		}
		returnValue.Lb.Frontendportd = append(returnValue.Lb.Frontendportd, frontEndPort)
		returnValue.Lb.Backendports = append(returnValue.Lb.Backendports, backEndPort)
	}

	return returnValue, nil
}

func GetFirewallRule(ctx context.Context, dbconn *sql.DB, record *pb.AdminClusterID) (*pb.GetAdminFirewallRuleResponse, error) {
	friendlyMessage := "Could not get Security Rules. Please try again"
	failedFunction := "GetSecurityRule."

	returnError := &pb.GetAdminFirewallRuleResponse{}
	returnValue := &pb.GetAdminFirewallRuleResponse{
		Getfirewallresponse: []*pb.AdminFirewallRuleResponse{},
	}
	var clusterId int32
	clusterId, err := utils.ValidateClusterExistance(ctx, dbconn, record.Clusteruuid)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"ValidateClusterExistance", friendlyMessage)
	}
	if clusterId == -1 {
		return returnError, grpc_status.Errorf(codes.NotFound, "Cluster not found: %s", record.Clusteruuid)
	}

	var vip_id int32
	// Get vip id from clusteruuid
	rows, err := dbconn.QueryContext(ctx, query.GetVipidByClusterQuery, clusterId)
	if err != nil {
		return returnError, utils.ErrorHandler(ctx, err, failedFunction+"Getvipidbycluster", friendlyMessage)
	}

	defer rows.Close()

	for rows.Next() {
		var sourceips []byte
		var protocol []byte
		var sourceipsresult []string
		var protocolresult []string

		fwresponse := &pb.AdminFirewallRuleResponse{
			Destinationip: "",
			State:         "Not Specified",
			Sourceip:      []string{},
			Vipid:         0,
			Port:          0,
			Vipname:       "",
			Viptype:       "",
			Protocol:      []string{},
			Internalport:  0,
		}
		err = rows.Scan(&vip_id)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"scanvipidsfromClusterid.rows.scan", friendlyMessage)
		}
		err = dbconn.QueryRowContext(ctx, query.GetPublicVipsForFirewallQuery, clusterId, vip_id).Scan(&fwresponse.Vipid, &fwresponse.Destinationip, &fwresponse.Port, &fwresponse.Internalport, &sourceips, &fwresponse.State, &fwresponse.Viptype, &protocol, &fwresponse.Vipname)
		if err != nil {
			return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetPublicVipsForFirewallQuery", friendlyMessage)
		}
		if sourceips != nil {
			err = json.Unmarshal(sourceips, &sourceipsresult)
			if err != nil {
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetMetadata.unmarshall.souceip", friendlyMessage)
			}
			fwresponse.Sourceip = sourceipsresult
		} else {
			fwresponse.Sourceip = []string{}
		}
		if protocol != nil {
			err = json.Unmarshal(protocol, &protocolresult)
			if err != nil {
				return returnError, utils.ErrorHandler(ctx, err, failedFunction+"GetMetadata.unmarshall.protocol", friendlyMessage)
			}
			fwresponse.Protocol = protocolresult
		}
		returnValue.Getfirewallresponse = append(returnValue.Getfirewallresponse, fwresponse)
	}
	return returnValue, nil
}
