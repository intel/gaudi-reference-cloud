-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
UPDATE nodeprovider SET is_default = 'false';
UPDATE nodeprovider SET is_default = 'true' WHERE nodeprovider_name = 'Compute';

do
$$
begin
  if exists (select datname FROM pg_database WHERE datname='main') AND not exists (select * from pg_roles where rolname = 'iks_user1') then
    create user iks_user1 password 'OnlyUseForTest!';
  end if;
end
$$
;

do
$$
begin
  if exists (select datname FROM pg_database WHERE datname='main') AND not exists (select * from pg_roles where rolname = 'iks_user2') then
    create user iks_user2 password 'OnlyUseForTest!';
  end if;
end
$$
;

do
$$
begin
  if exists (select datname FROM pg_database WHERE datname='main') then
    ALTER user iks_user1 WITH LOGIN;
    GRANT select, insert, update, delete on all tables in schema public TO iks_user1;
    ALTER default privileges in schema public GRANT select,insert, update, delete on tables TO iks_user1;
    ALTER user iks_user2 WITH LOGIN;
    GRANT select, insert, update, delete on all tables in schema public TO iks_user2;
    ALTER default privileges in schema public GRANT select,insert, update, delete on tables TO iks_user2;
  end if;
end
$$
;

do
$$
begin
  if exists (select * from pg_roles where rolname = 'dbuser') then
     ALTER default privileges in schema public GRANT select,insert, update, delete on tables TO dbuser;
  end if;
end
$$
;