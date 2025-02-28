-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
create table if not exists workspace_hw_gateways (
    cloud_account_id text not null,
    lb_fqdn text primary key not null,

    lb_created boolean not null,
    fw_created BOOLEAN NOT NULL,
    gatewayNodeport int not null,

    workspace_id text not null,
    created_at timestamptz NOT NULL DEFAULT (now()),
    updated_at timestamptz NOT NULL DEFAULT (now()),

    is_active bool not null default true
    -- only keep this for testing delete 
    -- FOREIGN KEY (workspace_id) REFERENCES workspace(id)  ON DELETE CASCADE
);

create table if not exists workspace_service_gateways(
    cloud_account_id text not null,

    dns_fqdn text primary key not null,
    gateway_istio_name text not null, 
    gateway_selector_istio_labels text not null, 
    gateway_istio_secret_name text not null, 
    lb_fqdn text not null,
    created_at timestamptz NOT NULL DEFAULT (now()),
    updated_at timestamptz NOT NULL DEFAULT (now()),

    is_active bool not null default true,
    FOREIGN KEY (lb_fqdn) REFERENCES workspace_hw_gateways(lb_fqdn) ON DELETE CASCADE
);

create index if not exists workspace_hw_gateways_dns_lb_fqdn on workspace_hw_gateways(lb_fqdn);
create index if not exists workspace_service_gateways_dns_fqdn on workspace_service_gateways(dns_fqdn);