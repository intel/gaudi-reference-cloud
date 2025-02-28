-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
CREATE TYPE resource_states AS ENUM (
    'REQUESTED',
    'ACCEPTED',
    'PROVISIONING',
    'FAILED',
    'READY',
    'UPDATING'
);

CREATE TYPE registration_states AS ENUM (
    'REQUESTED',
    'ACCEPTED',
    'PROVISIONING',
    'FAILED',
    'READY',
    'DEREGISTERED'
);

CREATE TYPE cluster_types AS ENUM (
    'SLURMSTATIC',
    'PRIVATE'
);

CREATE TYPE machine_types AS ENUM (
    'SML_VM_TYPE',
    'MED_VM_TYPE',
    'LRG_VM_TYPE',
    'PVC_BM_1100_4',
    'PVC_BM_1100_8',
    'PVC_BM_1550_8',
    'GAUDI_BM_TYPE',
    'UNKNOWN'
);

CREATE TYPE node_roles AS ENUM (
    'JUPYTERHUB_NODE',
    'LOGIN_NODE',
    'SLURM_CONTROLLER_NODE',
    'SLURM_COMPUTE_NODE',
    'UNKNOWN'
);

CREATE TYPE access_mode_types AS ENUM (
    'STORAGE_READ_WRITE',
    'STORAGE_READ_ONLY',
    'STORAGE_READ_WRITE_ONCE'
);

CREATE TYPE storage_mount_types AS ENUM (
    'STORAGE_WEKA'
);

CREATE TABLE IF NOT EXISTS cluster (
    id SERIAL PRIMARY KEY,
    cluster_id text NOT NULL,
    cloud_account_id text NOT NULL,
    name text not NULL,
    description text not NULL,
    requested_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz,
    status resource_states NOT NULL,
    events jsonb,
    ssh_key_names jsonb,

    UNIQUE(cluster_id, cloud_account_id)
);

CREATE TABLE IF NOT EXISTS node (
    id SERIAL PRIMARY KEY,
    node_id text NOT NULL,
    name text,
    cloud_account_id text NOT NULL,
    node_role node_roles NOT NULL,
    labels jsonb,
    region text,
    availability_zone text,
    machine_type text NOT NULL,
    image text,
    status resource_states NOT NULL,
    jumphost text,

    UNIQUE (node_id, cloud_account_id)
);

CREATE TABLE IF NOT EXISTS storage (
    id SERIAL PRIMARY KEY,
    fs_resource_id text NOT NULL,
    cloud_account_id text NOT NULL,
    name text NOT NULL,
    description text,
    capacity text NOT NULL,
    access access_mode_types NOT NULL,
    mount storage_mount_types NOT NULL,
    local_mount_dir text NOT NULL,
    remote_mount_dir text NOT NULL,

    UNIQUE (fs_resource_id, name, cloud_account_id)
);

CREATE TABLE IF NOT EXISTS user_training_registration (
    id SERIAL PRIMARY KEY,
    cloud_account_id text NOT NULL,
    registration_id text NOT NULL,
    training_id text NOT NULL,
    ssh_key_names jsonb,
    registered_at timestamptz NOT NULL DEFAULT NOW(), 
    status registration_states NOT NULL,
    jupyter_lab_url text,
    UNIQUE(cloud_account_id, registration_id, training_id)
);

CREATE TABLE IF NOT EXISTS cluster_node_mapping (
    id SERIAL PRIMARY KEY,
    node_id int references node(id) NOT NULL,
    cluster_id int references cluster(id) NOT NULL,
    registered_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_node_mapping (
    id SERIAL PRIMARY KEY,
    user_registration_id int references user_training_registration(id) NOT NULL,
    node_id int references node(id) NOT NULL,
    login_user text NOT NULL,
    ssh_key text NOT NULL,
    
    registered_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS cluster_storage_mapping (
    id SERIAL PRIMARY KEY,
    fs_resource_id int references storage(id) NOT NULL,
    cluster_id int references cluster(id) NOT NULL,
    registered_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS vnet_spec (
    id SERIAL PRIMARY KEY,
    cloud_account_id text NOT NULL,
    cluster_id text NOT NULL,
    region text NOT NULL,
    availabilityZone text NOT NULL,
    prefixLength int NOT NULL,
    UNIQUE (cluster_id, cloud_account_id)
);
