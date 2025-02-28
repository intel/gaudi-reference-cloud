# IDC Compute Test

Developer Cloud Services team: Intel Developer Cloud.

### Important Information Regarding ConfigMap Data Changes for Business Use Cases on Kind and Dev Environments. 

> Metering monitoring

ConfigMap name : us-dev-1a-compute-metering-monitor-manager-config

Parameter : nmaxUsageRecordSendInterval: "60m"

Podname (sub-string) : us-dev-1a-compute-metering-monitor-

Steps:
- Find the above configMap
- Inside the configMap, find the respective paramter and change the value as required
- Restart the pod


> Quota Enforcement

ConfigMap name : us-dev-1-compute-api-server

Parameter : cloudAccountQuota

Podname (sub-string) : us-dev-1-compute-api-server-

Steps:
- Find the above configMap
- Inside the configMap, find the respective parameter 
  - For the respective cloud account type, change the number of instances for each instance type 
    as required
- Restart the pod

