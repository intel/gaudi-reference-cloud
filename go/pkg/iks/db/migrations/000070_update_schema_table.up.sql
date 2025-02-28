-- Step 1: Alter tables cluster and nodegroup to add below new columns
ALTER TABLE cluster ADD COLUMN clustertype varchar(50);
ALTER TABLE nodegroup ADD COLUMN nodegrouptype varchar(50);

-- Step 2: Update existing records with default values
UPDATE cluster SET clustertype = 'generalpurpose';
UPDATE nodegroup SET nodegrouptype = 'gp';

-- Step 3: Alter tables cluster and nodegroup again to add set new columns to not null
ALTER TABLE cluster ALTER COLUMN clustertype SET NOT NULL;
ALTER TABLE nodegroup ALTER COLUMN nodegrouptype SET NOT NULL;