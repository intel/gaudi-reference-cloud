#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

set -ex

which go
go version

which mockgen
mockgen -version

which protoc
protoc --version

while [ ! -f go.mod ]; do
    dir=$(pwd)
    if [ "${dir}" = / ]; then
        echo "go.mod not found" 1>&2
        exit 1
    fi
    cd ..
done

module=$(grep ^module go.mod | sed -e 's/^module //')

validate="${GOPATH}/pkg/mod/github.com/envoyproxy/protoc-gen-validate@v0.10.1"

# Generate go protobuf and grpc code: go/pkg/pb/*.pb.go and go/pkg/pb/*_grpc.pb.go
# This protoc command must be first, because the go commands that follow need the
# go/pkg/pb/annotations.pb.go file the protoc command produces.
protoc --go_opt=module=$module --go_out=. \
       --go-grpc_opt=module=$module --go-grpc_out=. \
       --validate_out="lang=go,paths=source_relative:pkg/pb" \
       --proto_path ../public_api/proto \
       -I $validate \
       ../public_api/proto/*.proto

# calculate protorestfiles,the list of .proto files that contain http annotations for 
# the REST API.
#
# protorestfiles also enforces some requirements on IDC .proto files needed for
# grpc-proxy and grpc-rest-gateway to work.
protorestfiles=$(go run build/tools/protorestfiles/protorestfiles.go ../public_api/proto/*.proto)

# Generate grpc rest gateway code: go/pkg/pb/*.pb.gw.go
protoc --grpc-gateway_out=. \
       --grpc-gateway_opt logtostderr=true \
       --grpc-gateway_opt module=$module \
       --proto_path ../public_api/proto \
       -I $validate \
       $protorestfiles

# Generate grpc-reflect code
go run build/tools/prototempl/prototempl.go \
        --template-file svc/grpc-reflect/grpc-reflect.go.tmpl \
        --output-file svc/grpc-reflect/grpc-reflect.go \
        -I $validate \
        ../public_api/proto/*.proto

# Generate grpc-rest-gateway code
go run build/tools/prototempl/prototempl.go \
        --template-file svc/grpc-rest-gateway/register.go.tmpl \
        --output-file svc/grpc-rest-gateway/register.go \
        -I $validate \
        $protorestfiles

# Generate swagger files for grpc-rest-gateway
go run build/tools/genswagger/genswagger.go \
        --output-dir pkg/pb/swagger \
        -I $validate \
        $protorestfiles

# Generate grpc-proxy envoy config to route gRPC requests to the correct back end
# service
go run build/tools/prototempl/prototempl.go \
        --template-file svc/grpc-proxy/configmap.yaml.tmpl \
        --output-file svc/grpc-proxy/chart/grpc-proxy/templates/configmap.yaml \
        -I $validate \
        ../public_api/proto/*.proto

# Generate app-client-grpc-proxy envoy config to route gRPC requests to the correct back end
# service (limited access)
go run build/tools/prototempl/prototempl.go \
        --template-file svc/grpc-proxy/configmap.yaml.tmpl \
        --output-file svc/grpc-proxy/chart/grpc-proxy/templates/configmap-appclient.yaml \
        --output-configmap-name-suffix "-appclient" \
        -I $validate \
        -C \
        ../public_api/proto/*.proto

# Generate .rego policy for opa-envoy to use for authorizing gRPC requests in grpc-proxy
go run build/tools/protoauthzuser/protoauthzuser.go \
        --configmap svc/grpc-proxy/chart/grpc-proxy/templates/configmap-authzuser.yaml \
        -I $validate \
        ../public_api/proto/*.proto

go run build/tools/protoauthzservice/protoauthzservice.go \
        --configmap svc/grpc-proxy/chart/grpc-proxy/templates/configmap-authzservice.yaml \
        -I $validate \
        ../public_api/proto/*.proto

go run build/tools/protoauthzappclient/protoauthzappclient.go \
        --configmap svc/grpc-proxy/chart/grpc-proxy/templates/configmap-authzappclient.yaml \
        -I $validate \
        ../public_api/proto/*.proto

# Generate .pb files for opa-envoy to use to parse gRPC payloads for authorizing
# gRPC requests in grpc-proxy
go run build/tools/protopb/protopb.go \
        --configmap svc/grpc-proxy/chart/grpc-proxy/templates/configmap-pb.yaml \
        -I $validate \
        ../public_api/proto/*.proto

# Create OpenAPI Swagger spec for public Compute API.
protoc --openapiv2_out=. \
       --openapiv2_opt=generate_unbound_methods=true \
       --openapiv2_opt=allow_merge=true,merge_file_name=../public_api/proto/compute \
       --proto_path ../public_api/proto \
       -I $validate \
       ../public_api/proto/compute.proto

# Create billingdriverproxy.go
go run pkg/billing/build/genbillingdriverproxy.go \
    --output-file pkg/billing/driverproxy.go \
    -I $validate \
    --format ../public_api/proto/billing.proto

# Create mocks for Compute Service.
mockgen -source=pkg/pb/compute_grpc.pb.go -destination pkg/pb/mock_compute_grpc.pb.go -package pb
mockgen -source=pkg/pb/compute_private_grpc.pb.go -destination pkg/pb/mock_compute_private_grpc.pb.go -package pb

# Create mocks for Metering Service
mockgen -source=pkg/pb/metering_grpc.pb.go -destination pkg/pb/mock_metering_grpc.pb.go -package pb

# Create mocks for Product catalog Service
mockgen -source=pkg/pb/productcatalog_grpc.pb.go -destination pkg/pb/mock_productcatalog_grpc.pb.go -package pb

# Create mocks for Billing Service
mockgen -source=pkg/pb/billing_grpc.pb.go -destination pkg/pb/mock_billing_grpc.pb.go -package pb

# Create mocks for CloudAccount Service
mockgen -source=pkg/pb/cloudaccount_grpc.pb.go -destination pkg/pb/mock_cloudaccount_grpc.pb.go -package pb

# Create mocks for IKS service
mockgen -source=pkg/pb/iks_grpc.pb.go -destination pkg/pb/mock_iks_grpc.pb.go -package pb

# Create mocks for Storage service
mockgen -source=pkg/pb/storage_private_grpc.pb.go -destination pkg/pb/mock_storage_private_grpc.pb.go -package pb
mockgen -source=pkg/pb/storage_kms_grpc.pb.go -destination pkg/pb/mock_storage_kms_grpc.pb.go -package pb

# create mocks for Authz service
mockgen -source=pkg/pb/authz_grpc.pb.go -destination pkg/pb/mock_authz_grpc.pb.go -package pb

# Create mocks for Quota Management Service 
mockgen -source=pkg/pb/quota_admin_private_grpc.pb.go -destination pkg/pb/mock_quota_admin_private_grpc.pb.go -package pb
