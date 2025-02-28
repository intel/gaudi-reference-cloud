ALTER TABLE cloud_accounts 
ADD COLUMN IF NOT EXISTS paid_services_allowed BOOLEAN DEFAULT TRUE;
