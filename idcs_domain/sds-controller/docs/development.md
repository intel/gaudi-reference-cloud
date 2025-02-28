<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# Development
This document describe development workflow and how to work with the repository

## Prerequisite
Project uses `bazelisk` as the main entry point, it is required to have it installed
To install please follow guide for your platform [bazelisk](https://github.com/bazelbuild/bazelisk)

To generate protobuf source files:
```
bazelisk run //:update_gen
```

## Build

To fully build all packages execute following command:
```
bazelisk build //...
```

It will discover and build all available targets, to build just one target (for example storage-controller), one can execute following command:
```
bazelisk build //services/storage_controller/...
```

### Generate BUILD files

We use [gazelle](https://github.com/bazel-contrib/bazel-gazelle) to simplify and speed-up the process of managing bazel manifests.
If you add some new dependencies or new source files, you don't have to make any changes there manually.

```bash
# Update go version
bazelisk run @io_bazel_rules_go//go -- mod edit -go=1.24
# Update all the dependencies
bazelisk run @io_bazel_rules_go//go -- get -u
bazelisk run @io_bazel_rules_go//go -- mod tidy

bazelisk mod tidy
```

Adding new build targets is as simple as running the following command

```
bazelisk run //:gazelle
```



To get more familiar with bazel targets use bazel [documentation](https://bazel.build/docs).

### Generate Proto file
For IDE support generation of protofiles supported by:
```
bazelisk run //:update_go_pb
```

This command will generate pb files and put them alongside proto files, this files should not be checked in the git repo

## Run

To run specific target execute `run` command, for example for storage controller:
```
bazelisk run //services/storage_controller/cmd/storage_controller
```

To pass parameters to executable use `--` and params after the bazel target:
```
bazelisk run //ci/k6/cmd -- run $(pwd)/services/storage_controller/tests/integration/proto_api.js
```

Some executable require specific ENV variables to be set, they can be passed as part of the shell command:
```
WEKA_CREDS=login:password STORAGE_CONTROLLER_CONFIG_FILE=$(pwd)/services/storage_controller/tests/configs/storage-conf.yaml bazelisk run //services/storage_controller/cmd/storage_controller
```

## Test

To run all tests execute following command:
```
bazelisk test //...
```

To run specific test specify test target:
```
bazelisk test //services/storage_controller/pkg/server:server_test
```

### Coverage

To gather coverage info one can run following command:
```
bazelisk coverage //...
```

Command above will generate lcov coverage file, to generate human readable document, `lcov` tools can be used (need to be installed on the host machine):
```
genhtml --branch-coverage --output genhtml "$(bazelisk info output_path)/_coverage/_coverage_report.dat"
```

Command above will create genhtml folder and put coverage data in here.

## <a name="oci-images"></a> OCI Images

Repository uses OCI images as the container format, in order to build OCI and push images there is extra dependencies required,
to run OCI images you will need install one of the container runtimes like `podman` or `docker`.

Please note, if you plan to run container, specify `target`` platform for build (e.g. linux_x86), or binaries will be build for `host` platform,
e.g. windows/mac/linux, depending on your host OS.

### Building

To build OCI image locally execute build command:
```
bazelisk build //services/storage_controller:image --platforms=//:linux_x86_64
```

To push it to the registry execute following command (you need to authenticated with registry):
```
bazelisk run //services/storage_controller:image_push --platforms=//:linux_x86_64 -- -r "<REGISTRY_ADDRESS>"  -t "<IMAGE_TAG>
```

### Running
To run OCI image execute following commands:
```
bazelisk run //services/storage_controller:tarball --platforms=//:linux_x86_64
docker run -rm localtest.localhost/storage_controller:latest
```

Following command will build image which will work on host (exec) platform:
```
bazelisk build //services/storage_controller:image
```

## Platforms

Build system (bazel) distinguish between `host` and `target` platform, this repo is configured so `host` platform is always
auto-detected, no need to specify it. It is required to make sure tools used to build works and execute on the `host` platform.

Target platform is auto-detected as the `host` platform if there is no flag specified during target, so it will run by default on host,
one can specify target platform, for example if building for linux under macos by `target_platforms=//:<platform>` arg.

Build system support following target platforms:
- `linux_arm64`
- `linux_x86_64`
- `osx_arm64`
- `osx_x86_64`
- `windows_arm64`
- `windows_x86_64`

In most cases only `linux` platform is usable.
