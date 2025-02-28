-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation

CREATE TABLE IF NOT EXISTS cloudmonitorregistrationstable
(
    registration_id character varying(20) COLLATE pg_catalog."default" NOT NULL,
    name character varying(20) COLLATE pg_catalog."default" NOT NULL,
    source character varying(20) COLLATE pg_catalog."default" NOT NULL,
    destination character varying(30) COLLATE pg_catalog."default" NOT NULL,
    status character varying(30) COLLATE pg_catalog."default",
    cloud_account_id VARCHAR(12) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    CONSTRAINT cloudmonitorregistrationstable_pkey PRIMARY KEY (registration_id),
    CONSTRAINT cloudmonitorregistrationstable_ukey UNIQUE (cloud_account_id)
);


CREATE TABLE IF NOT EXISTS cloudmonitorresourcestable
(
    cloudmonitor_resource_id VARCHAR(40) COLLATE pg_catalog."default" UNIQUE NOT NULL,
    resource_id VARCHAR(40) COLLATE pg_catalog."default" NOT NULL,
    resource_type VARCHAR(12) NOT NULL,
    registration_id character varying(20) COLLATE pg_catalog."default" NOT NULL,
    cloud_account_id VARCHAR(12) NOT NULL,
    category character varying(30) COLLATE pg_catalog."default" NOT NULL,
    status character varying(30) COLLATE pg_catalog."default",
    created_at TIMESTAMP NOT NULL,
    CONSTRAINT cloudmonitorresourcestable_pkey PRIMARY KEY (cloudmonitor_resource_id,resource_type,category)
);

CREATE TABLE IF NOT EXISTS resourcetypes (
   resourcetype VARCHAR(27) PRIMARY KEY,
   resourcetype_id INT NOT NULL,
   CONSTRAINT resourcetypes_ukey UNIQUE (resourcetype_id)
);

INSERT INTO resourcetypes (resourcetype, resourcetype_id)
VALUES('VM',1) ON CONFLICT DO NOTHING;


CREATE TABLE IF NOT EXISTS intervals (
   interval VARCHAR(63) ,
   resourcetype_id INT NOT NULL,
    FOREIGN KEY (resourcetype_id)
       REFERENCES resourcetypes (resourcetype_id)
);
INSERT INTO intervals (interval, resourcetype_id)
VALUES('Last 6 hours',1),('Last 12 hours',1),('Last 18 hours',1),('Last 24 hours',1),('Last 7 days',1) ON CONFLICT DO NOTHING;


CREATE TABLE IF NOT EXISTS metricstypes (
   metricstype VARCHAR(63) ,
    resourcetype_id INT NOT NULL,
    FOREIGN KEY (resourcetype_id)
       REFERENCES resourcetypes (resourcetype_id)
);
INSERT INTO metricstypes (metricstype, resourcetype_id)
VALUES('cpu',1),('memory',1),('network_receive_bytes',1),('network_transmit_bytes','1'),('storage_read_traffic_bytes','1'),('storage_write_traffic_bytes','1'),('storage_iops_read_total','1'),('storage_iops_write_total','1'),('storage_read_times_ms_total','1'),('storage_read_times_ms_total','1') ON CONFLICT DO NOTHING;