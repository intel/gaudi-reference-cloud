
## Test Argo CD with Deploy All In Kind

Run the [Manual Test Process](../../README.md#manual-test-process)
(`make deploy-all-in-kind` through *basic end-to-end tests*).
This deployment will include Gitea (a local instance of Github) and Argo CD.

Configure a new Git repo in Gitea. Initially, the Git repo will not have any IDC services.

```bash
mkdir -p local/deploy-all-in-kind
GITEA_ADMIN_PASSWORD=$(cat local/secrets/gitea_admin_password)
rm -rf local/deploy-all-in-kind/idc-argocd-local-repo
cp -rv deployment/argocd/idc-argocd-initial-data local/deploy-all-in-kind/idc-argocd-local-repo

(cd local/deploy-all-in-kind/idc-argocd-local-repo &&
git init . &&
git branch -m main &&
git config --local http.proxy "" &&
git remote add origin http://gitea_admin:${GITEA_ADMIN_PASSWORD}@dev.gitea.cloud.intel.com.kind.local/gitea_admin/idc-argocd.git)

(cd local/deploy-all-in-kind/idc-argocd-local-repo &&
git add --all --verbose &&
git commit -m "update" &&
git push -u origin main)
```

Ensure that the app-of-apps Application is synced.

```bash
kubectl get -n argocd applications
```

Expected output:

```
NAME          SYNC STATUS   HEALTH STATUS
app-of-apps   Synced        Healthy
```

Generate Argo CD manifests for IDC services.

```bash
make helmfile-generate-argocd-values
```

Add Argo CD manifests to Git repo in Gitea.

```bash
mkdir -p local/deploy-all-in-kind/idc-argocd-local-repo/applications
cp -rv local/secrets/helm-values/* local/deploy-all-in-kind/idc-argocd-local-repo/applications

(cd local/deploy-all-in-kind/idc-argocd-local-repo &&
git add --all --verbose &&
git commit -m "update" &&
git push -u origin main)
```

View Applications.

```bash
watch kubectl get -n argocd applications
```

Wait for Application resources to be created and achieve Synced status as shown below.
Note that Argo CD will recognize the existing Kubernetes resources deployed by Helmfile.
Argo CD will not recreate these resources.

```
NAME                                                        SYNC STATUS   HEALTH STATUS
app-of-apps                                                 Synced        Healthy
kind-idc-global-argo-cd-resources                           Unknown       Healthy
kind-idc-global-argocd                                      Synced        Healthy
kind-idc-global-billing                                     Synced        Healthy
kind-idc-global-billing-aria                                Synced        Healthy
kind-idc-global-billing-db                                  Synced        Healthy
kind-idc-global-billing-intel                               Synced        Healthy
kind-idc-global-billing-schedulers                          Synced        Healthy
kind-idc-global-billing-standard                            Synced        Healthy
kind-idc-global-cloudaccount                                Synced        Healthy
kind-idc-global-cloudaccount-db                             Synced        Healthy
kind-idc-global-cloudaccount-enroll                         Synced        Healthy
kind-idc-global-external-secrets                            Synced        Healthy
kind-idc-global-gitea                                       Synced        Healthy
kind-idc-global-grpc-proxy-external                         Synced        Healthy
kind-idc-global-grpc-proxy-internal                         Synced        Healthy
kind-idc-global-grpc-reflect                                Synced        Healthy
kind-idc-global-grpc-rest-gateway                           Synced        Healthy
kind-idc-global-metering                                    Synced        Healthy
kind-idc-global-metering-db                                 Synced        Healthy
kind-idc-global-oidc                                        Synced        Healthy
kind-idc-global-productcatalog                              Synced        Healthy
kind-idc-global-productcatalog-crds                         Synced        Healthy
kind-idc-global-productcatalog-operator                     Synced        Healthy
kind-idc-global-trade-scanner                               Synced        Healthy
kind-idc-global-us-dev-1-compute-api-server                 Synced        Healthy
kind-idc-global-us-dev-1-compute-db                         Synced        Healthy
kind-idc-global-us-dev-1-grpc-proxy-external                Synced        Healthy
kind-idc-global-us-dev-1-grpc-proxy-internal                Synced        Healthy
kind-idc-global-us-dev-1-grpc-reflect                       Synced        Healthy
kind-idc-global-us-dev-1-grpc-rest-gateway                  Synced        Healthy
kind-idc-global-us-dev-1-iks                                Synced        Progressing
kind-idc-global-us-dev-1-kfaas                              Synced        Healthy
kind-idc-global-us-dev-1-kfaas-db                           Synced        Healthy
kind-idc-global-us-dev-1-populate-instance-type             Synced        Healthy
kind-idc-global-us-dev-1-populate-machine-image             Synced        Healthy
kind-idc-global-us-dev-1-populate-subnet                    Synced        Healthy
kind-idc-global-us-dev-1-storage-api-server                 Synced        Healthy
kind-idc-global-us-dev-1-storage-db                         Synced        Healthy
kind-idc-global-us-dev-1-storage-kms                        Synced        Healthy
kind-idc-global-us-dev-1-storage-scheduler                  Synced        Healthy
kind-idc-global-us-dev-1a-compute-crds                      Synced        Healthy
kind-idc-global-us-dev-1a-compute-metering-monitor          Synced        Healthy
kind-idc-global-us-dev-1a-instance-replicator               Synced        Healthy
kind-idc-global-us-dev-1a-metal3-crds                       Synced        Healthy
kind-idc-global-us-dev-1a-ssh-proxy-operator                Synced        Healthy
kind-idc-global-us-dev-1a-storage-metering-monitor          Synced        Healthy
kind-idc-global-us-dev-1a-storage-operator                  Synced        Healthy
kind-idc-global-us-dev-1a-storage-replicator                Synced        Healthy
kind-idc-global-us-dev-1a-vm-instance-operator-harvester1   Synced        Healthy
kind-idc-global-us-dev-1a-vm-instance-scheduler             Synced        Healthy
kind-idc-global-vault                                       OutOfSync     Healthy
```

## Test Moving Applications from Argo CD to Helm

For reference, see https://argocd-applicationset.readthedocs.io/en/stable/Application-Deletion/.

Edit `local/deploy-all-in-kind/idc-argocd-local-repo/manifests/base/applications/base-appset.yaml`.
Set `preserveResourcesOnDeletion: true`.

```bash
(cd local/deploy-all-in-kind/idc-argocd-local-repo &&
git add --all --verbose &&
git commit -m "update" &&
git push -u origin main)
```

Confirm that all Applications no longer have a finalizer.

Delete all regional Applications.

```bash
rm -rf local/deploy-all-in-kind/idc-argocd-local-repo/applications/idc-regional/us-dev-1/kind-idc-global

(cd local/deploy-all-in-kind/idc-argocd-local-repo &&
git add --all --verbose &&
git commit -m "update" &&
git push -u origin main)
```

Confirm that Applications are deleted but that the pods continue to run.
