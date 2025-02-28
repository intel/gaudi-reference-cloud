-- Update the default value of the column to False.
ALTER TABLE cloud_accounts 
ADD COLUMN IF NOT EXISTS restricted BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS admin_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS access_limited_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
