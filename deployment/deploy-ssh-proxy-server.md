# Deploy Tenant SSH Proxy Server

## Create SSH key pair for SSH Proxy Operator

```bash
REGION=us-dev3-1
export SSH_PROXY_OPERATOR_KEY_COMMENT=${REGION}-ssh-proxy-operator-$(uuidgen | head -c 8)
rm local/secrets/${IDC_ENV}/ssh-proxy-operator/id_rsa*
make secrets
cat local/secrets/${IDC_ENV}/ssh-proxy-operator/id_rsa
cat local/secrets/${IDC_ENV}/ssh-proxy-operator/id_rsa.pub
```

Load secrets into Vault path controlplane/show/us-staging-1a-ssh-proxy-operator/ssh.

## Create guest user account in SSH Proxy Server

```bash
devcloud@pdx03-c08-azbs002-vm-1:~$
NEW_USER=guest
sudo adduser ${NEW_USER}
echo "AllowUsers ${NEW_USER}" | sudo tee /etc/ssh/sshd_config.d/${NEW_USER}.conf
sudo systemctl restart sshd
sudo -i -u ${NEW_USER}

guest@pdx03-c08-azbs002-vm-1:~$
mkdir .ssh
chmod 700 .ssh
touch .ssh/authorized_keys
chmod 664 .ssh/authorized_keys
```

Copy contents of local/secrets/${IDC_ENV}/ssh-proxy-operator/id_rsa.pub to .ssh/authorized_keys.

## Create bmo user account in SSH Proxy Server

Repeat above steps for `bmo` user.
Secrets are in https://internal-placeholder.com/ui/vault/secrets/controlplane/show/us-region-1a-bm-instance-operator/ssh.
