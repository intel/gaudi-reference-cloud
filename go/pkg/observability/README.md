<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# IDC Observability

## How to add tracing to an application

### Initialize Tracing

Add the following to main.go.

```go
import (
    "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

func main() {
    ...
    obs := observability.New(ctx)
    tracerProvider := obs.InitTracer(ctx)
    defer tracerProvider.Shutdown(ctx)
    ...
}
```

### Tracing of GRPC Server

```go
import (
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

grpcServer := grpc.NewServer(
    grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
    grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
)
```

### Tracing of GRPC Client

```go
import (
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

clientConn, err := grpc.Dial(
      ...
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
      grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		)
```

### Helm Chart Template

Add the following to the Helm chart `deployment.yaml` in [deployment/charts](../../../deployment/charts).

```
spec:
  template:
    metadata:
      annotations:
        {{- include "idc-common.vaultAnnotations" . | nindent 8 }}
        {{- include "idc-common.otelAnnotations" . | nindent 8 }}
    spec:
      containers:
        - env:
            {{- include "idc-common.proxyEnv" . | nindent 12 }}
            {{- include "idc-common.otelEnv" . | nindent 12 }}
```

### Helm Chart Values for Kind

Add the following to the Helm chart values used for kind in [deployment/kind/chart-values](../../../deployment/kind/chart-values).

```yaml
otel:
  deployment:
    environment: $(OTEL_DEPLOYMENT_ENVIRONMENT)
```

### BUILD.bazel for Kind

```
expand_template(
    ...
    substitutions = {
        ...
        "$(OTEL_DEPLOYMENT_ENVIRONMENT)": "$(OTEL_DEPLOYMENT_ENVIRONMENT)",
    },
)
```

### Vault Server

Vault is used to store the public CA certificate [otel-ca.pem](otel-ca.pem) used to validate the tracing server.
Although this is not a secret, in the future, a secret will be required and it will be stored in Vault.

Edit [deployment/kind/vault/load-secrets.sh](../../../deployment/kind/vault/load-secrets.sh).
Add a section like the following.

```bash
kubectl exec  -n vault vault-0 -- vault write auth/kubernetes/role/compute-api-server-role \
    bound_service_account_names=compute-api-server \
    bound_service_account_namespaces=idcs-system \
    policies=public \
    ttl=1h
```

### Network access

The connection to the OTEL server requires access through the Intel proxy.
The template `proxyEnv` sets the environment variable `https_proxy`.
This forces all traffic to use the Intel proxy, unless it has been excluded by the `no_proxy` environment variable.
To exclude cluster-local traffic, use an FQDN such as `compute-api-server.idcs-system.svc.cluster.local`.
Connections to private IP addresses are also excluded.

## Combined Logging and Tracing

LogAndSpan is a convenience type for using a Logger and Span together.

```go
import (
    obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("Instance.Create").WithValues("ResourceId", ResourceId).Start()
defer span.End()
log.Info("Creating instance", "request", request)
```

## Enabling traces in tests

Add following snippet to your BUILD.bazel target responsible for building your tests:
```
go_test(
...
    data = [
        "//go/pkg/observability:configdata",
    ],
    env = {
        # Following setting enable tests to send traces to ELK.
        # If not set, traces will not be sent.
        "OTEL_SERVICE_NAME": "<YOUR_SERVICE_NAME>", # e.g. compute-api-server-test
        "OTEL_EXPORTER_OTLP_CERTIFICATE": "../../observability/otel-ca.pem", # if needed, update to match directory structure
        "OTEL_EXPORTER_OTLP_ENDPOINT": "internal-placeholder.com:80",
        "https_proxy": "http://internal-placeholder.com:912",
    },
    env_inherit = [
        "OTEL_RESOURCE_ATTRIBUTES",
    ],
...
)
```
