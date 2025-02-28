<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# Secrets Vault Server

We are using Hashicorp's
[Vault](https://developer.hashicorp.com/vault "What is Vault?")
product for secret management.

Documentation:
https://developer.hashicorp.com/vault/docs

Training:
https://developer.hashicorp.com/vault/tutorials


## Production Server

[Add URL Here]

## Development Server

https://internal-placeholder.com


## Using the UI

Go to the server UI login page
https://internal-placeholder.com/ui/vault/auth?with=oidc

- Method:
  - Select `OIDC`
- Role:
  - <leave empty> for default

- <click> "Sign in with OIDC Provider" button

Authenticated will be against the Intel SSO. If prompted,
use your Intel Window's login credentials.

## CLI Work / Debug / Deployment

### Install the Client

If you plan to use the CLI for debug and/or testing, first install the client:
https://developer.hashicorp.com/vault/docs/install

### Configure the Environment

Set the `VAULT_ADDR` environment variable to point to the
correct Vault server. i.e. Development:

```shell
export VAULT_ADDR="https://internal-placeholder.com"
```

Get an Authentication Token using the Web UI from the appropriate server.

- Log into the Web UI and go to the Secrets pages
  i.e. Development:  https://internal-placeholder.com/ui/vault/secrets
- Under the Person Icon in the upper right of the page
  - Select "Copy Token" from the pull down menu
  - Note: The default is for the token to expire in 1 hour
- Set the `VAULT_TOKEN` environment variable to the toke value:
```shell
export VAULT_TOKEN='<paste token>'
```

### Intel Proxies

Most internal environment at Intel set the `no_proxy` or `NO_PROXY`
environment variables to include the `.intel.com` domain so that hosts
in the `.intel.com` domain are not attempted to be routed through the
Intel proxy servers.

The current Development Vault server is deployed in AWS, but resolves to
a domain name ending in `eglb.intel.com`.

To enable the use of the `vault` CLI, the No Proxy environment variable(s)
need to be updated to ensure they do not contain a match for the Vault server.
Because Vault is written in Go language, it will look for the Upper Case
`NO_PROXY` first; if that is undefined or empty, it will then use `no_proxy`.

If the case of the Development server `internal-placeholder.com`, we need
to ensure that `NO_PROXY` does not contain any of the following:
```
internal-placeholder.com
eglb.intel.com
intel.com
com
```
Settings `NO_PROXY` to an empty string is not enough as then Go will default
to using `no_proxy`. Either define `NO_PROXY` and ensure the host name and no
portion of its domain name are include, or a simple solution is to set both
environment variable versions to empty strings for that command only:
```shell
NO_PROXY="" no_proxy="" vault policy read default
```
This ensures that all network connections will go through the defined
Proxy Servers for that execution of the Vault client.

[Read This artivle](https://about.gitlab.com/blog/2021/01/27/we-need-to-talk-no-proxy/)
for a good tutorial on Proxy, the environment variables, and implementations
by variable tools and languages.

