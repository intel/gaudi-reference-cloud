// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import wekaV4 from 'k6/x/intel_internal/weka/v4';
import { check } from 'k6';
import { sleep } from 'k6';
import grpc from 'k6/net/grpc';

const client = new grpc.Client();
client.load([
    '../../../../../_main/api',
    '../../../../../googleapis~override',
    '../../../../../_main~non_module_dependencies~com_github_bufbuild_protovalidate/proto/protovalidate/buf/validate/_virtual_imports/validate_proto',
    '../../../../../_main~non_module_dependencies~com_github_bufbuild_protovalidate/proto/protovalidate/buf/validate/_virtual_imports/expression_proto',
    '../../../../../_main~non_module_dependencies~com_github_bufbuild_protovalidate/proto/protovalidate/buf/validate/priv/_virtual_imports/priv_proto',
    '../../../..'
],
    __ENV.S3_PROTO,
    __ENV.USER_PROTO,
    __ENV.WEKA_FILESYSTEM_PROTO,
    __ENV.WEKA_STATEFULCLIENT_PROTO,
);

export let options = {
    iterations: 1,
    thresholds: {
        "checks": [{ threshold: "rate==1.0", abortOnFail: false }]
    }
};

const ns1Auth = {
    "basic": {
        "principal": "nsUser",
        "credentials": "nsPassword",
    }
}

const ns2Auth = {
    "basic": {
        "principal": "nsUser2",
        "credentials": "nsPassword2",
    }
}


export default function () {
    const wekaPort = 14000;
    wekaV4.startAPI(wekaPort, false);

    // Give server time to get up
    sleep(0.3);

    client.connect('localhost:50051', {
        plaintext: true,
        timeout: "2s",
    });


    let ns1Id
    let ns2Id

    let ns1Fs1Id
    let ns1Fs2Id
    let ns2Fs1Id

    let ns1User1Id
    let ns1User2Id
    let ns2User1Id

    let s3UserId
    let bucketId
    let lrId

    let sc1Id
    let sc2Id

    check(client.invoke('intel.storagecontroller.v1.NamespaceService/CreateNamespace',
        {
            "cluster_id": {
                "uuid": "00000000-0000-0000-0000-000000000000",
            },
            "name": "NS_1",
            "quota": {
                "total_bytes": 500000000000000
            },
            "admin_user": {
                "name": ns1Auth.basic.principal,
                "password": ns1Auth.basic.credentials,
            }
        }), {
        "Is ok": (r) => {
            ns1Id = r.message.namespace.id
            return r && r.status === grpc.StatusOK
        }
    });

    check(client.invoke('intel.storagecontroller.v1.NamespaceService/CreateNamespace',
        {
            "cluster_id": {
                "uuid": "00000000-0000-0000-0000-000000000000",
            },
            "name": "NS_2",
            "quota": {
                "total_bytes": 100000000000000
            },
            "admin_user": {
                "name": ns2Auth.basic.principal,
                "password": ns2Auth.basic.credentials,
            }
        }), {
        "Is ok": (r) => {
            ns2Id = r.message.namespace.id
            return r && r.status === grpc.StatusOK
        }
    });

    check(client.invoke('intel.storagecontroller.v1.NamespaceService/GetNamespace',
        {
            "namespace_id": ns1Id
        }), {
        'Namespace1 added': (r) => r.message.namespace.quota.totalBytes === "500000000000000"
    });

    check(client.invoke('intel.storagecontroller.v1.NamespaceService/GetNamespace',
        {
            "namespace_id": ns2Id
        }), {
        'Namespace2 added': (r) => r.message.namespace.quota.totalBytes === "100000000000000"
    });

    check(client.invoke('intel.storagecontroller.v1.NamespaceService/UpdateNamespace',
        {
            "namespace_id": ns1Id,
            "quota": {
                "total_bytes": 40000
            }
        }), {
        "Is ok": (r) => r && r.status === grpc.StatusOK,
    });

    check(client.invoke('intel.storagecontroller.v1.NamespaceService/GetNamespace',
        {
            "namespace_id": ns1Id
        }), {
        'Namespace modified': (r) => r.message.namespace.quota.totalBytes === "40000"
    });


    check(client.invoke('intel.storagecontroller.v1.NamespaceService/ListNamespaces',
        {
            "cluster_id": {
                "uuid": "00000000-0000-0000-0000-000000000000",
            },
        }), {
        'Correct number of namespaces': (r) => r.message.namespaces.length === 2
    });

    check(client.invoke('intel.storagecontroller.v1.NamespaceService/ListNamespaces',
        {
            "cluster_id": {
                "uuid": "00000000-0000-0000-0000-000000000000",
            },
            "filter": {
                "names": ["NS_1"]
            }
        }), {
        'Correct number of namespaces': (r) => r.message.namespaces.length === 1
    });


    check(client.invoke('intel.storagecontroller.v1.weka.FilesystemService/CreateFilesystem',
        {
            "namespace_id": ns1Id,
            "name": "fs1",
            "total_bytes": 4000,
            "encrypted": true,
            "auth_required": true,
            "auth_ctx": ns1Auth
        }), {
        "Is ok": (r) => {
            ns1Fs1Id = r.message.filesystem.id
            return r && r.status === grpc.StatusOK
        }
    })

    check(client.invoke('intel.storagecontroller.v1.weka.FilesystemService/CreateFilesystem',
        {
            "namespace_id": ns1Id,
            "name": "fs2",
            "total_bytes": 2000,
            "encrypted": false,
            "auth_required": false,
            "auth_ctx": ns1Auth
        }), {
        "Is ok": (r) => {
            ns1Fs2Id = r.message.filesystem.id
            return r && r.status === grpc.StatusOK
        }
    })

    check(client.invoke('intel.storagecontroller.v1.weka.FilesystemService/CreateFilesystem',
        {
            "namespace_id": ns1Id,
            "name": "fs1",
            "total_bytes": 4000,
            "encrypted": true,
            "auth_required": true,
            "auth_ctx": ns1Auth
        }), {
        "Is ok": (r) => {
            return r && r.status !== grpc.StatusOK
        }
    })

    check(client.invoke('intel.storagecontroller.v1.weka.FilesystemService/CreateFilesystem',
        {
            "namespace_id": ns2Id,
            "name": "fs1",
            "total_bytes": 4000,
            "encrypted": true,
            "auth_required": true,
            "auth_ctx": ns2Auth
        }), {
        "Is ok": (r) => {
            ns2Fs1Id = r.message.filesystem.id
            return r && r.status === grpc.StatusOK
        }
    })

    check(client.invoke('intel.storagecontroller.v1.weka.FilesystemService/GetFilesystem',
        {
            "filesystem_id": ns1Fs1Id,
            "auth_ctx": ns1Auth
        }), {
        'fs1 added correctly in correct format': (r) => {
            let filesystem = r.message.filesystem
            return filesystem.capacity.totalBytes == "4000" && filesystem.authRequired && filesystem.isEncrypted && filesystem.name === "fs1"
        }
    });

    check(client.invoke('intel.storagecontroller.v1.weka.FilesystemService/GetFilesystem',
        {
            "filesystem_id": ns1Fs2Id,
            "auth_ctx": ns1Auth
        }), {
        'fs2 added correctly in correct format': (r) => {
            let filesystem = r.message.filesystem
            return filesystem.capacity.totalBytes == "2000" && !filesystem.authRequired && !filesystem.isEncrypted && filesystem.name === "fs2"
        }
    });

    check(client.invoke('intel.storagecontroller.v1.weka.FilesystemService/GetFilesystem',
        {
            "filesystem_id": ns2Fs1Id,
            "auth_ctx": ns2Auth
        }), {
        'fs1 added correctly in correct format': (r) => {
            let filesystem = r.message.filesystem
            return filesystem.capacity.totalBytes == "4000" && filesystem.authRequired && filesystem.isEncrypted && filesystem.name === "fs1"
        }
    });


    check(client.invoke('intel.storagecontroller.v1.weka.FilesystemService/UpdateFilesystem',
        {
            "filesystem_id": ns2Fs1Id,
            "auth_ctx": ns2Auth,
            "new_name": "fs3"
        }), {
        'Is ok': (r) => {
            return r && r.status === grpc.StatusOK
        }
    });

    check(client.invoke('intel.storagecontroller.v1.weka.FilesystemService/GetFilesystem',
        {
            "filesystem_id": ns2Fs1Id,
            "auth_ctx": ns2Auth
        }), {
        'fs1 name changed': (r) => {
            let filesystem = r.message.filesystem
            return filesystem.capacity.totalBytes == "4000" && filesystem.authRequired && filesystem.isEncrypted && filesystem.name === "fs3"
        }
    });

    check(client.invoke('intel.storagecontroller.v1.weka.FilesystemService/ListFilesystems',
        {
            "namespace_id": ns1Id,
            "auth_ctx": ns1Auth
        }), {
        'Correct number of filesystems': (r) => r.message.filesystems.length === 2
    });

    check(client.invoke('intel.storagecontroller.v1.weka.FilesystemService/ListFilesystems',
        {
            "namespace_id": ns1Id,
            "auth_ctx": ns1Auth,
            "filter": {
                "names": ["fs2"]
            }
        }), {
        'Correct number of filesystems': (r) => r.message.filesystems.length === 1
    });


    check(client.invoke('intel.storagecontroller.v1.UserService/CreateUser',
        {
            "namespace_id": ns1Id,
            "auth_ctx": ns1Auth,
            "user_name": "user1",
            "user_password": "password1",
        }), {
        "Is ok": (r) => {
            ns1User1Id = r.message.user.id
            return r && r.status === grpc.StatusOK
        }
    });

    check(client.invoke('intel.storagecontroller.v1.UserService/CreateUser',
        {
            "namespace_id": ns1Id,
            "auth_ctx": ns1Auth,
            "user_name": "user2",
            "user_password": "password2",
        }), {
        "Is ok": (r) => {
            ns1User2Id = r.message.user.id
            return r && r.status === grpc.StatusOK
        }
    });

    check(client.invoke('intel.storagecontroller.v1.UserService/CreateUser',
        {
            "namespace_id": ns2Id,
            "auth_ctx": ns2Auth,
            "user_name": "user1",
            "user_password": "password1",
        }), {
        "Is ok": (r) => {
            ns2User1Id = r.message.user.id
            return r && r.status === grpc.StatusOK
        }
    });

    check(client.invoke('intel.storagecontroller.v1.UserService/ListUsers',
        {
            "namespace_id": ns1Id,
            "auth_ctx": ns1Auth
        }), {
        'User added NS1': (r) => r.message.users.length === 3
    });

    check(client.invoke('intel.storagecontroller.v1.UserService/ListUsers',
        {
            "namespace_id": ns1Id,
            "auth_ctx": ns1Auth,
            "filter": {
                "names": ["user1"]
            }
        }), {
        'User added NS1': (r) => r.message.users.length === 1
    });

    check(client.invoke('intel.storagecontroller.v1.UserService/GetUser',
        {
            "user_id": ns1User1Id,
            "auth_ctx": ns1Auth,
        }), {
        'User added NS1': (r) => r.message.user.name === "user1"
    });

    check(client.invoke('intel.storagecontroller.v1.UserService/GetUser',
        {
            "user_id": ns1User2Id,
            "auth_ctx": ns1Auth,
        }), {
        'User added NS1': (r) => r.message.user.name === "user2"
    });

    check(client.invoke('intel.storagecontroller.v1.UserService/GetUser',
        {
            "user_id": ns2User1Id,
            "auth_ctx": ns2Auth,
        }), {
        'User added NS2': (r) => r.message.user.name === "user1"
    });

    check(client.invoke('intel.storagecontroller.v1.UserService/GetUser',
        {
            "user_id": ns1User1Id,
            "auth_ctx": {
                "basic": {
                    "principal": "user1",
                    "credentials": "password1",
                }
            },
        }), {
        'Password correct': (r) => r.message.user.name === "user1"
    });

    check(client.invoke('intel.storagecontroller.v1.UserService/UpdateUserPassword',
        {
            "user_id": ns1User1Id,
            "auth_ctx": ns1Auth,
            "new_password": "password11"
        }), {
        'UpdateUserPassword status is OK': (r) => r && r.status === grpc.StatusOK,
    });

    check(client.invoke('intel.storagecontroller.v1.UserService/GetUser',
        {
            "user_id": ns1User1Id,
            "auth_ctx": {
                "basic": {
                    "principal": "user1",
                    "credentials": "password1",
                }
            },
        }), {
        'Password changed': (r) => r.status !== grpc.StatusOK
    });

    check(client.invoke('intel.storagecontroller.v1.UserService/GetUser',
        {
            "user_id": ns1User1Id,
            "auth_ctx": {
                "basic": {
                    "principal": "user1",
                    "credentials": "password11",
                }
            },
        }), {
        'Password changed': (r) => r.message.user.name === "user1"
    });

    check(client.invoke('intel.storagecontroller.v1.NamespaceService/DeleteNamespace',
        {
            "namespace_id": ns1Id,
        }), {
        'DeleteNamespace status is OK': (r) => r && r.status === grpc.StatusOK,
    });

    check(client.invoke('intel.storagecontroller.v1.NamespaceService/ListNamespaces',
        {
            "cluster_id": {
                "uuid": "00000000-0000-0000-0000-000000000000",
            },
        }), {
        'Namespace is not found': (r) => r.message.namespaces.length === 1
    });

    check(client.invoke('intel.storagecontroller.v1.S3Service/CreateBucket',
        {
            "cluster_id": {
                "uuid": "00000000-0000-0000-0000-000000000000",
            },
            "name": "bucketname",
            "access_policy": 1
        }), {
        "Is ok": (r) => {
            bucketId = r.message.bucket.id
            return r && r.status === grpc.StatusOK
        }
    });

    check(client.invoke('intel.storagecontroller.v1.S3Service/CreateBucket',
        {
            "cluster_id": {
                "uuid": "00000000-0000-0000-0000-000000000000",
            },
            "name": "bucketName",
            "access_policy": 1
        }), {
        "Fail create bucket on invalid name": (r) => {
            return r && r.status !== grpc.StatusOK
        }
    });

    check(client.invoke('intel.storagecontroller.v1.S3Service/ListBuckets',
        {
            "cluster_id": {
                "uuid": "00000000-0000-0000-0000-000000000000",
            },
        }), {
        'Bucket added': (r) => r.message.buckets.length === 1
    });

    check(client.invoke('intel.storagecontroller.v1.S3Service/GetBucketPolicy',
        {
            "bucket_id": bucketId,
        }), {
        'Bucket policy correct': (r) => r.message.policy === "BUCKET_ACCESS_POLICY_NONE"
    });

    check(client.invoke('intel.storagecontroller.v1.S3Service/UpdateBucketPolicy',
        {
            "bucket_id": bucketId,
            "access_policy": 2
        }), {
        'UpdateBucketPolicy status is OK': (r) => r && r.status === grpc.StatusOK,
    });

    check(client.invoke('intel.storagecontroller.v1.S3Service/GetBucketPolicy',
        {
            "bucket_id": bucketId,
        }), {
        'Bucket policy updated': (r) => r.message.policy === "BUCKET_ACCESS_POLICY_READ"
    });

    check(client.invoke('intel.storagecontroller.v1.S3Service/CreateLifecycleRules',
        {
            "bucket_id": bucketId,
            "lifecycle_rules": [{
                "prefix": "pref",
                "expire_days": 5
            }]
        }), {
        "Is ok": (r) => {
            return r && r.status === grpc.StatusOK
        }
    });

    check(client.invoke('intel.storagecontroller.v1.S3Service/UpdateLifecycleRules',
        {
            "bucket_id": bucketId,
            "lifecycle_rules": [{
                "prefix": "pref",
                "expire_days": 5
            }]
        }), {
        'UpdateLifecycleRules status is OK': (r) => {
            return r && r.status === grpc.StatusOK
        }
    });

    check(client.invoke('intel.storagecontroller.v1.S3Service/ListLifecycleRules',
        {
            "bucket_id": bucketId,
        }), {
        'lr listed': (r) => {
            return r.message.lifecycleRules.length === 1
        }
    });

    check(client.invoke('intel.storagecontroller.v1.S3Service/CreateLifecycleRules',
        {
            "bucket_id": bucketId,
            "lifecycle_rules": [{
                "prefix": "pref",
                "expire_days": 5
            },
            {
                "prefix": "pref",
                "expire_days": 5
            }]
        }), {
        "Is ok": (r) => {
            return r && r.status === grpc.StatusOK
        }
    });

    check(client.invoke('intel.storagecontroller.v1.S3Service/ListLifecycleRules',
        {
            "bucket_id": bucketId,
        }), {
        'lr listed': (r) => {
            return r.message.lifecycleRules.length === 2
        }
    });

    check(client.invoke('intel.storagecontroller.v1.S3Service/DeleteLifecycleRules',
        {
            "bucket_id": bucketId,
        }), {
        'DeleteLifecycleRules status is OK': (r) => r && r.status === grpc.StatusOK,
    });

    check(client.invoke('intel.storagecontroller.v1.S3Service/ListLifecycleRules',
        {
            "bucket_id": bucketId,
        }), {
        'lr listed': (r) => {
            return r.message.lifecycleRules.length === 0
        }
    });

    check(client.invoke('intel.storagecontroller.v1.S3Service/CreateS3Principal',
        {
            "cluster_id": {
                "uuid": "00000000-0000-0000-0000-000000000000",
            },
            "name": "s3Principal",
            "credentials": "credentials",
        }), {
        "Is ok": (r) => {
            s3UserId = r.message.s3Principal.id
            return r && r.status === grpc.StatusOK
        }
    });

    check(client.invoke('intel.storagecontroller.v1.S3Service/GetS3Principal',
        {
            "principal_id": s3UserId,
        }), {
        'principal created': (r) => {
            let principal = r.message.s3Principal
            return principal.name === "s3Principal" && principal.policies.length === 0
        }
    });

    check(client.invoke('intel.storagecontroller.v1.S3Service/UpdateS3PrincipalPolicies',
        {
            "principal_id": s3UserId,
            "policies": [
                {
                    "bucket_id": bucketId,
                    "read": true
                }
            ]
        }), {
        "Is ok": (r) => {
            return r && r.status === grpc.StatusOK
        }
    });

    check(client.invoke('intel.storagecontroller.v1.S3Service/GetS3Principal',
        {
            "principal_id": s3UserId,
        }), {
        'policies added': (r) => {
            let principal = r.message.s3Principal
            return principal.policies.length === 1
        }
    });

    check(client.invoke('intel.storagecontroller.v1.weka.StatefulClientService/CreateStatefulClient',
        {
            "cluster_id": {
                "uuid": "00000000-0000-0000-0000-000000000000",
            },
            "name": "SC_1",
            // IP address range 192.0.2.0/24, also known as TEST-NET-1,
            // is reserved for use in documentation and example code.
            // More info RFC 5737
            "ip": "192.0.2.10"
        }), {
        "Is SC_1 statefulclient created - status is OK": (r) => {
            sc1Id = r.message.statefulClient.id
            return r && r.status === grpc.StatusOK
        }            
    });

    check(client.invoke('intel.storagecontroller.v1.weka.StatefulClientService/CreateStatefulClient',
        {
            "cluster_id": {
                "uuid": "00000000-0000-0000-0000-000000000000",
            },
            "name": "SC_2",
            // IP address range 192.0.2.0/24, also known as TEST-NET-1,
            // is reserved for use in documentation and example code.
            // More info RFC 5737
            "ip": "192.0.2.210"
        }), {
        "Is SC_2 statefulclient created - status is OK": (r) => {
            sc2Id = r.message.statefulClient.id
            return r && r.status === grpc.StatusOK
        }
    });

    check(client.invoke('intel.storagecontroller.v1.weka.StatefulClientService/GetStatefulClient',
        {
            "stateful_client_id": sc1Id
        }), {
        'statefulclient SC_1 added - status is OK': (r) => r.message.statefulClient.name === "SC_1"
    });

    check(client.invoke('intel.storagecontroller.v1.weka.StatefulClientService/GetStatefulClient',
        {
            "stateful_client_id": sc2Id
        }), {
        'statefulclient SC_2 added - status is OK': (r) => r.message.statefulClient.name === "SC_2"
    });

    check(client.invoke('intel.storagecontroller.v1.weka.StatefulClientService/ListStatefulClients',
        {
            "cluster_id": {
                "uuid": "00000000-0000-0000-0000-000000000000",
            },
        }), {
        'Correct number of lists of stateful clients -> 2': (r) => r.message.statefulClients.length === 2
    });

    check(client.invoke('intel.storagecontroller.v1.weka.StatefulClientService/ListStatefulClients',
        {
            "cluster_id": {
                "uuid": "00000000-0000-0000-0000-000000000000",
            },
            "filter": {
                "names": ["SC_1"]
            }
        }), {
        'Correct number of lists of stateful clients -> 1': (r) => r.message.statefulClients.length === 1
    });

    check(client.invoke('intel.storagecontroller.v1.weka.StatefulClientService/DeleteStatefulClient',
        {
            "stateful_client_id": sc2Id,
        }), {
        'Delete SC_2 statefulclient - status is OK': (r) => r && r.status === grpc.StatusOK,
    });
}
