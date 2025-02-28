<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
## IDC Compute API Server

See [/README.md](../../../README.md) for common information.

See [/go/pkg/manageddb/README.md](../../pkg/manageddb/README.md) for database information.

### Overview

The service compute_api_server runs the GRPC server.

The service compute_api_gateway runs the GRPC-REST gateway that receives incoming REST requests,
converts to GRPC, and forwards the GRPC request to the compute_api_server.

### Testing

#### Run Go tests

```bash
cd $(git rev-parse --show-toplevel)
make test
```

#### Deploy in kind

```bash
cd $(git rev-parse --show-toplevel)
make deploy-all-in-kind
```

#### Build and redeploy this service in an existing kind cluster

```bash
cd $(git rev-parse --show-toplevel)
make deploy-compute-api-server
```

#### Test connectivity to the GRPC-REST gateway

```
$ curl http://localhost/readyz
ok
```

```
$ curl -k https://localhost/readyz
ok
```

Check gateway logs.

```bash
kubectl logs -n idcs-system deployments/compute-api-gateway
```

Expected log output.

```
2022-12-21T19:49:59Z    INFO    RestService.readyz
2022-12-21T19:50:05Z    INFO    RestService.readyz
```

#### Test connectivity through the GRPC-REST gateway to the GRPC server

```
$ curl -k https://localhost/v1/ping
{}
```

Check server logs.

```bash
kubectl logs -n idcs-system deployments/compute-api-server
```

Expected log output.

```
2022-12-21T19:47:33Z    INFO    SshPublicKeyService.Ping        Ping
```

#### Test with Curl

```bash
export URL_PREFIX=https://localhost
test-scripts/sshpublickey_create_with_name.sh
test-scripts/sshpublickey_list.sh
test-scripts/instance_create_without_name.sh
test-scripts/instance_list.sh
```

## Running locally

This method is not recommended. Instead use kind.

```bash
go run cmd/compute_api_server/main.go --config config_server_run.yaml
go run cmd/compute_api_gateway/main.go --config config_gateway_run.yaml
curl http://localhost:8080/v1/ping
```

## Run Postgres Client in Kubernetes

See [/go/pkg/manageddb/README.md](../../pkg/manageddb/README.md) for database information.
