INSERT INTO cloud_accounts (id, name, owner, tid, oid, type) 
SELECT '000000000001', 'validator@intel.com', 'validator@intel.com','4015bb99-0522-4387-b47e-c821596dc736', 'd2c41d90-32ad-4312-ba8b-1b995a57475d', 5
WHERE NOT EXISTS
(
    SELECT * FROM cloud_accounts
    WHERE name='validator@intel.com' AND id='000000000001'
);

INSERT INTO cloud_accounts (id, name, owner, tid, oid, type) 
SELECT '000000000002', 'iks-user@intel.com', 'iks-user@intel.com','4015bb99-0522-4387-b47e-c821596dc736', '61befbee-0607-47c5-b140-c4509dfef836', 5
WHERE NOT EXISTS
(
    SELECT * FROM cloud_accounts
    WHERE name='iks-user@intel.com' AND id='000000000002'
);
