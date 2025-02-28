<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# grpc-rest-gateway

grpc-rest-gateway provides REST interfaces for all of the APIs
defined in public_api/*.proto, based on code produced by the
grpc-gateway protoc plugin. grpc-rest-gateway also serves
the devcloud-grpc.swagger.json OpenAPI definitions produced
by the openapiv2 protoc plugin.

grpc-rest-gateway translates the REST calls to gRPC and forwards
them to the grpc-proxy service.