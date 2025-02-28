-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation

drop index if exists workspace_hw_gateways_dns_lb_fqdn;
drop index if exists workspace_service_gateways_dns_fqdn;

drop table if exists workspace_service_gateways;
drop table if exists workspace_hw_gateways;
