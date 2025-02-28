-- Update the default value of the column to 1 -> which indicates UPGRADE_NOT_INITIATED.
ALTER TABLE cloud_accounts 
ADD COLUMN IF NOT EXISTS upgraded_to_premium SMALLINT DEFAULT 1,
ADD COLUMN IF NOT EXISTS upgraded_to_enterprise SMALLINT DEFAULT 1;
