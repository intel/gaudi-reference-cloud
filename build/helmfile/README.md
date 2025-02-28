
# Helmfile for IDC

A customized version of Helmfile is in https://github.com/ClaudioFaheyIntel/helmfile/tree/idc.

The following changes have been made:

- `helmfile write-values` no longer downloads Helm charts from the registry.
  This allows chart push and manifests generation steps to be independent.
  In particular, it allows unit testing of manifests generator.

## Local Build of Helmfile

```bash
cd
git clone https://github.com/ClaudioFaheyIntel/helmfile
cd helmfile
git checkout idc
go install golang.org/dl/go1.21.7@latest
go1.21.7 download
export GOROOT=$(go1.21.7 env GOROOT)
PATH=${GOROOT}/bin:${PATH} make build
./helmfile --version
```

To use in IDC, edit https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/blob/main/build/repositories/repositories.bzl.
Comment out `http_archive(name="helmfile_linux_amd64", ...)`.
Uncomment `native.new_local_repository(name="helmfile_linux_amd64", ...)`.

## Release Build of Helmfile

Create tag in format `v0.161.0-idc1`.

Create release in Github. This will run a Github Action that builds the release binaries.
