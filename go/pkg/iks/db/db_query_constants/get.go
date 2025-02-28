// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package db_query_constants

// Cluster Queries
var (
	GetClustersStatesByName = `
	SELECT unique_id, clusterstate_name
	FROM public.cluster
	WHERE name = $1 and cloudaccount_id = $2
	`
	GetClusterCounts = `
	SELECT count(*) FROM public.cluster WHERE cloudaccount_id = $1 AND clusterstate_name NOT IN ('Deleting','DeletePending','Deleted')
	`
	GetClusterIdQuery = `
	SELECT cluster_id FROM public.cluster where unique_id = $1
  	`
	GetClusterTypeQuery = `
	SELECT cluster_id, clustertype FROM public.cluster where unique_id = $1
	`
)

// Cloud Account ExtraSpec Table Queries
var (
	GetCloudAccountProvider = `
	SELECT coalesce(
		(	SELECT provider_name 
			FROM public.cloudaccountextraspec c 
			WHERE cloudaccount_id = $1
		),'iks') AS provider_name  
	`
)

// Firewall Rule Table Queries
var (
	GetPublicVipsForFirewallQuery = `
	SELECT v.vip_id, v.vip_ip , d.port, d.pool_port, v.sourceips, COALESCE(v.firewall_status,'Not Specified'), v.viptype_name ,d.protocol , d.vip_name
	FROM public.vip v
	    INNER JOIN public.vipdetails d ON
		v.vip_id = d.vip_id
	WHERE v.cluster_id = $1 AND v.vip_id = $2 AND v.viptype_name = 'public';
	`
)

// Instance Types Table Queries
var (
	GetDefaultCpInstanceTypeAndNodeProvider = `
	SELECT instancetype_name , nodeprovider_name
	FROM public.instancetype
	WHERE is_default = true and nodeprovider_name = (SELECT nodeprovider_name from public.nodeprovider where is_default = true)
	`
	GetInstanceTypeQuery = `
	SELECT i.instancetype_name
	FROM instancetype i 
	WHERE i.nodeprovider_name = (SELECT nodeprovider_name from public.nodeprovider where is_default = true) 
	`
	GetActiveInstanceTypeQuery = `
	SELECT i.instancetype_name FROM instancetype i where i.lifecyclestate_id = (
		SELECT lifecyclestate_id
		FROM lifecyclestate
		WHERE name = 'Active') and i.nodeprovider_name = (SELECT nodeprovider_name from public.nodeprovider where is_default = true)
	`
)

// OsImage Table Queries
var (
	GetDefaultCpOsImageQuery = `
	SELECT osimage_name
	FROM osimage
	WHERE cp_default='true' AND lifecyclestate_id=(SELECT lifecyclestate_id FROM lifecyclestate WHERE name='Active')
	`
)

// OsImage Instance Table Queries
var (
	GetImiArtifactQuery = `
	SELECT imiartifact FROM public.osimageinstance WHERE osimageinstance_name = $1
	`
)

// K8s Compatibility Queries
var (
	// Validation Queries
	// APPENDS ADDITIONAL 'AND' ON CREATE CLUSTER
	GetDefaultCpOsImageInstance = `
		SELECT kc.cp_osimageinstance_name, kc.k8sversion_name
		FROM public.k8scompatibility kc
			INNER JOIN public.k8sversion ks 
			ON kc.k8sversion_name = ks.k8sversion_name AND kc.provider_name = $2
		WHERE kc.runtime_name = $1 AND NOT ks.test_version AND kc.instancetype_name = $3 AND kc.osimage_name = $4
		AND ks.lifecyclestate_id = (SELECT l.lifecyclestate_id FROM lifecyclestate l WHERE l.name = 'Active') 
	`
	// APPENDS ADDITIONAL 'AND' ON CREATE NODEGROUP
	GetDefaultWrkOsImageInstance = `
		SELECT kc.wrk_osimageinstance_name
		FROM public.k8scompatibility kc
			INNER JOIN public.k8sversion ks 
			ON kc.k8sversion_name = ks.k8sversion_name AND kc.provider_name = $2
		WHERE kc.runtime_name = $1 AND NOT ks.test_version AND kc.instancetype_name = $3 AND kc.osimage_name = $4
	`
)

// Nodegroup Queries
var (
	GetNodeGroupCountByName = `
	SELECT count(nodegroup_id) as count
	FROM public.nodegroup n
	WHERE n.name = $2 and n.nodegrouptype_name = 'Worker' and cluster_id = (select cluster_id from public."cluster" c where c.unique_id = $1)
	`
	GetNodeGroupCounts = `
	  SELECT count(*) FROM public.nodegroup WHERE cluster_id = $1 and nodegrouptype_name = 'Worker'
	`
	GetNodeGroupTypeQuery = `
	SELECT n.cluster_id, n.nodegroup_id, n.nodegrouptype
	 FROM public.nodegroup n
	WHERE n.unique_id = $2 AND n.cluster_id = (
		SELECT c.cluster_id 
		FROM public.cluster  c
		WHERE c.unique_id = $1 AND c.clusterstate_name != 'Deleted'
	)
	`
	GetNodeGroupTypeAndInstanceTypeQuery = `
	SELECT n.cluster_id, n.nodegroup_id, n.nodegrouptype, n.instancetype_name
	 FROM public.nodegroup n
	WHERE n.unique_id = $2 AND n.cluster_id = (
		SELECT c.cluster_id 
		FROM public.cluster  c
		WHERE c.unique_id = $1 AND c.clusterstate_name != 'Deleted'
	)
	`
)

// Storage Table
var (
	GetDefaultStorageProvider = `
		SELECT storageprovider_name 
		FROM storageprovider
		WHERE is_default=true
	`
)

// Vip Provider Queries
var (
	GetVipProviderDefaults = `
	SELECT vipprovider_name FROM public.vipprovider where is_default = 'true'
	`
	GetVipidByClusterQuery = `
	SELECT vip_id FROM public.vip where cluster_id = $1 and viptype_name = 'public' and vipstate_name = 'Active'`
)
