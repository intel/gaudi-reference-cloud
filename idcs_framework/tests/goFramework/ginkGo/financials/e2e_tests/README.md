*********************************************
Runnning E2e Financial tests with changing deployment
*********************************************

Change the following config maps in IDC deployment to run e2e tests until Scheduler Ops APIs are availble

Billing Schedulers
"""""""""""""""""
Update billing-schedulers configmap with following values

 creditsInstallSchedulerInterval: 60
 reportUsageSchedulerInterval: 60
 creditsExpiryMinimumInterval: 1
 creditUsageSchedulerInterval: 1
 creditExpirySchedulerInterval: 1
 premiumCloudCreditThreshold: 80
 intelCloudCreditThreshold: 80
 enterpriseCloudCreditThreshold: 80
 premiumCloudCreditNotifyBeforeExpiry: 4320
 intelCloudCreditNotifyBeforeExpiry: 4320
 enterpriseCloudCreditNotifyBeforeExpiry: 4320
 servicesTerminationSchedulerInterval: 240
 servicesTerminationInterval: 1440
 eventExpirySchedulerInterval: 1440


 After updating restart deployment by scaling pods

kubectl scale --replicas=0 deployment billing-schedulers -n idcs-system
kubectl scale --replicas=1 deployment billing-schedulers -n idcs-system


Metering Monitor
===============
update metering monitor config map (us-dev-1a-compute-metering-monitor-manager-config) 

update maxUsageRecordSendInterval: "60m" with maxUsageRecordSendInterval: "1m"

After updating restart deployment by scaling pods

kubectl scale --replicas=0 deployment us-dev-1a-compute-metering-monitor -n idcs-system
kubectl scale --replicas=1 deployment us-dev-1a-compute-metering-monitor -n idcs-system
