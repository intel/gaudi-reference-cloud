<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# Development

## Prerequisites

Ubuntu 22.04 with latest updates

Install protobuf pkgs
```
   $ go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
```

Edit .bashrc  with the following exports:
```
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
export GOPRIVATE=github.com/intel-innersource/frameworks.cloud.devcloud.services.idc
```
   
## Build

To fully build all packages execute following command:
```
make build
```

### Generate Proto file
To generate code from the gRPC proto definition file execute the following command:

```
make api
```

This command will generate pb files and place them alongside proto file, these files should not be checked in the git repo

## Run

To run different packages:

```
TBD
```

## Test

### Unit tests
To run Go unit tests use the command 

```
make test
```
Each case tests a single handler using an in-memory OVSDB model and a temporary PostgreSQL Docker container. After each test, these components will be automatically purged.


### Docker compose

<details>
  <summary>Installation</summary>

It is recommended to get a recent version of `docker-compose` binary file from its official GitHub repository:
```
wget https://github.com/docker/compose/releases/download/<vX.Y.Z>/docker-compose-linux-x86_64
```

For example, to download `v2.29.2`:
```
wget https://github.com/docker/compose/releases/download/v2.29.2/docker-compose-linux-x86_64
sudo chmod +x docker-compose-linux-x86_64
sudo mv docker-compose-linux-x86_64 /usr/local/bin/docker-compose
```

Alternatively, `docker-compose` is usually also available in package managers such as `apt`. However, it can be a rather early version.

</details>

<details>
  <summary>Configuration</summary>

Pulling images from the Internet requires a separate proxy configuration file. If networking issues are encountered during building Docker images, please consider creating a configuration file at:

`/etc/systemd/system/docker.service.d/proxy.conf`

```
[Service]
Environment="HTTP_PROXY=http://proxy-chain.intel.com:912"
Environment="HTTPS_PROXY=http://proxy-chain.intel.com:912"
Environment="NO_PROXY=localhost,127.0.0.1,.example.com"
```
After creating or modifying the file, use the following commands to restart service:
```
systemctl daemon-reload
systemctl restart docker
```

</details>

<details open>
  <summary>Run Tests</summary>
First, make sure the latest binary files have been built, and secrets for SSL have been generated:

```
make
make secrets
```

When running tests for the first time, the Docker images should be built first:

```
docker-compose -f assets/docker/compose.yaml build
```

However, this step can be skipped if the user has access to the `idc-networking` Harbor project (https://amr-idc-registry-pre.infra-host.com/harbor/projects/9/repositories), where we maintain pre-built Docker images that can be directly pulled.

To run a local environment, use the command

```
docker-compose -f assets/docker/compose.yaml up --detach
```
It runs an OVN container and a PostgreSQL container locally with port mapping. With this environment, more test cases are available:

```
./sdncontroller
./tests/featureTests/test_nat.sh
```

Similarly, other scripts at `tests/featureTests/` such as `test_security.sh` can be tested in this way.

OVN topology and SQL rows can be dumped through Docker by using commands such as

```
docker exec -it docker-local-ovn-1 ovn-nbctl show
docker exec -it docker-local-psql-1 psql -U sdncontroller sdncontroller -c "SELECT * from vpc;"
```

Currently, each script at `tests/featureTests/` requires a manual cleanup before running another test. To clear the environment, use the command

```
docker-compose -f assets/docker/compose.yaml down
```

Alternatively, the user can choose to run the SDN Controller inside a container as well, by adding the flag `--profile integrated`. For example:
```
docker-compose -f assets/docker/compose.yaml --profile integrated build
docker-compose -f assets/docker/compose.yaml --profile integrated up
docker-compose -f assets/docker/compose.yaml --profile integrated down
```
This method opens the port `127.0.0.1:50051` for gRPC connection. Please note that:

- The SDN Controller binary must be built using `make` before building the Docker images.
- It may take longer time (around 10s) to fully set up the environment, since the SDN Controller container must wait for the other two containers to get ready first.

</details>

### Ghz

[Ghz](https://github.com/bojand/ghz) is a gRPC testing tool which may be useful as a client in addition to `sdnctl` in the following scenarios:

- When performance metrics such as throughput and latency need to be measured.
- When SSL is enabled for the gRPC server.

With the `docker-compose` setup above, the following examples show how `ghz` sends requests to the SDN Controller:

```bash
# When SSL is disabled.
ghz --insecure --proto ./api/sdn/v1/ovnnet.proto --call sdn.v1.Ovnnet.ListVPCs 127.0.0.1:50051

# When SSL is enabled.
ghz --proto ./api/sdn/v1/ovnnet.proto --call sdn.v1.Ovnnet.ListVPCs 127.0.0.1:50051 --cacert=./secrets/cacert.pem --cert=./secrets/ovn-central-cert.pem --key=./secrets/ovn-central-privkey.pem
```

The tool can also generate fake input data quickly. The following example creates a subnet with a random name and IP range:
```bash
ghz --insecure --proto ./api/sdn/v1/ovnnet.proto --call sdn.v1.Ovnnet.CreateSubnet -d '{"subnet_id":{"uuid":"{{newUUID}}"}, "name":"{{randomString 8}}", "cidr":"{{randomInt 1 80}}.{{randomInt 1 80}}.{{randomInt 1 80}}.0/24", "vpc_id":{"uuid":"3edf1cee-b7dd-424f-a854-a5eea28653ce"}, "availability_zone":"az1"}' 127.0.0.1:50051
```