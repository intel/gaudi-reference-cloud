// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package db_query_constants

// Cluster Queries
var (
	// insert queries
	InsertClusterRecordQuery = `
 		INSERT INTO public.cluster (clusterstate_name, name, description, region_name, networkservice_cidr,
			cluster_dns, networkpod_cidr, encryptionconig, advanceconfigs, provider_args, backup_args, provider_name,
			labels, tags, annotations, backuptype_name, created_date, unique_id, cloudaccount_id, kubernetes_status, clustertype) 
 		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING cluster_id
 	`
)

// Cluster Rev Queries
var (
	InsertRevQuery = `
	 INSERT INTO public.clusterrev (cluster_id, currentspec_json, desiredspec_json, component_typegrp, 
		component_typename, currentdata, desireddata, timestamp, change_applied) 
	 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING clusterrev_id
 	`
)

// Provisioning Log Queries
var (
	InsertProvisioningQuery = `
	INSERT INTO public.provisioninglog (cluster_id, logentry, loglevel_name,logobject,logtimestamp)
	VALUES ($1,$2,$3,$4,$5)
	`
)

// SSH Queries
var (
	InsertSshkeyQuery = `
	 INSERT INTO public.cluster_extraconfig (cluster_id,cluster_ssh_key,cluster_ssh_pub_key,cluster_ssh_key_name,nonce,encryptionkey_id) VALUES ($1,$2,$3,$4,$5,$6)
 	`
)

// Storage
var (
	InsertStorageTableQuery = `
		INSERT INTO public.storage (cluster_id, storageprovider_name, storagestate_name, size) values ($1, $2, $3, $4)
	`
)

// Nodegroup Queries
var (
	InsertControlPlaneQuery = `
	 INSERT INTO public.nodegroup (cluster_id, osimageinstance_name, k8sversion_name, 
		 nodegroupstate_name, nodegrouptype_name, name, description, 
		 networkinterface_name, instancetype_name, nodecount, sshkey, runtime_name,
		 tags, upgstrategydrainbefdel, upgstrategymaxnodes, statedetails, 
		 createddate, unique_id, lifecyclestate_id, kubernetes_status, vnets, nodegrouptype)
	 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21, $22)
	 RETURNING nodegroup_id
	`

	InsertNodeGroupQuery = `
	INSERT INTO public.nodegroup (cluster_id, osimageinstance_name, k8sversion_name,
		nodegroupstate_name, nodegrouptype_name, name, description, 
		networkinterface_name, instancetype_name, nodecount, sshkey, runtime_name,
		tags, upgstrategydrainbefdel, upgstrategymaxnodes, statedetails,
		createddate, unique_id, lifecyclestate_id ,annotations, kubernetes_status, vnets, userdata_webhook, nodegrouptype)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)
	RETURNING nodegroup_id
	`
)

// VIP Queries
var (
	InsertVipQuery = `
	 INSERT INTO public.vip (cluster_id, dns_aliases, vip_dns,viptype_name, vip_status, vipstate_name, owner,vipprovider_name,vipinstance_id)
	 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING vip_id`

	InsertVipDetailsQuery = `
		 INSERT INTO public.vipdetails (vip_id, vip_name, description, port, pool_name, pool_port)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING vip_id
	`
)
