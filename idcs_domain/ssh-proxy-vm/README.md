# SSH Proxy Virtual Machine for Development

## Description

This provides an SSH server that can be deployed as a Harvester VM.
This can be used as an SSH proxy server whose authorized_keys file is maintained by the SSH Proxy Controller.

Additionally, the proxy server has an secondary internal private IP address 172.16.0.1/16
with NAT enabled so that any host on 172.16.0.0/16 can make outbound connections to external networks.

## Adding Guest Accounts

Each instance of the SSH Proxy Operator should have a separate user account in the SSH proxy server
so it can manage its own `authorized_keys` file.
This means that each IDC developer that runs the SSH Proxy Operator locally should have a separate account.
The naming convention for these accounts is `guest-$USER`.

Run the following commands to generate an SSH key pair that will be used by the SSH Proxy Operator.

```bash
cd $(git rev-parse --show-toplevel)
export IDC_ENV=kind-2regions
make secrets
eval `make show-export`
cat ${SECRETS_DIR}/ssh-proxy-operator/id_rsa.pub
```

This will generate the SSH key pair in `local/secrets/ssh-proxy-operator` or `local/secrets/${IDC_ENV}/ssh-proxy-operator`.
The public key file will be named `id_rsa.pub` in this directory.
The contents of this file should then be copied to the users section of
[deployment/helm/ssh-proxy-vm/values.yaml](deployment/helm/ssh-proxy-vm/values.yaml).

For example:

```yaml
users:
  guest-claudiof:
    ssh_authorized_keys:
    - ssh-rsa AAAAB3NzaC...X2FhKK0= dev-ssh-proxy-operator-claudiof-92055de2@intel.com
```

A PR should then be submitted with the updated values.yaml file.
Once approved and merged, follow the steps in the next section.

## Deploying in Development

1. Set the environment variables for HARVESTER_CLUSTER_NAME and SECRETS_DIR.

```bash 
export HARVESTER_CLUSTER_NAME=harvester3
export SECRETS_DIR=$(git rev-parse --show-toplevel)/local/secrets
```

2. Login to the Harvester UI.

3. Click *Support* (bottom left).

4. Click *Download KubeConfig*.

5. Place the kubeconfig file in your secret directory, inside a directory named harvester-kubeconfig `$(SECRETS_DIR)/harvester-kubeconfig`.

6. Download `ssh_host_*` files from [Vault](https://internal-placeholder.com/ui/vault/secrets/dev-idc-env/show/shared/harvester1/proxy)
   and place in the `idcs_domain/ssh-proxy-vm/secrets` directory.

7. Run below, then verify that KUBECONFIG and OVERRIDES_FILES variables are correct.  

```bash
(cd idcs_domain/ssh-proxy-vm && make show-config)
```

8. Run below.

```bash
(cd idcs_domain/ssh-proxy-vm && make redeploy)
```

## How to Use

### SSH to an instance through a proxy server using your default identity

Replace the final IP address with the address of your instance.

```shell
ssh -J guest@10.165.62.252 ubuntu@172.16.0.51
```

### SSH to an instance through a proxy server using another identity file

Start SSH Agent and load an identity file.

```shell
eval `ssh-agent`
ssh-add ../operator/config/samples/sshkeys/testuser@example.com_id_rsa
```

Replace the final IP address with the address of your instance.

```shell
ssh -J guest@10.165.62.252 ubuntu@172.16.1.28
```

## Troubleshooting

### Verbose Logging

Run SSH with the `-v` parameter to show verbose logs.

```shell
ssh -v -J guest@10.165.62.252 ubuntu@172.16.0.51
```

### Expected Output Without SSH Agent

```
OpenSSH_8.2p1 Ubuntu-4ubuntu0.5, OpenSSL 1.1.1f  31 Mar 2020
debug1: Reading configuration data /etc/ssh/ssh_config
debug1: /etc/ssh/ssh_config line 19: include /etc/ssh/ssh_config.d/*.conf matched no files
debug1: /etc/ssh/ssh_config line 21: Applying options for *
debug1: Setting implicit ProxyCommand from ProxyJump: ssh -l guest -v -W '[%h]:%p' 10.165.62.252
debug1: Executing proxy command: exec ssh -l guest -v -W '[172.16.0.51]:22' 10.165.62.252
debug1: identity file /home/claudiof/.ssh/id_rsa type 0
debug1: identity file /home/claudiof/.ssh/id_rsa-cert type -1
debug1: identity file /home/claudiof/.ssh/id_dsa type -1
debug1: identity file /home/claudiof/.ssh/id_dsa-cert type -1
debug1: identity file /home/claudiof/.ssh/id_ecdsa type -1
debug1: identity file /home/claudiof/.ssh/id_ecdsa-cert type -1
debug1: identity file /home/claudiof/.ssh/id_ecdsa_sk type -1
debug1: identity file /home/claudiof/.ssh/id_ecdsa_sk-cert type -1
debug1: identity file /home/claudiof/.ssh/id_ed25519 type -1
debug1: identity file /home/claudiof/.ssh/id_ed25519-cert type -1
debug1: identity file /home/claudiof/.ssh/id_ed25519_sk type -1
debug1: identity file /home/claudiof/.ssh/id_ed25519_sk-cert type -1
debug1: identity file /home/claudiof/.ssh/id_xmss type -1
debug1: identity file /home/claudiof/.ssh/id_xmss-cert type -1
debug1: Local version string SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5
OpenSSH_8.2p1 Ubuntu-4ubuntu0.5, OpenSSL 1.1.1f  31 Mar 2020
debug1: Reading configuration data /etc/ssh/ssh_config
debug1: /etc/ssh/ssh_config line 19: include /etc/ssh/ssh_config.d/*.conf matched no files
debug1: /etc/ssh/ssh_config line 21: Applying options for *
debug1: Connecting to 10.165.62.252 [10.165.62.252] port 22.
debug1: Connection established.
debug1: identity file /home/claudiof/.ssh/id_rsa type 0
debug1: identity file /home/claudiof/.ssh/id_rsa-cert type -1
debug1: identity file /home/claudiof/.ssh/id_dsa type -1
debug1: identity file /home/claudiof/.ssh/id_dsa-cert type -1
debug1: identity file /home/claudiof/.ssh/id_ecdsa type -1
debug1: identity file /home/claudiof/.ssh/id_ecdsa-cert type -1
debug1: identity file /home/claudiof/.ssh/id_ecdsa_sk type -1
debug1: identity file /home/claudiof/.ssh/id_ecdsa_sk-cert type -1
debug1: identity file /home/claudiof/.ssh/id_ed25519 type -1
debug1: identity file /home/claudiof/.ssh/id_ed25519-cert type -1
debug1: identity file /home/claudiof/.ssh/id_ed25519_sk type -1
debug1: identity file /home/claudiof/.ssh/id_ed25519_sk-cert type -1
debug1: identity file /home/claudiof/.ssh/id_xmss type -1
debug1: identity file /home/claudiof/.ssh/id_xmss-cert type -1
debug1: Local version string SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5
debug1: Remote protocol version 2.0, remote software version OpenSSH_8.9p1 Ubuntu-3
debug1: match: OpenSSH_8.9p1 Ubuntu-3 pat OpenSSH* compat 0x04000000
debug1: Authenticating to 10.165.62.252:22 as 'guest'
debug1: SSH2_MSG_KEXINIT sent
debug1: SSH2_MSG_KEXINIT received
debug1: kex: algorithm: curve25519-sha256
debug1: kex: host key algorithm: ecdsa-sha2-nistp256
debug1: kex: server->client cipher: chacha20-poly1305@openssh.com MAC: <implicit> compression: none
debug1: kex: client->server cipher: chacha20-poly1305@openssh.com MAC: <implicit> compression: none
debug1: expecting SSH2_MSG_KEX_ECDH_REPLY
debug1: Server host key: ecdsa-sha2-nistp256 SHA256:mR53riUf4KKl6fmfxk/27zaO9M2qIAhOX1NyZdpwo0I
debug1: Host '10.165.62.252' is known and matches the ECDSA host key.
debug1: Found key in /home/claudiof/.ssh/known_hosts:9
debug1: rekey out after 134217728 blocks
debug1: SSH2_MSG_NEWKEYS sent
debug1: expecting SSH2_MSG_NEWKEYS
debug1: SSH2_MSG_NEWKEYS received
debug1: rekey in after 134217728 blocks
debug1: Will attempt key: /home/claudiof/.ssh/id_rsa RSA SHA256:GZUsjph1/KzRGt7Uxpwf3Tr02BEpFxxPPIibHig+QIs
debug1: Will attempt key: /home/claudiof/.ssh/id_dsa 
debug1: Will attempt key: /home/claudiof/.ssh/id_ecdsa 
debug1: Will attempt key: /home/claudiof/.ssh/id_ecdsa_sk 
debug1: Will attempt key: /home/claudiof/.ssh/id_ed25519 
debug1: Will attempt key: /home/claudiof/.ssh/id_ed25519_sk 
debug1: Will attempt key: /home/claudiof/.ssh/id_xmss 
debug1: SSH2_MSG_EXT_INFO received
debug1: kex_input_ext_info: server-sig-algs=<ssh-ed25519,sk-ssh-ed25519@openssh.com,ssh-rsa,rsa-sha2-256,rsa-sha2-512,ssh-dss,ecdsa-sha2-nistp256,ecdsa-sha2-nistp384,ecdsa-sha2-nistp521,sk-ecdsa-sha2-nistp256@openssh.com,webauthn-sk-ecdsa-sha2-nistp256@openssh.com>
debug1: kex_input_ext_info: publickey-hostbound@openssh.com (unrecognised)
debug1: SSH2_MSG_SERVICE_ACCEPT received
debug1: Authentications that can continue: publickey
debug1: Next authentication method: publickey
debug1: Offering public key: /home/claudiof/.ssh/id_rsa RSA SHA256:GZUsjph1/KzRGt7Uxpwf3Tr02BEpFxxPPIibHig+QIs
debug1: Server accepts key: /home/claudiof/.ssh/id_rsa RSA SHA256:GZUsjph1/KzRGt7Uxpwf3Tr02BEpFxxPPIibHig+QIs
debug1: Authentication succeeded (publickey).
Authenticated to 10.165.62.252 ([10.165.62.252]:22).
debug1: channel_connect_stdio_fwd 172.16.0.51:22
debug1: channel 0: new [stdio-forward]
debug1: getpeername failed: Bad file descriptor
debug1: Requesting no-more-sessions@openssh.com
debug1: Entering interactive session.
debug1: pledge: network
debug1: client_input_global_request: rtype hostkeys-00@openssh.com want_reply 0
debug1: Remote: /home/guest/.ssh/authorized_keys:2: key options: agent-forwarding command permitopen port-forwarding
debug1: Remote: /home/guest/.ssh/authorized_keys:2: key options: agent-forwarding command permitopen port-forwarding
debug1: Remote protocol version 2.0, remote software version OpenSSH_8.9p1 Ubuntu-3
debug1: match: OpenSSH_8.9p1 Ubuntu-3 pat OpenSSH* compat 0x04000000
debug1: Authenticating to 172.16.0.51:22 as 'ubuntu'
debug1: SSH2_MSG_KEXINIT sent
debug1: SSH2_MSG_KEXINIT received
debug1: kex: algorithm: curve25519-sha256
debug1: kex: host key algorithm: ecdsa-sha2-nistp256
debug1: kex: server->client cipher: chacha20-poly1305@openssh.com MAC: <implicit> compression: none
debug1: kex: client->server cipher: chacha20-poly1305@openssh.com MAC: <implicit> compression: none
debug1: expecting SSH2_MSG_KEX_ECDH_REPLY
debug1: Server host key: ecdsa-sha2-nistp256 SHA256:1qaIUQIIOrbTRHe9L8PyqkNyyoMxdIinMv/FfOx1Emk
debug1: Host '172.16.0.51' is known and matches the ECDSA host key.
debug1: Found key in /home/claudiof/.ssh/known_hosts:19
debug1: rekey out after 134217728 blocks
debug1: SSH2_MSG_NEWKEYS sent
debug1: expecting SSH2_MSG_NEWKEYS
debug1: SSH2_MSG_NEWKEYS received
debug1: rekey in after 134217728 blocks
debug1: Will attempt key: /home/claudiof/.ssh/id_rsa RSA SHA256:GZUsjph1/KzRGt7Uxpwf3Tr02BEpFxxPPIibHig+QIs
debug1: Will attempt key: /home/claudiof/.ssh/id_dsa 
debug1: Will attempt key: /home/claudiof/.ssh/id_ecdsa 
debug1: Will attempt key: /home/claudiof/.ssh/id_ecdsa_sk 
debug1: Will attempt key: /home/claudiof/.ssh/id_ed25519 
debug1: Will attempt key: /home/claudiof/.ssh/id_ed25519_sk 
debug1: Will attempt key: /home/claudiof/.ssh/id_xmss 
debug1: SSH2_MSG_EXT_INFO received
debug1: kex_input_ext_info: server-sig-algs=<ssh-ed25519,sk-ssh-ed25519@openssh.com,ssh-rsa,rsa-sha2-256,rsa-sha2-512,ssh-dss,ecdsa-sha2-nistp256,ecdsa-sha2-nistp384,ecdsa-sha2-nistp521,sk-ecdsa-sha2-nistp256@openssh.com,webauthn-sk-ecdsa-sha2-nistp256@openssh.com>
debug1: kex_input_ext_info: publickey-hostbound@openssh.com (unrecognised)
debug1: SSH2_MSG_SERVICE_ACCEPT received
debug1: Authentications that can continue: publickey
debug1: Next authentication method: publickey
debug1: Offering public key: /home/claudiof/.ssh/id_rsa RSA SHA256:GZUsjph1/KzRGt7Uxpwf3Tr02BEpFxxPPIibHig+QIs
debug1: Server accepts key: /home/claudiof/.ssh/id_rsa RSA SHA256:GZUsjph1/KzRGt7Uxpwf3Tr02BEpFxxPPIibHig+QIs
debug1: Authentication succeeded (publickey).
Authenticated to 172.16.0.51 (via proxy).
debug1: channel 0: new [client-session]
debug1: Requesting no-more-sessions@openssh.com
debug1: Entering interactive session.
debug1: pledge: proc
debug1: client_input_global_request: rtype hostkeys-00@openssh.com want_reply 0
debug1: Remote: /home/ubuntu/.ssh/authorized_keys:1: key options: agent-forwarding port-forwarding pty user-rc x11-forwarding
debug1: Remote: /home/ubuntu/.ssh/authorized_keys:1: key options: agent-forwarding port-forwarding pty user-rc x11-forwarding
debug1: Sending environment.
debug1: Sending env LANG = en_US.UTF-8
Welcome to Ubuntu 22.04.1 LTS (GNU/Linux 5.15.0-1018-kvm x86_64)

 * Documentation:  https://help.ubuntu.com
 * Management:     https://landscape.canonical.com
 * Support:        https://ubuntu.com/advantage

  System information as of Sat Oct  1 20:35:26 UTC 2022

  System load:  0.0               Processes:               106
  Usage of /:   15.6% of 9.51GB   Users logged in:         0
  Memory usage: 2%                IPv4 address for enp1s0: 172.16.0.51
  Swap usage:   0%


1 update can be applied immediately.
1 of these updates is a standard security update.
To see these additional updates run: apt list --upgradable


Last login: Sat Oct  1 20:34:04 2022 from 172.16.0.1
To run a command as administrator (user "root"), use "sudo <command>".
See "man sudo_root" for details.

ubuntu@claudiof2:~
```

### Expected Output With SSH Agent

```
OpenSSH_8.2p1 Ubuntu-4ubuntu0.5, OpenSSL 1.1.1f  31 Mar 2020
debug1: Reading configuration data /etc/ssh/ssh_config
debug1: /etc/ssh/ssh_config line 19: include /etc/ssh/ssh_config.d/*.conf matched no files
debug1: /etc/ssh/ssh_config line 21: Applying options for *
debug1: Setting implicit ProxyCommand from ProxyJump: ssh -l guest -v -W '[%h]:%p' 10.165.62.252
debug1: Executing proxy command: exec ssh -l guest -v -W '[172.16.1.28]:22' 10.165.62.252
debug1: identity file /home/claudiof/.ssh/id_rsa type 0
debug1: identity file /home/claudiof/.ssh/id_rsa-cert type -1
debug1: identity file /home/claudiof/.ssh/id_dsa type -1
debug1: identity file /home/claudiof/.ssh/id_dsa-cert type -1
debug1: identity file /home/claudiof/.ssh/id_ecdsa type -1
debug1: identity file /home/claudiof/.ssh/id_ecdsa-cert type -1
debug1: identity file /home/claudiof/.ssh/id_ecdsa_sk type -1
debug1: identity file /home/claudiof/.ssh/id_ecdsa_sk-cert type -1
debug1: identity file /home/claudiof/.ssh/id_ed25519 type -1
debug1: identity file /home/claudiof/.ssh/id_ed25519-cert type -1
debug1: identity file /home/claudiof/.ssh/id_ed25519_sk type -1
debug1: identity file /home/claudiof/.ssh/id_ed25519_sk-cert type -1
debug1: identity file /home/claudiof/.ssh/id_xmss type -1
debug1: identity file /home/claudiof/.ssh/id_xmss-cert type -1
debug1: Local version string SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5
OpenSSH_8.2p1 Ubuntu-4ubuntu0.5, OpenSSL 1.1.1f  31 Mar 2020
debug1: Reading configuration data /etc/ssh/ssh_config
debug1: /etc/ssh/ssh_config line 19: include /etc/ssh/ssh_config.d/*.conf matched no files
debug1: /etc/ssh/ssh_config line 21: Applying options for *
debug1: Connecting to 10.165.62.252 [10.165.62.252] port 22.
debug1: Connection established.
debug1: identity file /home/claudiof/.ssh/id_rsa type 0
debug1: identity file /home/claudiof/.ssh/id_rsa-cert type -1
debug1: identity file /home/claudiof/.ssh/id_dsa type -1
debug1: identity file /home/claudiof/.ssh/id_dsa-cert type -1
debug1: identity file /home/claudiof/.ssh/id_ecdsa type -1
debug1: identity file /home/claudiof/.ssh/id_ecdsa-cert type -1
debug1: identity file /home/claudiof/.ssh/id_ecdsa_sk type -1
debug1: identity file /home/claudiof/.ssh/id_ecdsa_sk-cert type -1
debug1: identity file /home/claudiof/.ssh/id_ed25519 type -1
debug1: identity file /home/claudiof/.ssh/id_ed25519-cert type -1
debug1: identity file /home/claudiof/.ssh/id_ed25519_sk type -1
debug1: identity file /home/claudiof/.ssh/id_ed25519_sk-cert type -1
debug1: identity file /home/claudiof/.ssh/id_xmss type -1
debug1: identity file /home/claudiof/.ssh/id_xmss-cert type -1
debug1: Local version string SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5
debug1: Remote protocol version 2.0, remote software version OpenSSH_8.9p1 Ubuntu-3
debug1: match: OpenSSH_8.9p1 Ubuntu-3 pat OpenSSH* compat 0x04000000
debug1: Authenticating to 10.165.62.252:22 as 'guest'
debug1: SSH2_MSG_KEXINIT sent
debug1: SSH2_MSG_KEXINIT received
debug1: kex: algorithm: curve25519-sha256
debug1: kex: host key algorithm: ecdsa-sha2-nistp256
debug1: kex: server->client cipher: chacha20-poly1305@openssh.com MAC: <implicit> compression: none
debug1: kex: client->server cipher: chacha20-poly1305@openssh.com MAC: <implicit> compression: none
debug1: expecting SSH2_MSG_KEX_ECDH_REPLY
debug1: Server host key: ecdsa-sha2-nistp256 SHA256:mR53riUf4KKl6fmfxk/27zaO9M2qIAhOX1NyZdpwo0I
debug1: Host '10.165.62.252' is known and matches the ECDSA host key.
debug1: Found key in /home/claudiof/.ssh/known_hosts:9
debug1: rekey out after 134217728 blocks
debug1: SSH2_MSG_NEWKEYS sent
debug1: expecting SSH2_MSG_NEWKEYS
debug1: SSH2_MSG_NEWKEYS received
debug1: rekey in after 134217728 blocks
debug1: Will attempt key: testuser@example.com RSA SHA256:EarVC2w6tTuQEJ1Ek44Zdvrf4ByWTAdh+Tg4Wr1g6PI agent
debug1: Will attempt key: /home/claudiof/.ssh/id_rsa RSA SHA256:GZUsjph1/KzRGt7Uxpwf3Tr02BEpFxxPPIibHig+QIs
debug1: Will attempt key: /home/claudiof/.ssh/id_dsa 
debug1: Will attempt key: /home/claudiof/.ssh/id_ecdsa 
debug1: Will attempt key: /home/claudiof/.ssh/id_ecdsa_sk 
debug1: Will attempt key: /home/claudiof/.ssh/id_ed25519 
debug1: Will attempt key: /home/claudiof/.ssh/id_ed25519_sk 
debug1: Will attempt key: /home/claudiof/.ssh/id_xmss 
debug1: SSH2_MSG_EXT_INFO received
debug1: kex_input_ext_info: server-sig-algs=<ssh-ed25519,sk-ssh-ed25519@openssh.com,ssh-rsa,rsa-sha2-256,rsa-sha2-512,ssh-dss,ecdsa-sha2-nistp256,ecdsa-sha2-nistp384,ecdsa-sha2-nistp521,sk-ecdsa-sha2-nistp256@openssh.com,webauthn-sk-ecdsa-sha2-nistp256@openssh.com>
debug1: kex_input_ext_info: publickey-hostbound@openssh.com (unrecognised)
debug1: SSH2_MSG_SERVICE_ACCEPT received
debug1: Authentications that can continue: publickey
debug1: Next authentication method: publickey
debug1: Offering public key: testuser@example.com RSA SHA256:EarVC2w6tTuQEJ1Ek44Zdvrf4ByWTAdh+Tg4Wr1g6PI agent
debug1: Server accepts key: testuser@example.com RSA SHA256:EarVC2w6tTuQEJ1Ek44Zdvrf4ByWTAdh+Tg4Wr1g6PI agent
debug1: Authentication succeeded (publickey).
Authenticated to 10.165.62.252 ([10.165.62.252]:22).
debug1: channel_connect_stdio_fwd 172.16.1.28:22
debug1: channel 0: new [stdio-forward]
debug1: getpeername failed: Bad file descriptor
debug1: Requesting no-more-sessions@openssh.com
debug1: Entering interactive session.
debug1: pledge: network
debug1: client_input_global_request: rtype hostkeys-00@openssh.com want_reply 0
debug1: Remote: /home/guest/.ssh/authorized_keys:1: key options: agent-forwarding command permitopen port-forwarding
debug1: Remote: /home/guest/.ssh/authorized_keys:1: key options: agent-forwarding command permitopen port-forwarding
debug1: Remote protocol version 2.0, remote software version OpenSSH_8.9p1 Ubuntu-3
debug1: match: OpenSSH_8.9p1 Ubuntu-3 pat OpenSSH* compat 0x04000000
debug1: Authenticating to 172.16.1.28:22 as 'ubuntu'
debug1: SSH2_MSG_KEXINIT sent
debug1: SSH2_MSG_KEXINIT received
debug1: kex: algorithm: curve25519-sha256
debug1: kex: host key algorithm: ecdsa-sha2-nistp256
debug1: kex: server->client cipher: chacha20-poly1305@openssh.com MAC: <implicit> compression: none
debug1: kex: client->server cipher: chacha20-poly1305@openssh.com MAC: <implicit> compression: none
debug1: expecting SSH2_MSG_KEX_ECDH_REPLY
debug1: Server host key: ecdsa-sha2-nistp256 SHA256:0zMihElfSKUsVNTo05ctmz6Fyu6OgbaAi+AFy1Lvty0
debug1: Host '172.16.1.28' is known and matches the ECDSA host key.
debug1: Found key in /home/claudiof/.ssh/known_hosts:13
debug1: rekey out after 134217728 blocks
debug1: SSH2_MSG_NEWKEYS sent
debug1: expecting SSH2_MSG_NEWKEYS
debug1: SSH2_MSG_NEWKEYS received
debug1: rekey in after 134217728 blocks
debug1: Will attempt key: testuser@example.com RSA SHA256:EarVC2w6tTuQEJ1Ek44Zdvrf4ByWTAdh+Tg4Wr1g6PI agent
debug1: Will attempt key: /home/claudiof/.ssh/id_rsa RSA SHA256:GZUsjph1/KzRGt7Uxpwf3Tr02BEpFxxPPIibHig+QIs
debug1: Will attempt key: /home/claudiof/.ssh/id_dsa 
debug1: Will attempt key: /home/claudiof/.ssh/id_ecdsa 
debug1: Will attempt key: /home/claudiof/.ssh/id_ecdsa_sk 
debug1: Will attempt key: /home/claudiof/.ssh/id_ed25519 
debug1: Will attempt key: /home/claudiof/.ssh/id_ed25519_sk 
debug1: Will attempt key: /home/claudiof/.ssh/id_xmss 
debug1: SSH2_MSG_EXT_INFO received
debug1: kex_input_ext_info: server-sig-algs=<ssh-ed25519,sk-ssh-ed25519@openssh.com,ssh-rsa,rsa-sha2-256,rsa-sha2-512,ssh-dss,ecdsa-sha2-nistp256,ecdsa-sha2-nistp384,ecdsa-sha2-nistp521,sk-ecdsa-sha2-nistp256@openssh.com,webauthn-sk-ecdsa-sha2-nistp256@openssh.com>
debug1: kex_input_ext_info: publickey-hostbound@openssh.com (unrecognised)
debug1: SSH2_MSG_SERVICE_ACCEPT received
debug1: Authentications that can continue: publickey
debug1: Next authentication method: publickey
debug1: Offering public key: testuser@example.com RSA SHA256:EarVC2w6tTuQEJ1Ek44Zdvrf4ByWTAdh+Tg4Wr1g6PI agent
debug1: Server accepts key: testuser@example.com RSA SHA256:EarVC2w6tTuQEJ1Ek44Zdvrf4ByWTAdh+Tg4Wr1g6PI agent
debug1: Authentication succeeded (publickey).
Authenticated to 172.16.1.28 (via proxy).
debug1: channel 0: new [client-session]
debug1: Requesting no-more-sessions@openssh.com
debug1: Entering interactive session.
debug1: pledge: proc
debug1: client_input_global_request: rtype hostkeys-00@openssh.com want_reply 0
debug1: Remote: /home/ubuntu/.ssh/authorized_keys:1: key options: agent-forwarding port-forwarding pty user-rc x11-forwarding
debug1: Remote: /home/ubuntu/.ssh/authorized_keys:1: key options: agent-forwarding port-forwarding pty user-rc x11-forwarding
debug1: Sending environment.
debug1: Sending env LANG = en_US.UTF-8
Welcome to Ubuntu 22.04.1 LTS (GNU/Linux 5.15.0-1018-kvm x86_64)

 * Documentation:  https://help.ubuntu.com
 * Management:     https://landscape.canonical.com
 * Support:        https://ubuntu.com/advantage

  System information as of Sat Oct  1 20:45:27 UTC 2022

  System load:  0.0               Processes:               104
  Usage of /:   15.6% of 9.51GB   Users logged in:         0
  Memory usage: 2%                IPv4 address for enp1s0: 172.16.1.28
  Swap usage:   0%


1 update can be applied immediately.
1 of these updates is a standard security update.
To see these additional updates run: apt list --upgradable


Last login: Sat Oct  1 20:45:27 2022 from 172.16.0.1
To run a command as administrator (user "root"), use "sudo <command>".
See "man sudo_root" for details.

ubuntu@my-tiny-vm-1:~$ 
```
