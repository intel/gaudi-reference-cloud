Secrity APIs:

User Input: Source Ips

Create/update API: Create Firewall Rules for a cluster.

Workflow:
1. Use source ip provided by user and update exising vip record using vip id.
2. Update cluster rev table with firewallrule crd


Validations:
1. Validate if cluster exists
2. Validate if cloudaccount has correct permissions
3. Validate if cluster is actionable.
4. Validate if user has existing vip_ip associated with cluster , if not ask user to create at least one vip.
5. Validate if vip port exists
6. Validate if vip is active
7. validate if fw rule is active
8. If vip_ip exists then update vip table with required values.
9. Update vip table for vip id.
10. Insert cluster rev entry
11. Generate crd for vipids
12. Update clsuter state to pending


Input:
CloudaccountID
ClusterID
SourceIPs
Internal ip
Port
Protocol

Output:
Sourceips
Destinationip
Port
Vipid
State
Protocol


Get API:

Workflow:
Get all the security tabl related columns for all vip type = "public" and Active vips

Validations:
1. Validate if cluster exists
2. Validate if cloudaccount has correct permissions

Input:
CloudaccountID
ClusterID

Output:
Array of
Sourceips
Destinationip
Port
Vipid
State



Delete API:

Workflow:
1. Update firewall_status information for vip id for cluster.
2. Update cluster rev table
3. Generate crd for vipids
4. Update clsuter state to pending

Validations:
1. Validate if cluster exists
2. Validate if vip exists
3. Validate if cloudaccount has correct permissions
4. Validate if cluster is actionable.


Input:
ClouadaccountID
ClusterID
VipID

output:
Empty