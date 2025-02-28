-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
CREATE TABLE IF NOT EXISTS kfaas_deployments
(
    deployment_id character varying(20) COLLATE pg_catalog."default" NOT NULL,
    deployment_name character varying(20) COLLATE pg_catalog."default" NOT NULL,
    kf_version character varying(20) COLLATE pg_catalog."default" NOT NULL,
    k8s_cluster_i_d character varying(30) COLLATE pg_catalog."default" NOT NULL,
    k8s_cluster_name character varying(30) COLLATE pg_catalog."default" NOT NULL,
    storage_class_name character varying(30) COLLATE pg_catalog."default" NOT NULL,
    cloud_account_id VARCHAR(12) NOT NULL,
    status character varying(30) COLLATE pg_catalog."default",
    created_date TIMESTAMP NOT NULL,
    CONSTRAINT kf_deployments_pkey PRIMARY KEY (deployment_id)
)
