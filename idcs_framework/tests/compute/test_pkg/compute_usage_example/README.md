## How to create compute instance using compute util

Make use of the compute util (/compute/framework_pkg/compute_util.go) to create any compute instance (VM or BM).

Sample payloads are added below, Refer swagger for more details.
Swagger link - https://<<envname>>-internal-placeholder.com/openapiv2

### Instance Creation (POST)

```
Endpoint- /v1/cloudaccounts/{metadata.cloudAccountId}/instances
```

#### sample payload
```{
"metadata": {
    "name": "<<instance-name>>"
},
"spec": {
    "availabilityZone": "<<availability-zone>>",  # Staging Example: us-staging-1a
    "instanceType": "<<instance-type>>",          # type of instance (vm, bm, lb)
    "machineImage": "<<machine-image>>",          # machine image for selected instance type
    "runStrategy": "RerunOnFailure",             
    "sshPublicKeyNames": [
    "<<ssh-key-name>>"                          # SSH key should be created as pre-requisite
    ],
    "interfaces": [
    {
        "name": "eth0",
        "vNet": "<<vnet-name>>"                   # VNET should be present, staging ex: us-staging-1a-default
    }
    ]
}
}
```

### Retrieve All Instance (GET)

```
Endpoint - /v1/cloudaccounts/{metadata.cloudAccountId}/instances
payload - Not Applicable
```

### Retrieve Instance by ID (GET)

```
Endpoint - /v1/cloudaccounts/{metadata.cloudAccountId}/instances/id/{metadata.resourceId}
payload - Not Applicable
```

### Retrieve Instance by Name (GET)

```
Endpoint - /v1/cloudaccounts/{metadata.cloudAccountId}/instances/id/{metadata.name}
payload - Not Applicable
```

### Delete an Instance by ID (DELETE)

```
Endpoint - /v1/cloudaccounts/{metadata.cloudAccountId}/instances/id/{metadata.resourceId}
payload - Not Applicable
```

### Delete an Instance by Name (DELETE)

```
Endpoint - /v1/cloudaccounts/{metadata.cloudAccountId}/instances/id/{metadata.name}
payload - Not Applicable
```