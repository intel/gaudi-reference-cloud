ALTER TABLE storage ALTER COLUMN size TYPE float USING CAST(size AS float);

ALTER TABLE cloudaccountextraspec ALTER COLUMN total_storage_size TYPE float USING CAST(total_storage_size AS float);

UPDATE storage SET size = size / 1000;

UPDATE cloudaccountextraspec SET total_storage_size = total_storage_size / 1000;
