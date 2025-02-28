drop index subnet_vlan_idx;
alter table subnet add column vlan_domain varchar(100) not null default '';
create unique index subnet_vlan_idx on subnet (region, availability_zone, vlan_domain, vlan_id);
