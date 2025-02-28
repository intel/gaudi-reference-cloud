-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
--- a table for k8s security insights service
CREATE TYPE vendors as ENUM(
    'unspecified',
    'oss',
    'rancher'
);

CREATE TYPE component_type as ENUM(
    'image',
    'compressed-bundle',
    'file',
    'gitrepo',
    'unspecified'
);

CREATE TABLE IF NOT EXISTS k8s_release (
    id BIGSERIAL PRIMARY KEY,
    version TEXT NOT NULL,
    vendor vendors NOT NULL,
    license TEXT NOT NULL,
    PURL TEXT NOT NULL,
    release_timestamp TIMESTAMP NOT NULL,
    eos_timestamp TIMESTAMP NOT NULL,
    eol_timestamp TIMESTAMP NOT NULL,
    properties jsonb,
    
    UNIQUE (vendor, version)
);

CREATE TABLE IF NOT EXISTS k8s_release_components (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    sha256 TEXT NOT NULL,
    license TEXT NOT NULL,
    type component_type NOT NULL,
    url TEXT NOT NULL,
    release_timestamp TIMESTAMP NOT NULL,
    release_id INT references k8s_release(id),
    UNIQUE (release_id,name,version)
);

CREATE TABLE IF NOT EXISTS k8s_release_sbom (
    id BIGSERIAL PRIMARY KEY,
    create_timestamp TIMESTAMP NOT NULL,
    format TEXT NOT NULL,
    release_id INT references k8s_release(id) NOT NULL,
    component_id INT references k8s_release_components(id),
    sbom TEXT
);

CREATE TABLE IF NOT EXISTS vulnerability (
    id BIGSERIAL PRIMARY KEY,
    cveid TEXT NOT NULL,
    severity TEXT NOT NULL,
    description TEXT NOT NULL,
    affected_package TEXT NOT NULL,
    affected_version TEXT NOT NULL,
    fixed_version TEXT NOT NULL,
    published_at TIMESTAMP NOT NULL,

    UNIQUE (cveid)
);

CREATE TABLE IF NOT EXISTS vulnerability_scan_schedule (
    id BIGSERIAL PRIMARY KEY,
    release_id INT references k8s_release(id) NOT NULL,
    component_id INT references k8s_release_components(id),
    scan_tool TEXT NOT NULL,
    scan_at TIMESTAMP NOT NULL 
);

CREATE TABLE IF NOT EXISTS vulnerability_scan_report (
    id BIGSERIAL PRIMARY KEY,
    scan_id INT references vulnerability_scan_schedule(id) NOT NULL,
    vulnerability_id INT references vulnerability(id) NOT NULL
);

CREATE TABLE IF NOT EXISTS update_policy (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    active boolean NOT NULL,
    policy jsonb NOT NULL,
    sha256 TEXT NOT NULL,

    UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS cis_report (
    id BIGSERIAL PRIMARY KEY,
    release_id INT references k8s_release(id) NOT NULL,
    scan_tool TEXT NOT NULL,
    scan_at TIMESTAMP NOT NULL,
    report jsonb NOT NULL
);

CREATE INDEX  IF NOT EXISTS k8s_release_version_idx ON k8s_release (version, vendor);

CREATE INDEX  IF NOT EXISTS k8s_release_components_idx ON k8s_release_components (release_id);

CREATE INDEX  IF NOT EXISTS k8s_release_sbom_idx ON k8s_release_sbom (release_id, component_id);

CREATE INDEX  IF NOT EXISTS vulnerability_idx ON vulnerability (cveid);
