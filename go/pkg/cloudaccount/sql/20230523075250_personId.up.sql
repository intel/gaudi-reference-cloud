ALTER TABLE cloud_accounts ADD COLUMN IF NOT EXISTS person_id VARCHAR(12);

CREATE INDEX IF NOT EXISTS cloud_accounts_person_id_idx ON cloud_accounts(person_id);
