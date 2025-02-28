#### Steps to deploy Vault agent injector in a remote cluster


###### Remote cluster commands
1) Deploy vault agent injector.
```
helm install vault hashicorp/vault --set "injector.externalVaultAddr=http://<vault-server-IP-address>:<vault-port>" --set injector.authPath="auth/kubernetes2" --set injector.logLevel=debug -n vault --create-namespace
```
2) create secret for the vault service account
```
cat > vault-secret.yaml <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: vault-token-g955r
  namespace: vault
  annotations:
    kubernetes.io/service-account.name: vault
type: kubernetes.io/service-account-token
EOF
```
3) get jwt token, kubernetes CA file and host

```
VAULT_HELM_SECRET_NAME=$(kubectl -n vault get secrets --output=json | jq -r '.items[].metadata | select(.name|startswith("vault-token-")).name')

echo "TOKEN_REVIEW_JWT=$(kubectl get secret $VAULT_HELM_SECRET_NAME -n vault --output='go-template={{ .data.token }}' | base64 --decode)" > vault.env
echo "KUBE_CA_CERT_B64=$(kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}')" >> vault.env
echo "KUBE_HOST=$(kubectl config view --raw --minify --flatten --output='jsonpath={.clusters[].cluster.server}')" >> vault.env
```

###### Vault Server commands

1) copy vault.env to the vault server
```
kubectl cp vault.env vault/vault-0:/home/vault/

```
2) run following script to enable kubernetes auth for the remote cluster
```
cat > vault-config.sh <<EOF
#!/bin/ash
set -e
cd /home/vault
source vault.env

# Get the kubernetes CA cert of the remote cluster 
echo $KUBE_CA_CERT_B64 > .cert.pem
base64 -d .cert.pem > cert.pem
rm .cert.pem

# kubernetes auth for the remote cluster
vault auth enable -path=kubernetes2 kubernetes
vault write auth/kubernetes2/config token_reviewer_jwt="$TOKEN_REVIEW_JWT" kubernetes_host="$KUBE_HOST" kubernetes_ca_cert=@cert.pem issuer="https://kubernetes.default.svc.cluster.local"
vault write auth/kubernetes2/role/enrollment-role bound_service_account_names=enrollment bound_service_account_namespaces=idcs-enrollment policies=enrollment ttl=1h
EOF

kubectl cp vault-config.sh vault/vault-0:/home/vault/

kubectl exec  -n vault vault-0 -- ash -c "cd /home/vault && chmod +x vault-config.sh && ash vault-config.sh"
```

