ALTER TABLE ssh_public_key
ADD COLUMN owner_email varchar(255) not null default '';
