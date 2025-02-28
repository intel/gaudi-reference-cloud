.. _enable_disable_grcp_service:

Enable/Disable GRCP Service
###########################

This document outlines the steps to enable or disable a GRCP Service for a specific environment.

Enable/Disable Service
**********************

To enable a specific API service in a given environment, the deployment variable  ```enabledServices``` must include an entry with the service name set to true.
If the service name is not added or the value of the entry is false, the necessary endpoints will not be added to the Envoy configuration of the gRPC proxy.
This allows fine-grained control over which services are exposed in each environment.


Enabling a Service in All Environments
**************************************

If you need to enable a service across all environments, or if you require the service to be available in the Jenkins test pipeline,
add the service name to ```deployment/helmfile/defaults.yaml.gotmpl``` and set its value to true under the enabledServices key.

Enabling a Service in Only a Specific Environment
*************************************************

If you need to enable a service in a specific environment, you have two options:

1. Remove the service entry from ```deployment/helmfile/defaults.yaml.gotmpl```. Then, add it to the appropriate environment-specific file located in deployment/helmfile/environments.
This is the most straightforward solution if the service is not required for testing in the Jenkins test pipeline.

2. Enable the service in ```deployment/helmfile/defaults.yaml.gotmpl``` and set it to false in each environment where it should be disabled.