CREATE TABLE IF NOT EXISTS members (
    id BIGSERIAL PRIMARY KEY,
    admin_account_id VARCHAR(12) NOT NULL,
    member_email VARCHAR(255) NOT NULL CHECK(user != ''),
    invitation_code VARCHAR(12) NOT NULL CHECK(invitation_code != '') UNIQUE,
    invitation_state SMALLINT CHECK(invitation_state > 0),
    expiry TIMESTAMP,
    notes TEXT,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    revoked BOOLEAN DEFAULT FALSE,
    UNIQUE (admin_account_id, member_email)
);

CREATE INDEX IF NOT EXISTS members_account_id_idx ON members(admin_account_id);