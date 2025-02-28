--------------------------------------------------------------------------------
-- subnet
--------------------------------------------------------------------------------

create sequence subnet_id_seq as int minvalue 1;

create table subnet (
    subnet_id int primary key default nextval('subnet_id_seq'),
    region varchar(100) not null,
    availability_zone varchar(100) not null,
    address_space varchar(100) not null,
    -- first address in subnet such as "172.16.11.0/24"
    subnet cidr not null,
    prefix_length smallint not null,
    gateway inet not null,
    vlan_id int not null,
    subnet_consumer_id varchar(255) null
);

create unique index subnet_consumer_idx on subnet (subnet_consumer_id);
create unique index subnet_subnet_idx on subnet (region, availability_zone, address_space, subnet);
create unique index subnet_vlan_idx on subnet (region, availability_zone, vlan_id);

--------------------------------------------------------------------------------
-- address
--------------------------------------------------------------------------------

create table address (
    subnet_id int not null references subnet (subnet_id),
    address inet not null,
    address_consumer_id varchar(255) null,
    primary key (subnet_id, address)
);

create unique index address_consumer_idx on address (subnet_id, address_consumer_id);
