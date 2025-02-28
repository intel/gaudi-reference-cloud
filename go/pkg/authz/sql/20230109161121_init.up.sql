-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE effect_type AS ENUM ('allow', 'deny');

CREATE TABLE cloud_account_roles (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP WITH TIME ZONE,
    cloud_account_id VARCHAR(255),
    alias VARCHAR(255),
    effect          effect_type,
    users           JSONB NOT NULL
);

CREATE UNIQUE INDEX idx_cloud_account_roles_cloud_account_id_alias_unique ON cloud_account_roles (cloud_account_id, alias);
CREATE INDEX idx_cloud_account_roles_users ON cloud_account_roles USING GIN (users);

CREATE TABLE permissions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    cloud_account_role_id UUID REFERENCES cloud_account_roles(id),
    cloud_account_id    VARCHAR(255) NOT NULL,
    resource_type     VARCHAR(255) NOT NULL,
    resource_id       VARCHAR(255) NOT NULL,
    actions         JSONB NOT NULL
);

CREATE UNIQUE INDEX idx_permissions_cloud_account_role_id_resource_type_id_unique ON permissions (cloud_account_role_id, resource_type, resource_id);
CREATE INDEX idx_permissions_actions ON permissions USING GIN (actions);

CREATE TYPE event_type AS ENUM ('check', 'create', 'update', 'add_resource', 'add_user', 'remove_resource', 'remove_user', 'assign_sys_role', 'remove_sys_role', 'delete', 'lookup', 'query', 'get','add_user_ca_roles','remove_user_ca_roles');

CREATE TABLE audit (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_date              TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    cloud_account_id        VARCHAR(255),
    cloud_account_role_ids   UUID[],
    event_type              event_type,
    additional_info         JSONB
);

CREATE TABLE default_role_assignments (
    cloud_account_id        VARCHAR(255) PRIMARY KEY,
    assigned_at             TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    admins                  JSONB NOT NULL,
    members                 JSONB NOT NULL DEFAULT '[]'::JSONB
);
