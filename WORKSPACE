workspace(name = "com_intel_devcloud")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive", "http_file")

# See https://github.com/bazelbuild/rules_go
http_archive(
    name = "io_bazel_rules_go",
    sha256 = "f4a9314518ca6acfa16cc4ab43b0b8ce1e4ea64b81c38d8a3772883f153346b8",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.50.1/rules_go-v0.50.1.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.50.1/rules_go-v0.50.1.zip",
    ],
)

# Download the dependency using http_file
http_file(
    name = "trivy_dep",
    sha256 = "b8fc3f1f817ef8e17bb6fbd91f7aa87ebd8f996006ce4d5f4c11e9c4c6a1a7bd",
    url = "https://github.com/aquasecurity/trivy/releases/download/v0.18.3/trivy_0.18.3_Linux-64bit.tar.gz",
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains(version = "1.23.2")

load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

http_archive(
    name = "rules_proto",
    sha256 = "6fb6767d1bef535310547e03247f7518b03487740c11b6c6adb7952033fe1295",
    strip_prefix = "rules_proto-6.0.2",
    url = "https://github.com/bazelbuild/rules_proto/releases/download/6.0.2/rules_proto-6.0.2.tar.gz",
)

load("@rules_proto//proto:repositories.bzl", "rules_proto_dependencies")

rules_proto_dependencies()

load("@rules_proto//proto:setup.bzl", "rules_proto_setup")

rules_proto_setup()

load("@rules_proto//proto:toolchains.bzl", "rules_proto_toolchains")

rules_proto_toolchains()

git_repository(
    name = "com_google_protobuf",
    commit = "c9869dc7803eb0a21d7e589c40ff4f9288cd34ae",
    remote = "https://github.com/protocolbuffers/protobuf",
    shallow_since = "1658780535 -0700",
)

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()

http_archive(
    name = "bazel_gazelle",
    integrity = "sha256-MpOL2hbmcABjA1R5Bj2dJMYO2o15/Uc5Vj9Q0zHLMgk=",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.35.0/bazel-gazelle-v0.35.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.35.0/bazel-gazelle-v0.35.0.tar.gz",
    ],
)

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

load("//:deps.bzl", "go_dependencies")

# gazelle:repository_macro deps.bzl%go_dependencies
go_dependencies()

gazelle_dependencies()

http_archive(
    name = "io_bazel_rules_docker",
    sha256 = "b1e80761a8a8243d03ebca8845e9cc1ba6c82ce7c5179ce2b295cd36f7e394bf",
    urls = ["https://github.com/bazelbuild/rules_docker/releases/download/v0.25.0/rules_docker-v0.25.0.tar.gz"],
)

load(
    "@io_bazel_rules_docker//repositories:repositories.bzl",
    container_repositories = "repositories",
)

container_repositories()

load(
    "@io_bazel_rules_docker//go:image.bzl",
    _go_image_repos = "repositories",
)

_go_image_repos()

load("//build/toolchains:idc_toolchain.bzl", "register_idc_toolchains")

register_idc_toolchains()

http_archive(
    name = "io_bazel_rules_k8s",
    sha256 = "ce5b9bc0926681e2e7f2147b49096f143e6cbc783e71bc1d4f36ca76b00e6f4a",
    strip_prefix = "rules_k8s-0.7",
    urls = ["https://github.com/bazelbuild/rules_k8s/archive/refs/tags/v0.7.tar.gz"],
)

load("@io_bazel_rules_k8s//k8s:k8s.bzl", "k8s_repositories")

k8s_repositories()

# Bazel rules to install and manipulate Helm charts with Bazel
# See https://github.com/masmovil/bazel-rules
git_repository(
    name = "com_github_masmovil_bazel_rules",
    commit = "5d23e9e2f8eb350d6fb179e811067351f6574233",
    remote = "https://github.com/masmovil/bazel-rules.git",
)

load(
    "@com_github_masmovil_bazel_rules//repositories:repositories.bzl",
    mm_repositories = "repositories",
)

mm_repositories()

# https://github.com/aspect-build/bazel-lib/releases/tag/v1.39.0
http_archive(
    name = "aspect_bazel_lib",
    sha256 = "4d6010ca5e3bb4d7045b071205afa8db06ec11eb24de3f023d74d77cca765f66",
    strip_prefix = "bazel-lib-1.39.0",
    url = "https://github.com/aspect-build/bazel-lib/releases/download/v1.39.0/bazel-lib-v1.39.0.tar.gz",
)

load("@aspect_bazel_lib//lib:repositories.bzl", "aspect_bazel_lib_dependencies")

aspect_bazel_lib_dependencies()

http_archive(
    name = "helm3_linux_amd64",
    build_file = "//build/helm:helm.BUILD.bazel",
    sha256 = "f5355c79190951eed23c5432a3b920e071f4c00a64f75e077de0dd4cb7b294ea",
    urls = ["https://get.helm.sh/helm-v3.16.3-linux-amd64.tar.gz"],
)

http_archive(
    name = "helm3_linux_arm64",
    build_file = "//build/helm:helm.BUILD.bazel",
    urls = ["https://get.helm.sh/helm-v3.13.3-linux-arm64.tar.gz"],
)

# https://github.com/bazelbuild/rules_pkg/releases/tag/0.9.1
http_archive(
    name = "rules_pkg",
    sha256 = "8f9ee2dc10c1ae514ee599a8b42ed99fa262b757058f65ad3c384289ff70c4b8",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_pkg/releases/download/0.9.1/rules_pkg-0.9.1.tar.gz",
        "https://github.com/bazelbuild/rules_pkg/releases/download/0.9.1/rules_pkg-0.9.1.tar.gz",
    ],
)

load("@rules_pkg//:deps.bzl", "rules_pkg_dependencies")

rules_pkg_dependencies()

# Below excludes the directory idcs_core from bazel build //...
# https://github.com/bazelbuild/bazel/issues/4888#issuecomment-407890682
local_repository(
    name = "ignore_idcs_core",
    path = "idcs_core",
)

local_repository(
    name = "ignore_local_go",
    path = "local/go",
)

http_file(
    name = "postgresql_chart",
    sha256 = "b44a3836970aacc9cd2a7c1982e1a8f6c07fb52b456e15520ce8527748f473b5",
    url = "https://charts.bitnami.com/bitnami/postgresql-12.1.5.tgz",
)

http_file(
    name = "timescaledb_chart",
    sha256 = "f081413886ef8cabb1ca6b7ca77c7970c061ae4a42bf753947577fe5021431b1",
    url = "https://github.com/timescale/helm-charts/releases/download/timescaledb-single-0.27.5/timescaledb-single-0.27.5.tgz",
)



http_archive(
    name = "kubebuilder_tools",
    build_file = "//build/kubebuilder:kubebuilder-tools.BUILD.bazel",
    sha256 = "c9796a0a13ccb79b77e3d64b8d3bb85a14fc850800724c63b85bf5bacbe0b4ba",
    strip_prefix = "kubebuilder",
    url = "https://storage.googleapis.com/storage/v1/b/kubebuilder-tools/o/kubebuilder-tools-1.25.0-linux-amd64.tar.gz?alt=media",
)

http_archive(
    name = "kubebuilder_tools_arm64",
    build_file = "//build/kubebuilder:kubebuilder-tools.BUILD.bazel",
    sha256 = "f048b9ba53cb722b5f3b6acc8f8bf4101aa77f0a3062cbd1b0c4cb0d8dd1eaf3",
    strip_prefix = "kubebuilder",
    url = "https://storage.googleapis.com/storage/v1/b/kubebuilder-tools/o/kubebuilder-tools-1.25.0-linux-arm64.tar.gz?alt=media",
)

http_archive(
    name = "kubebuilder_tools_arm64_darwin",
    build_file = "//build/kubebuilder:kubebuilder-tools.BUILD.bazel",
    sha256 = "e5ae7aaead02af274f840693131f24aa0506b0b44ccecb5f073847b39bef2ce2",
    strip_prefix = "kubebuilder",
    url = "https://storage.googleapis.com/storage/v1/b/kubebuilder-tools/o/kubebuilder-tools-1.25.0-darwin-arm64.tar.gz?alt=media",
)

load("//build/repositories:repositories.bzl", "idc_repositories")

idc_repositories()

# https://github.com/ash2k/bazel-tools/tree/master/multirun
git_repository(
    name = "com_github_ash2k_bazel_tools",
    commit = "2add5bb84c2837a82a44b57e83c7414247aed43a",
    remote = "https://github.com/ash2k/bazel-tools.git",
    shallow_since = "1679573490 +1100",
)

load("@com_github_ash2k_bazel_tools//multirun:deps.bzl", "multirun_dependencies")

multirun_dependencies()

# https://github.com/bazel-contrib/rules_oci/releases/tag/v1.4.3
http_archive(
    name = "rules_oci",
    sha256 = "d41d0ba7855f029ad0e5ee35025f882cbe45b0d5d570842c52704f7a47ba8668",
    strip_prefix = "rules_oci-1.4.3",
    url = "https://github.com/bazel-contrib/rules_oci/releases/download/v1.4.3/rules_oci-v1.4.3.tar.gz",
)

load("@rules_oci//oci:dependencies.bzl", "rules_oci_dependencies")

rules_oci_dependencies()

load("@rules_oci//oci:repositories.bzl", "LATEST_CRANE_VERSION", "oci_register_toolchains")

oci_register_toolchains(
    name = "oci",
    crane_version = LATEST_CRANE_VERSION,
    # Uncommenting the zot toolchain will cause it to be used instead of crane for some tasks.
    # Note that it does not support docker-format images.
    # zot_version = LATEST_ZOT_VERSION,
)

# You can pull your base images using oci_pull like this:
load("@rules_oci//oci:pull.bzl", "oci_pull")

oci_pull(
    name = "distroless_base",
    digest = "sha256:ccaef5ee2f1850270d453fdf700a5392534f8d1a8ca2acda391fbb6a06b81c86",
    image = "gcr.io/distroless/base",
    platforms = [
        "linux/amd64",
        "linux/arm64",
    ],
)

# https://github.com/bazelbuild/rules_python/releases/tag/0.27.0
http_archive(
    name = "rules_python",
    sha256 = "9acc0944c94adb23fba1c9988b48768b1bacc6583b52a2586895c5b7491e2e31",
    strip_prefix = "rules_python-0.27.0",
    url = "https://github.com/bazelbuild/rules_python/releases/download/0.27.0/rules_python-0.27.0.tar.gz",
)

load("@rules_python//python:repositories.bzl", "python_register_toolchains")

# Use a hermetic Python version.
# This is required to produce reproducible tar files on systems running different versions of Python due to https://bugs.python.org/issue18819.
# Available versions are listed in @rules_python//python:versions.bzl.
# https://rules-python.readthedocs.io/en/stable/getting-started.html#using-a-workspace-file
python_register_toolchains(
    name = "python_3_10_2",
    python_version = "3.10.2",
)

# https://github.com/bazel-contrib/rules_bazel_integration_test
http_archive(
    name = "rules_bazel_integration_test",
    sha256 = "b079b84278435441023f03de1a72baff9e4e4fe2cb1092ed4c9b60dc8b42e732",
    urls = [
        "https://github.com/bazel-contrib/rules_bazel_integration_test/releases/download/v0.25.0/rules_bazel_integration_test.v0.25.0.tar.gz",
    ],
)

load("@rules_bazel_integration_test//bazel_integration_test:deps.bzl", "bazel_integration_test_rules_dependencies")

bazel_integration_test_rules_dependencies()

load("@cgrindel_bazel_starlib//:deps.bzl", "bazel_starlib_dependencies")

bazel_starlib_dependencies()

load("@bazel_skylib//:workspace.bzl", "bazel_skylib_workspace")

bazel_skylib_workspace()

load("@rules_bazel_integration_test//bazel_integration_test:defs.bzl", "bazel_binaries")

bazel_binaries(versions = [
    "//:.bazelversion",
    "5.4.0",
])

# IDC Universe Deployer
load("//build/git_archive:git_archive.bzl", "git_archive")
load("//build/bzl_file_embed:bzl_file_embed.bzl", "bzl_file_embed")

load("//universe_deployer/main_universe:main_universe_defs.bzl", "main_universe_srcs")

git_archive(
    name = "git_archives",
    srcs = main_universe_srcs,
    commit_metadata_source_files = {
        "universe_deployer_commit_metadata": "deployment/universe_deployer/commit_metadata.json",
    },
    remote = "https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc",
)

bzl_file_embed(
    name = "main_universe_config_file_embed",
    srcs = main_universe_srcs,
)

load("//universe_deployer/create_releases:create_releases_defs.bzl", "create_releases_srcs")

git_archive(
    name = "create_releases_git_archives",
    srcs = create_releases_srcs,
    commit_metadata_source_files = {
        "universe_deployer_commit_metadata": "deployment/universe_deployer/commit_metadata.json",
    },
    remote = "https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc",
)

bzl_file_embed(
    name = "create_releases_universe_config_file_embed",
    srcs = create_releases_srcs,
)

local_repository(
    name = "rules_sphinx",
    path = "build/rules_sphinx",
)

load("@rules_sphinx//sphinx:direct_repositories.bzl", "rules_sphinx_direct_deps")

rules_sphinx_direct_deps()

load("@rules_sphinx//sphinx:indirect_repositories.bzl", "rules_sphinx_indirect_deps")

rules_sphinx_indirect_deps()
