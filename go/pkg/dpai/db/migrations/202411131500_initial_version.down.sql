-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
drop table if exists workspace_service_gateways;

drop table if exists workspace_hw_gateways;

drop index if exists workspace_hw_gateways_dns_lb_fqdn;
drop index if exists workspace_service_gateways_dns_fqdn;

drop table if exists workspace;

drop table if exists workspace_size;

drop table if exists deployment;

drop table if exists deployment_task;

-- postgres
drop table if exists postgres_size;

drop table if exists postgres_version;

DROP table if exists postgres;

-- hms
drop table if exists hms_conf_group;

drop table if exists hms_conf;

drop table if exists hms_size;

drop table if exists hms_version;

drop table if exists hms;

-- airflow
drop table if exists airflow_size;

drop table if exists airflow_version;

drop table if exists airflow_conf;

drop table if exists airflow;