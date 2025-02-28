-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
create table if not exists workspace_size(
    id text primary key,
    name text unique not null,
    size text not null,
    description text,

    is_active bool not null default true,
    created_at timestamptz NOT NULL DEFAULT (now()),
    created_by text not null,
    updated_at timestamptz NOT NULL DEFAULT (now()),
    updated_by text
);

create table if not exists workspace(
    cloud_account_id text not null,
    id text primary key,
    name text not null,
    region text,
    description text,
    tags jsonb,
    
    deployment_id text not null,
    iks_id text,
    management_nodegroup_id text,
    deployment_status_state text not null DEFAULT 'DPAI_ACCEPTED',
    deployment_status_display_name text,
    deployment_status_message text,

    is_active bool not null default true,
    created_at timestamptz NOT NULL DEFAULT (now()),
    created_by text not null,
    updated_at timestamptz NOT NULL DEFAULT (now()),
    updated_by text,
    UNIQUE(cloud_account_id, name)
);

create table if not exists deployment(
    cloud_account_id text,
    workspace_id text, -- will be empty for the service type workspace
    id text primary key,
    parent_deployment_id text,
    service_id text, -- for update and delete service_id will be added
    service_type text not null,
    change_indicator text not null,
    input_payload jsonb,
    output_payload jsonb,
    status_state text not null DEFAULT 'DPAI_ACCEPTED',
    status_display_name text,
    status_message text,
    error_message text,
    node_group_id text,

    created_at timestamptz NOT NULL DEFAULT (now()),
    created_by text not null,
    updated_at timestamptz NOT NULL DEFAULT (now())
);

create table if not exists deployment_task(
    id text not null primary key,
    deployment_id text not null, -- for create acion request_id will be added
    name text not null, -- for update and delete service_id will be added
    description text,
    timeout_in_mins INT DEFAULT 5,
    status_state text not null DEFAULT 'DPAI_ACCEPTED',
    status_display_name text,
    status_message text,
    input_payload jsonb,
    output_payload jsonb,
    error_message text,

    created_at timestamptz NOT NULL DEFAULT (now()),
    updated_at timestamptz,
    started_at timestamptz,
    ended_at timestamptz
);

-- postgres

create table if not exists postgres_size (
    id text PRIMARY key,
    name text not null UNIQUE,
    description text,
    number_of_instances_default int not null,
    number_of_instances_min int,
    number_of_instances_max int,
    resource_cpu_limit text,
    resource_cpu_request text,
    resource_memory_limit text,
    resource_memory_request text,
    number_of_pgpool_instances_default int not null,
    number_of_pgpool_instances_min int,
    number_of_pgpool_instances_max int,
    resource_pgpool_cpu_limit text,
    resource_pgpool_cpu_request text,
    resource_pgpool_memory_limit text,
    resource_pgpool_memory_request text,
    disk_size_in_gb_default int not null,
    disk_size_in_gb_min int,
    disk_size_in_gb_max int,
    storage_class_name text,

    is_active bool not null default true,
    created_at timestamptz NOT NULL DEFAULT (now()),
    created_by text not null,
    updated_at timestamptz,
    updated_by text
);

create table if not exists postgres_version (
    id text PRIMARY KEY,
    name text not null,
    description text,
    version text not null unique,
    postgres_version text not null,
    image_reference text,
    chart_reference jsonb,
    backward_compatible_from text,

    is_active bool not null default true,
    created_at timestamptz NOT NULL DEFAULT (now()),
    created_by text not null,
    updated_at timestamptz,
    updated_by text
);

create table if not exists postgres (
    cloud_account_id text not null,
    workspace_id text not null,
    id text primary key,
    name text not null,
    description text,
    version_id text not null,
    size_id text not null,
    number_of_instances INT,
    number_of_pgpool_instances int,
    disk_size_in_gb int,
    admin_username text not null,
    admin_password_secret_reference jsonb not null,
    advance_configuration jsonb,
    initial_database_name text,
    tags jsonb,
    server_url text,
    
    node_group_id text not null,
    deployment_id text not null,
    deployment_status_state text not null DEFAULT 'DPAI_ACCEPTED',
    deployment_status_display_name text,
    deployment_status_message text,

    is_active bool not null default true,
    created_at timestamptz NOT NULL DEFAULT (now()),
    created_by text not null,
    updated_at timestamptz,
    updated_by text
);

-- hms
create table if not exists hms_conf_group(
    id text not null primary key,
    name text not null unique,
    description text,
    
    is_active bool not null DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT (now()),
    created_by text not null,
    updated_at timestamptz,
    updated_by text
);

create table if not exists hms_conf(
    id text not null primary key,
    hms_id text not null,
    group_id text not null,
    key text not null,
    value text,
    
    is_active bool not null DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT (now()),
    created_by text not null,
    updated_at timestamptz,
    updated_by text
);


create table if not exists hms_size(
    id text not null primary key,
    name text not null,
    description text,
    instance_type_id text,
    number_of_instances_default int not null,
    number_of_instances_min int,
    number_of_instances_max int,

    resource_cpu_limit text,
    resource_cpu_request text,
    resource_memory_limit text,
    resource_memory_request text,

    backend_database_size_id text,
    
    is_active bool not null DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT (now()),
    created_by text not null,
    updated_at timestamptz,
    updated_by text
);

create table if not exists hms_version (
    id text PRIMARY KEY,
    name text not null,
    description text,
    version text not null unique,
    hms_version text not null,
    backend_database_version_id text not null,
    image_reference jsonb,
    chart_reference jsonb,
    backward_compatible_from text,

    is_active bool not null DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT (now()),
    created_by text not null,
    updated_at timestamptz,
    updated_by text
);

create table if not exists hms (
    cloud_account_id text not null,
    workspace_id text not null,
    id text primary key,
    name text not null,
    version_id text not null,
    size_id text not null,
    number_of_instances int,
    description text,
    tags jsonb,
    endpoint text not null,
    object_store_storage_endpoint text,
    object_store_warehouse_directory text not null,
    object_store_storage_access_key_secret_reference jsonb not null,
    object_store_storage_access_secret_secret_reference jsonb not null,
    
    deployment_id text not null,
    backend_database_id text not null,
    node_group_id text not null,
    deployment_status_state text not null DEFAULT 'DPAI_ACCEPTED',
    deployment_status_display_name text,
    deployment_status_message text,
    
    is_active bool not null DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT (now()),
    created_by text not null,
    updated_at timestamptz,
    updated_by text

);

-- airflow 
CREATE TABLE IF NOT EXISTS airflow_size
(
    id text PRIMARY key,
    name text NOT NULL DEFAULT 'small',
    description character varying,
    number_of_nodes_default integer NOT NULL DEFAULT 3,
    node_size_id character varying NOT NULL,
    backend_database_size_id text,
    webserver_count integer DEFAULT 1,
    webserver_cpu_limit character varying,
    webserver_memory_limit character varying,
    webserver_cpu_request character varying,
    webserver_memory_request character varying,
    log_directory_disk_size character varying,
    redis_disk_size character varying,
    schedular_count_default integer DEFAULT 2,
    scheduler_count_min integer,
    scheduler_count_max integer,
    scheduler_cpu_limit character varying,
    scheduler_memory_limit character varying,
    scheduler_memory_request character varying,
    scheduler_cpu_request character varying,
    worker_count_default integer DEFAULT 3,
    worker_count_min integer,
    worker_count_max integer,
    worker_memory_limit character varying,
    worker_memory_request character varying,
    worker_cpu_limit character varying,
    worker_cpu_request character varying,
    trigger_count integer,
    trigger_memory_limit character varying,
    trigger_memory_request character varying,
    trigger_cpu_limit character varying,
    trigger_cpu_request character varying,

    is_active bool not null default true,
    created_at timestamptz NOT NULL DEFAULT (now()),
    created_by text not null,
    updated_at timestamptz,
    updated_by text
);



CREATE TABLE IF NOT EXISTS airflow_version
(
    id text PRIMARY key,
    name character varying NOT NULL,
    version text not null unique,
    backend_database_version_id text not null,
    airflow_version text not null,
    python_version character varying NOT NULL,
    postgres_version character varying NOT NULL,
    redis_version character varying,
    executor_type character varying DEFAULT 'celery',
    image_reference jsonb,
    chart_reference jsonb,
    description character varying,
    backward_compatible_from text,

    is_active bool not null default true,
    created_at timestamptz NOT NULL DEFAULT (now()),
    created_by text not null,
    updated_at timestamptz,
    updated_by text
);

CREATE TABLE IF NOT EXISTS airflow_conf
(
    id text PRIMARY key,
    cloud_account_id text NOT NULL,
    airflow_id character varying NOT NULL,
    key character varying NOT NULL,
    value character varying,

    is_active bool not null default true,
    created_at timestamptz NOT NULL DEFAULT (now()),
    created_by text not null,
    updated_at timestamptz,
    updated_by text
);

CREATE TABLE IF NOT EXISTS airflow
(
    cloud_account_id text NOT NULL,
    id text NOT NULL PRIMARY key,
    name text NOT NULL,
    description text,
    version text NOT NULL,
    tags jsonb,
    bucket_id text,
    bucket_principal text,
    dag_folder_path text,
    plugin_folder_path text,
    requirement_path text,
    log_folder text,
    endpoint  text,
    webserver_admin_username  text,
    webserver_admin_password  text,
    size text NOT NULL,
    number_of_nodes integer,
    number_of_workers integer,
    number_of_schedulers integer,
    deployment_id text not null,
    backend_database_id text not null,
    iks_cluster_id text not null,
    workspace_id text NOT NULL,
    workspace_name text NOT NULL,
    node_group_id text not null,
    deployment_status_state text not null DEFAULT 'DPAI_ACCEPTED',
    deployment_status_display_name text,
    deployment_status_message text,
    is_active bool not null default true,
    created_at timestamptz NOT NULL DEFAULT (now()),
    created_by text not null,
    updated_at timestamptz,
    updated_by text
);

CREATE TABLE IF NOT EXISTS secret
(
    id serial,
    value text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT (now())
);