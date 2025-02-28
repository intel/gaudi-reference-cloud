# Universe Deployer

## Contents

- [Universe Deployer Documentation](../../docs/source/private/guides/universe_deployer.rst)
- [IDC Services Upgrade Procedure](../../docs/source/private/guides/services_upgrade_procedure.rst)
- [Demo Script](#demo-1)
- [Local Test Procedure](#local-test-procedure-1)
- [Compare Manifests Generator Output](#compare-manifests-generator-output)
- [References](#references)

## Demo 1

This demonstrates how Universe Deployer runs in Jenkins and creates a new branch in idc-argocd.

Checkout your working branch (feature/twc4727-1282-argocd-manifest-in-bazel2).

```bash
make test-universe-deployer-git-push && \
git branch -D feature/twc4727-846-main-authority ; \
git branch feature/twc4727-846-main-authority && \
git checkout feature/twc4727-846-main-authority && \
git commit -m "force build without PR" --allow-empty && \
git push --force -u origin feature/twc4727-846-main-authority
```

View Jenkins pipeline stage "Universe Deployer" for the branch:
https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/tree/feature/twc4727-846-main-authority

Review the log for text similar to the following:

```
Create a pull request for 'universe-deployer-2024-01-16T21-55-35' on GitHub by visiting:
https://github.com/ClaudioFaheyIntel/frameworks.cloud.devcloud.services.idc-argocd/pull/new/universe-deployer-2024-01-16T21-55-35
```

Click the link in the log to create a PR to merge into the *development* fork of idc-argocd.
https://github.com/ClaudioFaheyIntel/frameworks.cloud.devcloud.services.idc-argocd

Merge the PR.

Next, simulate a code change.
Edit `go/pkg/manageddb/manageddb.go` to log the text `TODO-CLAUDIO-V1` in New().

Commit change.

Edit `environments/staging.json`.
Update commit value of the following to the latest commit:

- environments.staging.regions.us-staging-1.components.compute
- environments.staging.regions.us-staging-1.components.computeApiServer

Commit change.
Push to Github.

Jenkins pipeline will show the git diff which should contain only the change shown below.
Note that only the compute-api-server Helm chart will change because it is the only changed component that uses manageddb.

```
Running {"Args": ["git", "diff", "--staged", "--exit-code"], "Dir": "/tmp/universe_deployer_git_pusher_180024741", "Env": []}
stdout	diff --git a/applications/idc-regional/us-staging-1/pdx05-k01-rgcp/us-staging-1-compute-api-server/config.json b/applications/idc-regional/us-staging-1/pdx05-k01-rgcp/us-staging-1-compute-api-server/config.json
stdout	index 69c80b1..be173cf 100644
stdout	--- a/applications/idc-regional/us-staging-1/pdx05-k01-rgcp/us-staging-1-compute-api-server/config.json
stdout	+++ b/applications/idc-regional/us-staging-1/pdx05-k01-rgcp/us-staging-1-compute-api-server/config.json
stdout	@@ -2,7 +2,7 @@
stdout	   "envconfig": {
stdout	     "releaseName": "us-staging-1-compute-api-server",
stdout	     "chartName": "intelcloud/compute-api-server",
stdout	-    "chartVersion": "0.0.1-fe7b6f8abecceee875dbdeebfe418ace1b424de03a514cbcc3dc2889a5f75a0e",
stdout	+    "chartVersion": "0.0.1-c534bf212f7eb1e13a68aa375215f091084c36808adc4681c6decfa30088b838",
stdout	     "chartRegistry": "amr-idc-registry-pre.infra-host.com",
stdout	     "namespace": "idcs-system"
stdout	   }
Completed {"Args[0]": "git", "err": "exit status 1"}
```

Jenkins pipeline will create a new branch in https://github.com/ClaudioFaheyIntel/frameworks.cloud.devcloud.services.idc-argocd.

Click the link in the log to create a PR to merge into the *development* fork of idc-argocd.
https://github.com/ClaudioFaheyIntel/frameworks.cloud.devcloud.services.idc-argocd

## Local Test Procedure 1

This runs Universe Deployer in a local workstation.

```bash
time make test-universe-deployer-git-push universe-deployer |& tee local/universe_deployer_v0.log
```

Extract the generated Argo CD manifests.

```bash
rm -rf local/universe_deployer
mkdir -p local/universe_deployer
tar -C local/universe_deployer -xf bazel-bin/deployment/universe_deployer/main_universe_deployer.tar
rm -rf local/universe_deployer_v0
mv local/universe_deployer local/universe_deployer_v0
```

Edit `go/pkg/manageddb/manageddb.go` to log the text `TODO-CLAUDIO-V1`.

Edit `deployment/helmfile/environments/staging.yaml.gotmpl`.
Set `log.encoder` to `console`.

Commit change.

Edit `universe_deployer/environments/staging.json`.
Set `environments.staging.regions.us-staging-1.components.compute.commit` to the commit hash.

Generate manifests using the following command.

```bash
time make universe-deployer |& tee local/universe_deployer_v2.log
```

Compare with previous version using the following command.

```bash
rm -rf local/universe_deployer
mkdir -p local/universe_deployer
tar -C local/universe_deployer -xf bazel-bin/deployment/universe_deployer/main_universe_deployer.tar
diff -r local/universe_deployer_v0 local/universe_deployer
```

Compare with `make helmfile-generate-argocd-values` using the following commands.

```bash
IDC_ENV=staging make helmfile-generate-argocd-values
diff -r local/secrets/staging/helm-values local/universe_deployer/ | less -S
```

Create a new Git branch with generated Argo CD manifests.

```bash
make main-universe-deployer-git-pusher
```

## Compare Manifests Generator Output

Use these steps to compare the output of Universe Deployer Manifests Generator with different commits or branches.
This will generate manifests for all environments listed in test-manifests-generator.sh.
This can be used to ensure that no unexpected changes occur.

```bash
git checkout main
deployment/helmfile/scripts/test-manifests-generator.sh local/manifests-generator_v0
```

Checkout a different commit and run the same steps.

```bash
git checkout ...
deployment/helmfile/scripts/test-manifests-generator.sh local/manifests-generator_v1
```

Compare with the previous version using the following command.

```bash
diff --color=always -r \
local/manifests-generator_v0/manifests \
local/manifests-generator_v1/manifests | less -R -S
```

You may also want to use the VSCode Compare Folders extension to compare the output of the folders.

## References

- Some components of Universe Deployer are modeled after https://github.com/abrisco/rules_helm/blob/main/helm/private/helm_package.bzl.
