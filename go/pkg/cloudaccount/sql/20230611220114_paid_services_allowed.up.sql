-- Update the default value of the column to False.
ALTER TABLE cloud_accounts ALTER COLUMN paid_services_allowed SET DEFAULT FALSE;
