ALTER TABLE members 
ADD COLUMN IF NOT EXISTS person_id VARCHAR(12);

CREATE INDEX IF NOT EXISTS members_person_id_idx ON members(person_id);
