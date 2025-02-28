// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import wekaV4 from 'k6/x/intel_internal/weka/v4';
import { check } from 'k6';
import { sleep } from 'k6';
import grpc from 'k6/net/grpc';
import {
  deepEquals,
} from './util.js';

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
);

export let options = {
  iterations: 1,
  thresholds: {
    "checks": [{ threshold: "rate==1.0", abortOnFail: false }]
  }
};

const adminLogin = { "path": "/api/v2/login", "method": "POST", "headers": { "User-Agent": ["Go-http-client/1.1"], "Content-Length": ["60"], "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"] }, "body": { "org": "Root", "password": "adminPassword", "username": "admin" } }

const clusterStatus = {
  "name": "test",
  "location": "localhost",
  "type": "TYPE_WEKA",
  "labels": { "category": "testing-infra", "operation": "testing", "wekaName": "stewie", "wekaGuid": "9724f5ec-a68c-437d-8411-03c8425c06b8" },
  "supportsApi": ["API_TYPE_WEKA_FILESYSTEM", "API_TYPE_OBJECT_STORE"],
  "capacity": {
    "storage": { "totalBytes": "1829454741504", "availableBytes": "0" },
    "namespaces": { "totalCount": 256, "availableCount": 254 },
    "filesystems": null, "users": null
  },
  "health": { "status": "STATUS_HEALTHY" },
  "id": { "uuid": "00000000-0000-0000-0000-000000000000" }
};

const authProperties = {
  "basic": {
    "principal": "admin",
    "credentials": "testPassword",
  }
}

const namespace = {
  "id": { "clusterId": { "uuid": "00000000-0000-0000-0000-000000000000" }, "id": "uid_string" },
  "name": "ORG-1",
  "quota": { "totalBytes": "20000002048" },
  "ipFilters": []
}


const filesystem = {
  "id": { "id": "uid_string", "namespaceId": { "clusterId": { "uuid": "00000000-0000-0000-0000-000000000000" }, "id": "uid_string" } },
  "name": "fs1",
  "status": "STATUS_CREATING",
  "isEncrypted": true,
  "authRequired": true,
  "capacity": { "totalBytes": "1073741824", "availableBytes": "1073741824" },
  "backend": "localhost",
}

const user = {
  "id": { "namespaceId": { "clusterId": { "uuid": "00000000-0000-0000-0000-000000000000" }, "id": "uid_string" }, "id": "uid_string" },
  "name": "admin",
  "role": "ROLE_UNSPECIFIED"
}

let authHeader = ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"]

export default function () {
  const wekaPort = 14000;
  wekaV4.startAPI(wekaPort, true);

  // Give server time to get up
  sleep(0.3);

  client.connect('localhost:50051', {
    plaintext: true,
    timeout: "2s",
  });

  check(client.invoke('intel.storagecontroller.v1.ClusterService/ListClusters', {}), {
    'ListClusters status is OK': (r) => r && r.status === grpc.StatusOK,
    'ListClusters in correct format': (r) => deepEquals(r.message.clusters[0], clusterStatus),
    'ListClusters should call correct requests': () => {

      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        adminLogin,
        { "path": "/api/v2/cluster", "method": "GET", "headers": { "User-Agent": ["Go-http-client/1.1"], "Authorization": authHeader, "Accept-Encoding": ["gzip"] }, "body": null },
        { "path": "/api/v2/organizations", "method": "GET", "headers": { "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"], "Authorization": authHeader }, "body": null }])
    }
  });


  check(client.invoke('intel.storagecontroller.v1.ClusterService/GetCluster',
    {
      "cluster_id": {
        "uuid": "00000000-0000-0000-0000-000000000000",
      },
    }), {
    'GetCluster status is OK': (r) => r && r.status === grpc.StatusOK,
    'GetCluster in correct format': (r) => deepEquals(r.message.cluster, clusterStatus),
    'GetCluster should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        { "path": "/api/v2/login", "method": "POST", "headers": { "Content-Length": ["60"], "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"] }, "body": { "password": "adminPassword", "username": "admin", "org": "Root" } },
        { "path": "/api/v2/cluster", "method": "GET", "headers": { "User-Agent": ["Go-http-client/1.1"], "Authorization": authHeader, "Accept-Encoding": ["gzip"] }, "body": null },
        { "path": "/api/v2/organizations", "method": "GET", "headers": { "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"], "Authorization": authHeader }, "body": null }])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.NamespaceService/GetNamespace',
    {
      "namespace_id": {
        "cluster_id": {
          "uuid": "00000000-0000-0000-0000-000000000000",
        },
        "id": "uid_string"
      }
    }), {
    'GetNamespace status is OK': (r) => r && r.status === grpc.StatusOK,
    'GetNamespace in correct format': (r) => deepEquals(r.message.namespace, namespace),
    'GetNamespace should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs,
        [
          adminLogin,
          { "path": "/api/v2/organizations/uid_string", "method": "GET", "headers": { "User-Agent": ["Go-http-client/1.1"], "Authorization": authHeader, "Accept-Encoding": ["gzip"] }, "body": null }
        ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.NamespaceService/UpdateNamespace',
    {
      "namespace_id": {
        "cluster_id": {
          "uuid": "00000000-0000-0000-0000-000000000000",
        },
        "id": "uid_string"
      },
      "quota": {
        "total_bytes": 40000
      }
    }), {
    'UpdateNamespace status is OK': (r) => r && r.status === grpc.StatusOK,
    'UpdateNamespace in correct format': (r) => deepEquals(r.message.namespace, namespace),
    'UpdateNamespace should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        { "path": "/api/v2/login", "method": "POST", "headers": { "User-Agent": ["Go-http-client/1.1"], "Content-Length": ["60"], "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"] }, "body": { "password": "adminPassword", "username": "admin", "org": "Root" } },
        { "path": "/api/v2/organizations/uid_string/limits", "method": "PUT", "headers": { "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"], "Content-Length": ["21"], "Authorization": authHeader }, "body": { "total_quota": 40000 } }
      ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.NamespaceService/CreateNamespace',
    {
      "cluster_id": {
        "uuid": "00000000-0000-0000-0000-000000000000",
      },
      "name": "NS1",
      "admin_user": {
        "name": "username",
        "password": "password",
      },
      "quota": {
        "total_bytes": 100000000000000
      },
    }), {
    'CreateNamespace status is OK': (r) => r && r.status === grpc.StatusOK,
    'CreateNamespace in correct format': (r) => deepEquals(r.message.namespace, namespace),
    'CreateNamespace should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        adminLogin,
        { "path": "/api/v2/organizations", "method": "POST", "headers": { "Content-Length": ["102"], "Authorization": authHeader, "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"] }, "body": { "total_quota": 100000000000000, "username": "username", "name": "NS1", "password": "password", "ssd_quota": 0 } }])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.NamespaceService/DeleteNamespace',
    {
      "namespace_id": {
        "cluster_id": {
          "uuid": "00000000-0000-0000-0000-000000000000",
        },
        "id": "uid_string"
      },
    }), {
    'DeleteNamespace status is OK': (r) => r && r.status === grpc.StatusOK,
    'DeleteNamespace should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        { "path": "/api/v2/login", "method": "POST", "headers": { "User-Agent": ["Go-http-client/1.1"], "Content-Length": ["60"], "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"] }, "body": { "password": "adminPassword", "username": "admin", "org": "Root" } },
        { "path": "/api/v2/organizations/uid_string", "method": "DELETE", "headers": { "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"], "Authorization": authHeader }, "body": null }
      ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.weka.FilesystemService/GetFilesystem',
    {
      "filesystem_id": {
        "namespace_id": {
          "cluster_id": {
            "uuid": "00000000-0000-0000-0000-000000000000",
          },
          "id": "uid_string"
        },
        "id": "fs_uid_string",
      },
      "auth_ctx": authProperties,
    }), {
    'GetFilesystem status is OK': (r) => r && r.status === grpc.StatusOK,
    'GetFilesystem in correct format': (r) => deepEquals(r.message.filesystem, filesystem),
    'GetFilesystem should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        { "path": "/api/v2/login", "method": "POST", "headers": { "User-Agent": ["Go-http-client/1.1"], "Content-Length": ["60"], "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"] }, "body": { "org": "ORG-1", "password": "testPassword", "username": "admin" } },
        { "path": "/api/v2/fileSystems/fs_uid_string", "method": "GET", "headers": { "User-Agent": ["Go-http-client/1.1"], "Authorization": authHeader, "Accept-Encoding": ["gzip"] }, "body": null }
      ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.weka.FilesystemService/UpdateFilesystem',
    {
      "filesystem_id": {
        "namespace_id": {
          "cluster_id": {
            "uuid": "00000000-0000-0000-0000-000000000000",
          },
          "id": "uid_string"
        },
        "id": "fs_uid_string"
      },
      "new_name": "fs22",
      "auth_ctx": authProperties,
    }), {
    'UpdateFilesystem status is OK': (r) => r && r.status === grpc.StatusOK,
    'UpdateFilesystem in correct format': (r) => deepEquals(r.message.filesystem, filesystem),
    'UpdateFilesystem should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        { "path": "/api/v2/login", "method": "POST", "headers": { "User-Agent": ["Go-http-client/1.1"], "Content-Length": ["60"], "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"] }, "body": { "org": "ORG-1", "password": "testPassword", "username": "admin" } },
        { "path": "/api/v2/fileSystems/fs_uid_string", "method": "PUT", "headers": { "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"], "Content-Length": ["19"], "Authorization": authHeader }, "body": { "new_name": "fs22" } }
      ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.weka.FilesystemService/CreateFilesystem',
    {
      "namespace_id": {
        "cluster_id": {
          "uuid": "00000000-0000-0000-0000-000000000000",
        },
        "id": "uid_string"
      },
      "name": "fs2",
      "total_bytes": 4000,
      "encrypted": true,
      "auth_required": true,
      "auth_ctx": authProperties,
    }), {
    'CreateFilesystem status is OK': (r) => r && r.status === grpc.StatusOK,
    'CreateFilesystem in correct format': (r) => deepEquals(r.message.filesystem, filesystem),
    'CreateFilesystem should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        { "path": "/api/v2/login", "method": "POST", "headers": { "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"], "Content-Length": ["60"] }, "body": { "org": "ORG-1", "password": "testPassword", "username": "admin" } },
        { "path": "/api/v2/fileSystems", "method": "POST", "headers": { "Authorization": authHeader, "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"], "Content-Length": ["103"] }, "body": { "total_capacity": 4000, "auth_required": true, "encrypted": true, "group_name": "tenantfsgroup", "name": "fs2" } }
      ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.weka.FilesystemService/DeleteFilesystem',
    {
      "filesystem_id": {
        "namespace_id": {
          "cluster_id": {
            "uuid": "00000000-0000-0000-0000-000000000000",
          },
          "id": "uid_string"
        },
        "id": "fs_uid_string"
      },
      "auth_ctx": authProperties,
    }), {
    'DeleteFileSystem status is OK': (r) => r && r.status === grpc.StatusOK,
    'DeleteFileSystem should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        { "path": "/api/v2/login", "method": "POST", "headers": { "User-Agent": ["Go-http-client/1.1"], "Content-Length": ["60"], "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"] }, "body": { "password": "testPassword", "username": "admin", "org": "ORG-1" } },
        { "path": "/api/v2/fileSystems/fs_uid_string?purge_from_obs=true", "method": "DELETE", "headers": { "Authorization": authHeader, "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"] }, "body": null },
        { "path": "/api/v2/fileSystems/fs_uid_string", "method": "GET", "headers": { "User-Agent": ["Go-http-client/1.1"], "Authorization": authHeader, "Accept-Encoding": ["gzip"] }, "body": null }
      ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.UserService/GetUser',
    {
      "user_id": {
        "namespace_id": {
          "cluster_id": {
            "uuid": "00000000-0000-0000-0000-000000000000",
          },
          "id": "uid_string"
        },
        "id": "uid_string"
      },
      "auth_ctx": authProperties,
    }), {
    'GetUser status is OK': (r) => r && r.status === grpc.StatusOK,
    'GetUser in correct format': (r) => deepEquals(r.message.user, user),
    'GetUser should call correct requests': () => {
      const reqs = wekaV4.requests

      wekaV4.requests = []
      return deepEquals(reqs, [
        { "path": "/api/v2/login", "method": "POST", "headers": { "Content-Length": ["60"], "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"] }, "body": { "username": "admin", "org": "ORG-1", "password": "testPassword" } },
        { "path": "/api/v2/users", "method": "GET", "headers": { "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"], "Authorization": authHeader }, "body": null }
      ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.UserService/CreateUser',
    {
      "namespace_id": {
        "cluster_id": {
          "uuid": "00000000-0000-0000-0000-000000000000",
        },
        "id": "uid_string"
      },
      "auth_ctx": authProperties,
      "user_name": "user1",
      "user_password": "password1",
    }), {
    'CreateUser status is OK': (r) => r && r.status === grpc.StatusOK,
    'CreateUser in correct format': (r) => deepEquals(r.message.user, user),
    'CreateUser should call correct requests': () => {
      const reqs = wekaV4.requests

      wekaV4.requests = []
      return deepEquals(reqs, [
        { "path": "/api/v2/login", "method": "POST", "headers": { "Content-Length": ["60"], "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"] }, "body": { "username": "admin", "org": "ORG-1", "password": "testPassword" } },
        { "path": "/api/v2/users", "method": "POST", "headers": { "User-Agent": ["Go-http-client/1.1"], "Content-Length": ["60"], "Authorization": authHeader, "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"] }, "body": { "password": "password1", "role": "Regular", "username": "user1" } }
      ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.UserService/DeleteUser',
    {
      "user_id": {
        "namespace_id": {
          "cluster_id": {
            "uuid": "00000000-0000-0000-0000-000000000000",
          },
          "id": "uid_string"
        },
        "id": "user_uid_string"
      },
      "auth_ctx": authProperties,
    }), {
    'DeleteUser status is OK': (r) => r && r.status === grpc.StatusOK,
    'DeleteUser should call correct requests': () => {
      const reqs = wekaV4.requests

      wekaV4.requests = []
      return deepEquals(reqs, [
        { "path": "/api/v2/login", "method": "POST", "headers": { "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"], "Content-Length": ["60"], "Content-Type": ["application/json"] }, "body": { "org": "ORG-1", "password": "testPassword", "username": "admin" } },
        { "path": "/api/v2/users/user_uid_string", "method": "DELETE", "headers": { "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"], "Authorization": authHeader }, "body": null }
      ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.UserService/UpdateUser',
    {
      "user_id": {
        "namespace_id": {
          "cluster_id": {
            "uuid": "00000000-0000-0000-0000-000000000000",
          },
          "id": "uid_string"
        },
        "id": "user_uid_string"
      },
      "role": 1,
      "auth_ctx": authProperties,
    }), {
    'UpdateUser status is OK': (r) => r && r.status === grpc.StatusOK,
    'UpdateUser in correct format': (r) => deepEquals(r.message.user, user),
    'UpdateUser should call correct requests': () => {
      const reqs = wekaV4.requests

      wekaV4.requests = []
      return deepEquals(reqs, [
        { "path": "/api/v2/login", "method": "POST", "headers": { "Content-Length": ["60"], "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"] }, "body": { "org": "ORG-1", "password": "testPassword", "username": "admin" } },
        { "path": "/api/v2/users/user_uid_string", "method": "PUT", "headers": { "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"], "Content-Length": ["19"], "Authorization": authHeader, "Content-Type": ["application/json"] }, "body": { "role": "OrgAdmin" } }
      ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.UserService/UpdateUserPassword',
    {
      "user_id": {
        "namespace_id": {
          "cluster_id": {
            "uuid": "00000000-0000-0000-0000-000000000000",
          },
          "id": "uid_string"
        },
        "id": "user_uid_string"
      },
      "new_password": "password2",
      "auth_ctx": authProperties,
    }), {
    'UpdateUserPassword status is OK': (r) => r && r.status === grpc.StatusOK,
    'UpdateUserPassword should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        { "path": "/api/v2/login", "method": "POST", "headers": { "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"], "Content-Length": ["60"] }, "body": { "org": "ORG-1", "password": "testPassword", "username": "admin" } },
        { "path": "/api/v2/users/user_uid_string/password", "method": "PUT", "headers": { "User-Agent": ["Go-http-client/1.1"], "Content-Length": ["24"], "Authorization": authHeader, "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"] }, "body": { "password": "password2" } }
      ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.S3Service/ListBuckets',
    {
      "cluster_id": {
        "uuid": "00000000-0000-0000-0000-000000000000",
      },
    }), {
    'ListBuckets status is OK': (r) => r && r.status === grpc.StatusOK,
    'ListBuckets should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        adminLogin,
        { "path": "/api/v2/s3/buckets", "method": "GET", "headers": { "User-Agent": ["Go-http-client/1.1"], "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"], "Accept-Encoding": ["gzip"] }, "body": null }
      ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.S3Service/GetBucketPolicy',
    {
      "bucket_id": {
        "cluster_id": {
          "uuid": "00000000-0000-0000-0000-000000000000",
        },
        "id": "uid_string"
      },
    }), {
    'GetBucketPolicy status is OK': (r) => r && r.status === grpc.StatusOK,
    'GetBucketPolicy should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        adminLogin,
        { "path": "/api/v2/s3/buckets/uid_string/policy", "method": "GET", "headers": { "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"], "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"] }, "body": null }
      ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.S3Service/CreateBucket',
    {
      "cluster_id": {
        "uuid": "00000000-0000-0000-0000-000000000000",
      },
      "name": "bucketname",
      "access_policy": 1
    }), {
    'CreateBucket status is OK': (r) => r && r.status === grpc.StatusOK,
    'CreateBucket should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        adminLogin,
        { "path": "/api/v2/s3/buckets", "method": "POST", "headers": { "Content-Length": ["62"], "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"], "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"] }, "body": { "policy": "none", "bucket_name": "bucketname", "hard_quota": "0B" } },
      ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.S3Service/UpdateBucketPolicy',
    {
      "bucket_id": {
        "cluster_id": {
          "uuid": "00000000-0000-0000-0000-000000000000",
        },
        "id": "uid_string"
      },
      "access_policy": 2
    }), {
    'UpdateBucketPolicy status is OK': (r) => r && r.status === grpc.StatusOK,
    'UpdateBucketPolicy should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        adminLogin,
        { "path": "/api/v2/s3/buckets/uid_string/policy", "method": "PUT", "headers": { "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"], "Content-Length": ["28"], "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"] }, "body": { "bucket_policy": "download" } }
      ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.S3Service/DeleteBucket',
    {
      "bucket_id": {
        "cluster_id": {
          "uuid": "00000000-0000-0000-0000-000000000000",
        },
        "id": "uid_string"
      },
    }), {
    'DeleteBucket status is OK': (r) => r && r.status === grpc.StatusOK,
    'DeleteBucket should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        adminLogin,
        { "path": "/api/v2/s3/buckets/uid_string", "method": "DELETE", "headers": { "User-Agent": ["Go-http-client/1.1"], "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"], "Accept-Encoding": ["gzip"] }, "body": null }
      ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.S3Service/ListLifecycleRules',
    {
      "bucket_id": {
        "cluster_id": {
          "uuid": "00000000-0000-0000-0000-000000000000",
        },
        "id": "uid_string"
      },
    }), {
    'ListLifecycleRules status is OK': (r) => r && r.status === grpc.StatusOK,
    'ListLifecycleRules should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        adminLogin,
        { "path": "/api/v2/s3/buckets/uid_string/lifecycle/rules", "method": "GET", "headers": { "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"], "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"] }, "body": null }])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.S3Service/CreateLifecycleRules',
    {
      "bucket_id": {
        "cluster_id": {
          "uuid": "00000000-0000-0000-0000-000000000000",
        },
        "id": "uid_string"
      },
      "lifecycle_rules": [{
        "prefix": "pref",
        "expire_days": 5
      }]
    }), {
    'CreateLifecycleRules status is OK': (r) => r && r.status === grpc.StatusOK,
    'CreateLifecycleRules should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        adminLogin,
        { "path": "/api/v2/s3/buckets/uid_string/lifecycle/rules", "method": "DELETE", "headers": { "User-Agent": ["Go-http-client/1.1"], "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"], "Accept-Encoding": ["gzip"] }, "body": null },
        { "path": "/api/v2/s3/buckets/uid_string/lifecycle/rules", "method": "POST", "headers": { "User-Agent": ["Go-http-client/1.1"], "Content-Length": ["35"], "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"], "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"] }, "body": { "expiry_days": "5", "prefix": "pref" } }])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.S3Service/UpdateLifecycleRules',
    {
      "bucket_id": {
        "cluster_id": {
          "uuid": "00000000-0000-0000-0000-000000000000",
        },
        "id": "uid_string"
      },
      "lifecycle_rules": [{
        "id": {
          "id": "lrId",
        },
        "prefix": "pref",
        "expire_days": 5
      }]
    }), {
    'UpdateLifecycleRules status is OK': (r) => r && r.status === grpc.StatusOK,
    'UpdateLifecycleRules should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        adminLogin,
        { "path": "/api/v2/s3/buckets/uid_string/lifecycle/rules", "method": "DELETE", "headers": { "User-Agent": ["Go-http-client/1.1"], "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"], "Accept-Encoding": ["gzip"] }, "body": null },
        { "path": "/api/v2/s3/buckets/uid_string/lifecycle/rules", "method": "POST", "headers": { "User-Agent": ["Go-http-client/1.1"], "Content-Length": ["35"], "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"], "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"] }, "body": { "expiry_days": "5", "prefix": "pref" } }
      ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.S3Service/DeleteLifecycleRules',
    {
      "bucket_id": {
        "cluster_id": {
          "uuid": "00000000-0000-0000-0000-000000000000",
        },
        "id": "uid_string"
      },
    }), {
    'DeleteLifecycleRules status is OK': (r) => r && r.status === grpc.StatusOK,
    'DeleteLifecycleRules should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        adminLogin,
        { "path": "/api/v2/s3/buckets/uid_string/lifecycle/rules", "method": "DELETE", "headers": { "User-Agent": ["Go-http-client/1.1"], "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"], "Accept-Encoding": ["gzip"] }, "body": null },
      ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.S3Service/GetS3Principal',
    {
      "principal_id": {
        "cluster_id": {
          "uuid": "00000000-0000-0000-0000-000000000000",
        },
        "id": "uid_string"
      },
    }), {
    'GetS3Principal status is OK': (r) => r && r.status === grpc.StatusOK,
    'GetS3Principal should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        adminLogin,
        { "path": "/api/v2/users", "method": "GET", "headers": { "User-Agent": ["Go-http-client/1.1"], "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"], "Accept-Encoding": ["gzip"] }, "body": null },
        { "path": "/api/v2/s3/policies/uid_string", "method": "GET", "headers": { "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"], "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"] }, "body": null }
      ])
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
    'CreateS3Principal status is OK': (r) => r && r.status === grpc.StatusOK,
    'CreateS3Principal should call correct requests': (r) => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        adminLogin,
        { "path": "/api/v2/users", "method": "POST", "headers": { "Content-Length": ["63"], "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"], "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"] }, "body": { "password": "credentials", "role": "S3", "username": "s3Principal" } }
      ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.S3Service/SetS3PrincipalCredentials',
    {
      "principal_id": {
        "cluster_id": {
          "uuid": "00000000-0000-0000-0000-000000000000",
        },
        "id": "uid_string"
      },
      "credentials": "credentials",
    }), {
    'SetS3PrincipalCredentials status is OK': (r) => r && r.status === grpc.StatusOK,
    'SetS3PrincipalCredentials should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        adminLogin,
        { "path": "/api/v2/users/uid_string/password", "method": "PUT", "headers": { "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"], "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"], "Content-Length": ["26"] }, "body": { "password": "credentials" } }
      ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.S3Service/UpdateS3PrincipalPolicies',
    {
      "principal_id": {
        "cluster_id": {
          "uuid": "00000000-0000-0000-0000-000000000000",
        },
        "id": "uid_string"
      },
      "policies": [
        {
          "bucket_id": {
            "cluster_id": {
              "uuid": "00000000-0000-0000-0000-000000000000",
            },
            "id": "uid_string",
          },
          "prefix": "pref/Data/1sdf.sdff",
        }
      ]
    }), {
    'UpdateS3PrincipalPolicies status is OK': (r) => r && r.status === grpc.StatusOK,
    'UpdateS3PrincipalPolicies should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        adminLogin,
        { "path": "/api/v2/users", "method": "GET", "headers": { "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"], "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"] }, "body": null },
        { "path": "/api/v2/s3/policies/uid_string", "method": "GET", "headers": { "User-Agent": ["Go-http-client/1.1"], "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"], "Accept-Encoding": ["gzip"] }, "body": null },
        { "path": "/api/v2/s3/policies/uid_string", "method": "DELETE", "headers": { "User-Agent": ["Go-http-client/1.1"], "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"], "Accept-Encoding": ["gzip"] }, "body": null },
        { "path": "/api/v2/s3/policies", "method": "POST", "headers": { "Content-Type": ["application/json"], "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"], "Content-Length": ["341"], "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"] }, "body": { "policy_file_content": { "Statement": [{ "Resource": ["arn:aws:s3:::uid_string"], "Sid": "bucket", "Action": ["s3:GetBucketLocation", "s3:GetBucketPolicy", "s3:ListBucket", "s3:ListBucketMultipartUploads", "s3:ListMultipartUploadParts", "s3:GetBucketTagging", "s3:ListBucketVersions"], "Effect": "Allow" }], "Version": "2012-10-17" }, "policy_name": "uid_string" } },
        { "path": "/api/v2/s3/policies/attach", "method": "POST", "headers": { "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"], "Content-Length": ["48"], "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"], "Content-Type": ["application/json"] }, "body": { "policy_name": "uid_string", "user_name": "admin" } }
      ])
    }
  });

  check(client.invoke('intel.storagecontroller.v1.S3Service/DeleteS3Principal',
    {
      "principal_id": {
        "cluster_id": {
          "uuid": "00000000-0000-0000-0000-000000000000",
        },
        "id": "uid_string"
      },
    }), {
    'DeleteS3Principal status is OK': (r) => r && r.status === grpc.StatusOK,
    'DeleteS3Principal should call correct requests': () => {
      const reqs = wekaV4.requests
      wekaV4.requests = []
      return deepEquals(reqs, [
        adminLogin,
        { "path": "/api/v2/s3/policies/uid_string", "method": "DELETE", "headers": { "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"], "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"] }, "body": null },
        { "path": "/api/v2/users/uid_string", "method": "DELETE", "headers": { "Accept-Encoding": ["gzip"], "User-Agent": ["Go-http-client/1.1"], "Authorization": ["Bearer eyJhbGciOiJFUzI1NiIsImtpZCI6ImZha2Uta2V5LWlkIiwidHlwIjoiSldUIn0.eyJhdWQiOlsiZXhhbXBsZS11c2VycyJdLCJpc3MiOiJmYWtlLWlzc3VlciIsInBlcm0iOltdfQ.4PyhTHLl89rWbaKiWzZjJR7h_NZYQ-1yYLlxSt47DWBwES_HCNgd5vFql_Z8P1UYmf7a29KZ4_pblVyXKqcneQ"] }, "body": null }
      ])
    }
  });
}
