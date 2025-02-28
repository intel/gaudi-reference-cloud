CREATE TABLE IF NOT EXISTS cloud_accounts_product_access (
    cloudaccount_id VARCHAR(12) NOT NULL,
    product_id VARCHAR(64) NOT NULL,
    family_id VARCHAR(64) NOT NULL,
    vendor_id VARCHAR(64) NOT NULL,
    admin_name VARCHAR(255) NOT NULL CHECK(admin_name != ''),
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (cloudaccount_id, vendor_id, product_id)
);
