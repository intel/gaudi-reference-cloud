-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
-- CREATE TABLE IF NOT EXISTS cloudmonitorbmmapping
-- (
--     resource_id VARCHAR(40) COLLATE pg_catalog."default" NOT NULL,
--     job_name character varying(20) COLLATE pg_catalog."default" NOT NULL,
--     opt_in_status character varying(12) COLLATE pg_catalog."default",
--     cloud_account_id VARCHAR(12) NOT NULL,
--     created_at TIMESTAMP NOT NULL,
--     CONSTRAINT cloudmonitorbmmapping_pkey PRIMARY KEY (resource_id,cloud_account_id),
--     CONSTRAINT cloudmonitorbmmapping_ukey UNIQUE (cloud_account_id)
-- );


CREATE TABLE IF NOT EXISTS cloudmonitorcloudaccountmapping
(
    cloudmonitorid serial,
    cloud_account_id VARCHAR(12) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    CONSTRAINT cloudmonitorcloudaccountmapping_pkey PRIMARY KEY (cloudmonitorid),
    CONSTRAINT cloudmonitorcloudaccountmapping_ukey UNIQUE (cloud_account_id)
);
