-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
-- Make sure to update the cloud account id, workspaceid, and the iks cluster id

INSERT INTO workspace (cloud_account_id, id, name, region, description, tags, deployment_id, iks_id, management_nodegroup_id, deployment_status_state, deployment_status_display_name, deployment_status_message, is_active, created_at, created_by, updated_at, updated_by)
VALUES ('513913963588', 'workspaceid', 'blue-bird-02', '', 'this is a demo cluster', NULL, '034eb9f6a622-4244-9c36-9ec9eb7778cb', 'cl-6zobp3cd2i', 'ng-znwuexqome', 'DPAI_SUCCESS', 'Success', 'Successfully provisioned the DPAI Workspace', TRUE, '2024-09-18 11:34:05.77885+00', ' ', '2024-09-18 11:50:38.17094+00', ' ');

-- Make sure to update the cloud account id, workspaceid

INSERT INTO deployment(cloud_account_id, workspace_id, id, service_type, change_indicator, input_payload, status_state, status_display_name, status_message, error_message, node_group_id, created_at, created_by, updated_at)
VALUES ('513913963588', ' ', '034eb9f6a622-4244-9c36-9ec9eb7778cb', 'DPAI_WORKSPACE', 'DPAI_CREATE', '{"name": "blue-bird-02", "description": "this is a demo cluster", "cloudAccountId": "513913963588"}', 'DPAI_SUCCESS', 'Deployment in progress', 'Deployment: CreateAirflow:airflow-aftest-b0c7c40972431a961f7c in progress', ' ', ' ', '2024-09-05 16:33:18.144093+00', 'Joydeep', '2024-09-05 16:33:18.144093+00');


INSERT INTO postgres_size
VALUES ('fda46be1-ca28-4a99-ac76-fd2cea3f568b', 'small', 'Best suited for dev and test environment', 3, 2, 5, '500m', '250m', '1G', '512Mi', 2, 1, 5, '500m', '250m', '512Mi', '256Mi', 1, 1, 5, 'StorageClass', TRUE, '2024-02-29 16:33:18.144093+00', 'CallerOfTheAPI', '2024-02-29 16:33:18.144093+00', 'CallerOfTheAPI');


INSERT INTO postgres_version
VALUES ('d72c9b86-feb9-4985-b711-b083be1fb885', '0.1', 'Patch Version upgrade', '0.1', '16.1.0', ' ', '{"repoUrl": https://charts.bitnami.com/bitnami, "version": "12.3.7", "repoName": "bitnami", "chartName": "bitnami/postgresql-ha"}', ' ', TRUE, '2024-02-29 16:33:18.144093+00', 'CallerOfTheAPI', '2024-02-29 16:33:18.144093+00', 'CallerOfTheAPI');