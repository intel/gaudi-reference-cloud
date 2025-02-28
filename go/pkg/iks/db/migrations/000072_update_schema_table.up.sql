ALTER TABLE vip ADD COLUMN sourceips json NULL;
ALTER TABLE vip ADD COLUMN firewall_status VARCHAR(20) NULL;
ALTER TABLE vipdetails ADD COLUMN protocol json NULL;

INSERT INTO defaultconfig(name, value)
VALUES ('firewall_protocol','TCP');