# SDN end-to-end test

## structure of the end-to-end test

```
├── README.md
├── allscfabrics # test cases folder for allfabrics containerlab topology. One folder for each containerlab topology.
│   └── allscfabrics.go
├── frontendonly # test cases folder for frontendonly containerlab topology
│   └── frontendonly.go # the main test cases code are stored in this file.
├── frontendonly-tiny # a containerlab topology that includes one frontend switch and one server, it's for test env with limited resources.
│   ├── README.md
│   └── frontendonly-tiny.go
└── suite_test.go # prepare global resources/components in this file.
```

## How to run

#### Prerequisites
Install Docker (not covered here)

Install Containerlab `bash -c "$(curl -sL https://get.containerlab.dev)"`

Install ginkgo:
```
go install github.com/onsi/ginkgo/v2/ginkgo@v2.19.0
cp ~/go/bin/ginkgo /usr/local/bin/
```

Bring up all in kind `make deploy-all-in-kind`

Ensure kind cluster is up and that the kubeconfig file will connect to the cluster where it is running `kubectl get switches -n idcs-system`.

Connect kind container to the clab network (may require a containerlab topology to be deployed first): `docker network connect clab idc-global-control-plane`

Check that /etc/hosts on your VM contains an entry for `us-dev-1-provider-sdn-controller-rest.idcs-system.svc.cluster.local` (should be added by make deploy-all-in-kind, but may not if you have an old deployment)

Make sure that either no proxy env vars are set, or .cluster.local is part of your no_proxy env var.

Create the eAPI secret file under `/vault/secrets/eapi` with the below format (we usually use "admin" for both user and password for containerlab switches)
```
credentials:
  username: <user>
  password: <password>
```

Set eapi username/password in kind Vault:
```
vi local/secrets/EAPI_USERNAME
vi local/secrets/EAPI_PASSWD
vault kv put -mount=controlplane us-dev-1/us-dev-1a/nw-sdn-controller/eapi     password="@local/secrets/EAPI_PASSWD" username="@local/secrets/EAPI_USERNAME"
vault kv put -mount=controlplane us-dev-1/provider-sdn-controller/eapi password="@local/secrets/EAPI_PASSWD" username="@local/secrets/EAPI_USERNAME"
```

The linux `diff` tool is used to compare the switches configuration, make sure it's installed. (Running in Windows system is not supported yet)
```
sudo apt update
sudo apt install diffutils
``` 

#### Run test-cases

A tag `ginkgo_only` has been added to the e2e test files, so they won't be executed during `normal go test`. 

```
# go to the <sdn-controller-project-root>/tests/e2e folder

# run all test cases
ginkgo -tags=ginkgo_only

# run frontendonly only
ginkgo -tags=ginkgo_only --label-filter="frontendonly"

# run a specific testcase only
  ginkgo -tags=ginkgo_only --label-filter="frontendonly && case1"

# run Tenant-SDN tests only
ginkgo -tags=ginkgo_only --label-filter="tsdn"

# run Provider-SDN / REST API tests only
ginkgo -tags=ginkgo_only --label-filter="psdn"
```


## Troubleshooting

"Error: the 'frontendonly' lab has already been deployed. Destroy the lab before deploying a lab with the same name"

Try to `clab destroy` manually in the clab directory (networking/clab/frontendonly for the above error).
If that doesn't work, stop & remove the docker containers manually.
