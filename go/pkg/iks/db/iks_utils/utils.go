// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package iks_utils

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/google/uuid"
	db_query "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/db_query_constants"
	clusterv1alpha "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"golang.org/x/crypto/sha3"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	superComputeClusterType     = "supercompute"
	superComputeAINodegroupType = "supercompute-ai"
	superComputeGPNodegroupType = "supercompute-gp"
	superComputeConstant        = "sc"
	ClusterUUIDType             = "Cluster"
	NodegroupUUIDType           = "Nodegroup"
)

const (
	GetClusterStorageEnableQuery = `
		SELECT storage_enable
		FROM cluster
		WHERE unique_id=$1
	`
	GetClusterStorageStatusQuery = `
		SELECT storagestate_name, size, storageprovider_name, kubernetes_status
		FROM storage
		WHERE  cluster_id=$1
	`
	GetAddonsQuery = `
		select av.name, av.install_type, av.artifact_repo from  addoncompatibilityk8s ac
			inner join addonversion av
			on av.addonversion_name = ac.addonversion_name
		where av.admin_only=$1 and av.onbuild = $2 and k8sversion_name=$3 and addonversion_type=$4;
	`

	GetClusterUuidCheck = `
		SELECT count(cluster_id) as count
		FROM public.cluster c
		WHERE c.unique_id like $1
	`
	GetNodeGroupUuidCheck = `
		SELECT count(nodegroup_id) as count
		FROM public.nodegroup ng
		WHERE ng.unique_id like $1
	`
	GetDefaultConfigUtilsQuery = `
		SELECT value
		FROM  defaultconfig
		WHERE name = $1
	`
	GetCloudAccountCreate = `
		SELECT coalesce(
				(SELECT active_account_create_cluster
				FROM public.cloudaccountextraspec c
				WHERE cloudaccount_id = $1), false
		)as active_account_create_cluster
		`
	GetCloudAccountStorage = `
	SELECT coalesce(
		(SELECT allow_create_storage
		FROM public.cloudaccountextraspec c
		WHERE cloudaccount_id = $1), false
		)as allow_create_storage
        `
	getClusterCompatDetails = `
		SELECT c.provider_name, n.k8sversion_name, n.runtime_name, n.instancetype_name
		FROM  public.cluster c
			INNER JOIN nodegroup n
			ON c.cluster_id = n.cluster_id AND n.nodegrouptype_name = 'ControlPlane'
		WHERE c.unique_id = $1
	`
	getNodeGroupCompatDetails = `
		SELECT c.provider_name, n.k8sversion_name, n.runtime_name,  n.osimageinstance_name, n.nodegrouptype_name, n.instancetype_name
		FROM  public.nodegroup n
			INNER JOIN cluster c
			ON c.cluster_id = n.cluster_id
		WHERE n.unique_id = $1
	`
	GetNodeGroupStateQuery = `
		SELECT n.nodegroupstate_name
		FROM  public.nodegroup n
		WHERE n.unique_id = $1
	`
	getActiveK8sVersionAndImi = `
		SELECT
			ks.k8sversion_name,
			CASE WHEN $1 = 'ControlPlane'
				THEN kc.cp_osimageinstance_name
				ELSE kc.wrk_osimageinstance_name
			END
		FROM k8sversion AS ks
		INNER JOIN k8scompatibility AS kc ON kc.k8sversion_name = ks.k8sversion_name
		WHERE kc.provider_name = $2 AND kc.runtime_name = $3 AND kc.instancetype_name=$4 AND kc.osimage_name=$5
		AND ks.lifecyclestate_id = (SELECT l.lifecyclestate_id FROM lifecyclestate l WHERE l.name = 'Active')
	`
	getK8sVersionPatchVersions = `
		SELECT
			ks.k8sversion_name,
			CASE WHEN $1 = 'ControlPlane'
				THEN kc.cp_osimageinstance_name
				ELSE kc.wrk_osimageinstance_name
			END
		FROM k8sversion AS ks
			INNER JOIN k8scompatibility kc ON kc.k8sversion_name = ks.k8sversion_name
		WHERE kc.provider_name = $2 AND kc.runtime_name = $3 AND kc.instancetype_name=$5 AND kc.osimage_name=$6
		AND ks.minor_version = (SELECT minor_version from public.k8sversion WHERE k8sversion_name = $4)
		AND ks.lifecyclestate_id = (SELECT lifecyclestate_id from public.lifecyclestate WHERE name = 'Active')
	`
	getImiFromK8sVersion = `
		SELECT
			CASE WHEN $1 = 'ControlPlane'
				THEN kc.cp_osimageinstance_name
				ELSE kc.wrk_osimageinstance_name
			END
		FROM k8scompatibility AS kc
		WHERE kc.provider_name = $2 AND kc.runtime_name = $3 AND kc.k8sversion_name = $4
			AND kc.instancetype_name =$5 AND kc.osimage_name=$6
	`
	GetDefaultsIKSQuery = `
    SELECT to_jsonb(array_agg(t)) as defaultvalues FROM public.defaultconfig t
	`

	GetAddonsfork8sVersionsQuery = `
		SELECT addonversion_name FROM public.addoncompatibilityk8s where k8sversion_name = $1
	`

	GetDefaultAddonsQuery = `
		SELECT t.name, t.install_type, t.artifact_repo
		FROM public.addonversion t
		WHERE addonversion_name = $1 AND admin_only = 'true' AND onbuild='true'
	`

	getMinorVersionQuery = `
		SELECT ks.minor_version
		FROM public.k8sversion ks
		WHERE k8sversion_name = $1
	`
	GetImiArtifactQuery = `
		SELECT imiartifact
		FROM public.osimageinstance
		WHERE osimageinstance_name = $1
	`
	ClusterExistanceQuery = `
		SELECT cluster_id
		FROM public.cluster
		WHERE unique_id = $1 AND clusterstate_name != 'Deleted'
	`
	NodeGroupExistanceQuery = `
		SELECT n.cluster_id, n.nodegroup_id
	 	FROM public.nodegroup n
		WHERE n.unique_id = $2 AND n.cluster_id = (
			SELECT c.cluster_id
			FROM public.cluster  c
			WHERE c.unique_id = $1 AND c.clusterstate_name != 'Deleted'
		)
	`

	VipExistanceQuery = `
	  SELECT v.cluster_id, v.vip_id
	  FROM public.vip v
  	WHERE v.vip_id = $2 AND v.cluster_id = (
	    SELECT cluster_id
	    FROM public.cluster  c
	    WHERE c.unique_id = $1 AND c.clusterstate_name != 'Deleted'
		)
	`
	getClusterControlPlaneQuery = `
		SELECT n.unique_id
	 	FROM public.nodegroup n
		WHERE n.nodegrouptype_name = 'ControlPlane' AND n.cluster_id = (
			SELECT c.cluster_id
			FROM public.cluster  c
			WHERE c.unique_id = $1
		)
	`
	GetClusterCloudAccountQuery = `
	  SELECT cloudaccount_id
		FROM public.cluster
		WHERE unique_id = $1
	`
	GetClusterStateQuery = `
		SELECT clusterstate_name
		FROM public.cluster
		WHERE unique_id = $1
	`
	GetVipStateQuery = `
	   SELECT vipstate_name FROM public.vip
	   WHERE vip_id = $1
	   `

	GetClusterRevChangeAppliedQuery = `
		SELECT change_applied
		FROM public.clusterrev cr
			inner join public.cluster c
			on c.cluster_id = cr.cluster_id
		WHERE c.unique_id  = $1
		ORDER BY clusterrev_id desc
		LIMIT 1
	`
	GetLatestClusterRevQuery = `
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
	GetCloudAccountExtraSpecMaxValues = `
		SELECT coalesce(maxclusters_override, -1) as maxclusters_override,
			coalesce(maxclusterng_override, -1) as maxclusterng_override,
			coalesce(maxclusterilb_override, -1) as maxclusterilb_override,
			coalesce(maxclustervm_override, -1) as maxclustervm_override,
			coalesce(maxnodegroupvm_override, -1) as maxnodegroupvm_override
		FROM cloudaccountextraspec c
		WHERE cloudaccount_id = $1
	`

	GetDefaultLifeCycleStatesQuery = `
	SELECT name AS state FROM public.lifecyclestate
	`

	GetDefaultNodeProviderQuery = `
	SELECT nodeprovider_name AS nodeprovider FROM public.nodeprovider WHERE lifecyclestate_id = (SELECT lifecyclestate_id from public.lifecyclestate WHERE name = 'Active')
	`

	GetInstanceTypeQuery = `
	SELECT instancetype_name FROM public.instancetype
	WHERE instancetype_name = $1
	`

	GetSecStateQuery = `
	SELECT COALESCE(firewall_status,'Not Specified') FROM vip WHERE vip_id = $1
	`

	GetsourceipQueryCount = `
	SELECT count(sourceips) FROM vip WHERE vip_id = $1
	`

	GetStorageSizeQuery = `
		SELECT size
		FROM public.storage
		WHERE cluster_id=$1
	`

	GetCloudAccountStorageSizeQuery = `
		SELECT total_storage_size from public.cloudaccountextraspec WHERE cloudaccount_id = $1
	`
)

const (
	iksProvider      = "iks"
	rke2Provider     = "rke2"
	cpNodegroupType  = "ControlPlane"
	wkNodegroupType  = "Worker"
	cpLifecycleState = "Active"
)

type Defaultvalues struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type DefaultAddons struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Artifact string `json:"artifact"`
}

type EncryptionStruct struct {
	Data map[string]string `json:"data"`
}

type OperatorMessage struct {
	ErrorCode int32  `json:"errorCode"`
	Message   string `json:"message"`
}

type FileProductInfo struct {
	MinSize          int64
	MaxSize          int64
	UpdatedTimestamp time.Time
}

/********** PRIVATE FUNCTIONS **********/
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

/********** PUBLIC FUNCTIONS **********/

func ErrorHandler(ctx context.Context, err error, failedCall string, friendlyMessage string) error {
	/* SET UP THE IDC LOGGER */
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("ErrorHandler")
	/* PRINT THE DETAILED LOG */
	log.Error(err, failedCall)
	/* RETURN USER FRIENDLY ERROR */
	return errors.New(friendlyMessage)
}

// ErrorHandlerWithGrpcCode logs the error and returns a friendly error message with a grpc status code
// To be used when non-default status code is required in GRPC response
func ErrorHandlerWithGrpcCode(ctx context.Context, err error, failedCall string, friendlyMessage string, code codes.Code) error {
	/* SET UP THE IDC LOGGER */
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("ErrorHandler")
	/* PRINT THE DETAILED LOG */
	log.Error(err, failedCall)
	/* RETURN USER FRIENDLY ERROR */
	return status.Error(code, friendlyMessage)
}

func ParseOperatorMessage(ctx context.Context, operatorMessage []byte) (string, int32, error) {
	var operatorMessageStruct OperatorMessage
	err := json.Unmarshal(operatorMessage, &operatorMessageStruct)
	if err != nil {
		return string(operatorMessage), 0, nil
	}
	return operatorMessageStruct.Message, operatorMessageStruct.ErrorCode, nil
}

func Base64EncodeString(ctx context.Context, byteArr []byte) (string, error) {
	encodedByte := base64.StdEncoding.EncodeToString(byteArr)
	return encodedByte, nil
}
func Base64DecodeString(ctx context.Context, str string) ([]byte, error) {
	decodedByte, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return make([]byte, 0), err
	}
	return decodedByte, nil
}

func GetLatestEncryptionKey(ctx context.Context, filePath string) ([]byte, int, error) {
	body, err := ioutil.ReadFile(filePath)
	if err != nil {
		return make([]byte, 0), 1, nil
	}
	var content EncryptionStruct
	err = json.Unmarshal(body, &content)
	if err != nil {
		return make([]byte, 0), 1, err
	}
	key := 0
	value := ""
	for k, v := range content.Data {
		i, err := strconv.Atoi(k)
		if err != nil {
			return make([]byte, 0), 1, err
		}

		if i > key {
			key = i
			value = v
		}
	}
	if key == 0 || value == "" {
		return make([]byte, 0), 1, errors.New("Please check encryption key values")
	}

	return []byte(value), key, nil
}
func GetSpecificEncryptionKey(ctx context.Context, filePath string, encryptionKeyId int32) ([]byte, error) {
	body, err := ioutil.ReadFile(filePath)
	if err != nil {
		return make([]byte, 0), err
	}
	var content EncryptionStruct
	err = json.Unmarshal(body, &content)
	if err != nil {
		return make([]byte, 0), err
	}
	value := ""
	value = content.Data[fmt.Sprintf("%v", encryptionKeyId)]
	if value == "" {
		return make([]byte, 0), errors.New("Please check encryption key values")
	}

	return []byte(value), nil
}

func AesEncryptSecret(ctx context.Context, unencryptedValue string, key []byte, nonce []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintextBytes := []byte(unencryptedValue)
	ciphertext := aesgcm.Seal(nil, nonce, plaintextBytes, nil)
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	return encoded, nil
}

func AesDecryptSecret(ctx context.Context, encryptedValue string, key []byte, nonce []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedValue)
	if err != nil {
		return "", err
	}
	plaintextBytes, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintextBytes), err

}

func GenerateUuid(ctx context.Context, dbconn *sql.DB, UUIDType string) (string, error) {
	log := log.FromContext(ctx).WithName("GenerateUuid")

	var generatedUuid string
	var attempt = 0
	for attempt < 5 {
		log.V(0).Info("Generating UUID", logkeys.UUIDType, UUIDType, logkeys.Attempt, attempt)
		//  6 bytes long = 12-bit collision resistance.
		hash := make([]byte, 6)

		// Compute a 6-byte hash of buffer.
		sha3.ShakeSum256(hash, []byte(uuid.NewString()))

		// Append as string to prefix
		generatedUuid = strings.ToLower(
			strings.Trim(base32.StdEncoding.EncodeToString(hash), "="),
		)

		var sqlQuery = GetClusterUuidCheck
		var arg = "cl-" + generatedUuid[:7] + "%"
		if UUIDType == NodegroupUUIDType {
			sqlQuery = GetNodeGroupUuidCheck
			arg = "%" + generatedUuid + "%"
		}

		var count int
		if err := dbconn.QueryRowContext(ctx, sqlQuery, arg).Scan(&count); err != nil {
			return "", ErrorHandler(ctx, err, "GenerateUuid", "Could not Generate UUID")
		}

		if count == 0 {
			log.V(0).Info("Generated UUID is unique", logkeys.GeneratedUuid, generatedUuid)
			return generatedUuid, nil
		}

		log.V(0).Info("Generated UUID is not unique, trying again", logkeys.GeneratedUuid, generatedUuid)
		attempt++
	}

	return generatedUuid, ErrorHandler(ctx, errors.New("breaking GenerateUuid() loop after 5 tries"), "GenerateUuid", "Could not Generate UUID")
}

func ParseKeyValuePairIntoMap(ctx context.Context, keyValuePairByte []byte) (map[string]string, error) {
	returnError := make(map[string]string, 0)
	returnValue := make(map[string]string, 0)

	if keyValuePairByte == nil || len(keyValuePairByte) == 0 {
		return returnError, nil
	}
	/* UNMARSHALL DB DEFINITIONS */
	var unmarshalledKeyValuePairs []map[string]string
	err := json.Unmarshal(keyValuePairByte, &unmarshalledKeyValuePairs)
	if err != nil {
		fmt.Println("Unable to parse Cluster Definition", err)
		return returnError, errors.New("Could not parse key value pairs")
	}

	/* CONVERT TO MAP */
	for _, keyValuePair := range unmarshalledKeyValuePairs {
		returnValue[keyValuePair["key"]] = keyValuePair["value"]
	}
	return returnValue, nil
}

func GetClusterCompatDetails(ctx context.Context, dbconn *sql.DB, clusterUuid string) (string, string, string, string, string, error) {
	var (
		clusterProvider     string
		clusterK8sVersion   string
		clusterRuntime      string
		clusterInstanceType string
		clusterOsImageName  string
	)
	err := dbconn.QueryRowContext(ctx, getClusterCompatDetails, clusterUuid).Scan(&clusterProvider, &clusterK8sVersion, &clusterRuntime, &clusterInstanceType)
	if err != nil {
		fmt.Println("\n", (err), "\n .. GetClusterCompateDetails.getCLusterCompatDetails: Unexpected error")
		return "", "", "", "", "", errors.New("Unable to get Cluster Information")
	}
	osImageNameQuery := `
		SELECT o.osimage_name
		FROM  osimageinstance o
		WHERE o.osimageinstance_name = (
    	SELECT n.osimageinstance_name
    	FROM nodegroup n where n.nodegrouptype_name ='ControlPlane' AND n.cluster_id =(
        SELECT c.cluster_id
        FROM cluster c
        WHERE unique_id= $1
    	)
 		)
	`
	err = dbconn.QueryRowContext(ctx, osImageNameQuery, clusterUuid).Scan(&clusterOsImageName)
	if err != nil {
		fmt.Println("\n", (err), "\n .. GetClusterCompateDetails.getCLusterCompatDetails: Unexpected error")
		return "", "", "", "", "", errors.New("Unable to get Cluster Information")
	}

	return clusterProvider, clusterK8sVersion, clusterRuntime, clusterInstanceType, clusterOsImageName, nil
}

func GetNodeGroupCompatDetails(ctx context.Context, dbconn *sql.DB, nodeGroupUuid string) (string, string, string, string, string, string, string, error) {
	var (
		clusterProvider       string
		nodeGroupK8sVersion   string
		nodeGroupRuntime      string
		nodeGroupImi          string
		nodeGroupType         string
		nodeGroupInstanceType string
	)
	err := dbconn.QueryRowContext(ctx, getNodeGroupCompatDetails,
		nodeGroupUuid,
	).Scan(&clusterProvider, &nodeGroupK8sVersion, &nodeGroupRuntime, &nodeGroupImi, &nodeGroupType, &nodeGroupInstanceType)
	if err != nil {
		fmt.Println("\n", (err), "\n .. GetNodeGroupCompatDetails.getNodeGroupCompatDetails: Unexpected Error!")
		return "", "", "", "", "", "", "", errors.New("Unexpected Error: Could not get information about nodegroup")
	}
	osImageNameQuery := `
		SELECT o.osimage_name
		FROM  osimageinstance o
		WHERE o.osimageinstance_name = (
    	SELECT n.osimageinstance_name
    	FROM nodegroup n
		WHERE unique_id= $1
 		)
	`
	var nodeGroupOsImage string
	err = dbconn.QueryRowContext(ctx, osImageNameQuery, nodeGroupUuid).Scan(&nodeGroupOsImage)
	if err != nil {
		fmt.Println("\n", (err), "\n .. GetClusterCompateDetails.getCLusterCompatDetails: Unexpected error")
		return "", "", "", "", "", "", "", errors.New("Unable to get Cluster Information")
	}
	return clusterProvider, nodeGroupK8sVersion, nodeGroupRuntime, nodeGroupImi, nodeGroupType, nodeGroupInstanceType, nodeGroupOsImage, nil
}

func GetImiArtifact(ctx context.Context, dbconn *sql.DB, osImageInstanceName string) (string, error) {
	emptyReturn := ""

	var imiartifact string
	/* GET IMI ARTIFACT FOR CLUSTER REV DESIRED JSON */
	err := dbconn.QueryRowContext(ctx, GetImiArtifactQuery, osImageInstanceName).Scan(&imiartifact)
	if err != nil {
		fmt.Println("\n", (err), "\n .. GetImiArtifact.GetImiArtifactQuery: Unexpected Error!")
		return emptyReturn, errors.New("Unexected Error: Unable to get Image Artifact for k8sversion")
	}
	return imiartifact, nil
}

func GetMinorVersion(ctx context.Context, dbconn *sql.DB, k8sVersionName string) (string, error) {
	emptyReturn := ""
	var minorVersion string
	err := dbconn.QueryRowContext(ctx, getMinorVersionQuery, k8sVersionName).Scan(&minorVersion)
	if err != nil {
		fmt.Println("\n", (err), "\n .. GetMinorVersion.getMinorVersion: Unexpected Error!")
		return emptyReturn, errors.New("Unexected Error: Unable to get minor version")
	}

	return minorVersion, nil
}

func GetAvailableClusterVersionUpgrades(ctx context.Context, dbconn *sql.DB, clusterUuid string) ([]string, error) {
	var availableVersions []string
	returnError := make([]string, 0)

	/* VALIDATE CLUSTER IS ACTIONABLE*/
	actionableState, err := ValidaterClusterActionable(ctx, dbconn, clusterUuid)
	if err != nil {
		return returnError, errors.New("Unexpected Error: Could not validate state of cluster")
	}
	if !actionableState {
		return availableVersions, nil
	}

	/* GET INFO ABOUT CLUSTER*/
	clusterProvider, currentK8sVersion, clusterRuntime, clusterInstanceType, clusterOsImage, err := GetClusterCompatDetails(ctx, dbconn, clusterUuid)
	if err != nil {
		return returnError, errors.New("Unable to get Active K8sversion")
	}

	/* Get Active K8sVersions and Active IMI for the nodegroupType (ControlPlane) */
	activeK8sVersionsAndImi := make(map[string]string, 0)
	rows, err := dbconn.QueryContext(ctx, getActiveK8sVersionAndImi,
		cpNodegroupType,
		clusterProvider,
		clusterRuntime,
		clusterInstanceType,
		clusterOsImage,
	)
	if err != nil {
		fmt.Println("Unexpected Error: " + err.Error())
		return returnError, errors.New("Unable to get Active K8sversion")
	}
	defer rows.Close()
	for rows.Next() {
		var (
			k8sversion string // Patch name for k8sversion
			activeImi  string // IMI for the patch name
		)
		err = rows.Scan(&k8sversion, &activeImi)
		if err != nil {
			return returnError, errors.New("Unable to scan Active K8sversion and Active imi")
		}
		activeK8sVersionsAndImi[k8sversion] = activeImi
	}

	/* Convert into comparable version string using semver package and sort */
	semverCurrentVersion, err := parseVersions(currentK8sVersion)
	if err != nil {
		fmt.Println(err.Error())
		return returnError, errors.New("Unable to get Current K8sversion")
	}
	semverActiveK8sVersions, err := parseVersions(maps.Keys(activeK8sVersionsAndImi)...)
	if err != nil {
		return returnError, errors.New("Unable to get Active K8sversion")
	}
	semver.Sort(semverActiveK8sVersions)

	/* For every active k8sversion, compare if newer k8sversion exists */
	for _, activeK8sVersion := range semverActiveK8sVersions {
		if clusterProvider == iksProvider {
			if semverCurrentVersion[0].Minor < activeK8sVersion.Minor {
				availableVersions = append(availableVersions, fmt.Sprintf("%d.%d", activeK8sVersion.Major, activeK8sVersion.Minor))
				break
			}
		} else if clusterProvider == rke2Provider {
			if semverCurrentVersion[0].Minor == activeK8sVersion.Minor && semverCurrentVersion[0].Patch < activeK8sVersion.Patch {
				availableVersions = append(availableVersions, activeK8sVersion.String())
			} else if semverCurrentVersion[0].Minor < activeK8sVersion.Minor {
				availableVersions = append(availableVersions, activeK8sVersion.String())
				break
			}
		}
	}

	return availableVersions, nil
}

func GetAvailableWorkerImiUpgrades(ctx context.Context, dbconn *sql.DB, clusterUuid string, nodeGroupUuid string) ([]string, []string, error) {
	emptyReturn := make([]string, 0)
	var availableVersionsImi []string
	var availableVersionsKeys []string

	/* VALIDATIONS */
	// Get state of cluster (returns no upgrads available if not in actionable state)
	actionableState, err := ValidaterClusterActionable(ctx, dbconn, clusterUuid)
	if err != nil {
		return emptyReturn, emptyReturn, errors.New("Unexpected Error: Could not validate state of cluster")
	}
	if !actionableState {
		return emptyReturn, emptyReturn, nil
	}
	// Get state of nodegroup (returns no upgrads available if not in actionable state)
	var state string
	err = dbconn.QueryRowContext(ctx, GetNodeGroupStateQuery, nodeGroupUuid).Scan(&state)
	if err != nil {
		fmt.Println("GetAvailableWorkerImiUpgrades.GetNodeGroupStateQuery: Could not nodegroup state.", err.Error())
		return emptyReturn, emptyReturn, errors.New("Unexpcted Error: Could not get available nodegroup upgrades")
	}
	if state == "Updating" || state == "Creating" || state == "Deleting" {
		fmt.Println("GetAvailableWorkerImiUpgrades.GetNodeGroupStateQuery: Nodegroup is updating.")
		return emptyReturn, emptyReturn, nil
	}

	/* GET INFO NEEDED FOR COMPARISONS [Current Wrk IMI, Current K8s Active IMI, Available K8s PatcVersion] */
	// Get Control Plane UUID
	var controlPlaneUuid string
	err = dbconn.QueryRowContext(ctx, getClusterControlPlaneQuery, clusterUuid).Scan(&controlPlaneUuid)
	if err != nil {
		fmt.Println("GetAvailableWorkerImiUpgrades.getClusterControlPlaneQuery: Could not cluster control plane.", err.Error())
		return emptyReturn, emptyReturn, errors.New("Unexpcted Error: Could not get available nodegroup upgrades")
	}
	// Get Control Plane Info IMI
	_, cpK8sVersion, _, controlPlaneImi, _, _, _, err := GetNodeGroupCompatDetails(ctx, dbconn, controlPlaneUuid)
	// Get current Nodegroup Info and IMI
	wkProvider, wkK8sVersion, wkRuntime, nodeGroupImi, _, wkInstanceType, wkOsImage, err := GetNodeGroupCompatDetails(ctx, dbconn, nodeGroupUuid)

	if wkProvider == rke2Provider {
		// APPEND IF RKE VERSIONS DON'T MATCH
		if nodeGroupImi != controlPlaneImi {
			availableVersionsImi = append(availableVersionsImi, controlPlaneImi)
			availableVersionsKeys = append(availableVersionsImi, cpK8sVersion)
		}
	} else if wkProvider == iksProvider {
		// CONVERT VERSIONS TO COMPARABLE SEMVER VERSIONS
		semverControlPlaneVersion, err := parseVersions(cpK8sVersion)
		if err != nil {
			fmt.Println("GetAvailableWorkerImiUpgrades.parseversions: Could not parse control plane k8sverison.", err.Error())
			return emptyReturn, emptyReturn, errors.New("Unexpcted Error: Could not get available nodegroup upgrades")
		}
		semverNodeGroupVersion, err := parseVersions(wkK8sVersion)
		if err != nil {
			fmt.Println("GetAvailableWorkerImiUpgrades.parseversions: Could not parse nodegroup k8sverison.", err.Error())
			return emptyReturn, emptyReturn, errors.New("Unexpcted Error: Could not get available nodegroup upgrades")
		}
		if semverNodeGroupVersion[0].Minor == semverControlPlaneVersion[0].Minor && semverNodeGroupVersion[0].Patch < semverControlPlaneVersion[0].Patch {
			// GET CURRENT WORKER IMI FOR WORKER K8SVERSION
			var availableImi string
			err = dbconn.QueryRowContext(ctx, getImiFromK8sVersion,
				wkNodegroupType,
				wkProvider,
				wkRuntime,
				cpK8sVersion,
				wkInstanceType,
				wkOsImage,
			).Scan(&availableImi)
			if err != nil {
				fmt.Println("GetAvailableWorkerImiUpgrades.getImiFromK8sVersion: Could not get Worker IMI for provided k8sverison.", err.Error())
				return emptyReturn, emptyReturn, errors.New("Unexpcted Error: Could not get available nodegroup upgrades")
			}
			// ADD TO AVAILABLE UPGRADES
			availableVersionsImi = append(availableVersionsImi, availableImi)
			availableVersionsKeys = append(availableVersionsImi, cpK8sVersion)
		}
	}

	return availableVersionsKeys, availableVersionsImi, nil
}

/* HELPER FUNCTION TO OBTAIN ANY AVAILABLE UPGRADES FOR A NODEGROUP ***NEED TO DOUBLE CHECK RKE@ LOGIC*** */
func GetAvailableControlPlaneImiUpgrades(ctx context.Context, dbconn *sql.DB, clusterUuid string, controlPlaneUuid string) ([]string, []string, error) {
	var availableVersionsImi []string
	var availableVersionsKeys []string
	emptyReturn := make([]string, 0)

	/* VALIDATIONS */
	// Get state of cluster (returns no upgrads available if not in actionable state)
	actionableState, err := ValidaterClusterActionable(ctx, dbconn, clusterUuid)
	if err != nil {
		return emptyReturn, emptyReturn, errors.New("Unexpected Error: Could not validate state of cluster")
	}
	if !actionableState {
		return emptyReturn, emptyReturn, nil
	}
	// Get state of nodegroup (returns no upgrads available if not in actionable state)
	var state string
	err = dbconn.QueryRowContext(ctx, GetNodeGroupStateQuery, controlPlaneUuid).Scan(&state)
	if err != nil {
		fmt.Println("GetAvailableWorkerImiUpgrades.GetNodeGroupStateQuery: Could not nodegroup state.", err.Error())
		return emptyReturn, emptyReturn, errors.New("Unexpcted Error: Could not get available nodegroup upgrades")
	}
	if state == "Updating" || state == "Creating" || state == "Deleting" {
		fmt.Println("GetAvailableWorkerImiUpgrades.GetNodeGroupStateQuery: Nodegroup is updating.")
		return emptyReturn, emptyReturn, nil
	}

	/* GET INFORMATION NEEDED FOR COMPARISONS [Current CP IMI, Current K8s Active IMI, Available K8s PatchVersion] */
	// Get Needed information for K8sCompatability query [ClusterProvider, ClusterRuntime, K8sVersionName, InstanceType, OsImage]
	cpProvider, currentCpK8sVersion, cpRuntime, currentCpImi, _, cpInstanceeType, cpOsImage, err := GetNodeGroupCompatDetails(ctx, dbconn, controlPlaneUuid)
	// Get current Active IMI for the K8sVersion (to compare if Patch IMI has changed)
	var currentActiveImi string
	err = dbconn.QueryRowContext(ctx, getImiFromK8sVersion,
		cpNodegroupType,
		cpProvider,
		cpRuntime,
		currentCpK8sVersion,
		cpInstanceeType,
		cpOsImage,
	).Scan(&currentActiveImi)
	if err != nil {
		fmt.Println("GetAvailableControlPlaneImiUpgrades.getImiFromK8sVersion: Could not get current active IMI.", err.Error())
		return emptyReturn, emptyReturn, errors.New("Unexpcted Error: Could not get available nodegroup upgrades")
	}
	// Get all Active K8sVersion and Patche IMIs that are in Active state
	k8sPatchVersionsAndImi := make(map[string]string, 0)
	rows, err := dbconn.QueryContext(ctx, getK8sVersionPatchVersions,
		cpNodegroupType,
		cpProvider,
		cpRuntime,
		currentCpK8sVersion,
		cpInstanceeType,
		cpOsImage,
	)
	if err != nil {
		fmt.Println(err.Error())
		return emptyReturn, emptyReturn, errors.New("Unable to get Active K8sversion")
	}
	defer rows.Close()
	for rows.Next() {
		var (
			k8sversion string
			activeImi  string
		)
		err = rows.Scan(&k8sversion, &activeImi)
		if err != nil {
			return emptyReturn, emptyReturn, errors.New("Unable to get k8sversion and active imi")
		}
		k8sPatchVersionsAndImi[k8sversion] = activeImi
	}

	/* COMPARE TO SEE IF UPGRADE IS AVAILABLE */
	// Convert into comparable version string using semver package and sort
	semverCurrentVersion, err := parseVersions(currentCpK8sVersion)
	if err != nil {
		fmt.Println(err.Error())
		return emptyReturn, emptyReturn, errors.New("Unable to get Current K8sversion")
	}

	semverK8sPatchVersions, err := parseVersions(maps.Keys(k8sPatchVersionsAndImi)...)
	if err != nil {
		return emptyReturn, emptyReturn, errors.New("Unable to get Active K8sversion")
	}
	semver.Sort(semverK8sPatchVersions)

	/* If cluster provider is rke2, compare if current IMI is not the same as the Active IMI for the current K8s Version */
	if cpProvider == rke2Provider {
		// Check if current k8sversion IMI is latest
		if k8sPatchVersionsAndImi[currentCpK8sVersion] != "" && currentCpImi != k8sPatchVersionsAndImi[currentCpK8sVersion] {
			availableVersionsKeys = append(availableVersionsKeys, currentCpK8sVersion)
			availableVersionsImi = append(availableVersionsImi, k8sPatchVersionsAndImi[currentCpK8sVersion])
			// Check if current k8sversion is deprecated
		} else if currentCpImi != currentActiveImi {
			availableVersionsKeys = append(availableVersionsKeys, currentCpK8sVersion)
			availableVersionsImi = append(availableVersionsImi, currentActiveImi)
		}
		/* If cluster provider is Iks, compare if current Patch version is not the same as the Next Patch for the current K8s Version */
	} else if cpProvider == iksProvider {
		for _, activeK8sVersion := range semverK8sPatchVersions {
			if semverCurrentVersion[0].Minor == activeK8sVersion.Minor && semverCurrentVersion[0].Patch < activeK8sVersion.Patch {
				availableVersionsImi = append(availableVersionsImi, k8sPatchVersionsAndImi[activeK8sVersion.String()])
				availableVersionsKeys = append(availableVersionsKeys, activeK8sVersion.String())
			}
		}
	}
	return availableVersionsKeys, availableVersionsImi, nil
}

/*
	 Validate Cluster Existance
		- Searches for cluster using Unique ID
		- Return Cluster Id
*/
func ValidateClusterExistance(ctx context.Context, dbconn *sql.DB, clusterUuid string) (int32, error) {
	var clusterId int32
	returnError := int32(-1)

	err := dbconn.QueryRowContext(ctx, ClusterExistanceQuery, clusterUuid).Scan(&clusterId)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, nil
		}
		fmt.Println("\n", (err), "\n .. ValidateClusterExistance.ClusterExistanceQuery. Unexpexted error")
		return returnError, errors.New("Unexpected Error: Cluster not found: " + clusterUuid)
	}

	return clusterId, nil
}

/*
	 Validate Nodegroup Existance
		- Searches for cluster and nodegroup using Unique ID
		- Return Cluster Id and Nodegroup Id
*/
func ValidateNodeGroupExistance(ctx context.Context, dbconn *sql.DB, clusterUuid string, nodegroupUuid string) (int32, int32, error) {
	var clusterId, nodeGroupId int32
	returnError := int32(-1)
	err := dbconn.QueryRowContext(ctx, NodeGroupExistanceQuery, clusterUuid, nodegroupUuid).Scan(&clusterId, &nodeGroupId)
	if err != nil {
		if err == sql.ErrNoRows {
			return returnError, returnError, nil
		}
		fmt.Println("\n", (err), "\n .. ValidateNodegroupExistance.NodeGroupExistanceQuery. Unexpected error")
		return returnError, returnError, errors.New("Unexpected Error: Nodegroup not found: " + clusterUuid)
	}

	return clusterId, nodeGroupId, nil
}

/*
	 Validate VIP Existance
		- Searches for cluster and vip using vip_id
		- Return Cluster Id and vip Id
*/
func ValidateVipExistance(ctx context.Context, dbconn *sql.DB, clusterUuid string, vipIdParam int32, isPublicRequest bool) (int32, int32, error) {
	var clusterId, vipId int32
	returnError := int32(-1)
	query := VipExistanceQuery
	if isPublicRequest {
		query = query + " AND v.owner = $3"
		err := dbconn.QueryRowContext(ctx, query, clusterUuid, vipIdParam, "customer").Scan(&clusterId, &vipId)
		if err != nil {
			if err == sql.ErrNoRows {
				return returnError, returnError, nil
			}
			fmt.Println("\n", (err), "\n .. ValidateVipExistance.VipExistanceQuery. Unexpexted error")
			return returnError, returnError, errors.New("Unexpected Error: Vip not found: " + clusterUuid)
		}
	} else {
		err := dbconn.QueryRowContext(ctx, query, clusterUuid, vipIdParam).Scan(&clusterId, &vipId)
		if err != nil {
			if err == sql.ErrNoRows {
				return returnError, returnError, nil
			}
			fmt.Println("\n", (err), "\n .. ValidateVipExistance.VipExistanceQuery. Unexpexted error")
			return returnError, returnError, errors.New("Unexpected Error: Vip not found: " + clusterUuid)
		}
	}
	return clusterId, vipId, nil
}

/*
	 Validate Super Compute Cluster Type Existance
		- Searches for cluster type as super compute
		- If it is super compute it will check for general purpose nodegroup type using Unique ID
		- Return bool and error
*/
func ValidateSuperComputeClusterType(ctx context.Context, dbconn *sql.DB, clusterUuid string) (bool, error) {
	var clusterId int32
	var clustertype string
	err := dbconn.QueryRowContext(ctx, db_query.GetClusterTypeQuery, clusterUuid).Scan(&clusterId, &clustertype)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, errors.New("Unexpected Error: Cluster not found: " + clusterUuid)
		}
		fmt.Println("\n", (err), "\n .. ValidateSuperComputeClusterAndGPNodeGroupExistance.GetClusterTypeQuery. Unexpected error")
		return false, errors.New("Unexpected Error: Cluster not found: " + clusterUuid)
	}

	if clustertype != superComputeClusterType {
		return false, nil
	}

	return true, nil
}

/*
	 Validate Super Compute General Purpose Nodegroup Type Existance
		- Searches for cluster type as super compute
		- If it is super compute it will check for general purpose nodegroup type using Unique ID
		- Return bool and error
*/
func ValidateSuperComputeGPNodegroupType(ctx context.Context, dbconn *sql.DB, clusterUuid string, nodegroupUuid string) (bool, error) {
	var clusterId, nodeGroupId int32
	var nodegrouptype string
	err := dbconn.QueryRowContext(ctx, db_query.GetNodeGroupTypeQuery, clusterUuid, nodegroupUuid).Scan(&clusterId, &nodeGroupId, &nodegrouptype)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, errors.New("Unexpected Error: Nodegroup not found: " + clusterUuid)
		}
		fmt.Println("\n", (err), "\n .. ValidateNodegroupExistance.NodeGroupExistanceQuery. Unexpected error")
		return false, errors.New("Unexpected Error: Nodegroup not found: " + clusterUuid)
	}

	if nodegrouptype == superComputeAINodegroupType {
		return false, errors.New("Invalid Nodegroup Type Error: Cannot Modify or Delete AI Node Group type")
	}

	return true, nil
}

/*
	 Validate Super Compute General Purpose Nodegroup Instance Type Existance
		- Searches for cluster type as super compute
		- If it is super compute it will check for general purpose nodegroup type using Unique ID
		- Return bool and error
*/
func ValidateSuperComputeGPNodegroupAndInstanceType(ctx context.Context, dbconn *sql.DB, clusterUuid string, nodegroupUuid string) (bool, error) {
	var clusterId, nodeGroupId int32
	var nodegrouptype, instancetype string

	/*Default Values for IKS */
	defaultvalues, err := GetDefaultValues(ctx, dbconn)
	if err != nil {
		return false, errors.New("GetDefaultValues: Failed to get the default value")
	}
	aiNodegroupTypeList := strings.Split(defaultvalues["sc_ai_nodegrouptype"], ",")

	err = dbconn.QueryRowContext(ctx, db_query.GetNodeGroupTypeAndInstanceTypeQuery, clusterUuid, nodegroupUuid).Scan(&clusterId, &nodeGroupId, &nodegrouptype, &instancetype)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, errors.New("Unexpected Error: Nodegroup not found: " + clusterUuid)
		}
		fmt.Println("\n", (err), "\n .. ValidateNodegroupExistance.NodeGroupExistanceQuery. Unexpected error")
		return false, errors.New("Unexpected Error: Nodegroup not found: " + clusterUuid)
	}

	if nodegrouptype == superComputeAINodegroupType {
		return false, errors.New("Invalid Nodegroup Type Error: Cannot Modify or Delete AI Node Group type")
	}

	if nodegrouptype == superComputeGPNodegroupType {
		isValidSCGPInstanceType := ValidateSCGPNodegroupInstanceType(instancetype, aiNodegroupTypeList)
		if !isValidSCGPInstanceType {
			return false, errors.New(fmt.Sprintf("Invalid Instance Type Error: Cannot Have %s for General Purposes nodes", instancetype))
		}
	}

	if !ValidateScConstant(instancetype) {
		return false, fmt.Errorf("SuperCompute Cluster: Invalid Instance type for Nodegroup Name %s", instancetype)
	}

	return true, nil
}

func ValidateClusterCloudRestrictions(ctx context.Context, dbconn *sql.DB, cloudAccountId string) (bool, error) {
	/* GET RESTRICTIONS DEFAULT */
	var restricted bool
	err := dbconn.QueryRow(GetDefaultConfigUtilsQuery, "restrict_create_cluster").Scan(&restricted)
	if err != nil {
		fmt.Println("\n", (err), "\n .. ValidateClusterCloudRestrictions.GetDefaultconfig.restrict_create_cluster. Unexpexted error")
		return false, err
	}
	if restricted == false { // If not restricted, everyone can create
		return true, nil
	}
	/* GET ACTIVE ACCOUNT FROM EXTRASPEC */
	activeAccount := false
	err = dbconn.QueryRow(GetCloudAccountCreate, cloudAccountId).Scan(&activeAccount)
	if err != nil {
		fmt.Println("\n", (err), "\n .. ValidateClusterCloudRestrictions.GetActiveAccountCreateCluster.restrict_create_cluster. Unexpexted error")
		return false, err
	}

	return activeAccount, nil
}

func ValidateStorageRestrictions(ctx context.Context, dbconn *sql.DB, cloudAccountId string) (bool, error) {
	/* GET ACTIVE ACCOUNT FROM EXTRASPEC */
	var activeAccount bool
	err := dbconn.QueryRow(GetCloudAccountCreate, cloudAccountId).Scan(&activeAccount)
	if err != nil {
		fmt.Println("\n", (err), "\n .. ValidateStorageRestrictions.GetActiveAccountCreateCluster.allow_create_storage. Unexpexted error")
		return false, err
	}

	if activeAccount == false {
		return activeAccount, nil
	}

	var storageAllowed bool
	err = dbconn.QueryRow(GetCloudAccountStorage, cloudAccountId).Scan(&storageAllowed)
	if err != nil {
		fmt.Println("\n", (err), "\n .. ValidateStorageRestrictions.GetActiveAccountStorage.allow_create_storage. Unexpexted error")
		return false, err
	}

	return storageAllowed, nil
}

/*
Validate Cluster Permissions
  - Validates that the Cloud Account passed in has access to the cluster
  - Returns Bool. True if is owner
*/
func ValidateClusterCloudAccount(ctx context.Context, dbconn *sql.DB, clusterUuid string, cloudAccountId string) (bool, error) {
	var cloudaccountId string
	err := dbconn.QueryRowContext(ctx, GetClusterCloudAccountQuery, clusterUuid).Scan(&cloudaccountId)
	if err != nil {
		fmt.Println("\n", (err), "\n .. ValidateClusterCloudAccount.GetClusterCloudAccountQuery: Unexpected error!")
		return false, errors.New("Unexpected Error: Could not validate Cloud Account permissions")
	}
	if cloudaccountId != cloudAccountId {
		return false, nil
	}
	return true, nil
}

/*
	 Validate Cluster Actionable
		- Validates that cluster is in actionable state (Cluster Rev applied and not in Pending or Deleted State)
		- Returns Bool. True if actionable
*/
func ValidaterClusterActionable(ctx context.Context, dbconn *sql.DB, clusterUuid string) (bool, error) {
	// Validate Cluster State
	var clusterState string
	err := dbconn.QueryRowContext(ctx, GetClusterStateQuery, clusterUuid).Scan(&clusterState)
	if err != nil {
		fmt.Println("\n", (err), "\n .. ValidateClusterActionable.GetClusterStateQuery: Unexpected error")
		return false, errors.New("Unexpected Error: Could not validate state of cluster")
	}
	if clusterState == "Updating" || clusterState == "Pending" || clusterState == "DeletePending" || clusterState == "Deleting" || clusterState == "Deleted" {
		return false, nil
	}

	// Validate Cluster Revision
	var changeApplied bool
	err = dbconn.QueryRowContext(ctx, GetClusterRevChangeAppliedQuery, clusterUuid).Scan(&changeApplied)
	if err != nil {
		fmt.Println("\n", (err), "\n .. ValidateClusterActionable.GetClusterRevChangeAppliedQuery: Unexpected error")
		return false, errors.New("Unexpected Error: Could not validate state of cluster")
	}
	if !changeApplied {
		return false, nil
	}

	return true, nil
}

func ValidateNodegroupDelete(ctx context.Context, dbconn *sql.DB, nodegroupUuid string) (bool, error) {
	// Validate nodegroup deleting state
	var nodegroupstate string
	err := dbconn.QueryRowContext(ctx, GetNodeGroupStateQuery, nodegroupUuid).Scan(&nodegroupstate)
	if err != nil {
		fmt.Println("\n", (err), "\n .. ValidateNodegroupDelete.GetNodeGroupStateQuery: Unexpected error")
		return false, errors.New("Unexpected Error: Could not obtain nodegroup state")
	}

	if nodegroupstate == "Deleting" {
		return true, nil
	}
	return false, nil
}

func ValidateClusterDelete(ctx context.Context, dbconn *sql.DB, clusterUuid string) (bool, error) {
	// Validate cluster deleting state
	var clusterstate string
	err := dbconn.QueryRowContext(ctx, GetClusterStateQuery, clusterUuid).Scan(&clusterstate)
	if err != nil {
		fmt.Println("\n", (err), "\n .. ValidateClusterDelete.GetClusterStateQuery: Unexpected error")
		return false, errors.New("Unexpected Error: Could not obtain cluster state")
	}

	if clusterstate == "Deleting" {
		return true, nil
	}
	return false, nil
}

func ValidateVipDelete(ctx context.Context, dbconn *sql.DB, vipid int32) (bool, error) {
	// Validate vip deleting state
	var vipstate string
	err := dbconn.QueryRowContext(ctx, GetVipStateQuery, vipid).Scan(&vipstate)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		fmt.Println("\n", (err), "\n .. ValidateVipDelete.GetVipStatequery: Unexpected error")
		return false, errors.New("Unexpected Error: Could not obtain vip state")
	}
	if vipstate == "Deleting" {
		return true, nil
	}
	return false, nil
}

func ValidateSecActive(ctx context.Context, dbconn *sql.DB, vipid int32) (bool, error) {
	// Validate security state
	var secstate string
	err := dbconn.QueryRowContext(ctx, GetSecStateQuery, vipid).Scan(&secstate)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		fmt.Println("\n", (err), "\n .. ValidatSecDelete.GetSecStatequery: Unexpected error")
		return false, errors.New("Unexpected Error: Could not obtain security state")
	}
	if secstate == "Deleting" || secstate == "Reconciling" || secstate == "Deleted" {
		return true, nil
	}
	return false, nil
}

/*
	 Validate Instance Type Existance
		- Searches for instance type based on name
		- Return instance type name
*/
func ValidateInstanceTypeExistance(ctx context.Context, dbconn *sql.DB, instanceName string) (bool, error) {
	// Validate instance type existance
	var name string
	err := dbconn.QueryRowContext(ctx, GetInstanceTypeQuery, instanceName).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		fmt.Println("\n", (err), "\n .. ValidateInstanceType.GetInstanceTypeQuery: Unexpected error")
		return false, errors.New("Unexpected Error: Could not obtain instance type name")
	}
	return true, nil
}

func ValidateSecRuleExistance(ctx context.Context, dbconn *sql.DB, vipid int32) (bool, error) {
	var count int32
	err := dbconn.QueryRowContext(ctx, GetsourceipQueryCount, vipid).Scan(&count)
	if err != nil {
		return false, errors.New("Unexpected Error: Could not obtain source ip for vip")
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

func GetLatestClusterRev(ctx context.Context, dbconn *sql.DB, clusterUuid string) ([]byte, clusterv1alpha.Cluster, error) {
	emptyCluster := clusterv1alpha.Cluster{}
	emptyJson := []byte{}
	var (
		currentJson []byte
		clusterCrd  clusterv1alpha.Cluster
	)
	err := dbconn.QueryRowContext(ctx, GetLatestClusterRevQuery, clusterUuid).Scan(&currentJson)
	if err != nil {
		fmt.Println("\n", (err), "\n .. GetLatestClusterRev.GetLatestClusterRevQuery: Unexpected error")
		return emptyJson, emptyCluster, errors.New("Unexpected Error: Could not obtain latest revision of the cluster")
	}
	err = json.Unmarshal([]byte(currentJson), &clusterCrd)
	if err != nil {
		fmt.Println("\n", (err), "\n .. GetLatestClusterRev.json.unmarshal: Unexpected error")
		return emptyJson, emptyCluster, errors.New("Unexpected Error: Could not obtain latest revision of the cluster")
	}
	return currentJson, clusterCrd, nil
}

func GetDefaultValues(ctx context.Context, dbconn *sql.DB) (map[string]string, error) {
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("GetDefaultValues")

	var (
		data        []Defaultvalues
		getdefaults string
	)
	rows, err := dbconn.QueryContext(ctx, GetDefaultsIKSQuery)
	if err != nil {
		log.Error(err, "\n .. Get default values Transaction rollback")
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&getdefaults)
		if err != nil {
			log.Error(err, "\n .. Scan default values error !")
			return nil, err
		}

		err = json.Unmarshal([]byte(getdefaults), &data)
		if err != nil {
			log.Error(err, "\n .. Unmarshall default values error")
			return nil, err
		}
	}

	defaultvalues := make(map[string]string)
	for _, d := range data {
		defaultvalues[d.Name] = d.Value
	}

	return defaultvalues, nil
}

func GetDefaultAddons(ctx context.Context, dbconn *sql.DB, k8sversion string) ([]DefaultAddons, error) {
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("GetDefaultAddons")

	var addonsnames []string
	var addonname string
	rows, err := dbconn.QueryContext(ctx, GetAddonsfork8sVersionsQuery, k8sversion)
	if err != nil {
		log.Error(err, "\n .. Get addons compatability Transaction rollback")
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&addonname)
		if err != nil {
			log.Error(err, "\n .. Scan default add on names error !")
			return nil, err
		}
		addonsnames = append(addonsnames, addonname)
	}

	var (
		addondata   DefaultAddons
		addonvalues []DefaultAddons
	)

	for _, addonname := range addonsnames {
		rows, err := dbconn.QueryContext(ctx, GetDefaultAddonsQuery, addonname)
		if err != nil {
			log.Error(err, "\n .. Get default values Transaction rollback")
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&addondata.Name, &addondata.Type, &addondata.Artifact)
			if err != nil {
				log.Error(err, "\n .. Scan default values error !")
				return nil, err
			}
		}

		addonvalues = append(addonvalues, addondata)
	}

	return addonvalues, nil
}

func IsInstanceGroup(instancetype string) bool {
	identifier := "cluster-"
	isInstanceGroup := strings.Contains(instancetype, identifier)
	return isInstanceGroup
}

func GetPrivateFilesystemNS(ctx context.Context, fsClient pb.FilesystemOrgPrivateServiceClient, cloudaccount string, iksFilesystem string) (bool, *pb.AssignedNamespace, error) {
	returnError := &pb.AssignedNamespace{}
	returnNamespace := &pb.AssignedNamespace{}
	exists := false

	fsOrgGetRequest := &pb.FilesystemOrgGetRequestPrivate{
		Metadata: &pb.FilesystemMetadataReference{
			NameOrId: &pb.FilesystemMetadataReference_Name{
				Name: iksFilesystem,
			},
			CloudAccountId: cloudaccount,
		},
		//Prefix: "", // WHAT IS THIS
	}

	filesystem, err := fsClient.GetFilesystemOrgPrivate(ctx, fsOrgGetRequest)
	if err != nil && status.Code(err) != codes.NotFound {
		return exists, returnError, err
	}
	// Doesn't exists
	if err != nil && status.Code(err) == codes.NotFound {
		return exists, returnError, nil
	}
	// Exists
	if filesystem == nil || filesystem.Spec == nil || filesystem.Spec.Scheduler == nil || filesystem.Spec.Scheduler.Namespace == nil || filesystem.Spec.Scheduler.Namespace.CredentialsPath == "" || filesystem.Spec.Scheduler.Namespace.Name == "" {
		return exists, returnError, errors.New("could not get Namespace")
	}
	exists = true
	returnNamespace = filesystem.Spec.Scheduler.Namespace

	return exists, returnNamespace, nil
}

func CreatePrivateFilesystemNS(ctx context.Context, fsClient pb.FilesystemOrgPrivateServiceClient, cloudaccount string, iksFilesystem string, availabilityZone string, storageSize string) (*pb.AssignedNamespace, error) {
	returnError := &pb.AssignedNamespace{}

	fsCreateRequest := &pb.FilesystemOrgCreateRequestPrivate{
		Metadata: &pb.FilesystemMetadataPrivate{
			Name:           iksFilesystem,
			CloudAccountId: cloudaccount,
			Description:    fmt.Sprintf("IKS FS for user: %s", cloudaccount),
			SkipQuotaCheck: true,
		},
		Spec: &pb.FilesystemSpecPrivate{
			AvailabilityZone: availabilityZone,
			FilesystemType:   2, // 2 is the ComputeKubernetes type
			Request: &pb.FilesystemCapacity{
				Storage: storageSize,
			},
		},
	}

	filesystem, err := fsClient.CreateFilesystemOrgPrivate(ctx, fsCreateRequest)
	if err != nil {
		return returnError, err
	}

	if filesystem == nil || filesystem.Spec == nil || filesystem.Spec.Scheduler == nil || filesystem.Spec.Scheduler.Namespace == nil || filesystem.Spec.Scheduler.Namespace.CredentialsPath == "" || filesystem.Spec.Scheduler.Namespace.Name == "" {
		return returnError, errors.New("could not get Namespace")
	}

	return filesystem.Spec.Scheduler.Namespace, nil // THIS IS TEMPORARY UNTIL WE GET THE ACTUAL RETURN
}

func GetCloudAccountMaxValues(ctx context.Context, dbconn *sql.DB, cloudaccount string) (int, int, int, int, int, error) {
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("GetCloudAccountMaxValues")

	// Get the Cloud Account max values
	maxClustersPerCloudAccount := -1
	maxNodegroupsPerCluster := -1
	maxIlbsPerCluster := -1
	maxNodesPerCluster := -1
	maxNodesPerCloudAccount := -1
	err := dbconn.QueryRowContext(ctx, GetCloudAccountExtraSpecMaxValues, cloudaccount).Scan(
		&maxClustersPerCloudAccount,
		&maxNodegroupsPerCluster,
		&maxIlbsPerCluster,
		&maxNodesPerCluster,
		&maxNodesPerCloudAccount,
	)
	if err != nil && err != sql.ErrNoRows {
		log.Error(err, "\n .. Scan cloud account max values error!")
		return maxClustersPerCloudAccount, maxNodegroupsPerCluster, maxIlbsPerCluster, maxNodesPerCluster, maxNodesPerCloudAccount, err
	}

	return maxClustersPerCloudAccount, maxNodegroupsPerCluster, maxIlbsPerCluster, maxNodesPerCluster, maxNodesPerCloudAccount, nil
}

func IsIksAdminAuthenticatedUser(ctx context.Context, filePath string, userInputKey string) (bool, error) {
	vaultAdminUserKeyByte, _, err := GetLatestAdminUserKey(ctx, filePath)
	if err != nil {
		return false, err
	}

	userInputKeyByte, err := Base64DecodeString(ctx, userInputKey)
	if err != nil {
		return false, err
	}

	if string(userInputKeyByte) == string(vaultAdminUserKeyByte) {
		return true, nil
	}

	return false, nil
}

func GetLatestAdminUserKey(ctx context.Context, filePath string) ([]byte, int, error) {
	body, err := ioutil.ReadFile(filePath)
	if err != nil {
		return make([]byte, 0), 1, nil
	}
	var content EncryptionStruct
	err = json.Unmarshal(body, &content)
	if err != nil {
		return make([]byte, 0), 1, err
	}
	key := 0
	value := ""
	for k, v := range content.Data {
		i, err := strconv.Atoi(k)
		if err != nil {
			return make([]byte, 0), 1, err
		}

		if i > key {
			key = i
			value = v
		}
	}
	if key == 0 || value == "" {
		return make([]byte, 0), 1, errors.New("Please check admin user key values")
	}

	return []byte(value), key, nil
}

func ValidateCreateCluster(ctx context.Context, dbconn *sql.DB, name string, cloudAccountId string, maxClusterCount int, failedFunction string, friendlyMessage string) (string, string, string, string, error) {
	/* VALIDATE CLOUD ACCOUNT RESTRICTIONS*/
	activeAccount, err := ValidateClusterCloudRestrictions(ctx, dbconn, cloudAccountId)
	if err != nil {
		return "", "", "", "", ErrorHandler(ctx, err, failedFunction+"Utils.ValidateClusterCloudRestrictions", friendlyMessage)
	}
	if !activeAccount {
		return "", "", "", "", errors.New("Due to restrictions, we are currently not allowing non-approved users to provision clusters.")
	}

	/* VALIDATE CLUSTER NAME UNIQUENESS*/
	rows, err := dbconn.QueryContext(ctx, db_query.GetClustersStatesByName, name, cloudAccountId)
	if err != nil {
		return "", "", "", "", ErrorHandler(ctx, err, failedFunction+"GetClustersStatesbyName", friendlyMessage)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			clusterState string
			clusterUuid  string
		)
		err = rows.Scan(&clusterUuid, &clusterState)
		if err != nil {
			return "", "", "", "", ErrorHandler(ctx, err, failedFunction+"GetClustersStatesbyName.rows.scan", friendlyMessage)
		}
		if clusterState != "Deleting" && clusterState != "Deleted" {
			return "", "", "", "", errors.New("Cluster name already in use")
		}
	}

	/* CHECK CLOUDACCOUNT DATA FOR CUSTOM NODE PROVIDER*/
	var cpProviderName string
	err = dbconn.QueryRowContext(ctx, db_query.GetCloudAccountProvider,
		cloudAccountId,
	).Scan(&cpProviderName)
	if err != nil {
		return "", "", "", "", ErrorHandler(ctx, err, failedFunction+"GetCloudAccountProvider", friendlyMessage)
	}

	/* Max Cluster Count for Cloud Account*/
	cloudAccountMaxClusters := -1
	cloudAccountMaxClusters, _, _, _, _, err = GetCloudAccountMaxValues(ctx, dbconn, cloudAccountId)
	if err != nil {
		return "", "", "", "", ErrorHandler(ctx, err, failedFunction+"GetCloudAccountMaxValues", friendlyMessage)
	}
	if cloudAccountMaxClusters > -1 {
		maxClusterCount = cloudAccountMaxClusters
	}

	/* Validate Cluster Count */
	var clustercount int
	err = dbconn.QueryRowContext(ctx, db_query.GetClusterCounts,
		cloudAccountId,
	).Scan(&clustercount)
	if err != nil {
		return "", "", "", "", ErrorHandler(ctx, err, failedFunction+"GetClusterCounts", friendlyMessage)
	}
	if clustercount >= maxClusterCount {
		return "", "", "", "", ErrorHandler(ctx, err, failedFunction+"iksclusterCount", fmt.Sprintf("Can not create more than %d clusters for this cloud account", maxClusterCount))
	}

	/* VALIDATE CONTROL PLANE COMPATABILITY TABLES*/
	// Get the default instance type of CP nodes
	var (
		cpInstanceType string // Default control plane instance type
		cpNodeProvier  string // Default control plane node provider
	)
	err = dbconn.QueryRowContext(ctx, db_query.GetDefaultCpInstanceTypeAndNodeProvider).Scan(&cpInstanceType, &cpNodeProvier)
	if err != nil {
		return "", "", "", "", ErrorHandler(ctx, err, failedFunction+"GetDefaultCpInstanceTypeAndNodeProvider", friendlyMessage)
	}

	// **TEMP** Obtain the OS Image name for CPs (22.04). Will change once customers can select OS Images
	var (
		cpOsImageName string
	)
	err = dbconn.QueryRowContext(ctx, db_query.GetDefaultCpOsImageQuery).Scan(&cpOsImageName)
	if err != nil {
		return "", "", "", "", ErrorHandler(ctx, err, failedFunction+"GetOsImageName", friendlyMessage)
	}

	return cpProviderName, cpInstanceType, cpNodeProvier, cpOsImageName, nil
}

func ValidateCreateNodegroup(ctx context.Context, dbconn *sql.DB, instancetypeid string, count int32, cloudAccountId string, clusteruuid string, ngName string, clusterType string, max_nodegroups int, failedFunction string, friendlyMessage string) (int32, error) {

	/*Validate IKS instance types */
	var instancetype string
	var activeinstances []string
	var instances []string
	var instanceexistsiks bool
	var instanceactive bool
	instanceexistsiks = false
	instanceactive = false

	/* Validate All IKS Instance Types */
	rows, err := dbconn.QueryContext(ctx, db_query.GetInstanceTypeQuery)
	if err != nil {
		return 0, ErrorHandler(ctx, err, failedFunction+"GetIKSInstanceTypesGetQuery", friendlyMessage)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&instancetype)
		if err != nil {
			return 0, ErrorHandler(ctx, err, failedFunction+"GetIKSInstanceTypesRollback", "Get Iks instance types Transaction rollback")
		}
		instances = append(instances, instancetype)
	}

	for _, ins := range instances {
		if ins == instancetypeid {
			instanceexistsiks = true
		}
	}

	if !instanceexistsiks {
		return 0, ErrorHandler(ctx, err, failedFunction+"GetIKSDBInstanceTypes", "Instance type is not supported by IKS")
	}
	/*Validate Active instance types */
	rows, err = dbconn.QueryContext(ctx, db_query.GetActiveInstanceTypeQuery)
	if err != nil {
		return 0, ErrorHandler(ctx, err, failedFunction+"GetIKSInstanceTypesActive", friendlyMessage)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&instancetype)
		if err != nil {
			return 0, err
		}
		activeinstances = append(activeinstances, instancetype)
	}

	for _, activeIns := range activeinstances {
		if activeIns == instancetypeid {
			instanceactive = true
			break
		}
	}

	if !instanceactive {
		return 0, ErrorHandler(ctx, err, failedFunction+"ValidateInstanceTypesActiveState", "Instance Type is not supported by iks")
	}

	/* VALIDATE INSTANCE GROUP */
	isInstanceGroup := IsInstanceGroup(instancetypeid)
	if isInstanceGroup && count > 1 {
		return 0, errors.New("Only one instance group is allowed per Nodegroup")
	}

	/* VALIDATE CLUSTER EXISTANCE only for NON Super Compute Clusters */
	var clusterId int32
	if clusterType != superComputeClusterType {
		clusterId, err = ValidateClusterExistance(ctx, dbconn, clusteruuid)
		if err != nil {
			return 0, ErrorHandler(ctx, err, failedFunction+"ValidateClusterExistance", friendlyMessage)
		}
		if clusterId == -1 {
			return 0, errors.New("Cluster not found: " + clusteruuid)
		}
	}

	/* VALIDATE CLUSTER CLOUD ACCOUNT PERMISSIONS */
	if clusterType != superComputeClusterType {
		isOwner, err := ValidateClusterCloudAccount(ctx, dbconn, clusteruuid, cloudAccountId)
		if err != nil {
			return 0, ErrorHandler(ctx, err, failedFunction+"ValidateClusterCloudAccount", friendlyMessage)
		}
		if !isOwner {
			return 0, status.Errorf(codes.NotFound, "Cluster not found: %s", clusteruuid) // return 404 to avoid leaking cluster existence
		}
	}
	/* VALIDATE CLUSTER IS ACTIONABLE only for NON Super Compute Clusters*/
	if clusterType != superComputeClusterType {
		actionableState := false
		actionableState, err = ValidaterClusterActionable(ctx, dbconn, clusteruuid)
		if err != nil {
			return 0, ErrorHandler(ctx, err, failedFunction+"ValidateClusterActionable", friendlyMessage)
		}
		if !actionableState {
			return 0, errors.New("Cluster not in actionable state")
		}
	}

	/* VALIDATE NODEGROUP UNIQUE NAME */
	ngNameCount := 0
	err = dbconn.QueryRowContext(ctx, db_query.GetNodeGroupCountByName, clusteruuid, ngName).Scan(&ngNameCount)
	if err != nil {
		return 0, ErrorHandler(ctx, err, failedFunction+"GetNodeGroupCountByName", friendlyMessage)
	}
	if ngNameCount != 0 {
		return 0, errors.New("Nodegroup name already in use.")
	}

	/* Get cluster id from cluster  uuid */
	var clusterid int
	if clusterType != superComputeClusterType {
		err = dbconn.QueryRowContext(ctx, db_query.GetClusterIdQuery,
			clusteruuid,
		).Scan(&clusterid)
		if err != nil {
			return 0, ErrorHandler(ctx, err, failedFunction+"Getclusteridfromuuid", friendlyMessage)
		}
	}

	/* Validate Nodegroup Count */
	var nodegroupcount int
	if clusterType != superComputeClusterType {
		err = dbconn.QueryRowContext(ctx, db_query.GetNodeGroupCounts,
			clusterid,
		).Scan(&nodegroupcount)
		if err != nil {
			return 0, ErrorHandler(ctx, err, failedFunction+"GetnodegroupCount", friendlyMessage)
		}

		if nodegroupcount >= max_nodegroups {
			return 0, ErrorHandler(ctx, err, failedFunction+"iksnodegroupCount", fmt.Sprintf("Can not create more than %d nodegroups for this cluster", max_nodegroups))
		}
	}

	return clusterId, nil
}

func ValidateSuperComputeRequestSpec(ctx context.Context, dbconn *sql.DB, req *pb.SuperComputeClusterCreateRequest) error {
	//Validate Cluster Type
	if req.Clustertype != superComputeClusterType {
		return errors.New("SuperCompute Cluster: Invalid Cluster Type in the request")
	}

	// Validate Worker Nodegroup count
	if len(req.Nodegroupspec) <= 0 {
		return errors.New("SuperCompute Cluster: Need Atleast one AI Nodegroup to create a super compute cluster")
	}

	// Validate Storagespec
	if req.Storagespec == nil {
		return errors.New("SuperCompute Cluster: storagespec is required")
	}

	// Validate Storage is Enabled
	if !req.Storagespec.Enablestorage {
		return errors.New("SuperCompute Cluster: Storage should be enabled to create a super compute cluster")
	}

	/*Default Values for IKS */
	defaultvalues, err := GetDefaultValues(ctx, dbconn)
	if err != nil {
		return ErrorHandler(ctx, err, "GetDefaultValues", "Failed to get the default value")
	}

	_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, max_nodegroups, _, _, _, _, err := ConvDefaultsToInt(ctx, defaultvalues)
	if err != nil {
		return ErrorHandler(ctx, err, "convDefaultsToInt", "Failed to convert defaults to int")
	}

	if len(req.Nodegroupspec) >= max_nodegroups {
		return ErrorHandler(ctx, err, "iksnodegroupCount", fmt.Sprintf("Can not create more than %d nodegroups for this cluster", max_nodegroups))
	}

	// Validate Worker Nodegroup Types in the Request
	workerNodesMap := make(map[string]map[string]int)
	ngUniqueName := make(map[string]bool)
	aiNodeFound := false
	if len(req.Nodegroupspec) > 0 {
		for _, v := range req.Nodegroupspec {
			nodegroupType := v.Nodegrouptype
			instanceType := v.Instancetypeid

			if !ValidateScConstant(instanceType) {
				return fmt.Errorf("SuperCompute Cluster: Invalid Instance type for Nodegroup Name %s", v.Name)
			}

			aiNodegroupTypeList := strings.Split(defaultvalues["sc_ai_nodegrouptype"], ",")
			isValidSCAIInstanceType := ValidateSCAiNodegroupInstanceType(instanceType, aiNodegroupTypeList)

			/* VALIDATE NODEGROUP UNIQUE NAME */
			// Update or Insert the Node group Unique Names Map
			if ngUniqueName[v.Name] {
				return fmt.Errorf("SuperCompute Cluster: Duplicate Nodegroup name for %s: Nodegroup Name should be unique", v.Name)
			}
			ngUniqueName[v.Name] = true

			// Check for valid node group type
			if nodegroupType != superComputeAINodegroupType && nodegroupType != superComputeGPNodegroupType {
				return fmt.Errorf("SuperCompute Cluster: Node group type should be either %s or %s", superComputeAINodegroupType, superComputeGPNodegroupType)
			}

			// Check AI Node with the Gaudi Instance Type
			if nodegroupType == superComputeAINodegroupType && isValidSCAIInstanceType {
				if aiNodeFound {
					return errors.New("SuperCompute Cluster: Cannot have more than one AI node with Gaudi Instance Type")
				}
				aiNodeFound = true
			}

			// Check if the instance is of type instance group
			isInstanceGroup := IsInstanceGroup(instanceType)

			// Additional Checks based on node group type
			if nodegroupType == superComputeAINodegroupType {
				// Only add Gaudi Instances to AI node group
				if !isValidSCAIInstanceType {
					return errors.New("SuperCompute Cluster: AI node must have only Gaudi Instance types")
				}

				// Only Instance Groups are accepted for AI nodegroup types
				if !isInstanceGroup {
					return errors.New("SuperCompute Cluster: Invalid Instance Type for AI Node")
				}

				if v.Count > 1 {
					return errors.New("SuperCompute Cluster: Cannot have count more than one instance for AI Node")
				}

				if count, ok := workerNodesMap[nodegroupType]; ok {
					if len(count) > 1 {
						return errors.New("SuperCompute Cluster: Cannot have more than one instance for AI Node")
					}
				}
			} else if nodegroupType == superComputeGPNodegroupType {
				// Gaudi Instances shouldn't be added to GC node group
				if isValidSCAIInstanceType {
					return errors.New("SuperCompute Cluster: Cannot use Gaudi Instance Type for GC nodes")
				}

				// Only NON Instance Groups are accepted for General Purpose nodegroup types
				if isInstanceGroup {
					return errors.New("SuperCompute Cluster: Invalid Instance Type for GC Node")
				}
			}

			// Update the workerNodesMap map
			if _, ok := workerNodesMap[nodegroupType]; !ok {
				workerNodesMap[nodegroupType] = make(map[string]int)
			}
			workerNodesMap[nodegroupType][instanceType]++
		}

		// Check with atleast one AI node with Gaudi instance type was found
		if !aiNodeFound {
			return errors.New("SuperCompute Cluster: Atleast one AI node with Gaudi Instance Type is required")
		}
	}

	// Validate Storage Volume
	if req.Storagespec.Enablestorage {
		if req.Storagespec.Storagesize == "" && !strings.HasSuffix(req.Storagespec.Storagesize, "TB") {
			return errors.New("SuperCompute Cluster: Invalid Volume Size: Volume size cannot be zero or lesser than zero or non TB Values")
		}
	}

	return nil
}

func ValidateSCAiNodegroupInstanceType(instanceType string, aiNodegroupList []string) bool {
	for _, v := range aiNodegroupList {
		if strings.Contains(instanceType, v) {
			return true
		}
	}
	return false
}

func ValidateSCGPNodegroupInstanceType(instanceType string, aiNodegroupList []string) bool {
	for _, v := range aiNodegroupList {
		if strings.Contains(instanceType, v) {
			return false
		}
	}
	return true
}

func ConvDefaultsToInt(ctx context.Context, defaultvalues map[string]string) (int, int, int, int, int, int, int, int, int, int, int, int, int, int, int, int, int, int, int, int, int, int, error) {
	failedFunction := "convDefaultsToIn"
	friendlyMessage := "Could not conver defaults to int"

	/* convert string to Int */
	ilbenv, err := strconv.Atoi(defaultvalues["ilb_environment"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_environment", friendlyMessage)
	}

	ilbusergroup, err := strconv.Atoi(defaultvalues["ilb_usergroup"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_usergroup", friendlyMessage)
	}

	ilbcustomerenv, err := strconv.Atoi(defaultvalues["ilb_customer_environment"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_customerenvironment", friendlyMessage)
	}

	ilbcustomerusergroup, err := strconv.Atoi(defaultvalues["ilb_customer_usergroup"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_customerusergroup", friendlyMessage)
	}

	minActiveMembers, err := strconv.Atoi(defaultvalues["ilb_minactivemembers"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_minactivemembers", friendlyMessage)
	}

	memberConnectionLimit, err := strconv.Atoi(defaultvalues["ilb_memberConnectionLimit"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues._ilb_memberConnectionLimit", friendlyMessage)
	}

	memberPriorityGroup, err := strconv.Atoi(defaultvalues["ilb_memberPriorityGroup"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_memberPriorityGroup", friendlyMessage)
	}

	memberRatio, err := strconv.Atoi(defaultvalues["ilb_memberratio"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_memberratio", friendlyMessage)
	}

	etcdport, err := strconv.Atoi(defaultvalues["ilb_etcdport"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_etcdport", friendlyMessage)
	}

	etcdpool_port, err := strconv.Atoi(defaultvalues["ilb_etcdpool_port"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_etcdpoolport", friendlyMessage)
	}

	apiport, err := strconv.Atoi(defaultvalues["ilb_apiserverport"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_apiport", friendlyMessage)
	}

	apiserverpool_port, err := strconv.Atoi(defaultvalues["ilb_apiserverpool_port"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_apiserverpoolport", friendlyMessage)
	}

	public_apiserverport, err := strconv.Atoi(defaultvalues["ilb_public_apiserverport"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_publicapiserverport", friendlyMessage)
	}

	public_apiserverpool_port, err := strconv.Atoi(defaultvalues["ilb_public_apiserverpool_port"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_publicapiserverpoolport", friendlyMessage)
	}

	konnectPort, err := strconv.Atoi(defaultvalues["ilb_konnectivityport"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_kinnectivityport", friendlyMessage)
	}

	konnectPoolPort, err := strconv.Atoi(defaultvalues["ilb_konnectivitypool_port"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilb_kinnectivitpoolyport", friendlyMessage)
	}

	max_cust_cluster_ilb, err := strconv.Atoi(defaultvalues["max_cust_cluster_ilb"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.max_cust_cluster_ilb", friendlyMessage)
	}

	max_cluster_ng, err := strconv.Atoi(defaultvalues["max_cluster_ng"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.max_cust_cluster_ilb", friendlyMessage)
	}

	max_nodegroup_vm, err := strconv.Atoi(defaultvalues["max_nodegroup_vm"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.max_nodegroup_vm", friendlyMessage)
	}

	split := strings.Split(defaultvalues["ilb_allowed_ports"], ",")

	var ilbportone int
	var ilbporttwo int
	for range split {
		ilbportone, err = strconv.Atoi(string(split[0]))
		if err != nil {
			return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilballowedport", friendlyMessage)
		}

		ilbporttwo, err = strconv.Atoi(split[1])
		if err != nil {
			return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.ilballowedport", friendlyMessage)
		}
	}

	max_cluster, err := strconv.Atoi(defaultvalues["max_cluster"])
	if err != nil {
		return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, ErrorHandler(ctx, err, failedFunction+"strconv.atoi.defaultvalues.max_nodegroup_vm", friendlyMessage)
	}

	return ilbenv, ilbusergroup, ilbcustomerenv, ilbcustomerusergroup, minActiveMembers, memberConnectionLimit, memberPriorityGroup, memberRatio, etcdport, etcdpool_port, apiport, apiserverpool_port, public_apiserverport, public_apiserverpool_port, konnectPort, konnectPoolPort, max_cust_cluster_ilb, max_cluster_ng, max_nodegroup_vm, ilbportone, ilbporttwo, max_cluster, nil
}

func GetAddons(ctx context.Context, dbconn *sql.DB, adminOnly bool, onBuild bool, k8sVersion string, addonType string) ([]DefaultAddons, error) {
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("GetAddons")
	var addons []DefaultAddons
	rows, err := dbconn.QueryContext(ctx, GetAddonsQuery, adminOnly, onBuild, k8sVersion, addonType)
	if err != nil {
		log.Error(err, "\n .. Get addons compatability Transaction rollback")
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		addon := DefaultAddons{}
		err := rows.Scan(&addon.Name, &addon.Type, &addon.Artifact)
		if err != nil {
			log.Error(err, "\n .. Scan addon error!")
			return nil, err
		}
		addons = append(addons, addon)
	}
	return addons, nil
}

func GetClusterStorageStatus(ctx context.Context, dbconn *sql.DB, clusterId int32) ([]*pb.ClusterStorageStatus, error) {
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("GetClusterStorageStatus")
	var storageStatuses []*pb.ClusterStorageStatus

	rows, err := dbconn.QueryContext(ctx, GetClusterStorageStatusQuery, clusterId)
	if err != nil {
		log.Error(err, "\n .. Get Cluster Storage Status")
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		storageStatus := pb.ClusterStorageStatus{}
		var statusCrd clusterv1alpha.StorageStatus
		var statusJson []byte
		err := rows.Scan(&storageStatus.State, &storageStatus.Size, &storageStatus.Storageprovider, &statusJson)
		if err != nil {
			log.Error(err, "\n .. Scan storageStatus error !")
			return nil, err
		}
		if storageStatus.State == "Deleted" {
			continue
		}
		if len(statusJson) > 1 {
			err = json.Unmarshal(statusJson, &statusCrd)
			if err != nil {
				log.Error(err, "\n .. Unmarshall storageStatusJson error")
				return nil, err
			}
			storageStatus.Message = statusCrd.Message
			storageStatus.Reason = statusCrd.Reason
		}
		storageStatus.Size = storageStatus.Size + "TB"
		storageStatuses = append(storageStatuses, &storageStatus)
	}

	return storageStatuses, nil
}

func ValidateStorageEnabled(ctx context.Context, dbconn *sql.DB, clusterUuid string) (bool, error) {
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("ValidateStorageEnabled")
	isEnabled := false

	err := dbconn.QueryRowContext(ctx, GetClusterStorageEnableQuery, clusterUuid).Scan(&isEnabled)
	if err != nil {
		log.Error(err, "\n .. Validate Cluster Storage is enabled")
		return false, err
	}

	return isEnabled, nil
}

func GetDefaultLifeCycleStates(ctx context.Context, dbconn *sql.DB) ([]string, error) {
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("GetDefaultLifeCycleStates")

	var states []string
	var name string
	rows, err := dbconn.QueryContext(ctx, GetDefaultLifeCycleStatesQuery)
	if err != nil {
		log.Error(err, "\n .. Get Default LifeCycle States rollback")
		return nil, err
	}
	defer rows.Close()

	// PARSE ROWSE
	for rows.Next() {
		err := rows.Scan(&name)
		if err != nil {
			log.Error(err, "\n .. Scan default get state name error !")
			return nil, err
		}
		states = append(states, name)
	}

	return states, nil
}

func GetDefaultNodeProviderNames(ctx context.Context, dbconn *sql.DB) ([]string, error) {
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("GetDefaultNodeProviderNames")

	var nodeProvider []string
	var name string
	rows, err := dbconn.QueryContext(ctx, GetDefaultNodeProviderQuery)
	if err != nil {
		log.Error(err, "\n .. Get Default NodeProvider Names rollback")
		return nil, err
	}
	defer rows.Close()

	// PARSE ROWSE
	for rows.Next() {
		err := rows.Scan(&name)
		if err != nil {
			log.Error(err, "\n .. Scan default get nodeprovider name error !")
			return nil, err
		}
		nodeProvider = append(nodeProvider, name)
	}

	return nodeProvider, nil
}

func ConvertToGB(input string) (int, error) {
	input = strings.ToLower(input)

	if strings.HasSuffix(input, "tb") {
		// Handling TB to GB Conversion
		tbValue := strings.TrimSuffix(input, "tb")

		value, err := strconv.ParseFloat(tbValue, 64)
		if err != nil {
			return 0, err
		}

		gb := value * 1000
		return int(math.Round(gb)), nil
	} else if strings.HasSuffix(input, "gb") {
		// Handling GB value directly
		gbValue := strings.TrimSuffix(input, "gb")

		value, err := strconv.ParseFloat(gbValue, 64)
		if err != nil {
			return 0, err
		}
		return int(math.Round(value)), nil
	}

	// If it doesn't match with TB or tb we will simply return zero value
	return 0, fmt.Errorf("input storage value does not contain 'TB' or 'GB'")
}

func ValidateScConstant(instanceType string) bool {
	return strings.Contains(instanceType, superComputeConstant)
}

// Get the current storage for the cluster id
func GetStorageSize(ctx context.Context, dbconn *sql.DB, ClusterId int32) (float64, error) {
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("GetStorageSize")

	var storageSize float64
	err := dbconn.QueryRowContext(ctx, GetStorageSizeQuery, ClusterId).Scan(&storageSize)
	if err != nil {
		log.Error(err, "\n .. Could not obtain the storage size")
		return 0, err
	}

	return storageSize, nil
}

// Get the current total storage size for the cloudaccount id
func GetCloudAccountStorageSize(ctx context.Context, dbconn *sql.DB, CloudAccountId string) (float64, error) {
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("GetCloudAccountStorageSize")

	var totalStorageSize float64
	err := dbconn.QueryRowContext(ctx, GetCloudAccountStorageSizeQuery, CloudAccountId).Scan(&totalStorageSize)
	if err != nil {
		log.Error(err, "\n .. Could not obtain the storage size")
		return 0, err
	}

	return totalStorageSize, nil
}

// Obtain the available sizes for storage from the product catalog
func UpdateAvailableFileSizes(ctx context.Context, producatcatalogClient pb.ProductCatalogServiceClient,
	fsProd *FileProductInfo, fileProductName string) error {
	logger := log.FromContext(ctx).WithName("updateAvailableFileSizes")
	logger.Info("updating storage product configs")

	productFilter := pb.ProductFilter{
		Name: &fileProductName,
	}
	productResponse, err := producatcatalogClient.AdminRead(context.Background(), &productFilter)
	if err != nil {
		return err
	}
	logger.Info("product catalog response", logkeys.Response, productResponse)

	for _, product := range productResponse.Products {
		val, found := product.Metadata["volume.size.min"]
		if !found {
			return fmt.Errorf("product metadata not found")
		}
		min, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fmt.Errorf("error parsing product size")
		}
		fsProd.MinSize = min

		val, found = product.Metadata["volume.size.max"]
		if !found {
			return fmt.Errorf("product metadata not found")
		}
		max, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fmt.Errorf("error parsing product size")
		}
		fsProd.MaxSize = max
	}
	fsProd.UpdatedTimestamp = time.Now()

	return nil
}

func ParseFileSize(size string) float64 {
	var unit string
	if strings.HasSuffix(size, "TB") {
		unit = "TB"
	} else {
		return -1
	}
	// split string to extract numeric value
	splits := strings.Split(size, unit)
	if len(splits) != 2 {
		return -1
	}
	// convert value to bytes
	sizeInt, err := strconv.ParseInt(splits[0], 10, 64)
	if err != nil {
		return -1
	}
	return float64(sizeInt)
}

func ValidateWithProductCatalog(ctx context.Context, productcatalogServiceClient pb.ProductCatalogServiceClient, fileProductName string, totalStorageSize float64) (bool, error) {
	//Obtain the limits from the product catalog
	fileProduct := FileProductInfo{}

	if err := UpdateAvailableFileSizes(ctx, productcatalogServiceClient, &fileProduct, fileProductName); err != nil {
		return false, fmt.Errorf("failed to update storage product details")
	}

	if totalStorageSize < float64(fileProduct.MinSize) || totalStorageSize > float64(fileProduct.MaxSize) {
		return false, fmt.Errorf("Invalid storage size, total size of storage for this CloudAccount ID is outside allowed range")

	}

	return true, nil
}
