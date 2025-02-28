-- Step 1: Alter tables cloudaccountextraspec and drop any existing constraint
ALTER TABLE cloudaccountextraspec DROP CONSTRAINT IF EXISTS maxclusters_override_max_check;
ALTER TABLE cloudaccountextraspec DROP CONSTRAINT IF EXISTS maxclusterng_override_max_check;
ALTER TABLE cloudaccountextraspec DROP CONSTRAINT IF EXISTS maxclusterilb_override_max_check;
ALTER TABLE cloudaccountextraspec DROP CONSTRAINT IF EXISTS maxclustervm_override_max_check;
ALTER TABLE cloudaccountextraspec DROP CONSTRAINT IF EXISTS maxnodegroupvm_override_max_check;

-- Step 2: Alter tables cloudaccountextraspec and new constraint for each column
ALTER TABLE cloudaccountextraspec ADD CONSTRAINT maxclusters_override_max_check CHECK (maxclusters_override <= 20);
ALTER TABLE cloudaccountextraspec ADD CONSTRAINT maxclusterng_override_max_check CHECK (maxclusterng_override <= 50);
ALTER TABLE cloudaccountextraspec ADD CONSTRAINT maxclusterilb_override_max_check CHECK (maxclusterilb_override <= 20);
ALTER TABLE cloudaccountextraspec ADD CONSTRAINT maxclustervm_override_max_check CHECK (maxclustervm_override <= 512);
ALTER TABLE cloudaccountextraspec ADD CONSTRAINT maxnodegroupvm_override_max_check CHECK (maxnodegroupvm_override <= 512);

-- Step 3: Update existing records with default values
UPDATE cloudaccountextraspec SET maxclusters_override = 3;
UPDATE cloudaccountextraspec SET maxclusterng_override = 5;
UPDATE cloudaccountextraspec SET maxclusterilb_override = 2;
UPDATE cloudaccountextraspec SET maxclustervm_override = 50;
UPDATE cloudaccountextraspec SET maxnodegroupvm_override = 10;

-- Step 4: Alter tables cloudaccountextraspec and to set below columns to not null
ALTER TABLE cloudaccountextraspec ALTER COLUMN maxclusters_override SET NOT NULL;
ALTER TABLE cloudaccountextraspec ALTER COLUMN maxclusterng_override SET NOT NULL;
ALTER TABLE cloudaccountextraspec ALTER COLUMN maxclusterilb_override SET NOT NULL;
ALTER TABLE cloudaccountextraspec ALTER COLUMN maxclustervm_override SET NOT NULL;
ALTER TABLE cloudaccountextraspec ALTER COLUMN maxnodegroupvm_override SET NOT NULL;