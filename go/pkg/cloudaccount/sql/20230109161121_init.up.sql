CREATE TABLE IF NOT EXISTS cloud_accounts (
    id VARCHAR(12) NOT NULL PRIMARY KEY,
    parent_id VARCHAR(12),
    tid VARCHAR(64) NOT NULL,
    oid VARCHAR(64) NOT NULL CHECK(oid != '') UNIQUE,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    name VARCHAR(255) NOT NULL CHECK(name != '') UNIQUE,
    type SMALLINT CHECK(type > 0),
    owner VARCHAR(255) NOT NULL CHECK(owner != ''),
    billing_account_created BOOLEAN DEFAULT FALSE,
    enrolled BOOLEAN DEFAULT FALSE,
    low_credits BOOLEAN DEFAULT FALSE,
    credits_depleted TIMESTAMP DEFAULT to_timestamp(0),
    terminate_paid_services BOOLEAN DEFAULT FALSE,
    terminate_message_queued BOOLEAN DEFAULT FALSE,
    delinquent BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS cloud_account_members (
    account_id VARCHAR(12) REFERENCES cloud_accounts(id) ON DELETE CASCADE,
    member VARCHAR(255) NOT NULL CHECK(member != ''),
    UNIQUE (account_id, member)
);

CREATE INDEX IF NOT EXISTS cloud_account_members_account_id_idx ON cloud_account_members(account_id);
CREATE INDEX IF NOT EXISTS cloud_account_members_member_idx ON cloud_account_members(member);
