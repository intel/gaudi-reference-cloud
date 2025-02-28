# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
load("//:repositories.bzl", "com_github_bazelbuild_buildtools", "com_github_bufbuild_protovalidate", "grpc_ecosystem_grpc_gateway", "weka_openapi")

def _non_module_dependencies_impl(_ctx):
    grpc_ecosystem_grpc_gateway()
    com_github_bazelbuild_buildtools()
    weka_openapi()
    com_github_bufbuild_protovalidate()

non_module_dependencies = module_extension(
    implementation = _non_module_dependencies_impl,
)
