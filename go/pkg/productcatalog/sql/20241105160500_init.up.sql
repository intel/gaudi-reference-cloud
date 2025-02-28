-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation

CREATE TABLE IF NOT EXISTS vendor (
    id uuid NOT NULL UNIQUE,
    name VARCHAR(255) PRIMARY KEY,
    description VARCHAR(255),
    organization_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS product_family (
    id uuid NOT NULL UNIQUE,
    name VARCHAR(255) PRIMARY KEY,
    vendor_name VARCHAR(255) NOT NULL,
    description VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (vendor_name) REFERENCES vendor(name)
);

CREATE TABLE IF NOT EXISTS intel_cloud_service (
    name VARCHAR(255) PRIMARY KEY,
    description VARCHAR(255),
    admin_name VARCHAR(255) DEFAULT 'system',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS product (
    id VARCHAR(36) UNIQUE,
    name VARCHAR(255),
    family_name VARCHAR(255) NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    status VARCHAR(255),
    usage VARCHAR(255),
    change_request_status VARCHAR(255),
    enabled BOOLEAN DEFAULT TRUE,
    version VARCHAR(255) DEFAULT 'v.1.0.0',
    environment VARCHAR(255),
    admin_name VARCHAR(255) DEFAULT 'system',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (name),
    FOREIGN KEY (service_name) REFERENCES intel_cloud_service(name),
    FOREIGN KEY (family_name) REFERENCES product_family(name)
);

CREATE TABLE IF NOT EXISTS region (
    id SERIAL UNIQUE,
    name VARCHAR(255) PRIMARY KEY,
    friendly_name VARCHAR(255),
    type VARCHAR(255),
    subnet VARCHAR(255),
    availability_zone VARCHAR(255),
    prefix INT,
    is_default BOOLEAN,
    api_dns VARCHAR(255),
    admin_name VARCHAR(255) DEFAULT 'system',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS product_region_set (
    product_name VARCHAR(255) NOT NULL,
    region_name VARCHAR(255) NOT NULL,
    FOREIGN KEY (product_name) REFERENCES product(name) ON DELETE CASCADE,
    FOREIGN KEY (region_name) REFERENCES region(name) ON DELETE CASCADE,
    PRIMARY KEY (product_name, region_name)
);

CREATE TABLE IF NOT EXISTS rate (
    id SERIAL PRIMARY KEY,
    account_type VARCHAR(255),
    usage_unit_type VARCHAR(255),
    rate DOUBLE PRECISION,
    admin_name VARCHAR(255) DEFAULT 'system',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS product_rate_set (
    product_name VARCHAR(255) NOT NULL,
    rate_id INT NOT NULL,
    FOREIGN KEY (product_name) REFERENCES product(name) ON DELETE CASCADE,
    FOREIGN KEY (rate_id) REFERENCES rate(id) ON DELETE CASCADE,
    PRIMARY KEY (product_name, rate_id)
);

CREATE TABLE IF NOT EXISTS metadata (
    id SERIAL PRIMARY KEY,
    key VARCHAR(255) NOT NULL,
    value VARCHAR NOT NULL,
    type VARCHAR(255),
    context VARCHAR(255),
    admin_name VARCHAR(255) DEFAULT 'system',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS metadata_set (
    product_name VARCHAR(36) NOT NULL,
    metadata_id INT NOT NULL,
    FOREIGN KEY (product_name) REFERENCES product(name) ON DELETE CASCADE,
    FOREIGN KEY (metadata_id) REFERENCES metadata(id) ON DELETE CASCADE,
    PRIMARY KEY (product_name, metadata_id)
);

CREATE TABLE IF NOT EXISTS cloud_accounts_product_access (
    cloudaccount_id VARCHAR(12) NOT NULL,
    product_id VARCHAR(64) NOT NULL,
    family_id uuid NOT NULL,
    admin_name VARCHAR(255) NOT NULL CHECK(admin_name != ''),
    created TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (product_id) REFERENCES product(id),
    FOREIGN KEY (family_id) REFERENCES product_family(id),
    PRIMARY KEY (cloudaccount_id, product_id)
);

CREATE TABLE IF NOT EXISTS cloud_accounts_region_access (
    cloudaccount_id VARCHAR(12) NOT NULL,
    region_name VARCHAR(255) NOT NULL,
    admin_name VARCHAR(255) NOT NULL CHECK(admin_name != ''),
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (region_name) REFERENCES region(name),
    PRIMARY KEY (cloudaccount_id, region_name)
);

