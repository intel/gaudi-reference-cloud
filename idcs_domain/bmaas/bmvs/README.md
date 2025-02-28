# Bare Metal Virtual Stack (BMVS)

The goal of this stack is to emulate the physical hardware use for the
Intel Developer Cloud deployment.

This is useful for development testing without having to have access to physical
hardware for most testing cases. This stack can also be useful for CI/CD.

## Business Impact
- Cost saving of not requiring the purchase of hardware
- Time savings to enable testing without waiting for hardware
- Efficiency in allow testing locally on the developers system without the need for on site

## Design

The BMVS consists of:
- Guest Host virtual machines (defaults to 3)
- Virtual BMC for each Guest Host
- Virtual Networking
- HTTP Server to support OS installs
- Bastion Server to front all Guest Host connections

For more details about this stack, see the [Wiki](https://internal-placeholder.com/display/andrekeedy/Baremetal+Virtual+stack).

## Requirements:
The project is deployed using [Ansible](https://docs.ansible.com/ansible/latest/index.html),
deploys [virtual machines](https://wiki.libvirt.org/page/Main_Page#libvirt_Wiki)
and [networking](https://libvirt.org/formatnetwork.html) using
[libvirt](https://libvirt.org/index.html),
and creates virtual BMCs using
[Virtual Redfish BMC](https://docs.openstack.org/sushy-tools/latest/user/dynamic-emulator.html).

```
make install-requirements
```

If using a Virtual Workstation configured by the group IT dept, you may have to
request that ports be opened for incoming connects.
For the Virtual BMCs via sushy-emulator:
```
-A INPUT -p tcp --match multiport --dports 8000:8010 -j ACCEPT
```

*Only if requiring VNC instead of the more reliable `virt-manager`*
```
-A INPUT -p tcp --match multiport --dports 5900:5910 -j ACCEPT
```

## Status
To get the overall status of the Bare Metal Virtual Stack, rub
the make command without arguments or passing the `status` option.

```
make
```

## Deployment

The number of guest host servers is controlled by the Ansible variable
`guest_host_deployments` and defaults to `3`.

```
make setup
```

The number of guest host VM servers can be overridden with

```
GUEST_HOST_DEPLOYMENTS=1 make setup-guest-test
```

### Clean up / Shutdown

To stop all of the virtual machines, libvirt networks, virtual BMCs, and
stop the HTTP server, used the `teardown` target.

```
make teardown
```

To ensure all allocated directories, images, etc are remove from the
deployment host, use the `clean` target.

```
make clean
```

## Development
Any new process, vm, services, or changes in general deployed to the developer's
system should have a match 'teardown' function implemented. This not only helps
keep the developer's system clean, but ensure a clear path to redeployment.

### Linters
After changes to the Ansible files, be sure to run the linters before committing.
```
make lint
```

## Usage

### Networks

The Bare Metal Virtual Stack creates a Data Network, which hosts the
Bastion server (SSH Proxy server), and a VLAN network for each Guest
Host server deployed. Each Guest Host server is isolated on its own
VLAN which can only be access via the Bastion server with SSH.

This means the Bastion server has a `data` (192.168.150.1) interface and
one additional vlan interface for each guest node/vlan deployed.

| Host         | IP Address / Mask    | Interface | Default Router/DNS Server |
| ----------   |:-----------------:   |:---------:| :------------------------:|
| bmvs-bastion | 192.168.150.1 / 24   | data      | Set by DHCP               |
| "          " | 192.168.101.254 / 24 | vlan1     | "         "               |
| "          " | 192.168.102.254 / 24 | vlan2     | "         "               |
| "          " | 192.168.103.254 / 24 | vlan3     | "         "               |
| "          " | ...                  | ...       | "         "               |
| "          " | 192.168.10n.254 / 24 | vlann     | "         "               |
|              |                      |           |                           |
| gh-node-1    | 192.168.101.1 / 24   | vlan1     | 192.168.101.254           |
| gh-node-2    | 192.168.102.1 / 24   | vlan2     | 192.168.102.254           |
| gh-node-3    | 192.168.103.1 / 24   | vlan3     | 192.168.103.254           |
| ...          | ...           / 24   | ...       | ...                       |
| gh-node-n    | 192.168.10n.1 / 24   | vlann     | 192.168.10n.254           |

The Guest Host nodes should be configured to use the Bastion server as their
default router and DNS server.

Below is visual diagram of the network connectivity in the virtual stack.

```
                                    ^
                                    |
                         Intel or Public Network
                                    |
                                    |
                           +--------v--------+
                           |     [eth0]      |
                           |                 |
                           |  :Deployment:   |
                           |  :  Server  :   |
                           |                 |
                           | 192.168.150.254 |
                           |     [data]      |
                           +--------^--------+
                                    |
                                    |
                               Data Network
                             192.168.150.x/24
                                    |
                                    |
                        +-----------v-----------+
                        |        [data]         |
                        |     192.168.150.1     |
                        |                       |
                        |  :Bastion SSH Proxy:  |
                        |                       |
                        |x.101.254     x.10n.254|
         +-------------->[vlan1]         [vlanN]<----------------+
         |              |                       |                |
         |              |                       |                |
         |              |x.102.254     x.103.254|                |
  Virtual Network 1     |  [vlan2]     [vlan3]  |         Virtual Network N
  192.168.101.x/24      +-----^-----------^-----+         192.168.10n.x/24
         |                    |           |                      |
         |                    |           |                      |
         |       Virtual Network 2      Virtual Network 3        |
         |       192.168.102.x/24       192.168.103.x/24         |
         |                    |           |                      |
         |                    |           |                      |
         |                 +--+           +--+                   |
         |                 |                 |                   |
         |                 |                 |                   |
  +------v------+   +------v------+   +------v------+     +------v------+
  |   [vlan1]   |   |   [vlan2]   |   |   [vlan3]   |     |   [vlanN]   |
  |  xxx.101.1  |   |  xxx.102.1  |   |  xxx.103.1  |     |  xxx.10n.1  |
  |             |   |             |   |             |     |             |
  | :gh-node-1: |   | :gh-node-2: |   | :gh-node-3: |     | :gh-node-n: |
  |             |   |             |   |             |     |             |
  +------+------+   +------+------+   +------+------+     +------+------+
         :                 :                 :                   :
      libvirt           libvirt           libvirt             libvirt
         :                 :                 :                   :
  +------+------+   +------+------+   +------+------+     +------+------+
  |             |   |             |   |             |     |             |
  |  |vBMC 1|   |   |  |vBMC 2|   |   |  |vBMC 3|   |     |  |vBMC n|   |
  |             |   |             |   |             |     |             |
  | Deployment  |   | Deployment  |   | Deployment  |     | Deployment  |
  |   Server    |   |   Server    |   |   Server    |     |   Server    |
  |   [eth0]    |   |   [eth0]    |   |   [eth0]    |     |   [eth0]    |
  | :Port 8001: |   | :Port 8002: |   | :Port 8003: |     | :Port 800n: |
  +-------------+   +-------------+   +-------------+     +-------------+

```
Note that the virtual BMCs are only a process running on the Deployment
server, so their IP address is the same as the "public" (interface eth0)
IP address of the Deployment server. Each has a unique port number.

Currently, all of the access is via the Deployment server.

| Guest Host | IP Address / VLAN | VNC Console  | BMC Port on Deployment server |
| ---------- |:------------: |:------------:| :--------------:|
| gh-node-1  | 192.168.101.1 | 5901         | 8001            |
| gh-node-2  | 192.168.102.1 | 5902         | 8002            |
| gh-node-3  | 192.168.103.1 | 5903         | 8003            |
| ...        | ...           | ...          | ...             |
| gh-node-n  | 192.168.10n.1 | 5900+n       | 8000+n          |

### Bastion Server

The Bastion server is hardcoded to be 192.168.150.1. This IP is only accessible
by first connecting to the Deployment Hosts public IP address. This implies that
you should run your code for using this virtual stack on the Deployment server
or account for the required extra jump from Deployment -> Bastion -> Guest Host.

TODO: Add an iptable rule on the Deployment host that forwards all traffic
to a port on the Deployment host to the Bastion server on Port 22?

All access to the Guest Hosts is done through the Bastion server.

The SSH keys for the Bastion server are generated and stored in the
`ssh_keys` directory under `playbooks` directory.
```
ssh -i playbooks/ssh_keys/bmvs-bastion bastion@192.168.150.1
```

There is also a helper script in the `ssh_keys` directory called
`deploy_bastion_keys.sh` which will install the bastion keys and
update the user's `~/.ssh/config` for easy usage.
Example:
```
# Added by Bare Metal Virtual Stack scripting
Host bastion bmvs-bastion 192.168.150.1
    User guest
    HostName 192.168.150.1
    IdentityFile ~/.ssh/bmvs-bastion

```

This enables the user of simple ssh commands as the identies, user,
and hosts are defined in the `~/.ssh.config` file.
```
ssh bastion
```

### Guest Host
The guest hosts are deployed as hardware, without an OS installed.
For testing VMS, see [Guest Test Host](### Guest Testing Host) below.

The expectation is Dev Cloud tools/APIs will complete the installation
by mounting an ISO via the BMC and rebooting the VM.

Once the guest hosts are installed and configured, they can be accessed
by SSH Proxy through the Bastion server with something similar to:
```
ssh -i <path_to_key>/<guest_host_identity> -oProxyCommand="ssh -i playbooks/ssh_keys/bmvs-bastion guest@192.168.150.1 -W %h:%p" sdp@192.168.101.1
```
`sdp` being the deployed user on the guest-host. The identify for username
should also be installed on the Bastion server.

It is useful to update the `~/.ssh/config` file with the guest host's
username, host, and identity with something similar to:
```
# Guest Host 1
Host guest-1 guest-test-1 192.168.101.2
    ProxyJump bastion
    HostName 192.168.101.1
    User sdp
    Port 22
    IdentityFile ~/.ssh/guest-test-1

# Guest Host 2
Host guest-2 guest-test-2 192.168.102.2
    ProxyJump bastion
    HostName 192.168.102.1
    User sdp
    Port 22
    IdentityFile ~/.ssh/guest-test-2

# Added by Bare Metal Virtual Stack scripting
Host bastion bmvs-bastion 192.168.150.1
    User guest
    HostName 192.168.150.1
    IdentityFile ~/.ssh/bmvs-bastion

```

This enables the user of simple ssh commands as the identies, user,
and hosts are defined in the `~/.ssh.config` file.
```
ssh guest-2
```

#### Console Access
The graphical console can be access using `virt-manager` on the Deployment Host.
virt-manager does support remote connections using the --connect option. This is
the preferred method of monitoring and interacting with the system console.

##### Download Virtual Manager for Workstation/Laptop
Linux:
Virt-manager should be installed as part installing the BMVS requirements, but
can be installed via APT.
```
sudo apt install virt-manager
```

Windows:
A Windows client for Virtual Manager is available from the
[Virt Manager website](https://virt-manager.org/download/).

MacOS:
The simplest method for MacOS requires the use of `brew` to install.
See the [Brew Package website](https://formulae.brew.sh/formula/virt-manager#default)
```
brew install virt-manager
```
See the [Brew main page](https://brew.sh/) for brew installation directions.

##### Using VNC

A secondary method to access VNC graphic console of a running system is to
use the following process.
Use ssh to tunnel VNC remote access:
1. On your local system ssh with a tunnel
    ssh -L <VNC_PORT>:localhost:<VNC_PORT> REMOTE_IP
i.e. ssh -L 5901:localhost:5901 internal-placeholder.com
2. Open VNC client to locahost:<VNC_PORT>
i.e. tightvncconnect localhost:5901

##### Serial Console

The serial console tty can also be accessed using 'virsh console <node>' on the Deployment Host.

### Virtual BMC
The virtual BMCs are Redfish compliant and simulated using the
[sushy-emulator](https://docs.openstack.org/sushy-tools/latest/user/dynamic-emulator.html).

#### vBMC Access
The vBMC can be access on the Deployment server using a port address of 8000 + the number
of the Guest Host node.
```
http://{{deployhost}}:8001/redfish/v1/
```

#### vBMC Postman Collection
To enable exploring/testing the virtual BMC, the repo includes a
[Postman collection](postman_collection.json) which can be imported into the
[Postman](https://www.postman.com/) using `File --> Import` and selecting
the JSON file.

Before using the collection, select the collection name, click the `Variables` table, and
set the proper current value for the `deployment` host, and save.

Next use the `Guest Host - Systems` API call to get the UUID of the guest host VM.
Then update the current value for the `gh-node` in the variables tab, and save.


#### Authorization
Sushy-emulator currently only supports "Basic Auth" authorization.
```
Username: admin
Password: password
```
Note: This password will be change in the future and stored in https://passmanagement.intel.com/
To change the password now, see the
[README.md in the vBMC role](playbooks/roles/vbmc/files/README.md) below.

#### vBMC Sushy-emulator Log
Each virtual BMC has an emulator running on a unique port and storing the
operational results into a log file. The log file is stored in /var/log as
sushy-gh-<n>-emulator.log where n is the number of the Guest Host node.

### Guest Testing Host
To facilitate testing of the environment, the stack can deploy multiple test
guest host VMs with ubuntu 22.04 installed using the same base server image
deployed for the Bastion server.

```
make setup-guest-test
```

guest_host_deployments: 3
The number of guest test server is controlled by the Ansible variable
`guest_test_host_deployments` and defaults to `2`. This can be overridden with

```
GUEST_TEST_HOST_DEPLOYMENTS=1 make setup-guest-test
```

The Guest Testing servers are hardcoded to be 192.168.10n.2.

The SSH keys for each Guest Testing server are generated and stored in the
`ssh_keys` directory under `playbooks` directory.

The Guest Test server can be access using the bastion as an intermediary:

```
ssh -i playbooks/ssh_keys/guest-test-1 -oProxyCommand="ssh -i playbooks/ssh_keys/bmvs-bastion guest@192.168.150.1 -W %h:%p" sdp@192.168.101.2
```

There is also a helper script in the `ssh_keys` directory called
`deploy_guest_keys.sh` which will install the guest test server keys to both
the user on the deployment server as well as on the bastion server, and update
`.ssh/config` file easy usage.
Example:
```
# Added by Bare Metal Virtual Stack scripting
Host guest-1 guest-test-1 192.168.101.2
    ProxyJump bastion
    HostName 192.168.101.2
    User sdp
    Port 22
    IdentityFile ~/.ssh/guest-test-1

# Added by Bare Metal Virtual Stack scripting
Host guest-2 guest-test-2 192.168.102.2
    ProxyJump bastion
    HostName 192.168.102.2
    User sdp
    Port 22
    IdentityFile ~/.ssh/guest-test-2

# Added by Bare Metal Virtual Stack scripting
Host bastion bmvs-bastion 192.168.150.1
    User guest
    HostName 192.168.150.1
    IdentityFile ~/.ssh/bmvs-bastion
```

After the deployment script is run, the jump proxy shorthand can be used:
```
ssh guest-test-1
```

#### Caching of ubuntu server image

The Bastion server VM and all of the Test Guest Hosts VM are created from
the latest ubuntu Cloud Image for 22.04. To speed up deployment and testing
the image is downloaded into a shared cache directory the first time it is
needed, and then reused. To expedite testing or demos, the cache can be
preloaded ahead of time using:

```
make cache
```

## Difference from Physical BMC
- vBMC [sushy-tools](https://docs.openstack.org/sushy-tools/latest/user/dynamic-emulator.html)
only supports `Basic Auth` authentication.
  - Hardware BMC also support Sessions. No workflows truly need sessions.
  - Can we just use `Basic Auth` for vBMC and production hardware?
  - Or do we invest effort into enabling Sessions for sushy-tools emulator?
- vBMC [sushy-tools](https://docs.openstack.org/sushy-tools/latest/user/dynamic-emulator.html)
does not support NFS for mounting ISOs; only HTTP

## TODO
- Implement better network isolation
  - what are the minimum requirements for dev testing?
