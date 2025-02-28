# ARTIFACTS contains the complete set of containers and Helm charts in the IDC monorepo.
# Keys of ARTIFACTS must match the Helm chart name.
# If a record has a "chart" key, its value should be the Bazel target for the Helm chart.
# If a record has a "components" key, its value should be the list of components that use this Helm chart.
# If a record has an "image" key, its value should be the Bazel target for the container image used by the Helm chart.
# If a record has an "images" key, its value should be the list of Bazel targets for the container images used by the Helm chart.
# Records can have both an image and a chart, or just one of them.
# Keep this file in alphabetical order.

ARTIFACTS = {
    "argo-cd-resources": {
        "chart": "//deployment/charts/argo-cd-resources:chart",
    },
    "armada": {
        "chart": "//deployment/charts/armada:chart",
        "components": ["armada"],
        "image": "//go/pkg/training/armada/cmd/armada:armada_image",
    },
    "authz": {
        "chart": "//deployment/charts/authz:chart",
        "components": ["authz"],
        "image": "//go/svc/authz:authz_image",
    },
    "baremetal-enrollment-api": {
        "chart": "//deployment/charts/baremetal-enrollment-api:chart",
        "components": ["computeBmEnrollment"],
        "image": "//go/svc/baremetal_enrollment:baremetal_enrollment_api_image",
    },
    "baremetal-enrollment-operator": {
        "chart": "//deployment/charts/baremetal-enrollment-operator:chart",
        "components": ["computeBmEnrollment"],
        "image": "//go/pkg/baremetal_enrollment/operator:baremetal_enrollment_operator_image",
    },
    "baremetal-enrollment-task": {
        "chart": "//deployment/charts/baremetal-enrollment-task:chart",
        "components": ["computeBmEnrollment"],
    },
    "baremetal-operator": {
        "chart": "//deployment/charts/baremetal-operator:chart",
        "components": ["computeBmMetal3"],
    },
    "baremetal-operator-ns": {
        "chart": "//deployment/charts/baremetal-operator-ns:chart",
        "components": ["computeBmMetal3"],
    },
    "billing": {
        "chart": "//deployment/charts/billing:chart",
        "components": ["billing"],
        "image": "//go/svc/billing:billing_image",
    },
    "billing-aria": {
        "chart": "//deployment/charts/billing-aria:chart",
        "components": ["billing"],
        "image": "//go/svc/billing_driver_aria/cmd:billing-aria_image",
    },
    "billing-intel": {
        "chart": "//deployment/charts/billing-intel:chart",
        "components": ["billing"],
        "image": "//go/svc/billing_driver_intel/cmd:billing-intel_image",
    },
    "billing-schedulers": {
        "chart": "//deployment/charts/billing-schedulers:chart",
        "components": ["billing"],
        "image": "//go/svc/billing:billing_image",
    },
    "billing-standard": {
        "chart": "//deployment/charts/billing-standard:chart",
        "components": ["billing"],
        "image": "//go/svc/billing_driver_standard/cmd:billing-standard_image",
    },
    "bm-dnsmasq": {
        "chart": "//deployment/charts/bm-dnsmasq:chart",
        "components": ["computeBm"],
    },
    "bm-instance-operator": {
        "chart": "//deployment/charts/bm-instance-operator:chart",
        "components": ["computeBmInstanceOperator"],
        "image": "//go/pkg/instance_operator/bm:bm_instance_operator_image",
    },
    "bm-validation-operator": {
        "chart": "//deployment/charts/bm-validation-operator:chart",
        "components": ["computeBmValidationOperator"],
        "image": "//go/pkg/instance_operator/baremetal-validation-operator:bm_validation_operator_image",
    },
    "bucket-metering-monitor": {
        "chart": "//deployment/charts/bucket-metering-monitor:chart",
        "components": ["storage"],
        "image": "//go/pkg/storage/bucket_metering_monitor:bucket-metering-monitor_image",
    },
    "bucket-replicator": {
        "chart": "//deployment/charts/bucket-replicator:chart",
        "components": ["storage"],
        "image": "//go/pkg/storage/bucket_replicator/cmd/main:bucket-replicator_image",
    },
    "cloudaccount": {
        "chart": "//deployment/charts/cloudaccount:chart",
        "components": ["cloudaccount"],
        "image": "//go/svc/cloudaccount:cloudaccount_image",
    },
    "cloudaccount-enroll": {
        "chart": "//deployment/charts/cloudaccount-enroll:chart",
        "components": ["cloudaccount"],
        "image": "//go/svc/cloudaccount-enroll:cloudaccount-enroll_image",
    },
    "cloudcredits": {
        "chart": "//deployment/charts/cloudcredits:chart",
        "components": ["cloudcredits"],
        "image": "//go/svc/cloud_credits:cloudcredits_image",
    },
    "cloudcredits-worker": {
        "chart": "//deployment/charts/cloudcredits-worker:chart",
        "components": ["cloudcredits"],
        "image": "//go/svc/cloudcredits_worker:cloudcredits-worker_image",
    },
    "cloudmonitor": {
        "chart": "//deployment/charts/cloudmonitor:chart",
        "components": ["cloudmonitor"],
        "image": "//go/svc/cloudmonitor/cmd/cloudmonitor:cloudmonitor_image",
    },
    "cloudmonitor-logs-api-server": {
        "chart": "//deployment/charts/cloudmonitor-logs-api-server:chart",
        "components": ["cloudmonitorLogs"],
        "image": "//go/pkg/cloudmonitor_logs/api_server:cloudmonitor_logs_api_server_image"
    },
    "compute-api-server": {
        "chart": "//deployment/charts/compute-api-server:chart",
        "components": ["computeApiServer"],
        "image": "//go/svc/compute_api_server:compute_api_server_image",
    },
    "compute-crds": {
        "chart": "//deployment/charts/compute-crds:chart",
        "components": ["compute"],
    },
    "compute-metering-monitor": {
        "chart": "//deployment/charts/compute-metering-monitor:chart",
        "components": ["computeMeteringMonitor"],
        "image": "//go/pkg/compute_metering_monitor:compute_metering_monitor_image",
    },
    "console": {
        "chart": "//deployment/charts/console:chart",
        "image": "//go/svc/console:console_image",
    },
    "dataloader": {
        "chart": "//deployment/charts/dataloader:chart",
        "components": ["dataloader"],
        "image": "//go/pkg/dataloader:dataloader_image",
    },
    "dpai": {
        "chart": "//deployment/charts/dpai:chart",
        "components": ["dpai"],
        "image": "//go/svc/dpai/cmd/dpai:dpai_image",
    },
    "debug-tools": {
        "chart": "//deployment/charts/debug-tools:chart",
    },
    "dhcp-proxy": {
        "chart": "//deployment/charts/dhcp-proxy:chart",
        "components": ["computeBm"],
    },
    "external-secrets": {
        "components": ["externalSecrets"],
    },
    "firewall-operator": {
        "chart": "//deployment/charts/firewall-operator:chart",
        "components": ["firewall"],
        "image": "//go/pkg/firewall_operator:firewall_operator_image",
    },
    "firewall-crds": {
        "chart": "//deployment/charts/firewall-crds:chart",
        "components": ["firewall"],
    },
    "fleet-admin-api-server": {
        "chart": "//deployment/charts/fleet-admin-api-server:chart",
        "components": ["fleetAdmin"],
        "image": "//go/pkg/fleet_admin/api_server:fleet_admin_api_server_image",
    },
    "fleet-admin-ui-server": {
        "chart": "//deployment/charts/fleet-admin-ui-server:chart",
        "components": ["fleetAdmin"],
        "image": "//go/pkg/fleet_admin_ui_server/api_server:fleet_admin_ui_server_image",
    },
    "fleet-node-reporter": {
        "chart": "//deployment/charts/fleet-node-reporter:chart",
        "components": ["fleetAdmin"],
        "image": "//go/pkg/fleet_node_reporter:fleet_node_reporter_image",
    },
    "git-to-grpc-synchronizer": {
        "chart": "//deployment/charts/git-to-grpc-synchronizer:chart",
        "components": ["computePopulateInstanceType", "computePopulateMachineImage", "computePopulateSubnet", "populateProductCatalog"],
        "image": "//go/pkg/git_to_grpc_synchronizer:git_to_grpc_synchronizer_image",
    },
    "grpc-proxy": {
        "chart": "//go/svc/grpc-proxy/chart/grpc-proxy:chart",
        "components": ["grpc"],
        "images": {
            "opa": "//go/svc/opa:opa_image",
        },
    },
    "grpc-reflect": {
        "chart": "//deployment/charts/grpc-reflect:chart",
        "components": ["grpc"],
        "image": "//go/svc/grpc-reflect:grpc-reflect_image",
    },
    "grpc-rest-gateway": {
        "chart": "//deployment/charts/grpc-rest-gateway:chart",
        "components": ["grpc"],
        "image": "//go/svc/grpc-rest-gateway:grpc-rest-gateway_image",
    },
    "idc-versions": {
        "chart": "//deployment/charts/idc-versions:chart",
    },
    "idcs-init-k8s-resources": {
        "chart": "//deployment/charts/idcs-init-k8s-resources:chart",
    },
    "idcs-istio-custom-resources": {
        "chart": "//deployment/charts/idcs-istio-custom-resources:chart",
    },
    "iks": {
        "chart": "//deployment/charts/iks:chart",
        "components": ["iks"],
        "image": "//go/svc/iks/cmd/iks:iks_image",
    },
    "ilb-crds": {
        "chart": "//deployment/charts/ilb-crds:chart",
        "components": ["iks"],
    },
    "ilb-operator": {
        "chart": "//deployment/charts/ilb-operator:chart",
        "components": ["iks"],
        "image": "//go/pkg/ilb_operator:ilb_operator_image",
    },
    "infaas-dispatcher": {
        "chart": "//deployment/charts/infaas-dispatcher:chart",
        "components": ["maasCompute"],
        "image": "//go/pkg/infaas-dispatcher:infaas_dispatcher_image"
    },
    "infaas-inference": {
        "chart": "//deployment/charts/infaas-inference:chart",
        "components": ["maasCompute"],
        "image": "//go/pkg/infaas-agent:infaas-inference_image"
    },
    "infaas-resources": {
        "chart": "//deployment/charts/infaas-resources:chart",
        "components": ["maasComputeDependencies"]
    },
    "infaas-safeguard": {
        "chart": "//deployment/charts/infaas-safeguard:chart",
        "components": ["maasCompute"],
    },
    "instance-replicator": {
        "chart": "//deployment/charts/instance-replicator:chart",
        "components": ["computeInstanceReplicator"],
        "image": "//go/pkg/instance_replicator:instance_replicator",
    },
    "intel-device-plugin": {
        "chart": "//deployment/charts/intel-device-plugin:chart",
        "components": ["intelDevicePlugin"],
        "image": "//go/pkg/kubernetes_device_plugins/intel_device_plugin:intel_device_plugin_image",
    },
    "kfaas": {
        "chart": "//deployment/charts/kfaas:chart",
        "components": ["kfaas"],
        "image": "//go/svc/kfaas/cmd/kfaas:kfaas_image",
    },
    "kubernetes-crds": {
        "chart": "//deployment/charts/kubernetes-crds:chart",
        "components": ["iks"],
    },
    "kubernetes-operator": {
        "chart": "//deployment/charts/kubernetes-operator:chart",
        "components": ["iks"],
        "image": "//go/pkg/kubernetes_operator:kubernetes_operator_image",
    },
    "kubernetes-reconciler": {
        "chart": "//deployment/charts/kubernetes-reconciler:chart",
        "components": ["iks"],
        "image": "//go/pkg/kubernetes_reconciler:kubernetes_reconciler_image",
    },
    "kubescore": {
        "chart": "//deployment/charts/kubescore:chart",
        "components": ["insights"],
        "image": "//go/pkg/insights/kubescore/cmd/kubescore:kubescore_image",
    },
    "k8s-resource-patcher": {
        "chart": "//deployment/charts/k8s-resource-patcher:chart",
        "components": ["fleetAdmin"],
        "image": "//go/pkg/k8s_resource_patcher:k8s_resource_patcher_image",
    },
    "loadbalancer-crds": {
        "chart": "//deployment/charts/loadbalancer-crds:chart",
        "components": ["loadbalancer"],
    },
    "loadbalancer-operator": {
        "chart": "//deployment/charts/loadbalancer-operator:chart",
        "components": ["loadbalancer"],
        "image": "//go/pkg/loadbalancer_operator:loadbalancer_operator_image",
    },
    "loadbalancer-replicator": {
        "chart": "//deployment/charts/loadbalancer-replicator:chart",
        "components": ["loadbalancer"],
        "image": "//go/pkg/loadbalancer_replicator:loadbalancer_replicator_image",
    },
    "local-path-provisioner": {
        "chart": "//deployment/charts/local-path-provisioner:chart",
    },
    "maas-gateway": {
        "chart": "//deployment/charts/maas/maas-gateway:chart",
        "components": ["maas"],
        "image": "//go/pkg/maas-gateway:maas-gateway_image",
    },
    "metal3-crds": {
        "chart": "//deployment/charts/metal3-crds:chart",
        "components": ["computeBmMetal3"],
    },
    "metallb-custom-resources": {
        "chart": "//deployment/charts/metallb-custom-resources:chart",
        "components": ["computeBm"],
    },
    "metering": {
        "chart": "//deployment/charts/metering:chart",
        "components": ["metering"],
        "image": "//go/svc/metering/cmd/metering:metering_api_server_image",
    },
    "netbox": {
        "chart": "//deployment/charts/netbox:chart",
        "components": ["compute"],
    },
    "netbox-sso": {
        "chart": "//deployment/charts/netbox-azuread-sso:chart",
    },
    "network-api-server": {
        "chart": "//deployment/charts/network-api-server:chart",
        "image": "//go/pkg/network/api_server:network_api_server_image",
    },
    "network-operator": {
        "chart": "//deployment/charts/network-operator:chart",
        "image": "//go/pkg/network/operator:network_operator_image",
    },
    "network-crds": {
        "chart": "//deployment/charts/network-crds:chart",
        "components": ["network"],
    },
    "nginx-s3-gateway": {
        "chart": "//deployment/charts/nginx-s3-gateway:chart",
        "components": ["computeBm"],
    },
    "notification-gateway": {
        "chart": "//deployment/charts/notification-gateway:chart",
        "components": ["billing"],
        "image": "//go/svc/notification-gateway:notification-gateway_image",
    },
    "object-store-operator": {
        "chart": "//deployment/charts/object-store-operator:chart",
        "components": ["storage"],
        "image": "//go/pkg/storage/object_store_operator:object_store_operator_image",
    },
    "oidc": {
        "chart": "//deployment/charts/oidc:chart",
        "components": ["oidc"],
        "image": "//go/pkg/tools/oidc:oidc_image",
    },
    "portal": {
        "components": ["portal"],
    },
    "productcatalog": {
        "chart": "//deployment/charts/productcatalog:chart",
        "components": ["productCatalog"],
        "image": "//go/svc/productcatalog:productcatalog_image",
    },
    "productcatalog-crds": {
        "chart": "//deployment/charts/productcatalog-crds:chart",
        "components": ["productCatalog"],
    },
    "productcatalog-operator": {
        "chart": "//deployment/charts/productcatalog-operator:chart",
        "components": ["productCatalog"],
        "image": "//go/pkg/productcatalog_operator:productcatalog_operator_image",
    },
    "quick-connect-api-server": {
        "chart": "//deployment/charts/quick-connect-api-server:chart",
        "components": ["quickConnect"],
        "image": "//go/pkg/quick_connect/api_server:quick_connect_api_server_image",
    },
    "rate-limit": {
        "chart": "//deployment/charts/rate-limit:chart",
        "components": ["rateLimit"]
    },
    "rate-limit-redis": {
        "chart": "//deployment/charts/rate-limit-redis:chart",
        "components": ["rateLimit"]
    },
    "security-insights": {
        "chart": "//deployment/charts/security-insights:chart",
        "components": ["insights"],
        "image": "//go/pkg/insights/security-insights/cmd/insights:security-insights_image",
    },
    "security-scanner": {
        "chart": "//deployment/charts/security-scanner:chart",
        "components": ["insights"],
        "image": "//go/pkg/insights/security-scanner/cmd/main:security-scanner_image",
    },
    "sdn-controller": {
        "chart": "//deployment/charts/sdn-controller:chart",
        "components": ["sdn"],
        "image": "//go/pkg/sdn-controller:sdn-controller_image",
    },
    "sdn-controller-crds": {
        "chart": "//deployment/charts/sdn-controller-crds:chart",
        "components": ["sdn"],
    },
    "sdn-controller-rest": {
        "chart": "//deployment/charts/sdn-controller-rest:chart",
        "image": "//go/pkg/sdn-controller/rest-api:sdn-controller-rest_image",
    },
    "sdn-integrity-checker": {
        "chart": "//deployment/charts/sdn-integrity-checker:chart",
        "components": ["sdn"],
        "image": "//go/pkg/sdn-controller/tests/data_integrity_check:sdn-integrity-checker_image",
    },
    "sdn-restricted-sa": {
        "chart": "//deployment/charts/sdn-restricted-sa:chart",
        "components": ["sdn"],
    },
    "sdn-vn-controller": {
        "chart": "//deployment/charts/sdn-vn-controller:chart",
        "components": ["sdnVN"],
        "image": "//go/pkg/sdn-vn-controller:sdn-vn-controller_image",
    },
    "ssh-proxy-operator": {
        "chart": "//deployment/charts/ssh-proxy-operator:chart",
        "components": ["computeSshProxy"],
        "image": "//go/pkg/ssh_proxy_operator:ssh_proxy_operator_image",
    },
    "storage-admin-api-server": {
        "chart": "//deployment/charts/storage-admin-api-server:chart",
        "components": ["storage"],
        "image": "//go/pkg/storage/storage_admin_api_server/cmd/main:storage-admin-api-server_image",
    },
    "storage-api-server": {
        "chart": "//deployment/charts/storage-api-server:chart",
        "components": ["storage"],
        "image": "//go/pkg/storage/api_server/cmd/main:storage-api-server_image",
    },
    "storage-kms": {
        "chart": "//deployment/charts/storage-kms:chart",
        "components": ["storage"],
        "image": "//go/pkg/storage/storage_kms/cmd/main:storage-kms_image",
    },
    "storage-metering-monitor": {
        "chart": "//deployment/charts/storage-metering-monitor:chart",
        "components": ["storage"],
        "image": "//go/pkg/storage/storage_metering_monitor:storage-metering-monitor_image",
    },
    "storage-custom-metrics-service": {
        "chart": "//deployment/charts/storage-custom-metrics-service:chart",
        "components": ["storage"],
        "image": "//go/pkg/storage/storage_custom_metrics_service/cmd/main:storage-custom-metrics-service_image",
    },
    "storage-operator": {
        "chart": "//deployment/charts/storage-operator:chart",
        "components": ["storage"],
        "image": "//go/pkg/storage/storage_operator:storage_operator_image",
    },
    "storage-replicator": {
        "chart": "//deployment/charts/storage-replicator:chart",
        "components": ["storage"],
        "image": "//go/pkg/storage/storage_replicator/cmd/main:storage-replicator_image",
    },
    "storage-resource-cleaner": {
        "chart": "//deployment/charts/storage-resource-cleaner:chart",
        "components": ["storage"],
        "image": "//go/pkg/storage/storage_resource_cleaner/cmd/main:storage-resource-cleaner_image",
    },
    "storage-scheduler": {
        "chart": "//deployment/charts/storage-scheduler:chart",
        "components": ["storage"],
        "image": "//go/pkg/storage/storage_scheduler/cmd/main:storage-scheduler_image",
    },
    "storage-user": {
        "chart": "//deployment/charts/storage-user:chart",
        "components": ["storage"],
        "image": "//go/pkg/storage/storage_user/cmd/main:storage-user_image",
    },
    "quota-management-service": {
        "chart": "//deployment/charts/quota-management-service:chart",
        "components": ["quotaManagementService"],
        "image": "//go/pkg/quota_management/cmd/main:quota-management-service_image",
    },
    "switch-config-saver": {
        "chart": "//deployment/charts/switch-config-saver:chart",
        "components": ["sdn"],
        "image": "//go/pkg/switch-config-saver:switch-config-saver_image",
    },
    "telemetry": {
        "components": ["telemetry"],
    },
    "tftp-server": {
        "chart": "//deployment/charts/tftp-server:chart",
        "components": ["tftpServer"],
    },
    "trade-scanner": {
        "chart": "//deployment/charts/trade-scanner:chart",
        "components": ["tradeScanner"],
        "image": "//go/pkg/trade_scanner/cmd/trade_scanner:trade_scanner_image",
    },
    "training-api-server": {
        "chart": "//deployment/charts/training-api-server:chart",
        "components": ["training"],
        "image": "//go/pkg/training/api_server/cmd/main:training-api-server_image",
    },
    "usage": {
        "chart": "//deployment/charts/usage:chart",
        "components": ["metering"],
        "image": "//go/svc/usage/cmd:usage_image",
    },
    "vast-metering-monitor": {
        "chart": "//deployment/charts/vast-metering-monitor:chart",
        "components": ["storage"],
        "image": "//go/pkg/storage/vast_metering_monitor:vast-metering-monitor_image",
    },
    "vast-storage-operator": {
        "chart": "//deployment/charts/vast-storage-operator:chart",
        "components": ["storage"],
        "image": "//go/pkg/storage/vast_storage_operator:vast_storage_operator_image",
    },
    "user-credentials": {
        "chart": "//deployment/charts/user-credentials:chart",
        "components": ["cloudaccount"],
        "image": "//go/svc/user-credentials:user-credentials_image",
    },
    "vm-instance-operator": {
        "chart": "//deployment/charts/vm-instance-operator:chart",
        "components": ["computeVmInstanceOperator"],
        "image": "//go/pkg/instance_operator/vm:vm_instance_operator_image",
    },
    "vm-instance-scheduler": {
        "chart": "//deployment/charts/vm-instance-scheduler:chart",
        "components": ["computeVmInstanceScheduler"],
        "image": "//go/pkg/instance_scheduler/vm:vm_instance_scheduler_image",
    },
    "vm-machine-image-resources": {
        "chart": "//deployment/charts/vm-machine-image-resources:chart",
        "components": ["computePopulateMachineImage"],
    },
}
