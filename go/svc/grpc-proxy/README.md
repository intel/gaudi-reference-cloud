<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# gRPC Proxy

grpc-proxy is a router for gRPC calls. grpc-proxy is an envoy
configuration produced by the prototempl.go tool using the
gprc-proxy-configmap.yaml.tmpl template and the .proto files
in public_api/proto at the top of the tree.

The envoy configuration routes each gRPC API to the go
service that implements it. This requires the APIs
declared in a .proto file to be implemented by a go
service accessible by the name of the .proto file without
the .proto extention. For example, all of the APIs
declared in cloudaccount.proto must be implemented
by a service that other services can communicate with
by connecting to the DNS name "cloudaccount".

In addition the services declared in public_api/proto/\*.proto,
grpc-proxy routes APIs for /grpc.reflection.v1alpha.ServerReflection/
to the grpc-reflect service, which provides gRPC reflection
aggregating all of the APIs declared in public_api/proto/\*.proto

# To enable JWT validation 
```bash
BAZEL_EXTRA_OPTS="--define IDC_ENABLE_JWT_VALIDATION=true" make deploy-foundation
```