-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
-- name: GetWorkspace :one
SELECT 
  w.cloud_account_id, 
  w.id, 
  w.name, 
  w.region, 
  w.description, 
  w.tags, 
  w.deployment_id, 
  w.iks_id, 
  w.iks_cluster_name,
  w.ssh_key_name,
  w.management_nodegroup_id,
  d.status_state as deployment_status_state,
  d.status_display_name as deployment_status_display_name,
  d.status_message as deployment_status_message,
  w.created_at, 
  w.created_by, 
  d.updated_at, 
  w.updated_by
FROM workspace w
inner join deployment d 
on d.id = w.deployment_id
WHERE w.id = sqlc.arg('workspace_id')
  AND d.cloud_account_id = coalesce(sqlc.narg('cloud_account_id'), d.cloud_account_id)
  AND w.is_active = true;


-- name: ListWorkspaces :many
SELECT
  w.cloud_account_id, w.id, w.name, w.region, w.description, w.tags, w.deployment_id, w.iks_id, w.iks_cluster_name, w.ssh_key_name, w.management_nodegroup_id,
  d.status_display_name as deployment_status, d.status_message as deployment_status_message, d.status_state as deployment_status_state,
  w.created_at, w.created_by, w.updated_at, w.updated_by, w.is_active 
FROM workspace w 
join deployment d
on w.deployment_id = d.id
WHERE w.cloud_account_id = $1
AND w.name = coalesce(sqlc.narg('name'), w.name)
AND (d.status_display_name = coalesce(sqlc.narg('status_display_name'), d.status_display_name) or d.status_display_name is null)
AND w.is_active = true
ORDER BY name
LIMIT sqlc.arg('limit') 
OFFSET sqlc.arg('offset');

-- name: CountListWorkspaces :one
SELECT count(1)
FROM workspace w 
join deployment d
on w.deployment_id = d.id
WHERE w.cloud_account_id = $1
AND w.name = coalesce(sqlc.narg('name'), w.name)
AND (d.status_display_name = coalesce(sqlc.narg('status_display_name'), d.status_display_name) or d.status_display_name is null)
AND w.is_active = true;

-- -- name: CreateWorkspaceRequest :one
-- INSERT INTO workspace_form (
--   id, cloud_account_id, name, region, size_id, description, tags, created_by
-- ) VALUES (
--   $1, $2, $3, $4, $5, $6, $7, $8
-- )
-- RETURNING *;

-- -- name: GetWorkspaceFormById :one
-- SELECT * FROM workspace_form
-- WHERE id = $1 LIMIT 1;

-- name: CreateWorkspace :one
INSERT INTO workspace (
  id, cloud_account_id, name, region, description, tags, deployment_id, deployment_status_state, deployment_status_display_name, deployment_status_message, created_by
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: GetUpdateWorkspace :one
SELECT id, description, tags FROM workspace
WHERE  id = $1 
AND is_active = true
LIMIT 1;

-- name: UpdateWorkspace :one
UPDATE workspace
  set description = $2,
  tags = $3,
  updated_at = now()
WHERE id = $1
AND is_active = true
RETURNING *;

-- name: UpdateWorkspaceDeploymentStatus :one
UPDATE workspace w
  set deployment_status_state = coalesce(sqlc.narg('deployment_status_state'), d.status_state),
  deployment_status_display_name = coalesce(sqlc.narg('deployment_status_display_name'), d.status_display_name),
  deployment_status_message = coalesce(sqlc.narg('deployment_status_message'), d.status_message),
  iks_id = coalesce(sqlc.narg('iks_id'), iks_id),
  iks_cluster_name = coalesce(sqlc.narg('iks_cluster_name'), iks_cluster_name),
  management_nodegroup_id = coalesce(sqlc.narg('management_nodegroup_id'), management_nodegroup_id),
  ssh_key_name = coalesce(sqlc.narg('ssh_key_name'), ssh_key_name),
  updated_at = now()
FROM deployment d
WHERE (w.id = coalesce(sqlc.narg('id'), 'PLACEHOLDER') or w.deployment_id = coalesce(sqlc.narg('deployment_id'), 'PLACEHOLDER'))
and w.deployment_id = d.id
AND is_active = true
RETURNING *;

-- name: DeleteWorkspace :exec
UPDATE workspace
set is_active=false,
  updated_at = now()
WHERE id = $1
AND is_active = true;

-- name: ListWorkspaceServices :many

SELECT 
  s.id, 
  s.name, 
  s.version, 
  s.deployment_id, 
  d.service_type, 
  d.status_state,
  d.status_display_name,
  d.status_message,
  COALESCE(d.updated_at, d.created_at) AS service_updated_at
FROM (
  SELECT 
    id, 
    name, 
    version, 
    deployment_id
  FROM airflow
  WHERE cloud_account_id = coalesce(sqlc.narg('cloud_account_id'), cloud_account_id)
    AND workspace_id = sqlc.arg('workspace_id')
  UNION
  SELECT 
    id, 
    name, 
    version_id as version, 
    deployment_id
  FROM hms
  WHERE cloud_account_id = coalesce(sqlc.narg('cloud_account_id'), cloud_account_id)
    AND workspace_id = sqlc.arg('workspace_id')
) s 
INNER JOIN deployment d ON s.deployment_id = d.id
WHERE d.cloud_account_id = coalesce(sqlc.narg('cloud_account_id'), cloud_account_id)
  AND d.workspace_id = sqlc.arg('workspace_id');

-- name: CheckUniqueWorkspaceId :one
select count(*)  as cnt
from workspace
where id like $1 || '%';

-- name: GetWorkspaceIdFromServiceId :one
SELECT workspace_id FROM postgres
WHERE id = $1
union
SELECT workspace_id FROM hms
WHERE id = $1
union
  select workspace_id from airflow
  where id = $1;

-- name: GetClusterIdFromServiceId :one
SELECT cloud_account_id, iks_id FROM workspace w
inner join (
  SELECT workspace_id FROM postgres
  WHERE id = $1
  union
  SELECT workspace_id FROM hms
  WHERE id = $1
  union
  select workspace_id from airflow
  where id = $1
) s
ON s.workspace_id = w.id;

-- name: GetHwGatewaysByWorkspaceId :one
select * from workspace_hw_gateways where workspace_id = $1 and cloud_account_id = $2 limit 1; 

-- name: GetHwGatewaysByWorkspaceIdandLB :one
select * from workspace_hw_gateways where workspace_id = $1 and lb_fqdn = $2 and cloud_account_id = $3 limit 1;

-- name: GetGatewaysByWorkspaceIdandLBCreated :one
select * from workspace_hw_gateways where workspace_id = $1 and lb_fqdn = $2 and cloud_account_id = $3 and lb_created limit 1;

-- name: GetGatewaysByWorkspaceIdCloudAccountandDnsLBReady :one 
select * from workspace_hw_gateways where workspace_id = $1 and cloud_account_id = $2 and lb_created and fw_created limit 1;

-- name: GetGatewaysByWorkspaceIdandLBReady :one
select * from workspace_hw_gateways where workspace_id = $1 and cloud_account_id = $2 and lb_created and fw_created limit 1;

-- name: UpdateGatewayFwStatusForWorkspace :one
update workspace_hw_gateways set fw_created = $4 where workspace_id = $1 and lb_fqdn = $2 and cloud_account_id = $3 RETURNING *;

-- name: GetLbFqdnForWorkspace :one 
select distinct(lb_fqdn) from workspace_hw_gateways where cloud_account_id = $1 and workspace_id = $2 limit 1;

-- name: InsertGatewayForWorkspace :one
insert into workspace_hw_gateways
  (cloud_account_id, lb_fqdn, lb_created, fw_created, gatewayNodeport, workspace_id, created_at, updated_at, is_active)
VALUES (
    $1, $2, $3, $4, $5, $6, now(), now(), true
) RETURNING 1;

-- name: InsertGatewayForWorkspaceService :one
INSERT INTO workspace_service_gateways
  (cloud_account_id, dns_fqdn, gateway_istio_name, gateway_selector_istio_labels, gateway_istio_secret_name, lb_fqdn, created_at, updated_at, is_active)
VALUES (  
  $1, $2, $3, $4, $5, $6, now(), now(), $7
) RETURNING 1;


-- name: GetHwGatewayForWorkspaceService :one 
select workspace_hw_gateways.lb_fqdn, workspace_hw_gateways.gatewayNodeport
  from workspace_hw_gateways inner join workspace on workspace_hw_gateways.workspace_id = workspace.id
  WHERE workspace.id = $1
LIMIT 1;

-- name: GetServiceGatewayForWorkspaceService :one 
SELECT workspace_service_gateways.cloud_account_id, workspace_service_gateways.dns_fqdn, workspace_service_gateways.gateway_istio_name,
    workspace_service_gateways.gateway_selector_istio_labels, workspace_service_gateways.gateway_istio_secret_name, workspace_service_gateways.is_active FROM (
  SELECT workspace_hw_gateways.lb_fqdn
      FROM workspace_hw_gateways 
      INNER JOIN workspace ON workspace_hw_gateways.workspace_id = workspace.id
      WHERE workspace.id = $1
  ) as tm
  INNER JOIN workspace_service_gateways ON workspace_service_gateways.lb_fqdn = tm.lb_fqdn
LIMIT 1;


-- name: GetServiceGatewayForWorkspaceServiceFromDnsFqdn :one
select * from workspace_service_gateways 
  where cloud_account_id = $1 and dns_fqdn = $2
limit 1;

--  DOES A SINGLE ROW DELETE 

-- name: DeleteGatewayForWorkspaceService :exec 
delete from workspace_service_gateways 
  where cloud_account_id = $1 
    and dns_fqdn = $2;


-- name: GatGatewayForWorkspaceService :one 
select * from workspace_service_gateways
  where cloud_account_id = $1 and dns_fqdn = $2
  limit 1;


-- DOES A CASCADE DELETE TO REMOVE ALL THE ROWS WHICH THIS LB FQDN CREATED BY THE PARENT WORKSAPCE 

-- name: DeleteHwGatewayForWorkspace :exec
delete from workspace_hw_gateways
  where cloud_account_id = $1 and workspace_id = $2 and lb_fqdn = $3; 
  