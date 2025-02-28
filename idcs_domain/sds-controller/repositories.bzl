# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def grpc_ecosystem_grpc_gateway():
    http_archive(
        name = "grpc_ecosystem_grpc_gateway",
        sha256 = "21b4c4dade327216e5a8f4be66ced489789dfff108c30a7f556af9941b2d4ede",
        strip_prefix = "grpc-gateway-2.18.0",
        urls = [
            "https://github.com/grpc-ecosystem/grpc-gateway/archive/refs/tags/v2.18.0.tar.gz",
        ],
    )

def com_github_bazelbuild_buildtools():
    http_archive(
        name = "com_github_bazelbuild_buildtools",
        sha256 = "42968f9134ba2c75c03bb271bd7bb062afb7da449f9b913c96e5be4ce890030a",
        strip_prefix = "buildtools-6.3.3",
        urls = [
            "https://github.com/bazelbuild/buildtools/archive/v6.3.3.tar.gz",
        ],
    )

def weka_openapi():
    http_archive(
        name = "weka_openapi",
        sha256 = "98d9b6f6e18ca08a6d6e8496cd0e0b224005e0e9407af6a69342b7509a5838f9",
        build_file = "//:bazel/third_party/weka_api.BUILD",
        patch_args = ["-p1"],
        patches = [
            "bazel/patches/weka_openapi_4_2.patch",
        ],
        strip_prefix = "Weka-REST-API-Docs-c776cd3570d0aacdfbf6867070fab4071ded7d16",
        urls = [
            "https://github.com/weka/Weka-REST-API-Docs/archive/c776cd3570d0aacdfbf6867070fab4071ded7d16.tar.gz",
        ],
    )

def com_github_bufbuild_protovalidate():
    http_archive(
        name = "com_github_bufbuild_protovalidate",
        sha256 = "03ee49c344d350355c7e21b36456d709cbcba493f2cfa2029be99fc065aa7c48",
        strip_prefix = "protovalidate-0.5.3",
        urls = [
            "https://github.com/bufbuild/protovalidate/archive/refs/tags/v0.5.3.tar.gz",
        ],
    )
