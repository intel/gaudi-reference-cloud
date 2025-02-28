ALTER TABLE addonversion ADD COLUMN onbuild BOOLEAN DEFAULT false;
ALTER TABLE addonversion ADD COLUMN addonversion_type VARCHAR(30);
UPDATE addonversion SET addonversion_type='kubernetes', onbuild=true;
ALTER TABLE addonversion ALTER COLUMN addonversion_type SET NOT NULL;