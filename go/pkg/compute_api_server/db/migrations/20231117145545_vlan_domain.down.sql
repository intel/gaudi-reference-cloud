drop index subnet_vlan_idx;
alter table subnet drop column vlan_domain;
create unique index subnet_vlan_idx on subnet (region, availability_zone, vlan_id);
