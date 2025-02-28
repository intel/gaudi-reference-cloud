<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# S3 Flows

Contains main flows of S3 interactions

All commands assume stage environment and '918b5026-d516-48c8-bfd3-5998547265b2' cluster

## Bucket Creation and Access
### Resource Creation
Create Bucket
```bash
./grpcurl -cacert ca.crt  -proto s3.proto -d '{"cluster_id": {"uuid": "918b5026-d516-48c8-bfd3-5998547265b2"}, "name": "bucket10", "quota_bytes": 53687091200, "access_policy": 1}' internal-placeholder.com:443 intel.storagecontroller.v1.S3Service.CreateBucket
```

Create Principal
```bash
./grpcurl -cacert ca.crt  -proto s3.proto -d '{"cluster_id": {"uuid": "918b5026-d516-48c8-bfd3-5998547265b2"}, "name": "s3principal1", "credentials": "S3principalcreds" }' internal-placeholder.com:443 intel.storagecontroller.v1.S3Service.CreateS3Principal
```
Record output principal ID, it will be needed later

### Giving Access

```bash
./grpcurl -cacert ca.crt  -proto s3.proto -d '{"principal_id": {"cluster_id": {"uuid": "918b5026-d516-48c8-bfd3-5998547265b2"}, "id": "PRINCIPAL_ID"}, "policies": [{"bucket_id": {"cluster_id": {"uuid": "918b5026-d516-48c8-bfd3-5998547265b2"}, "id": "bucket10"}, "read": true, "write": true, "delete": true}]}' internal-placeholder.com:443 intel.storagecontroller.v1.S3Service.UpdateS3PrincipalPolicies
```

### S3 API Check

```python
#!/usr/bin/env/python
import boto3
import logging
from botocore.exceptions import ClientError
from botocore.client import Config
config = Config(
  signature_version = 's3v4'
)
s3 = boto3.resource('s3',
                    endpoint_url='https://s3w-pdx05-2.us-staging-1.cloud.intel.com:9000',
                    aws_access_key_id='s3principal1',
                    aws_secret_access_key='S3principalcreds',
                    config=config)
try:
  s3.Bucket('bucket10').upload_file('s3.proto','s3_proto_object')
  s3.Bucket('bucket10').download_file('s3_proto_object', 's3.proto.dl')
except ClientError as e:
        logging.error(e)
```

### Cleanup
Bucket
```bash
./grpcurl -cacert ca.crt  -proto s3.proto -d '{"bucket_id": {"cluster_id": {"uuid": "918b5026-d516-48c8-bfd3-5998547265b2"}, "id": "bucket10"}}' internal-placeholder.com:443 intel.storagecontroller.v1.S3Service.DeleteBucket
```

Principal
```bash
./grpcurl -cacert ca.crt  -proto s3.proto -d '{"principal_id": {"cluster_id": {"uuid": "918b5026-d516-48c8-bfd3-5998547265b2"}, "id": "PRINCIPAL_ID"}}' internal-placeholder.com:443 intel.storagecontroller.v1.S3Service.DeleteS3Principal
```

## Bucket List and Query
Buckets
```bash
./grpcurl -cacert ca.crt  -proto s3.proto -d '{"cluster_id": {"uuid": "918b5026-d516-48c8-bfd3-5998547265b2"}}' internal-placeholder.com:443 intel.storagecontroller.v1.S3Service.ListBuckets
```

## Principal Management
Create Principal
```bash
./grpcurl -cacert ca.crt  -proto s3.proto -d '{"cluster_id": {"uuid": "918b5026-d516-48c8-bfd3-5998547265b2"}, "name": "s3principal1", "credentials": "S3principalcreds" }' internal-placeholder.com:443 intel.storagecontroller.v1.S3Service.CreateS3Principal
```

Get Principal
```bash
./grpcurl -cacert ca.crt  -proto s3.proto -d '{"principal_id": {"cluster_id": {"uuid": "918b5026-d516-48c8-bfd3-5998547265b2"}, "id": "PRINCIPAL_ID"}}' internal-placeholder.com:443 intel.storagecontroller.v1.S3Service.GetS3Principal
```

Set Credentials
```bash
./grpcurl -cacert ca.crt  -proto s3.proto -d '{"principal_id": {"cluster_id": {"uuid": "918b5026-d516-48c8-bfd3-5998547265b2"}, "id": "PRINCIPAL_ID"}, "credentials": "S3principalcreds2"}' internal-placeholder.com:443 intel.storagecontroller.v1.S3Service.SetS3PrincipalCredentials
```

### Cleanup
Principal
```bash
./grpcurl -cacert ca.crt  -proto s3.proto -d '{"principal_id": {"cluster_id": {"uuid": "918b5026-d516-48c8-bfd3-5998547265b2"}, "id": "PRINCIPAL_ID"}}' internal-placeholder.com:443 intel.storagecontroller.v1.S3Service.DeleteS3Principal
```
