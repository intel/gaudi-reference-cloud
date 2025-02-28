# Deployment of IDC with Helmfile

## Troubleshooting Tips for Helmfile and Helm

### Show Makefile variables

```bash
export IDC_ENV=dev1
make show-config
```

### Install Helm and Helmfile

```bash
make install-interactive-tools
```

### Build and reinstall all Helm releases that use a specific chart

```bash
make deploy-grpc-proxy
```

### Build and reinstall a specific Helm release

```bash
make helm-push-grpc-proxy && \
HELMFILE_OPTS="destroy --selector name=grpc-proxy-external" make run-helmfile
HELMFILE_OPTS="apply --selector name=grpc-proxy-external" make run-helmfile
```

### View rendered Helm values

```bash
make helm-push-grpc-proxy && \
HELMFILE_OPTS="write-values --selector name=us-dev-1-grpc-proxy-internal \
--output-file-template '/tmp/helm-values/{{ .Release.Name }}.yaml' \
--skip-deps" make run-helmfile |& tee local/helmfile.log
```

### View rendered Kubernetes resources

```bash
make helm-push-grpc-proxy && \
HELMFILE_OPTS="template --selector name=us-dev-1-grpc-proxy-external \
--output-dir /tmp \
--skip-deps" make run-helmfile |& tee local/helmfile.log
```

### Common Errors

#### Error

```
Templating release=us-dev-1-grpc-proxy-internal, chart=/tmp/helmfile2411785538/idcs-system/kind-idc-us-dev-1/us-dev-1-grpc-proxy-internal/grpc-proxy/0.0.1-n1203-h6b843079/grpc-proxy
...
STDERR:
  Error: YAML parse error on grpc-proxy/templates/deployment.yaml: error converting YAML to JSON: yaml: line 38: could not find expected ':'
  Use --debug flag to render out invalid YAML
```

#### Cause

Helmfile rendering of values was successful but Helm failed to render the Kubernetes resources (`helm template`).

#### Resolution

Add `--debug` flag to HELMFILE_OPTS. Review log file immediately before error.

## References

- https://helmfile.readthedocs.io/en/latest/
- https://helmfile.readthedocs.io/en/latest/#templating
- https://helmfile.readthedocs.io/en/latest/templating_funcs/
- http://masterminds.github.io/sprig/
