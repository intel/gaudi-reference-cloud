load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive", "http_file")
load("@io_bazel_rules_docker//container:container.bzl", "container_pull")
load("@io_bazel_rules_docker//go:static.bzl", "DIGESTS")

def idc_repositories():
    http_file(
        name = "yq_v2.4.1_linux_arm64",
        urls = ["https://github.com/mikefarah/yq/releases/download/2.4.1/yq_linux_arm64"],
        sha256 = "7e10a955605ea174a79a27c591086bfc49b9b8a8c1b50c5a79844878ef90ee0c",
        executable = True
    )

    http_file(
        name = "grpc_health_probe-linux-amd64",
        downloaded_file_path = "grpc_health_probe",
        urls = ["https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.34/grpc_health_probe-linux-amd64"],
        sha256 = "3ddaf85583613c97693e9b8aaa251dac07e73e366e159a7ccadbcf553117fcef",
        executable = True
    )

    http_file(
        name = "yq_linux_amd64_file",
        downloaded_file_path = "yq",
        executable = True,
        sha256 = "8afd786b3b8ba8053409c5e7d154403e2d4ed4cf3e93c237462dc9ef75f38c8d",
        urls = ["https://github.com/mikefarah/yq/releases/download/v4.35.2/yq_linux_amd64"],
    )

    http_file(
        name = "jq_linux_amd64_file",
        downloaded_file_path = "jq",
        executable = True,
        urls = ["https://github.com/jqlang/jq/releases/download/jq-1.6/jq-linux64"],
        sha256 = "af986793a515d500ab2d35f8d2aecd656e764504b789b66d7e1a0b727a124c44",
    )

    http_file(
        name = "kubectl_linux_arm64",
        sha256 = "8366cd74910411dd9546117edd98b3248b6d33e8ea9b7e65de84168e0f162d47",
        urls = ["https://storage.googleapis.com/kubernetes-release/release/v1.16.1/bin/linux/arm64/kubectl"],
        executable = True
    )

    http_file(
        name = "kubectl_linux_amd64",
        downloaded_file_path = "kubectl",
        executable = True,
        sha256 = "4717660fd1466ec72d59000bb1d9f5cdc91fac31d491043ca62b34398e0799ce",
        urls = ["https://storage.googleapis.com/kubernetes-release/release/v1.28.0/bin/linux/amd64/kubectl"],
    )

    http_file(
        name = "kind_linux_amd64",
        downloaded_file_path = "kind",
        executable = True,
        sha256 = "b543dca8440de4273be19ad818dcdfcf12ad1f767c962242fcccdb383dff893b",
        urls = ["https://github.com/kubernetes-sigs/kind/releases/download/v0.19.0/kind-linux-amd64"],
    )

    http_archive(
        name = "helm_v2.17.0_linux_arm64",
        urls = ["https://get.helm.sh/helm-v2.17.0-linux-arm64.tar.gz"],
        sha256 = "c3ebe8fa04b4e235eb7a9ab030a98d3002f93ecb842f0a8741f98383a9493d7f",
        build_file = "@com_github_masmovil_bazel_rules//:helm.BUILD",
    )

    http_archive(
        name = "helm_v3.6.2_linux_arm64",
        sha256 = "957031f3c8cf21359065817c15c5226cb3082cac33547542a37cf3425f9fdcd5",
        urls = ["https://get.helm.sh/helm-v3.6.2-linux-arm64.tar.gz"],
        build_file = "@com_github_masmovil_bazel_rules//:helm.BUILD",
    )

    http_archive(
        name = "helmfile_linux_amd64",
        build_file = "//build/helmfile:helmfile.BUILD.bazel",
        sha256 = "34a5ca9c5fda733f0322f7b12a2959b7de4ab125bcf6531337751e263b027d58",
        urls = ["https://github.com/helmfile/helmfile/releases/download/v0.169.2/helmfile_0.169.2_linux_amd64.tar.gz"],
    )

    http_archive(
        name = "vault_linux_amd64",
        build_file = "//build/vault:vault.BUILD.bazel",
        sha256 = "cf1015d0b30806515120d4a86672ea77da1fb0559e3839ba88d8e02e94e796a6",
        urls = ["https://releases.hashicorp.com/vault/1.13.1/vault_1.13.1_linux_amd64.zip"],
    )

    http_archive(
        name = "trivy_linux_64",
        build_file = "//go/pkg/insights/security-scanner:trivy.BUILD.bazel",
        sha256 = "eb79a4da633be9c22ce8e9c73a78c0f57ffb077fb92cb1968aaf9c686a20c549",
        urls = ["https://github.com/aquasecurity/trivy/releases/download/v0.58.0/trivy_0.58.0_Linux-64bit.tar.gz"],
    )

    container_pull(
        name = "ansible",
        digest = "sha256:c56b4d9df72e515041e825ff0997f4a899fb5d9fdce22a68597b2cb8bc36a954",
        registry = "litmuschaos",
        repository = "ansible-runner",
    )

    container_pull(
        name = "image_static_arm64",
        # tag = "latest-arm64",
        digest = "sha256:e432dc668db28751374cd8792779786ff71e6ba8077610feeef15e9267bcda13",
        registry = "gcr.io",
        repository = "distroless/static",
    )

    container_pull(
        name = "image_static_arm64_debug",
        digest = "sha256:2ddee9b0745bce45a5296c27377712115406592416af8807de7df6985584896d",
        tag = "debug-arm64",
        registry = "gcr.io",
        repository = "distroless/static",
    )
 
    container_pull(
        name = "image_static_x86_64",
        digest = DIGESTS["latest"],
        registry = "gcr.io",
        repository = "distroless/static",
    )

    container_pull(
        name = "image_static_x86_64_debug",
        digest = DIGESTS["debug"],
        registry = "gcr.io",
        repository = "distroless/static",
    )

    container_pull(
        name = "image_alpine_x86_64",
        tag = "3.18.3",
        digest = "sha256:c5c5fda71656f28e49ac9c5416b3643eaa6a108a8093151d6d1afc9463be8e33",
        registry = "amr-idc-registry-pre.infra-host.com/cache",
        repository = "library/alpine",
    )
