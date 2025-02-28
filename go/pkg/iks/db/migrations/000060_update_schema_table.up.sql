ALTER TABLE cluster ADD COLUMN storage_enable BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE cluster ADD COLUMN storage_size INT DEFAULT 0;
ALTER TABLE cloudaccountextraspec ADD COLUMN total_storage_size INT DEFAULT 0;