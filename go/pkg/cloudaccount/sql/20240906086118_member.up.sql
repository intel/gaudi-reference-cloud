ALTER TABLE members 
ADD COLUMN IF NOT EXISTS cloud_account_role_ids TEXT[];
